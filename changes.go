package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Solicitudes de cambio: un agente (Command) que no es del equipo de quien actualizó una necesidad por última vez
// (el runner que estuvo en el lugar, u otro agente) no la cambia directamente. El cambio queda como solicitud para esa
// persona, que la acepta o la rechaza en su página (Runners o Command): quien estuvo ahí puede saber algo que el agente
// no sabe. Así ni un agente sin equipo ni uno de otro equipo pisa los datos de un runner o de un equipo.

var changeRequestsBucket = []byte("change_requests")

func ensureChangeRequests(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(changeRequestsBucket)
	return err
}

type ChangeLine struct {
	Label  string `json:"label"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type ChangeRequest struct {
	ID        string                     `json:"id"`
	NeedID    string                     `json:"needId"`
	CampID    string                     `json:"campId"`
	CampCode  string                     `json:"campCode,omitempty"`
	CampArea  string                     `json:"campArea,omitempty"`
	Need      string                     `json:"need"`
	FromID    string                     `json:"fromId"`
	FromName  string                     `json:"fromName"`
	FromTeam  string                     `json:"fromTeam,omitempty"`
	ToID      string                     `json:"toId"`
	ToName    string                     `json:"toName"`
	Fields    map[string]json.RawMessage `json:"fields"`
	Changes   []ChangeLine               `json:"changes"`
	Status    string                     `json:"status"`
	CreatedAt string                     `json:"createdAt"`
	DecidedAt string                     `json:"decidedAt,omitempty"`
}

var errNotYourRequest = errors.New("Esta solicitud no es para ti o ya se respondió.")

// ownerOf: quien actualizó la necesidad por última vez (o, en las anteriores a este registro, su runner o constructor).
func ownerOf(n CampNeed) string {
	switch {
	case n.UpdatedBy != "":
		return n.UpdatedBy
	case n.RunnerID != "":
		return n.RunnerID
	}
	return n.BuilderID
}

var statusLabel = map[string]string{"OPEN": "Abierta", "IN_PROGRESS": "En camino", "COMPLETED": "Completada", "DELETED": "Quitada (ya no necesita nada)"}

func nameOf(tx *bolt.Tx, id string) string {
	if id == "" {
		return "sin asignar"
	}
	if u, ok := getUser(tx, id); ok {
		return u.Name
	}
	return "?"
}

// describeChanges: qué cambiaría el pedido f en n, en palabras («Estado: En camino → Completada»). Vacío si nada.
func describeChanges(tx *bolt.Tx, n CampNeed, f map[string]json.RawMessage) []ChangeLine {
	var out []ChangeLine
	add := func(label, before, after string) {
		if before != after {
			out = append(out, ChangeLine{label, before, after})
		}
	}
	for key := range f {
		switch key {
		case "need":
			add("Pedido", n.Need, fieldText(f, "need"))
		case "status":
			add("Estado", statusLabel[n.Status], statusLabel[fieldText(f, "status")])
		case "urgency":
			add("Urgencia", normalizeUrgency(n.Urgency), normalizeUrgency(fieldText(f, "urgency")))
		case "runnerId":
			add("Runner", nameOf(tx, n.RunnerID), nameOf(tx, fieldText(f, "runnerId")))
		case "builderId":
			add("Constructor", nameOf(tx, n.BuilderID), nameOf(tx, fieldText(f, "builderId")))
		case "kind":
			add("Tipo", n.Kind, storedKind(fieldText(f, "kind")))
		default:
			out = append(out, ChangeLine{"Otros datos (" + key + ")", "", string(f[key])})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

// askOwner: si el agente no puede cambiar la necesidad directamente, guarda (o renueva) su solicitud para quien la
// actualizó por última vez y la devuelve. nil = el agente puede cambiarla él mismo.
func (a *App) askOwner(tx *bolt.Tx, id string, f map[string]json.RawMessage, agent User) (*ChangeRequest, error) {
	var n CampNeed
	v := tx.Bucket(needsBucket).Get([]byte(id))
	if v == nil || json.Unmarshal(v, &n) != nil {
		return nil, nil
	}
	owner := ownerOf(n)
	if owner == "" || owner == agent.ID {
		return nil, nil
	}
	// Solo si esa persona puede responder (Runners o Command); si no, se cambia como antes.
	ou, ok := getUser(tx, owner)
	if !ok || ou.Disabled || !(ou.has(roleRunner) || ou.has(roleAgent)) {
		return nil, nil
	}
	if t := teamOf(tx, agent.ID); t != "" && t == teamOf(tx, owner) {
		return nil, nil
	}
	changes := describeChanges(tx, n, f)
	if len(changes) == 0 {
		return nil, nil
	}
	req := ChangeRequest{ID: randomToken(), NeedID: n.ID, CampID: n.CampID, Need: n.Need, FromID: agent.ID, FromName: agent.Name, ToID: owner, ToName: ou.Name, Fields: f, Changes: changes, Status: "pending", CreatedAt: nowStamp()}
	if t, ok := getTeam(tx, teamOf(tx, agent.ID)); ok {
		req.FromTeam = t.Name
	}
	if pv := tx.Bucket(submissions).Get([]byte(n.CampID)); pv != nil {
		var p Point
		if json.Unmarshal(pv, &p) == nil {
			req.CampCode, req.CampArea = p.Code, p.Area
		}
	}
	b := tx.Bucket(changeRequestsBucket)
	// Una sola solicitud pendiente por agente y necesidad: la nueva reemplaza a la anterior.
	if e := b.ForEach(func(k, v []byte) error {
		var old ChangeRequest
		if json.Unmarshal(v, &old) == nil && old.Status == "pending" && old.NeedID == n.ID && old.FromID == agent.ID {
			req.ID = old.ID
		}
		return nil
	}); e != nil {
		return nil, e
	}
	return &req, putJSON(b, req.ID, req)
}

// changeRequests: las pendientes para me y las que me envió (de los últimos 7 días), las más nuevas primero.
func (a *App) changeRequests(me string) (forMe, sent []ChangeRequest) {
	forMe, sent = []ChangeRequest{}, []ChangeRequest{}
	since := time.Now().Add(-7 * 24 * time.Hour)
	a.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(changeRequestsBucket).ForEach(func(_, v []byte) error {
			var c ChangeRequest
			if json.Unmarshal(v, &c) != nil {
				return nil
			}
			if c.ToID == me && c.Status == "pending" {
				forMe = append(forMe, c)
			}
			if t, e := time.Parse(time.RFC3339, c.CreatedAt); c.FromID == me && (c.Status == "pending" || e == nil && t.After(since)) {
				c.Fields = nil
				sent = append(sent, c)
			}
			return nil
		})
	})
	newest := func(l []ChangeRequest) {
		sort.SliceStable(l, func(i, j int) bool { return l[i].CreatedAt > l[j].CreatedAt })
	}
	newest(forMe)
	newest(sent)
	return forMe, sent
}

// decideChange: quien recibió la solicitud la acepta (se aplica como la habría hecho el agente, y la necesidad sigue
// siendo suya) o la rechaza (no cambia nada).
func (a *App) decideChange(me, id string, accept bool) (ChangeRequest, error) {
	var c ChangeRequest
	err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(changeRequestsBucket)
		v := b.Get([]byte(id))
		if v == nil || json.Unmarshal(v, &c) != nil || c.ToID != me || c.Status != "pending" {
			return errNotYourRequest
		}
		c.Status, c.DecidedAt = "rejected", nowStamp()
		if accept {
			from, ok := getUser(tx, c.FromID)
			if !ok {
				return errNotYourRequest
			}
			n, e := a.patchNeed(tx, c.NeedID, c.Fields, &from)
			if e != nil {
				return e
			}
			n.UpdatedBy = me
			if e := putNeed(tx.Bucket(needsBucket), &n); e != nil {
				return e
			}
			c.Status = "accepted"
		}
		return putJSON(b, c.ID, c)
	})
	return c, err
}

func changeError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errNotYourRequest) {
		fail(w, 409, err.Error())
		return true
	}
	return runError(w, err)
}

func (a *App) changeRoutes(mux *http.ServeMux, runners *crew) {
	list := func(w http.ResponseWriter, me string) {
		forMe, sent := a.changeRequests(me)
		response(w, 200, map[string]any{"forMe": forMe, "sent": sent})
	}
	decide := func(w http.ResponseWriter, r *http.Request, me string) {
		d := r.PathValue("decision")
		if d != "accept" && d != "reject" {
			fail(w, 404, "not found")
			return
		}
		c, err := a.decideChange(me, r.PathValue("id"), d == "accept")
		if changeError(w, err) {
			return
		}
		response(w, 200, c)
	}
	mux.HandleFunc("GET /api/runner/change-requests", func(w http.ResponseWriter, r *http.Request) {
		if s, ok := a.crewSession(runners, w, r); ok {
			list(w, s.runnerID)
		}
	})
	mux.HandleFunc("POST /api/runner/change-requests/{id}/{decision}", func(w http.ResponseWriter, r *http.Request) {
		if s, ok := a.crewSession(runners, w, r); ok && a.guard(w, r) {
			decide(w, r, s.runnerID)
		}
	})
	mux.HandleFunc("GET /api/agent/change-requests", func(w http.ResponseWriter, r *http.Request) {
		if u, _, ok := a.agentSession(w, r); ok {
			list(w, u.ID)
		}
	})
	mux.HandleFunc("POST /api/agent/change-requests/{id}/{decision}", func(w http.ResponseWriter, r *http.Request) {
		if u, _, ok := a.agentSession(w, r); ok {
			decide(w, r, u.ID)
		}
	})
}
