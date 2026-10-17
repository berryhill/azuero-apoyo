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
	Kind      string                     `json:"kind,omitempty"` // necesidad (vacío en las anteriores) o punto
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

// teamDataError: un agente intenta algo que no se puede pedir como solicitud (borrar o cambiar fotos, fusionar,
// publicar un punto retenido) sobre datos de otra persona o equipo.
type teamDataError struct{ owner string }

func (e teamDataError) Error() string {
	return "Esto es de " + e.owner + ": solo esa persona, su equipo o administración lo pueden cambiar."
}

// mediaGuard: un agente cambia o borra una foto solo si la subió él, alguien de su equipo o administración.
func mediaGuard(tx *bolt.Tx, r *http.Request, m MediaItem) error {
	by := agentBy(r)
	if by == nil || m.AddedBy == "" || m.AddedBy == "admin" || m.AddedBy == by.ID {
		return nil
	}
	if t := teamOf(tx, by.ID); t != "" && t == teamOf(tx, m.AddedBy) {
		return nil
	}
	return teamDataError{nameOf(tx, m.AddedBy)}
}

var errNotYourRequest = errors.New("Esta solicitud no es para ti o ya se respondió.")

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

// whoDecides: entre los candidatos (quien actualizó por última vez y, si no puede responder, quien lleva la carrera),
// el primero que puede responder en su página (Runners, Constructores o Command). nil = el agente cambia directamente:
// es esa misma persona, es de su equipo (los datos entre compañeros de equipo son de confianza) o no hay nadie que
// pueda responder.
func whoDecides(tx *bolt.Tx, agent User, candidates ...string) *User {
	for _, id := range candidates {
		if id == "" {
			continue
		}
		if id == agent.ID {
			return nil
		}
		u, ok := getUser(tx, id)
		if !ok || u.Disabled || !(u.has(roleRunner) || u.has(roleAgent) || u.has(roleBuilder)) {
			continue
		}
		if t := teamOf(tx, agent.ID); t != "" && t == teamOf(tx, id) {
			return nil
		}
		return &u
	}
	return nil
}

// saveRequest guarda (o renueva: una pendiente por agente y registro) la solicitud de cambio.
func saveRequest(tx *bolt.Tx, req ChangeRequest, agent User, owner User) (*ChangeRequest, error) {
	req.FromID, req.FromName, req.ToID, req.ToName, req.Status, req.CreatedAt = agent.ID, agent.Name, owner.ID, owner.Name, "pending", nowStamp()
	req.ID = randomToken()
	if t, ok := getTeam(tx, teamOf(tx, agent.ID)); ok {
		req.FromTeam = t.Name
	}
	if pv := tx.Bucket(submissions).Get([]byte(req.CampID)); pv != nil {
		var p Point
		if json.Unmarshal(pv, &p) == nil {
			req.CampCode, req.CampArea = p.Code, p.Area
		}
	}
	b := tx.Bucket(changeRequestsBucket)
	if e := b.ForEach(func(k, v []byte) error {
		var old ChangeRequest
		if json.Unmarshal(v, &old) == nil && old.Status == "pending" && old.Kind == req.Kind && old.NeedID == req.NeedID && old.CampID == req.CampID && old.FromID == agent.ID {
			req.ID = old.ID
		}
		return nil
	}); e != nil {
		return nil, e
	}
	return &req, putJSON(b, req.ID, req)
}

// askOwner: si el agente no puede cambiar la necesidad directamente, guarda su solicitud para quien la actualizó por
// última vez (o quien lleva la carrera) y la devuelve. nil = el agente puede cambiarla él mismo.
func (a *App) askOwner(tx *bolt.Tx, id string, f map[string]json.RawMessage, agent User) (*ChangeRequest, error) {
	var n CampNeed
	v := tx.Bucket(needsBucket).Get([]byte(id))
	if v == nil || json.Unmarshal(v, &n) != nil {
		return nil, nil
	}
	owner := whoDecides(tx, agent, n.UpdatedBy, n.RunnerID, n.BuilderID)
	if owner == nil {
		return nil, nil
	}
	changes := describeChanges(tx, n, f)
	if len(changes) == 0 {
		return nil, nil
	}
	return saveRequest(tx, ChangeRequest{Kind: "necesidad", NeedID: n.ID, CampID: n.CampID, Need: n.Need, Fields: f, Changes: changes}, agent, *owner)
}

// askPointOwner: lo mismo para los datos de un punto: decide quien lo editó por última vez desde Command o, si no,
// quien lo agregó con su código.
func (a *App) askPointOwner(tx *bolt.Tx, id string, f map[string]json.RawMessage, agent User) (*ChangeRequest, error) {
	var p Point
	v := tx.Bucket(submissions).Get([]byte(id))
	if v == nil || json.Unmarshal(v, &p) != nil {
		return nil, nil
	}
	owner := whoDecides(tx, agent, p.UpdatedBy, p.AddedBy)
	if owner == nil {
		return nil, nil
	}
	// Qué cambiaría: se aplica a una copia y se compara campo por campo, como en el historial.
	before, _ := json.Marshal(p)
	after := p
	after.Members = append([]Member(nil), p.Members...)
	if e := applyCampFields(&after, f); e != nil {
		return nil, e
	}
	if e := uniqueCode(tx, &after); e != nil {
		return nil, e
	}
	after.History = p.History
	data, _ := json.Marshal(after)
	changes := diffJSON(tx, before, data)
	if len(changes) == 0 {
		return nil, nil
	}
	return saveRequest(tx, ChangeRequest{Kind: "punto", CampID: p.ID, Need: "Datos del punto", Fields: f, Changes: changes}, agent, *owner)
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
func (a *App) decideChange(r *http.Request, me, id string, accept bool) (ChangeRequest, error) {
	var c ChangeRequest
	err := a.write(r, func(tx *bolt.Tx) error {
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
			// En el historial queda como cambio del agente, aceptado por quien lo recibió; la necesidad o el punto sigue siendo suyo.
			who, _ := actorIn(tx)
			setActor(tx, Actor{from.ID, who.Via + " · aceptado por " + nameOf(tx, me)})
			var e error
			if c.Kind == "punto" {
				_, e = a.patchCamp(tx, c.CampID, c.Fields, &from, me)
			} else {
				_, e = a.patchNeed(tx, c.NeedID, c.Fields, &from, me)
			}
			if e != nil {
				return e
			}
			setActor(tx, who)
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

func (a *App) changeRoutes(mux *http.ServeMux, runners, builders *crew) {
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
		c, err := a.decideChange(r, me, r.PathValue("id"), d == "accept")
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
	mux.HandleFunc("GET /api/builder/change-requests", func(w http.ResponseWriter, r *http.Request) {
		if s, ok := a.crewSession(builders, w, r); ok {
			list(w, s.runnerID)
		}
	})
	mux.HandleFunc("POST /api/builder/change-requests/{id}/{decision}", func(w http.ResponseWriter, r *http.Request) {
		if s, ok := a.crewSession(builders, w, r); ok && a.guard(w, r) {
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
