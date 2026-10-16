package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Historial de cambios (rastro): cada cambio de un punto, de una necesidad (incluye runs, empaque, fotos, despacho y
// acopio) y de los centros de acopio queda anotado con quién lo hizo (persona, equipo y página), cuándo y qué cambió
// (antes → después). Así se puede seguir el rastro de cualquier dato. Lo ven administración y Command, en cada punto.

var auditBucket = []byte("audit")

// centersTrail: clave del rastro de los centros de acopio (nombre, archivar, recuperar, quién elige cuál).
const centersTrail = "~acopios"

func ensureAudit(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(auditBucket)
	return err
}

// Actor: quién hace la solicitud. ID vacío = público o el sistema; "admin" = administración (contraseña maestra).
type Actor struct {
	ID  string
	Via string
}

type AuditEntry struct {
	At     string `json:"at"`
	Kind   string `json:"kind"` // punto, necesidad, acopio
	Ref    string `json:"ref"`
	CampID string `json:"campId,omitempty"`
	Label  string `json:"label,omitempty"`
	Action string `json:"action"` // creó, cambió
	ByID   string `json:"byId,omitempty"`
	ByName string `json:"byName"`
	ByTeam string `json:"byTeam,omitempty"`
	// Equipo por su ID (para reconocer a los compañeros aunque el equipo cambie de nombre).
	ByTeamID string       `json:"byTeamId,omitempty"`
	Via      string       `json:"via,omitempty"`
	Changes  []ChangeLine `json:"changes"`
}

// El actor de la transacción de escritura en curso. bbolt hace una escritura a la vez, así que basta con recordar la
// última transacción que lo anotó; una escritura sin actor (migraciones al abrir) queda como «Sistema».
var txActor struct {
	sync.Mutex
	tx  *bolt.Tx
	who Actor
}

func setActor(tx *bolt.Tx, who Actor) {
	txActor.Lock()
	txActor.tx, txActor.who = tx, who
	txActor.Unlock()
}
func actorIn(tx *bolt.Tx) (Actor, bool) {
	txActor.Lock()
	defer txActor.Unlock()
	if txActor.tx == tx {
		return txActor.who, true
	}
	return Actor{}, false
}

// write es a.db.Update anotando quién hace la solicitud, para el historial.
func (a *App) write(r *http.Request, fn func(*bolt.Tx) error) error {
	who := a.actorFor(r)
	return a.db.Update(func(tx *bolt.Tx) error {
		setActor(tx, who)
		return fn(tx)
	})
}

// actorFor: la sesión de la página que hace la solicitud (sin rechazarla: eso lo hace el manejador).
func (a *App) actorFor(r *http.Request) Actor {
	if u := agentBy(r); u != nil {
		return Actor{u.ID, "Command"}
	}
	session := func(cookie string, m map[string]runnerSession) string {
		c, e := r.Cookie(cookie)
		if e != nil {
			return ""
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if s, ok := m[c.Value]; ok && time.Now().Before(s.expires) {
			return s.runnerID
		}
		return ""
	}
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/api/admin/"):
		if c, e := r.Cookie("relief_session"); e == nil {
			a.mu.Lock()
			_, ok := a.sessions[c.Value]
			a.mu.Unlock()
			if ok {
				return Actor{"admin", "Admin"}
			}
		}
	case strings.HasPrefix(p, "/api/agent/"):
		return Actor{session("agent_session", a.agentSessions), "Command"}
	case strings.HasPrefix(p, "/api/runner/"):
		return Actor{session("runner_session", a.runnerSessions), "Runners"}
	case strings.HasPrefix(p, "/api/builder/"):
		return Actor{session("builder_session", a.builderSessions), "Constructores"}
	case strings.HasPrefix(p, "/api/supplier/"):
		return Actor{session("supplier_session", a.supplierSessions), "Supply"}
	case strings.HasPrefix(p, "/api/cook/"):
		return Actor{session("cook_session", a.cookSessions), "Cocina"}
	case strings.HasPrefix(p, "/api/desk/"):
		return Actor{session("desk_session", a.deskSessions), "Donaciones"}
	case strings.HasPrefix(p, "/api/full/scout/"):
		return Actor{session("viewer_session", a.viewerSessions), "Scout"}
	case strings.HasPrefix(p, "/api/full/"):
		return Actor{session("viewer_session", a.viewerSessions), "Inicio"}
	}
	return Actor{"", "Inicio"}
}

// logAudit guarda una entrada con quien hizo la escritura en curso. Clave: punto + hora, para leer el rastro de un punto en orden.
func logAudit(tx *bolt.Tx, e AuditEntry) error {
	b := tx.Bucket(auditBucket)
	if b == nil || len(e.Changes) == 0 {
		return nil
	}
	who, ok := actorIn(tx)
	e.At = time.Now().UTC().Format(time.RFC3339Nano)
	switch {
	case !ok:
		e.ByName, e.Via = "Sistema", ""
	case who.ID == "admin":
		e.ByID, e.ByName, e.Via = "admin", "Administración", who.Via
	case who.ID == "":
		e.ByName, e.Via = "Público (sin sesión)", who.Via
	default:
		e.ByID, e.ByName, e.Via = who.ID, nameOf(tx, who.ID), who.Via
		if t, found := getTeam(tx, teamOf(tx, who.ID)); found {
			e.ByTeam, e.ByTeamID = t.Name, t.ID
		}
	}
	return putJSON(b, e.CampID+"\x00"+e.At+"\x00"+randomToken()[:6], e)
}

// Campos que se calculan o que solo ordenan: no son cambios de datos.
var auditSkip = map[string]bool{"id": true, "campId": true, "createdAt": true, "created": true, "source": true, "decided": true, "lastCheckup": true, "verifiedAt": true, "enRouteAt": true, "completedAt": true, "releasedAt": true, "supplierRequestedAt": true, "progressAt": true, "updatedAt": true, "runnerName": true, "campStatus": true, "highestUrgency": true, "needStatuses": true, "history": true, "team": true, "holdTeam": true, "row": true, "ordinal": true}

var auditLabel = map[string]string{
	// necesidades
	"need": "Pedido", "status": "Estado", "urgency": "Urgencia", "runnerId": "Runner", "builderId": "Constructor", "kind": "Tipo",
	"packing": "Lista de empaque", "media": "Fotos y videos", "supplierId": "Acopio", "dispatchedAt": "Despachada", "dispatchedBy": "Despachada por",
	"openOffers": "Acopios que ofrecen", "supplyOffers": "Respuestas de acopios", "supplyAskedAt": "Pidió a acopios", "sameAs": "Otro run de",
	"completedAt": "Completada", "enRouteAt": "En camino desde", "releasedAt": "Soltada", "releasedBy": "Soltada por", "updatedBy": "Última actualización de",
	"verifiedTeam": "Verificado por equipo", "verifiedBy": "Verificado por", "materials": "Materiales", "progress": "Avance", "scope": "Falta",
	"assessments": "Evaluaciones", "repeatDays": "Repite cada (días)", "nextDueAt": "Próxima vez", "doneLog": "Hecho", "createdBy": "Agregada por",
	// puntos
	"code": "Código", "area": "Zona", "place": "Lugar", "needs": "Necesidades", "people": "Personas", "coords": "Ubicación", "gps": "GPS",
	"members": "Personas del grupo", "contact": "Contacto", "phone": "Teléfono", "coordinator": "Reportado por", "coordinatorPhone": "Teléfono de quien reporta",
	"notes": "Notas", "directions": "Indicaciones", "heldBy": "Retenido por equipo", "heldUntil": "Retenido hasta", "checkups": "Checkups",
	"mergedInto": "Fusionado en", "children": "Niños", "babies": "Bebés", "adults": "Adultos", "olderAdults": "Adultos mayores", "families": "Familias",
	"disabled": "Con discapacidad", "lastCheckup": "Última revisión", "note": "Nota de revisión", "decided": "Decidido",
}

var pointStatusLabel = map[string]string{"approved": "Aprobado", "pending": "Por revisar", "archived": "Archivado", "rejected": "Rechazado"}

var idFields = map[string]bool{"runnerId": true, "builderId": true, "updatedBy": true, "releasedBy": true, "dispatchedBy": true, "createdBy": true}

// auditValue: un valor en palabras para el rastro (nombres en vez de IDs, resúmenes de listas).
func auditValue(tx *bolt.Tx, key string, raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` || string(raw) == "[]" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		switch {
		case idFields[key]:
			return nameOf(tx, s)
		case key == "status":
			if l := statusLabel[s]; l != "" {
				return l
			}
			if l := pointStatusLabel[s]; l != "" {
				return l
			}
		case key == "dispatchedAt" || key == "supplyAskedAt":
			return "sí"
		case key == "supplierId":
			if sup := supplierOf(tx, s); sup != nil {
				return strings.Trim(sup.Center+" · "+sup.Name, " ·")
			}
		case key == "heldBy" || key == "verifiedTeam":
			if t, ok := getTeam(tx, s); ok {
				return t.Name
			}
		}
		return s
	}
	switch key {
	case "packing", "materials":
		var lines []PackingLine
		if json.Unmarshal(raw, &lines) == nil {
			out := []string{}
			for _, l := range lines {
				out = append(out, fmt.Sprintf("%s ×%v", l.Item, l.Quantity))
			}
			return strings.Join(out, ", ")
		}
	case "media":
		var items []MediaItem
		if json.Unmarshal(raw, &items) == nil {
			out := []string{}
			for _, m := range items {
				out = append(out, strings.TrimSpace(m.Stage+" "+m.Name))
			}
			return fmt.Sprintf("%d: %s", len(items), strings.Join(out, ", "))
		}
	case "checkups", "doneLog":
		var list []string
		if json.Unmarshal(raw, &list) == nil {
			return fmt.Sprintf("%d", len(list))
		}
	}
	var buf bytes.Buffer
	json.Compact(&buf, raw)
	return buf.String()
}

// diffJSON compara dos registros campo por campo (antes nil = recién creado).
func diffJSON(tx *bolt.Tx, before, after []byte) []ChangeLine {
	var old, cur map[string]json.RawMessage
	json.Unmarshal(before, &old)
	json.Unmarshal(after, &cur)
	keys := map[string]bool{}
	for k := range old {
		keys[k] = true
	}
	for k := range cur {
		keys[k] = true
	}
	var out []ChangeLine
	for k := range keys {
		if auditSkip[k] || bytes.Equal(old[k], cur[k]) {
			continue
		}
		b, a := auditValue(tx, k, old[k]), auditValue(tx, k, cur[k])
		if b == a {
			continue
		}
		label := auditLabel[k]
		if label == "" {
			label = k
		}
		out = append(out, ChangeLine{label, b, a})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

// auditPut guarda el registro y anota en el historial lo que cambió respecto de lo guardado.
func auditPut(b *bolt.Bucket, kind, key, campID, label string, data []byte) error {
	before := b.Get([]byte(key))
	if before != nil {
		before = append([]byte(nil), before...)
	}
	if e := b.Put([]byte(key), data); e != nil {
		return e
	}
	if before != nil && bytes.Equal(before, data) {
		return nil
	}
	action := "cambió"
	if before == nil {
		action = "creó"
	}
	return logAudit(b.Tx(), AuditEntry{Kind: kind, Ref: key, CampID: campID, Label: label, Action: action, Changes: diffJSON(b.Tx(), before, data)})
}

// putPoint guarda un punto (ya serializado) con su rastro.
func putPoint(b *bolt.Bucket, key []byte, data []byte) error {
	var p Point
	json.Unmarshal(data, &p)
	label := strings.Trim(p.Code+" · "+p.Area, " ·")
	return auditPut(b, "punto", string(key), string(key), label, data)
}

// campTrail: el rastro de un punto y de sus necesidades, del más reciente al más antiguo.
func (a *App) campTrail(id string) []AuditEntry {
	out := []AuditEntry{}
	a.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(auditBucket).Cursor()
		prefix := []byte(id + "\x00")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e AuditEntry
			if json.Unmarshal(v, &e) == nil {
				out = append(out, e)
			}
		}
		return nil
	})
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func (a *App) auditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/camps/{id}/trail", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		response(w, 200, a.campTrail(r.PathValue("id")))
	})
	mux.HandleFunc("GET /api/admin/centers/trail", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		response(w, 200, a.campTrail(centersTrail))
	})
}

// putPointJSON: putPoint de un punto sin serializar.
func putPointJSON(b *bolt.Bucket, p Point) error {
	data, e := json.Marshal(p)
	if e != nil {
		return e
	}
	return putPoint(b, []byte(p.ID), data)
}
