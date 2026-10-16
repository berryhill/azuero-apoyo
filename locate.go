package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// «📍 Por ubicar»: entregas registradas sin saber a qué lugar fueron (p. ej. importadas de una lista de despacho sin
// destino). El punto está aprobado (guarda sus entregas y cuenta en el historial de runners y acopios) pero no es
// público. Quien participó —el runner, quien despachó, el acopio que lo despachó, o alguien de sus equipos— o
// administración lo ubica: le pone zona / lugar / enlace (y pasa a ser un punto normal) o lo une a un punto que ya
// existe (sus entregas y contactos pasan a ese punto y este queda archivado). Todo queda en el historial de cambios.

var errNotInvolved = errors.New("Solo quien llevó o despachó esta entrega, o alguien de su equipo, puede ubicarla.")

func (a *App) locateRoutes(mux *http.ServeMux) {
	// Desde Runners, Supply o Checkup (sesión de código).
	mux.HandleFunc("POST /api/full/camps/{id}/locate", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		a.locate(w, r, &u)
	})
	// Las entregas por ubicar en las que participó quien mira (o su equipo), con lo entregado, para ubicarlas.
	mux.HandleFunc("GET /api/full/tolocate", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		type delivery struct {
			Need       string        `json:"need"`
			RunnerName string        `json:"runnerName,omitempty"`
			At         string        `json:"at,omitempty"`
			Items      []PackingLine `json:"items,omitempty"`
		}
		type item struct {
			ID         string     `json:"id"`
			Area       string     `json:"area"`
			Place      string     `json:"place,omitempty"`
			Extra      string     `json:"extra,omitempty"`
			People     string     `json:"people,omitempty"`
			Members    []Member   `json:"members,omitempty"`
			Created    string     `json:"created,omitempty"`
			Deliveries []delivery `json:"deliveries"`
		}
		list := []item{}
		err := a.db.View(func(tx *bolt.Tx) error {
			byCamp := map[string][]CampNeed{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) == nil {
					byCamp[n.CampID] = append(byCamp[n.CampID], n)
				}
				return nil
			}); e != nil {
				return e
			}
			return tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if json.Unmarshal(v, &p) != nil || p.Status != "approved" || !p.ToLocate || !involvedIn(tx, u, p) {
					return nil
				}
				it := item{ID: p.ID, Area: p.Area, Place: p.Place, Extra: p.Extra, People: p.People, Members: p.Members, Created: p.Created, Deliveries: []delivery{}}
				for _, n := range byCamp[p.ID] {
					it.Deliveries = append(it.Deliveries, delivery{Need: n.Need, RunnerName: nameOf(tx, n.RunnerID), At: n.CompletedAt, Items: n.Packing})
				}
				list = append(list, it)
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	// Administración (y agentes por /api/agent, con la misma regla de participación).
	mux.HandleFunc("POST /api/admin/camps/{id}/locate", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		a.locate(w, r, agentBy(r))
	})
}

// involvedIn: la persona (o alguien de su equipo) agregó el punto, llevó una de sus entregas o la despachó su acopio.
func involvedIn(tx *bolt.Tx, u User, p Point) bool {
	tv := newTeamView(tx, u.ID)
	if tv.mate(p.AddedBy) {
		return true
	}
	found := false
	tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if !found && json.Unmarshal(v, &n) == nil && n.CampID == p.ID && (tv.mate(n.RunnerID) || tv.mate(n.SupplierID) || tv.mate(n.DispatchedBy)) {
			found = true
		}
		return nil
	})
	return found
}

// locate: by nil = administración (siempre puede); si no, tiene que haber participado (involvedIn).
func (a *App) locate(w http.ResponseWriter, r *http.Request, by *User) {
	f, ok := decodeFields(w, r, map[string]int{"targetId": 128, "area": 80, "place": 200, "directions": 8192, "gps": 1000})
	if !ok {
		return
	}
	targetID := fieldText(f, "targetId")
	area, place, directions, gps := strings.TrimSpace(fieldText(f, "area")), strings.TrimSpace(fieldText(f, "place")), strings.TrimSpace(fieldText(f, "directions")), strings.TrimSpace(fieldText(f, "gps"))
	var coords []float64
	if targetID == "" {
		if area == "" {
			failField(w, 400, "area", "Escribe la zona (pueblo o comunidad) o elige un punto que ya existe.")
			return
		}
		if gps != "" {
			c, e := parseMapLink(gps) // fuera de la transacción: puede seguir un enlace corto
			if e != nil {
				failField(w, 400, "gps", "Usa un enlace de Google Maps de la zona.")
				return
			}
			coords = c
		}
	}
	id := r.PathValue("id")
	var result Point
	err := a.write(r, func(tx *bolt.Tx) error {
		b := tx.Bucket(submissions)
		var p Point
		if v := b.Get([]byte(id)); v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" || !p.ToLocate {
			return errMissing
		}
		if by != nil && !involvedIn(tx, *by, p) {
			return errNotInvolved
		}
		if targetID == "" {
			p.Area, p.Place, p.Directions, p.ToLocate = area, place, directions, false
			if gps != "" {
				p.GPS, p.Coords = gps, coords
			}
			result = p
			return putPointJSON(b, p)
		}
		var target Point
		if v := b.Get([]byte(targetID)); v == nil || json.Unmarshal(v, &target) != nil || target.Status != "approved" || target.ToLocate || target.ID == p.ID {
			return userFieldError{"targetId", "Elige un punto aprobado que ya tenga ubicación."}
		}
		// Sus entregas pasan al punto elegido (cada una queda en el historial de cambios).
		var moving []CampNeed
		if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
			var n CampNeed
			if json.Unmarshal(v, &n) == nil && n.CampID == p.ID {
				moving = append(moving, n)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, n := range moving {
			n.CampID = target.ID
			if e := putNeed(tx.Bucket(needsBucket), &n); e != nil {
				return e
			}
		}
		// Sus contactos se suman a los del punto (sin repetir nombres), hasta 50.
		for _, m := range p.Members {
			dup := false
			for _, t := range target.Members {
				if runnerNameKey(t.Name) == runnerNameKey(m.Name) {
					dup = true
				}
			}
			if !dup && len(target.Members) < 50 {
				target.Members = append(target.Members, m)
			}
		}
		if e := putPointJSON(b, target); e != nil {
			return e
		}
		p.Status, p.MergedInto, p.ToLocate = "archived", target.ID, false
		result = target
		return putPointJSON(b, p)
	})
	switch {
	case errors.Is(err, errNotInvolved):
		fail(w, 403, err.Error())
		return
	case errors.Is(err, errMissing):
		fail(w, 404, "Esta entrega ya tiene ubicación o el punto no existe.")
		return
	}
	if serviceError(w, err) {
		return
	}
	response(w, 200, map[string]any{"id": result.ID, "code": result.Code, "area": result.Area, "place": result.Place, "merged": targetID != ""})
}
