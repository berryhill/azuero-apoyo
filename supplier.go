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

// Acopio (supplier.html): la persona del centro de acopio ve las carreras en camino. Cuando llega el corredor, busca
// su carrera, arma la lista de empaque, toma la «foto de partida» y la marca «Despachada»: un paso dentro de «En
// camino» (la necesidad sigue IN_PROGRESS hasta que el corredor marca «Completé»). No toma ni completa carreras.
// Entra con su código (rol «Acopio»); la sesión es propia (supplier_session en /api/supplier).
const roleSupplier = "supplier"

var errNotARun = errors.New("Esta carrera ya no está en camino.")

// supplierRun: la carrera como la ve acopio, con quién la lleva y si ya salió del acopio.
type supplierRun struct {
	runnerNeed
	RunnerName       string `json:"runnerName,omitempty"`
	RunnerPhone      string `json:"runnerPhone,omitempty"`
	DispatchedAt     string `json:"dispatchedAt,omitempty"`
	DispatchedByName string `json:"dispatchedByName,omitempty"`
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

// supplierCanSee: carreras de suministros en camino con corredor asignado.
func supplierCanSee(n CampNeed) bool {
	return !isRepair(n) && n.Status == "IN_PROGRESS" && n.RunnerID != ""
}

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
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/supplier/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.supplierSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "csrf": s.csrf})
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
	// Carreras en camino de todos los equipos (el acopio atiende a cualquiera que llegue), con corredor y teléfono.
	mux.HandleFunc("GET /api/supplier/runs", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := a.supplierSession(w, r); !ok {
			return
		}
		runs := []supplierRun{}
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
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				camp, ok := camps[n.CampID]
				if !ok || !supplierCanSee(n) {
					return nil
				}
				if n.Need, _ = splitRepair(n.Need); n.Need == "" {
					return nil
				}
				run := supplierRun{runnerNeed: runnerNeed{ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, Packing: n.Packing, Media: n.Media, Kind: n.Kind, RepeatDays: n.RepeatDays, Assigned: true, Camp: camp}, DispatchedAt: n.DispatchedAt, DispatchedByName: n.DispatchedByName}
				if u, found := getUser(tx, n.RunnerID); found {
					run.RunnerName, run.RunnerPhone = u.Name, u.Telephone
				}
				runs = append(runs, run)
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
		response(w, 200, map[string]any{"runs": runs})
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
			if !a.campExists(tx, n.CampID) || !supplierCanSee(n) {
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
	inRun := func(tx *bolt.Tx, n CampNeed) error {
		if !supplierCanSee(n) {
			return errMissing
		}
		return nil
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
		a.receiveMedia(w, r, u.ID, u.Name, inRun)
	})
	mux.HandleFunc("GET /api/supplier/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := a.supplierSession(w, r); !ok {
			return
		}
		a.serveMedia(w, r, inRun)
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
			if !a.campExists(tx, n.CampID) || !supplierCanSee(n) {
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
