package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Acopio / Supply (supplier.html): la persona del centro de acopio liga carreras a su centro y las despacha.
//   - «Pedir corredor»: de las necesidades abiertas pide un corredor; a los corredores les llega en «Pedidos» con su
//     nombre, centro y teléfono. El que acepta («¡Yo lo llevo!») queda ligado a ese acopio.
//   - O «Asignar en persona»: el corredor ya está en el acopio; se le asigna directo (le aparece en «Yo»).
//   - Las carreras ligadas en camino salen en «Por despachar»: lista de empaque, foto de partida y «Despachada» (un
//     paso dentro de «En camino»; sigue IN_PROGRESS hasta que el corredor marca «Completé»).
//
// Cada acopio ve y despacha solo sus carreras ligadas. Entra con su código (rol «Acopio»; sesión supplier_session).
const roleSupplier = "supplier"

var errNotARun = errors.New("Esta carrera ya no está en camino.")

// supplierInfo: el acopio de una carrera, como lo ven los corredores (para llamar y saber dónde recoger).
type supplierInfo struct {
	Name   string `json:"name"`
	Phone  string `json:"phone,omitempty"`
	Center string `json:"center,omitempty"`
}

func supplierOf(tx *bolt.Tx, id string) *supplierInfo {
	if id == "" {
		return nil
	}
	u, ok := getUser(tx, id)
	if !ok {
		return nil
	}
	return &supplierInfo{Name: u.Name, Phone: u.Telephone, Center: u.SupplyCenter}
}

// supplierRun: la carrera como la ve acopio, con quién la lleva y si ya salió del acopio.
type supplierRun struct {
	runnerNeed
	RunnerName       string `json:"runnerName,omitempty"`
	RunnerPhone      string `json:"runnerPhone,omitempty"`
	DispatchedAt     string `json:"dispatchedAt,omitempty"`
	DispatchedByName string `json:"dispatchedByName,omitempty"`
	// Pedir corredor: el pedido es de este acopio (mine) o de otro (Supplier).
	RequestedAt string `json:"requestedAt,omitempty"`
	RequestMine bool   `json:"requestMine,omitempty"`
}

// supplierRunner: corredores para «Asignar en persona» (nombre, teléfono y cuántas carreras lleva).
type supplierRunner struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Active    int    `json:"active"`
}

func (a *App) startSupplierSession(w http.ResponseWriter, r *http.Request, userID string) runnerSession {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour)}
	a.mu.Lock()
	a.supplierSessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "supplier_session", Value: token, Path: "/api/supplier", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
	return s
}

// supplierSession: sesión vigente de un usuario activo con el rol «Acopio» y código; lo que no es GET lleva el token CSRF.
func (a *App) supplierSession(w http.ResponseWriter, r *http.Request) (User, runnerSession, bool) {
	var u User
	c, e := r.Cookie("supplier_session")
	if e != nil {
		fail(w, 401, "login required")
		return u, runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := a.supplierSessions[c.Value]
	a.mu.Unlock()
	active := false
	if ok && time.Now().Before(s.expires) {
		a.db.View(func(tx *bolt.Tx) error {
			var found bool
			u, found = getUser(tx, s.runnerID)
			active = found && !u.Disabled && u.has(roleSupplier) && u.CodeHash != ""
			return nil
		})
	}
	if !active {
		fail(w, 401, "login required")
		return User{}, runnerSession{}, false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("X-CSRF-Token") != s.csrf {
		fail(w, 403, "CSRF token required")
		return User{}, runnerSession{}, false
	}
	a.touch(u.ID)
	return u, s, true
}

// supplierOwns: carrera de suministros en camino, con corredor, ligada a este acopio.
func supplierOwns(n CampNeed, me string) bool {
	return !isRepair(n) && n.Status == "IN_PROGRESS" && n.RunnerID != "" && n.SupplierID == me
}

// supplierCanRequest: necesidad de suministros abierta, sin corredor, que no es un servicio programado.
func supplierCanRequest(n CampNeed) bool {
	return !isRepair(n) && n.Status == "OPEN" && n.RunnerID == "" && !serviceScheduled(n, time.Now())
}

// openRequest: cuándo pidió acopio un corredor, solo mientras la necesidad sigue esperando (Abierta).
func openRequest(n CampNeed) string {
	if n.Status != "OPEN" {
		return ""
	}
	return n.SupplierRequestedAt
}

var errRequestedElsewhere = errors.New("Otro acopio ya pidió un corredor para esta necesidad.")
var errNotOpen = errors.New("Esta necesidad ya no está abierta: otro corredor la tomó o cambió de estado.")
var errRunnerBusy = errors.New("Ese corredor ya lleva 10 carreras en camino.")

func (a *App) supplierRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/supplier/session", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"accessCode": 64})
		if !ok {
			return
		}
		code := normalizeAccessCode(fieldText(f, "accessCode"))
		if code == "" {
			failField(w, 400, "accessCode", "Escribe tu código de acceso.")
			return
		}
		u, ok := a.signInWithCode(w, r, code, "Código de acceso incorrecto. Revísalo o pide uno nuevo a administración.")
		if !ok {
			return
		}
		if !u.has(roleSupplier) {
			failField(w, 403, "accessCode", "Tu código no tiene acceso a Acopio. Pide a administración el rol «Acopio».")
			return
		}
		s := a.startSupplierSession(w, r, u.ID)
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "center": u.SupplyCenter, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/supplier/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.supplierSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "center": u.SupplyCenter, "csrf": s.csrf})
		}
	})
	mux.HandleFunc("POST /api/supplier/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("supplier_session"); e == nil {
			a.mu.Lock()
			delete(a.supplierSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "supplier_session", Value: "", Path: "/api/supplier", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	// Carreras de este acopio (en camino; por despachar o despachadas), necesidades abiertas para pedir corredor y
	// corredores para asignar en persona.
	mux.HandleFunc("GET /api/supplier/runs", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		runs, open, runners := []supplierRun{}, []supplierRun{}, []supplierRunner{}
		err := a.db.View(func(tx *bolt.Tx) error {
			camps := map[string]runnerCamp{}
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status == "approved" {
					contacts := campContacts(p)
					p = publicPoint(p)
					camps[p.ID] = runnerCamp{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, People: p.People, Directions: p.Directions, Coordinator: p.Coordinator, GPS: p.GPS, Coords: p.Coords, Contacts: contacts}
				}
				return nil
			}); e != nil {
				return e
			}
			active := map[string]int{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				if n.Status == "IN_PROGRESS" && n.RunnerID != "" && !isRepair(n) {
					active[n.RunnerID]++
				}
				camp, ok := camps[n.CampID]
				if !ok || !(supplierOwns(n, me.ID) || supplierCanRequest(n)) {
					return nil
				}
				if n.Need, _ = splitRepair(n.Need); n.Need == "" {
					return nil
				}
				run := supplierRun{runnerNeed: runnerNeed{ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, Packing: n.Packing, Media: n.Media, Kind: n.Kind, RepeatDays: n.RepeatDays, Assigned: true, Camp: camp}, DispatchedAt: n.DispatchedAt, DispatchedByName: n.DispatchedByName}
				if n.Status == "OPEN" {
					run.Media, run.Packing = nil, nil
					if n.SupplierRequestedAt != "" {
						run.RequestedAt, run.RequestMine = n.SupplierRequestedAt, n.SupplierID == me.ID
						if !run.RequestMine {
							run.Supplier = supplierOf(tx, n.SupplierID)
						}
					}
					open = append(open, run)
					return nil
				}
				if u, found := getUser(tx, n.RunnerID); found {
					run.RunnerName, run.RunnerPhone = u.Name, u.Telephone
				}
				runs = append(runs, run)
				return nil
			}); e != nil {
				return e
			}
			return eachUser(tx, func(u User) error {
				if u.has(roleRunner) && !u.Disabled {
					runners = append(runners, supplierRunner{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Active: active[u.ID]})
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		// Primero lo que falta despachar (la carrera más reciente arriba: el corredor que acaba de llegar), luego lo despachado.
		sort.SliceStable(runs, func(i, j int) bool {
			if (runs[i].DispatchedAt == "") != (runs[j].DispatchedAt == "") {
				return runs[i].DispatchedAt == ""
			}
			if runs[i].DispatchedAt != "" {
				return runs[i].DispatchedAt > runs[j].DispatchedAt
			}
			return runs[i].EnRouteAt > runs[j].EnRouteAt
		})
		// Para pedir corredor: urgentes primero, luego lo que más tiempo lleva esperando.
		sort.SliceStable(open, func(i, j int) bool {
			if ri, rj := urgencyRank(open[i].Urgency), urgencyRank(open[j].Urgency); ri != rj {
				return ri > rj
			}
			return open[i].CreatedAt < open[j].CreatedAt
		})
		sort.SliceStable(runners, func(i, j int) bool { return runnerNameKey(runners[i].Name) < runnerNameKey(runners[j].Name) })
		response(w, 200, map[string]any{"runs": runs, "open": open, "runners": runners, "center": me.SupplyCenter})
	})
	// Nombre del centro de acopio de esta persona (texto; los corredores lo ven en los pedidos y en su carrera).
	mux.HandleFunc("POST /api/supplier/center", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"center": 120})
		if !ok {
			return
		}
		center := strings.Join(strings.Fields(fieldText(f, "center")), " ")
		err := a.db.Update(func(tx *bolt.Tx) error {
			u, found := getUser(tx, me.ID)
			if !found {
				return errMissing
			}
			u.SupplyCenter, u.UpdatedAt = center, nowStamp()
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]string{"center": center})
	})
	// Pedir corredor (o cancelar el pedido): la necesidad sigue Abierta y aparece en «Pedidos» de los corredores.
	mux.HandleFunc("POST /api/supplier/needs/{id}/request", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"cancel": 0})
		if !ok {
			return
		}
		var cancel bool
		if raw, has := f["cancel"]; has && json.Unmarshal(raw, &cancel) != nil {
			failField(w, 400, "cancel", "Valor inválido.")
			return
		}
		var n CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || !supplierCanRequest(n) {
				return errNotOpen
			}
			if n.SupplierRequestedAt != "" && n.SupplierID != me.ID {
				return errRequestedElsewhere
			}
			if cancel {
				n.SupplierID, n.SupplierRequestedAt = "", ""
			} else {
				n.SupplierID, n.SupplierRequestedAt = me.ID, nowStamp()
			}
			return putNeed(b, &n)
		})
		if errors.Is(err, errNotOpen) || errors.Is(err, errRequestedElsewhere) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "requestedAt": n.SupplierRequestedAt})
	})
	// Asignar en persona: el corredor ya está en el acopio; la carrera pasa a En camino, a su «Yo» y a «Por despachar».
	mux.HandleFunc("POST /api/supplier/needs/{id}/assign", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"runnerId": 128})
		if !ok {
			return
		}
		runnerID := fieldText(f, "runnerId")
		var n CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			runner, found := getUser(tx, runnerID)
			if !found || runner.Disabled || !runner.has(roleRunner) {
				return userFieldError{"runnerId", "Elige un corredor activo."}
			}
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || !supplierCanRequest(n) {
				return errNotOpen
			}
			if n.SupplierRequestedAt != "" && n.SupplierID != me.ID {
				return errRequestedElsewhere
			}
			busy := 0
			if e := b.ForEach(func(_, v []byte) error {
				var other CampNeed
				if e := json.Unmarshal(v, &other); e != nil {
					return e
				}
				if other.RunnerID == runnerID && other.Status == "IN_PROGRESS" && !isRepair(other) {
					busy++
				}
				return nil
			}); e != nil {
				return e
			}
			if busy >= 10 {
				return errRunnerBusy
			}
			n.RunnerID, n.Status, n.EnRouteAt, n.SupplierID = runnerID, "IN_PROGRESS", nowStamp(), me.ID
			if e := putNeed(b, &n); e != nil {
				return e
			}
			// Al final del orden del corredor, como si la hubiera tomado.
			runner.RunOrder = append(runner.RunOrder, n.ID)
			return putJSON(tx.Bucket(usersBucket), runner.ID, runner)
		})
		if errors.Is(err, errNotOpen) || errors.Is(err, errRequestedElsewhere) || errors.Is(err, errRunnerBusy) {
			fail(w, 409, err.Error())
			return
		}
		if userError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "status": n.Status})
	})
	mux.HandleFunc("GET /api/supplier/packing-items", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := a.supplierSession(w, r); !ok {
			return
		}
		var list []PackingItem
		err := a.db.View(func(tx *bolt.Tx) (e error) { list, e = packingItems(tx); return e })
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/supplier/needs/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		u, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"items": 0})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		var n CampNeed
		err = a.db.Update(func(tx *bolt.Tx) error {
			v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || !supplierOwns(n, u.ID) {
				return errNotARun
			}
			return addPacking(tx, &n, lines, u.ID, u.Name)
		})
		if errors.Is(err, errNotARun) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
	inRun := func(me string) func(tx *bolt.Tx, n CampNeed) error {
		return func(tx *bolt.Tx, n CampNeed) error {
			if !supplierOwns(n, me) {
				return errMissing
			}
			return nil
		}
	}
	// Foto de partida: solo esa etapa (la de entrega es del corredor en el punto).
	mux.HandleFunc("POST /api/supplier/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		u, _, ok := a.supplierSession(w, r)
		if !ok || !sameOrigin(w, r) {
			return
		}
		if r.URL.Query().Get("stage") != "pickup" {
			failField(w, 400, "stage", "Acopio sube solo fotos de partida.")
			return
		}
		a.receiveMedia(w, r, u.ID, u.Name, inRun(u.ID))
	})
	mux.HandleFunc("GET /api/supplier/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		a.serveMedia(w, r, inRun(me.ID))
	})
	// «Despachada»: salió del acopio con su carga (sigue en camino). undo la quita si se marcó por error.
	mux.HandleFunc("POST /api/supplier/needs/{id}/dispatch", func(w http.ResponseWriter, r *http.Request) {
		u, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"undo": 0})
		if !ok {
			return
		}
		var undo bool
		if raw, has := f["undo"]; has && json.Unmarshal(raw, &undo) != nil {
			failField(w, 400, "undo", "Valor inválido.")
			return
		}
		var n CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || !supplierOwns(n, u.ID) {
				return errNotARun
			}
			if undo {
				n.DispatchedAt, n.DispatchedBy, n.DispatchedByName = "", "", ""
			} else {
				n.DispatchedAt, n.DispatchedBy, n.DispatchedByName = nowStamp(), u.ID, strings.TrimSpace(u.Name)
			}
			return putNeed(b, &n)
		})
		if errors.Is(err, errNotARun) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "dispatchedAt": n.DispatchedAt, "dispatchedByName": n.DispatchedByName})
	})
}
