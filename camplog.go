package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Bitácora pública del punto (camp.html): toda la actividad desde que se reportó, en orden de tiempo, con quién la
// hizo. Sale del historial de cambios (audit.go), pero solo con los datos que ya son públicos: nunca contactos,
// teléfonos, personas del grupo, notas privadas, runners, empaque, fotos, acopios ni reparaciones. Quién: al público
// solo el papel («Un runner», «Un agente», «de un equipo»); a quien entró con código, el nombre completo de sus
// compañeros de equipo. Lo anterior al historial (empezó el 2026-10-15) se arma con las fechas guardadas.

type LogEntry struct {
	At      string       `json:"at"`
	Who     string       `json:"who"`
	Mate    bool         `json:"mate,omitempty"`
	What    string       `json:"what"`
	Changes []ChangeLine `json:"changes,omitempty"`
}

// Campos públicos de la bitácora, por tipo de registro. Los demás no salen.
var publicPointLog = map[string]bool{"Código": true, "Zona": true, "Lugar": true, "Necesidades": true, "Personas": true, "Adultos": true, "Niños": true, "Bebés": true, "Adultos mayores": true, "Con discapacidad": true, "Familias": true, "GPS": true, "Ubicación": true, "Notas": true, "Indicaciones": true, "Estado": true, "Checkups": true, "generator": true, "service": true, "serviceNotes": true, "extra": true}
var publicNeedLog = map[string]bool{"Pedido": true, "Estado": true, "Urgencia": true, "Despachada": true, "Runner": true}

// valores que se muestran como «actualizado» en vez del valor (públicos en el mapa, pero largos o poco claros).
var logLabel = map[string]string{"generator": "Generador", "service": "Servicio", "serviceNotes": "Notas de servicio", "extra": "Otras notas"}

func publicWho(e AuditEntry) string {
	who := ""
	switch {
	case e.ByName == "Sistema":
		return "Sistema"
	case e.ByID == "admin":
		return "Administración"
	case e.ByID == "":
		return "Alguien del público"
	case strings.HasPrefix(e.Via, "Runners"):
		who = "Un runner"
	case strings.HasPrefix(e.Via, "Constructores"):
		who = "Un constructor"
	case strings.HasPrefix(e.Via, "Command"):
		who = "Un agente"
	case strings.HasPrefix(e.Via, "Supply"):
		who = "Un acopio"
	default:
		who = "Un voluntario"
	}
	if e.ByTeam != "" {
		who += " de un equipo"
	}
	return who
}

// publicChanges deja solo los campos públicos; el runner sale como «asignado» o «sin asignar», sin nombre.
func publicChanges(e AuditEntry) []ChangeLine {
	allowed := publicNeedLog
	if e.Kind == "punto" {
		allowed = publicPointLog
	}
	var out []ChangeLine
	for _, c := range e.Changes {
		if !allowed[c.Label] {
			continue
		}
		if l := logLabel[c.Label]; l != "" {
			c.Label = l
		}
		switch c.Label {
		case "Runner":
			c.Before, c.After = assigned(c.Before), assigned(c.After)
			if c.Before == c.After {
				continue
			}
		case "Checkups":
			c = ChangeLine{"Checkup", "", "registrado"}
		case "Ubicación":
			c = ChangeLine{"Ubicación", "", "actualizada"}
		}
		out = append(out, c)
	}
	return out
}
func assigned(name string) string {
	if name == "" || name == "sin asignar" {
		return "sin asignar"
	}
	return "asignado"
}

// campLog arma la bitácora de un punto para quien mira (me vacío = público), del más reciente al más antiguo.
func (a *App) campLog(id, me string) ([]LogEntry, error) {
	out := []LogEntry{}
	err := a.db.View(func(tx *bolt.Tx) error {
		tv := newTeamView(tx, me)
		var p Point
		v := tx.Bucket(submissions).Get([]byte(id))
		if v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" || heldFrom(p, tv.team) {
			return errMissing
		}
		// Necesidades que se pueden mostrar: ni reparaciones, ni carreras reservadas de otro equipo (esas como el público).
		needs := map[string]CampNeed{}
		tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
			var n CampNeed
			if json.Unmarshal(v, &n) == nil && n.CampID == id {
				needs[n.ID] = n
			}
			return nil
		})
		first := map[string]string{}
		c := tx.Bucket(auditBucket).Cursor()
		prefix := []byte(id + "\x00")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e AuditEntry
			if json.Unmarshal(v, &e) != nil {
				continue
			}
			if _, seen := first[e.Ref]; !seen {
				first[e.Ref] = e.At
			}
			what := "📍 Punto"
			if e.Kind != "punto" {
				n, ok := needs[e.Ref]
				if !ok || isRepair(n) {
					continue
				}
				what = "📦 " + publicNeedText(n, tv)
			}
			changes := publicChanges(e)
			if len(changes) == 0 && e.Action != "creó" {
				continue
			}
			if e.Action == "creó" {
				what += " · creado"
			} else {
				what += " · cambio"
			}
			entry := LogEntry{At: e.At, Who: publicWho(e), What: what, Changes: changes}
			if me != "" && e.ByID != "" && e.ByID != "admin" && (e.ByID == me || tv.team != "" && e.ByTeamID == tv.team) {
				entry.Who, entry.Mate = e.ByName, true
			}
			out = append(out, entry)
		}
		// Antes del historial: lo que dicen las fechas guardadas del punto y de sus necesidades.
		before := func(ref, at string) bool {
			start, has := first[ref]
			return at != "" && (!has || at < start)
		}
		if before(p.ID, p.Created) {
			who := "Alguien del público"
			if p.AddedBy != "" {
				who = "Un voluntario"
				if t := teamOf(tx, p.AddedBy); t != "" {
					who += " de un equipo"
					if t == tv.team || p.AddedBy == me {
						who = nameOf(tx, p.AddedBy)
					}
				} else if p.AddedBy == me {
					who = nameOf(tx, p.AddedBy)
				}
			} else if p.Source == "admin" || p.Source == "sheet" {
				who = "Administración"
			}
			out = append(out, LogEntry{At: p.Created, Who: who, Mate: who == nameOf(tx, p.AddedBy) && p.AddedBy != "", What: "📍 Punto · reportado", Changes: []ChangeLine{{"Zona", "", p.Area}, {"Lugar", "", p.Place}}})
		}
		if before(p.ID, p.Decided) && p.Decided != p.Created {
			out = append(out, LogEntry{At: p.Decided, Who: "Administración", What: "📍 Punto · aprobado"})
		}
		for _, at := range p.Checkups {
			if before(p.ID, at) {
				out = append(out, LogEntry{At: at, Who: "Un agente", What: "📍 Punto · checkup", Changes: []ChangeLine{{"Checkup", "", "registrado"}}})
			}
		}
		for _, n := range needs {
			if isRepair(n) || n.Status == "DELETED" && !before(n.ID, n.CreatedAt) {
				continue
			}
			text := "📦 " + publicNeedText(n, tv)
			if before(n.ID, n.CreatedAt) {
				out = append(out, LogEntry{At: n.CreatedAt, Who: "—", What: text + " · reportada", Changes: []ChangeLine{{"Estado", "", "Abierta"}}})
			}
			if before(n.ID, n.EnRouteAt) {
				out = append(out, LogEntry{At: n.EnRouteAt, Who: "Un runner", What: text + " · en camino", Changes: []ChangeLine{{"Estado", "Abierta", "En camino"}}})
			}
			if before(n.ID, n.DispatchedAt) {
				out = append(out, LogEntry{At: n.DispatchedAt, Who: "Un acopio", What: text + " · despachada", Changes: []ChangeLine{{"Despachada", "", "sí"}}})
			}
			if before(n.ID, n.CompletedAt) {
				out = append(out, LogEntry{At: n.CompletedAt, Who: "Un runner", What: text + " · completada", Changes: []ChangeLine{{"Estado", "En camino", "Completada"}}})
			}
		}
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool { return stampTime(out[i].At).After(stampTime(out[j].At)) })
	return out, err
}

func stampTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

// publicNeedText: el pedido sin la parte de reparación; una carrera reservada de otro equipo sale sin detalle.
func publicNeedText(n CampNeed, tv *teamView) string {
	if tv.hidden(n) {
		return "Necesidad de un equipo"
	}
	text, _ := splitRepair(n.Need)
	if len([]rune(text)) > 80 {
		text = string([]rune(text)[:80]) + "…"
	}
	return text
}

// Scout (scout.html): los puntos que agregó quien entró con código o alguien de su equipo, con su estado.
func (a *App) campLogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/camps/{id}/log", func(w http.ResponseWriter, r *http.Request) {
		list, err := a.campLog(r.PathValue("id"), "")
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("GET /api/full/camps/{id}/log", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		list, err := a.campLog(r.PathValue("id"), u.ID)
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
}
