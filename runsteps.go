package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Carrera paso a paso (Runners → «Yo»): 1) recoger en el acopio — la lista de empaque y la foto de partida se arman
// ahí, runner y acopio a la vez, hasta que el acopio la marca «Despachada»; 2) ir al lugar; 3) entregar. Después de
// despachada, la carga ya no se cambia directo: el runner pide el cambio y el acopio lo acepta o no. Y si no puede
// seguir con la carga encima: la devuelve al acopio o la lleva a otro lugar (que se elige como en «Nuevo despacho»).

type PackingChange struct {
	Items  []PackingLine `json:"items"`
	Note   string        `json:"note,omitempty"`
	By     string        `json:"by,omitempty"`
	ByName string        `json:"byName,omitempty"`
	At     string        `json:"at"`
	// Al responder: "accepted" o "rejected", quién del acopio, cuándo y por qué (obligatorio al rechazar).
	Decision      string `json:"decision,omitempty"`
	DecidedByName string `json:"decidedByName,omitempty"`
	DecidedAt     string `json:"decidedAt,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

var errPackingDispatched = errors.New("Esta carga ya salió del acopio: pide el cambio al acopio («✏️ Pedir cambio»).")
var errNotDispatched = errors.New("Esta carrera no salió de un acopio: puedes cambiar la lista directamente.")

func (a *App) runStepRoutes(mux *http.ServeMux, runners *crew) {
	// Runners → 🕓 Historial: las carreras que esta persona ya entregó (la última primero), con el lugar, la carga y las
	// fotos, para revisarlas o corregir algo (p. ej. agregar fotos de entrega después).
	mux.HandleFunc("GET /api/runner/history", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		type place struct {
			ID    string `json:"id"`
			Code  string `json:"code,omitempty"`
			Area  string `json:"area,omitempty"`
			Place string `json:"place,omitempty"`
		}
		type item struct {
			ID           string        `json:"id"`
			Need         string        `json:"need"`
			CompletedAt  string        `json:"completedAt"`
			DispatchedAt string        `json:"dispatchedAt,omitempty"`
			SupplyCenter string        `json:"supplyCenter,omitempty"`
			Packing      []PackingLine `json:"packing,omitempty"`
			Media        []MediaItem   `json:"media,omitempty"`
			Camp         place         `json:"camp"`
		}
		list := []item{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil || n.RunnerID != s.runnerID || n.Status != "COMPLETED" || isRepair(n) {
					return nil
				}
				var p Point
				json.Unmarshal(tx.Bucket(submissions).Get([]byte(n.CampID)), &p)
				list = append(list, item{ID: n.ID, Need: n.Need, CompletedAt: n.CompletedAt, DispatchedAt: n.DispatchedAt, SupplyCenter: n.SupplyCenter, Packing: n.Packing, Media: n.Media, Camp: place{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place}})
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.Slice(list, func(i, j int) bool { return list[i].CompletedAt > list[j].CompletedAt })
		if len(list) > 200 {
			list = list[:200]
		}
		response(w, 200, list)
	})
	// Supply → Historial: alguien del centro completa una carrera que sigue «en camino» (el runner no la cerró), con el
	// mismo flujo que el runner: fotos de entrega (aparte, /media?stage=delivery) y qué necesitan ahora (need, opcional,
	// queda como necesidad nueva y abierta). Queda en el historial de cambios con quién la completó.
	mux.HandleFunc("POST /api/supplier/needs/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 20})
		if !ok {
			return
		}
		follow, urgency := strings.TrimSpace(fieldText(f, "need")), normalizeUrgency(fieldText(f, "urgency"))
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil || !centerOwns(tx, n, me.ID) {
				return errMissing
			}
			n.Status, n.CompletedAt = "COMPLETED", nowStamp()
			if e := putNeed(b, &n); e != nil {
				return e
			}
			if follow == "" {
				return nil
			}
			m := CampNeed{ID: randomToken(), CampID: n.CampID, Need: follow, Status: "OPEN", Urgency: urgency, CreatedAt: nowStamp(), CreatedBy: me.ID, CreatedByName: me.Name}
			return putNeed(b, &m)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "status": n.Status})
	})
	// El runner pide cambiar la carga ya despachada (la lista completa como debería quedar, y por qué).
	mux.HandleFunc("POST /api/runner/needs/{id}/packing/change", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"items": 0, "note": 600})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		var n CampNeed
		err = a.write(r, func(tx *bolt.Tx) error {
			u, _ := getUser(tx, s.runnerID)
			v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil || n.RunnerID != s.runnerID || n.Status != "IN_PROGRESS" {
				return errMissing
			}
			if n.SupplierID == "" || n.DispatchedAt == "" {
				return errNotDispatched
			}
			items := make([]PackingLine, 0, len(lines))
			for _, l := range lines {
				items = append(items, PackingLine{Item: l.Item, Quantity: l.Quantity, AddedBy: u.ID, AddedByName: u.Name})
			}
			note := strings.TrimSpace(fieldText(f, "note"))
			// Ya había un pedido sin responder: este lo reemplaza (lista nueva) y su motivo se suma al anterior.
			if old := n.PackingChange; old != nil && old.Note != "" {
				note = strings.TrimSpace(old.Note + "\n" + note)
			}
			n.PackingChange = &PackingChange{Items: items, Note: note, By: u.ID, ByName: u.Name, At: nowStamp()}
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if errors.Is(err, errNotDispatched) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packingChange": n.PackingChange})
	})
	// El acopio que la despachó acepta (la lista queda como la pidió el runner) o rechaza el cambio.
	mux.HandleFunc("POST /api/supplier/needs/{id}/packing/change", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"accept": 0, "reason": 600})
		if !ok {
			return
		}
		accept, reason := strings.TrimSpace(string(f["accept"])) == "true", strings.TrimSpace(fieldText(f, "reason"))
		if !accept && reason == "" {
			failField(w, 400, "reason", "Escribe por qué lo rechazas: el runner lo verá.")
			return
		}
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil || n.PackingChange == nil || !newTeamView(tx, me.ID).mate(n.SupplierID) {
				return errMissing
			}
			now := nowStamp()
			done := *n.PackingChange
			done.Decision, done.DecidedByName, done.DecidedAt, done.Reason = "rejected", me.Name, now, reason
			if accept {
				for i := range n.PackingChange.Items {
					n.PackingChange.Items[i].AddedAt = now
				}
				n.Packing = n.PackingChange.Items
				done.Decision = "accepted"
			}
			n.PackingChanges = append(n.PackingChanges, done)
			n.PackingChange = nil
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
	// «No puedo continuar» con carga despachada: action "return" (la devuelve al acopio: la necesidad vuelve a la lista
	// y el acopio la tiene otra vez en «Despachando» para darla a otro runner) o "redirect" (la lleva a otro lugar:
	// pointId de un punto, o point{area,…; toLocate} nuevo; esa carga sigue como una carrera nueva y la necesidad
	// original vuelve a la lista, sin carga).
	mux.HandleFunc("POST /api/runner/needs/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"action": 16, "pointId": 128, "point": 0, "note": 600})
		if !ok {
			return
		}
		action, note, pointID := fieldText(f, "action"), strings.TrimSpace(fieldText(f, "note")), fieldText(f, "pointId")
		var newPoint struct {
			Area, Place, Directions string
			ToLocate                bool
		}
		if action != "return" && action != "redirect" {
			failField(w, 400, "action", "Elige qué haces con la carga.")
			return
		}
		if note == "" {
			failField(w, 400, "note", "Cuenta qué pasó: el acopio lo necesita saber.")
			return
		}
		if action == "redirect" && pointID == "" && (f["point"] == nil || json.Unmarshal(f["point"], &newPoint) != nil || strings.TrimSpace(newPoint.Area) == "") {
			failField(w, 400, "point", "Elige el lugar o escribe al menos la zona.")
			return
		}
		var out CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			var n CampNeed
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil || n.RunnerID != s.runnerID || n.Status != "IN_PROGRESS" {
				return errMissing
			}
			u, _ := getUser(tx, s.runnerID)
			now := nowStamp()
			release := func(x *CampNeed) {
				x.RunnerID, x.Status, x.EnRouteAt, x.ReleasedAt, x.ReleasedBy, x.StopNote = "", "OPEN", "", now, s.runnerID, note
				x.DispatchedAt, x.DispatchedBy, x.DispatchedByName, x.PackingChange = "", "", "", nil
			}
			if action == "return" {
				// Vuelve a la lista como pedido de ese acopio (su oferta «Necesito runner»): otro runner la recoge ahí.
				release(&n)
				n.ReturnedAt, n.SupplierRequestedAt = now, now
				out = n
				return putNeed(b, &n)
			}
			dest := pointID
			if dest != "" {
				if !a.campExists(tx, dest) || dest == n.CampID {
					return userFieldError{"pointId", "Elige otro punto publicado."}
				}
			} else {
				p := Point{ID: randomToken(), Area: strings.Join(strings.Fields(newPoint.Area), " "), Place: strings.TrimSpace(newPoint.Place), Directions: strings.TrimSpace(newPoint.Directions),
					Source: "user", Status: "approved", Created: now, Decided: now, AddedBy: u.ID, AddedByName: u.Name, Note: "Creado por el runner al llevar ahí una carga que no pudo entregar en otro lugar.", ToLocate: newPoint.ToLocate}
				if p.ToLocate {
					p.Place = ""
				}
				codeFor(tx, &p)
				verifyStamp(tx, &u, &p.VerifiedTeam, &p.VerifiedBy, &p.VerifiedAt)
				if e := putPointJSON(tx.Bucket(submissions), p); e != nil {
					return e
				}
				dest = p.ID
			}
			m := CampNeed{ID: randomToken(), CampID: dest, Need: n.Need, Status: "IN_PROGRESS", Urgency: n.Urgency, Kind: n.Kind, RunnerID: n.RunnerID, SupplierID: n.SupplierID,
				CreatedAt: now, EnRouteAt: now, DispatchedAt: n.DispatchedAt, DispatchedBy: n.DispatchedBy, DispatchedByName: n.DispatchedByName, SupplyCenter: n.SupplyCenter,
				Packing: n.Packing, Media: n.Media, RedirectedFrom: n.ID, StopNote: note, CreatedBy: u.ID, CreatedByName: u.Name}
			if e := putNeed(b, &m); e != nil {
				return e
			}
			release(&n)
			n.Packing, n.Media, n.RedirectedTo = nil, nil, m.ID
			if e := putNeed(b, &n); e != nil {
				return e
			}
			for i, id := range u.RunOrder {
				if id == n.ID {
					u.RunOrder[i] = m.ID
				}
			}
			out = m
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": out.ID, "campId": out.CampID, "status": out.Status})
	})
}
