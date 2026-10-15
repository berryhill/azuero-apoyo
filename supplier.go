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

// Centros de acopio guardados: la lista compartida para elegir el nombre del centro (clave: nombre sin tildes ni
// mayúsculas, valor: el nombre como se escribió la primera vez).
var supplyCentersBucket = []byte("supply_centers")

func ensureSupplyCenters(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(supplyCentersBucket)
	return err
}

// supplyCenters: los centros guardados más los que ya tienen las personas de Supply, sin repetir, en orden alfabético.
func supplyCenters(tx *bolt.Tx) ([]string, error) {
	seen := map[string]string{}
	if b := tx.Bucket(supplyCentersBucket); b != nil {
		if e := b.ForEach(func(k, v []byte) error { seen[string(k)] = string(v); return nil }); e != nil {
			return nil, e
		}
	}
	if e := eachUser(tx, func(u User) error {
		if c := strings.TrimSpace(u.SupplyCenter); c != "" {
			if _, ok := seen[foldText(c)]; !ok {
				seen[foldText(c)] = c
			}
		}
		return nil
	}); e != nil {
		return nil, e
	}
	list := make([]string, 0, len(seen))
	for _, v := range seen {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool { return foldText(list[i]) < foldText(list[j]) })
	return list, nil
}

var errNotARun = errors.New("Este run ya no está en camino.")

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
	// El runner es de mi equipo (para «Solo mi equipo»; el punto lo dice camp.team).
	RunnerTeam bool `json:"runnerTeam,omitempty"`
	// Dónde va el corredor, si comparte su ubicación en vivo (reciente); solo en las carreras de este acopio.
	RunnerLive *runnerOnMap `json:"runnerLive,omitempty"`
	// «Corredores piden»: qué respondió este acopio («can» o «no»; vacío si no ha respondido) y cuántos más pueden.
	MyOffer   string `json:"myOffer,omitempty"`
	OfferedBy int    `json:"offeredBy,omitempty"`
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
	// Team: es de mi equipo (para «Solo mi equipo»).
	Team bool `json:"team,omitempty"`
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

var errRequestedElsewhere = errors.New("Otro acopio ya pidió un runner para esta necesidad.")
var errNotOpen = errors.New("Esta necesidad ya no está abierta: otro runner la tomó o cambió de estado.")
var errRunnerBusy = errors.New("Ese runner ya lleva 10 runs en camino.")

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
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "panels": userPanels(u), "center": u.SupplyCenter, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/supplier/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.supplierSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "panels": userPanels(u), "center": u.SupplyCenter, "csrf": s.csrf})
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
		runs, open, runners, asks, enroute := []supplierRun{}, []supplierRun{}, []supplierRunner{}, []supplierRun{}, []supplierRun{}
		now := time.Now()
		live := map[string]liveLocation{}
		a.mu.Lock()
		for id, loc := range a.locations {
			if now.Sub(loc.At) <= locationTTL {
				live[id] = loc
			}
		}
		a.mu.Unlock()
		err := a.db.View(func(tx *bolt.Tx) error {
			// «Solo mi equipo»: puntos que agregó alguien de mi equipo (o con un run de mi equipo) y runners de mi equipo.
			tv := newTeamView(tx, me.ID)
			team, e := teamCamps(tv)
			if e != nil {
				return e
			}
			groups, e := needGroups(tx)
			if e != nil {
				return e
			}
			camps := map[string]runnerCamp{}
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status == "approved" {
					contacts := campContacts(p)
					p = publicPoint(p)
					camps[p.ID] = runnerCamp{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, People: p.People, Directions: p.Directions, Coordinator: p.Coordinator, GPS: p.GPS, Coords: p.Coords, Contacts: contacts, Team: team[p.ID]}
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
				asking := supplyAskOpen(n)
				// En camino con otro acopio (o sin acopio): se puede mandar otro runner a lo mismo.
				other := !isRepair(n) && n.Status == "IN_PROGRESS" && n.RunnerID != "" && n.SupplierID != me.ID && !asking
				if !ok || !(supplierOwns(n, me.ID) || supplierCanRequest(n) || asking || other) {
					return nil
				}
				if n.Need, _ = splitRepair(n.Need); n.Need == "" {
					return nil
				}
				run := supplierRun{runnerNeed: runnerNeed{Also: alsoFor(tx, groups, n), ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, Packing: n.Packing, Media: n.Media, Kind: n.Kind, RepeatDays: n.RepeatDays, Assigned: true, Camp: camp}, DispatchedAt: n.DispatchedAt, DispatchedByName: n.DispatchedByName}
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
				if asking {
					// Un acopio que dijo que no, ya no la ve; los demás ven si ya respondieron y cuántos pueden.
					o := myOffer(n, me.ID)
					if o != nil && !o.Can {
						return nil
					}
					if o != nil {
						run.MyOffer = "can"
					}
					for _, x := range n.SupplyOffers {
						if x.Can {
							run.OfferedBy++
						}
					}
					run.Media, run.Packing, run.RequestedAt = nil, nil, n.SupplyAskedAt
				}
				run.RunnerTeam = tv.mate(n.RunnerID)
				if u, found := getUser(tx, n.RunnerID); found {
					run.RunnerName, run.RunnerPhone = u.Name, u.Telephone
					if loc, ok := live[n.RunnerID]; ok {
						run.RunnerLive = &runnerOnMap{ID: u.ID, Name: u.Name, Lat: loc.Lat, Lon: loc.Lon, Accuracy: loc.Accuracy, At: loc.At.UTC().Format(time.RFC3339), Seconds: int(now.Sub(loc.At).Seconds())}
					}
				}
				switch {
				case asking:
					asks = append(asks, run)
				case other:
					run.Media, run.Packing, run.RunnerLive = nil, nil, nil
					run.Supplier = supplierOf(tx, n.SupplierID)
					enroute = append(enroute, run)
				default:
					runs = append(runs, run)
				}
				return nil
			}); e != nil {
				return e
			}
			return eachUser(tx, func(u User) error {
				if u.has(roleRunner) && !u.Disabled {
					runners = append(runners, supplierRunner{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Active: active[u.ID], Team: tv.mate(u.ID)})
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
		var centers []string
		if e := a.db.View(func(tx *bolt.Tx) (e error) { centers, e = supplyCenters(tx); return e }); serviceError(w, e) {
			return
		}
		response(w, 200, map[string]any{"runs": runs, "open": open, "asks": asks, "enroute": enroute, "runners": runners, "center": me.SupplyCenter, "centers": centers})
	})
	// «Corredores piden»: este acopio puede surtir la carrera (can), no puede (can:false) o retira su oferta (withdraw).
	mux.HandleFunc("POST /api/supplier/needs/{id}/offer", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"can": 0, "withdraw": 0})
		if !ok {
			return
		}
		var can, withdraw bool
		json.Unmarshal(f["can"], &can)
		json.Unmarshal(f["withdraw"], &withdraw)
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			var n CampNeed
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || !supplyAskOpen(n) {
				return errNoAsk
			}
			kept := n.SupplyOffers[:0:0]
			for _, o := range n.SupplyOffers {
				if o.SupplierID != me.ID {
					kept = append(kept, o)
				}
			}
			if !withdraw {
				kept = append(kept, SupplyOffer{SupplierID: me.ID, Can: can, At: nowStamp()})
			}
			n.SupplyOffers = kept
			return putNeed(b, &n)
		})
		if errors.Is(err, errNoAsk) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]bool{"ok": true})
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
			// Un centro nuevo queda en la lista compartida; si ya existe con otras tildes o mayúsculas, se usa ese nombre.
			if center != "" {
				b := tx.Bucket(supplyCentersBucket)
				if v := b.Get([]byte(foldText(center))); v != nil {
					center = string(v)
				} else if e := b.Put([]byte(foldText(center)), []byte(center)); e != nil {
					return e
				}
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
	// Otro runner a lo mismo: de una necesidad ya en camino se crea otro envío ligado a este acopio; con runnerId se
	// asigna en persona (queda en camino), sin él se pide a los runners («Pedidos»).
	mux.HandleFunc("POST /api/supplier/needs/{id}/another", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"runnerId": 128})
		if !ok {
			return
		}
		runnerID := fieldText(f, "runnerId")
		var copy CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			var orig CampNeed
			if e := json.Unmarshal(v, &orig); e != nil {
				return e
			}
			if !a.campExists(tx, orig.CampID) || isRepair(orig) || orig.Status != "IN_PROGRESS" {
				return errNotARun
			}
			copy = CampNeed{ID: randomToken(), CampID: orig.CampID, Need: orig.Need, Urgency: orig.Urgency, Kind: orig.Kind, RequesterName: orig.RequesterName, RequesterPhone: orig.RequesterPhone, CreatedAt: nowStamp(), SameAs: groupOf(orig), SupplierID: me.ID}
			if runnerID == "" {
				copy.Status, copy.SupplierRequestedAt = "OPEN", nowStamp()
				return putNeed(b, &copy)
			}
			runner, found := getUser(tx, runnerID)
			if !found || runner.Disabled || !runner.has(roleRunner) {
				return userFieldError{"runnerId", "Elige un runner activo."}
			}
			busy, already := 0, false
			if e := b.ForEach(func(_, v []byte) error {
				var other CampNeed
				if e := json.Unmarshal(v, &other); e != nil {
					return e
				}
				if other.RunnerID == runnerID && other.Status == "IN_PROGRESS" && !isRepair(other) {
					busy++
					if groupOf(other) == groupOf(orig) {
						already = true
					}
				}
				return nil
			}); e != nil {
				return e
			}
			if already {
				return userFieldError{"runnerId", "Ese runner ya lleva esta necesidad."}
			}
			if busy >= 10 {
				return errRunnerBusy
			}
			copy.Status, copy.RunnerID, copy.EnRouteAt = "IN_PROGRESS", runnerID, nowStamp()
			if e := putNeed(b, &copy); e != nil {
				return e
			}
			runner.RunOrder = append(runner.RunOrder, copy.ID)
			return putJSON(tx.Bucket(usersBucket), runner.ID, runner)
		})
		if errors.Is(err, errNotARun) || errors.Is(err, errRunnerBusy) {
			fail(w, 409, err.Error())
			return
		}
		if userError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": copy.ID, "status": copy.Status})
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
				return userFieldError{"runnerId", "Elige un runner activo."}
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

// SupplyOffer: la respuesta de un acopio al pedido del corredor (puede surtirla o no).
type SupplyOffer struct {
	SupplierID string `json:"supplierId"`
	Can        bool   `json:"can"`
	At         string `json:"at"`
}

// supplyOffer: un acopio que puede surtir la carrera, como lo ve el corredor para elegir el más cercano.
type supplyOffer struct {
	SupplierID string `json:"supplierId"`
	Name       string `json:"name"`
	Phone      string `json:"phone,omitempty"`
	Center     string `json:"center,omitempty"`
	At         string `json:"at"`
}

// offersFor: solo los acopios que dijeron que pueden, mientras el pedido sigue abierto.
func offersFor(tx *bolt.Tx, n CampNeed) []supplyOffer {
	if n.SupplyAskedAt == "" || n.SupplierID != "" {
		return nil
	}
	out := []supplyOffer{}
	for _, o := range n.SupplyOffers {
		if !o.Can {
			continue
		}
		if u, ok := getUser(tx, o.SupplierID); ok && !u.Disabled && u.has(roleSupplier) {
			out = append(out, supplyOffer{SupplierID: u.ID, Name: u.Name, Phone: u.Telephone, Center: u.SupplyCenter, At: o.At})
		}
	}
	return out
}

// supplyAskOpen: el corredor pidió acopio y todavía no eligió uno.
func supplyAskOpen(n CampNeed) bool {
	return !isRepair(n) && n.Status == "IN_PROGRESS" && n.RunnerID != "" && n.SupplierID == "" && n.SupplyAskedAt != ""
}

func myOffer(n CampNeed, me string) *SupplyOffer {
	for i := range n.SupplyOffers {
		if n.SupplyOffers[i].SupplierID == me {
			return &n.SupplyOffers[i]
		}
	}
	return nil
}

var errNoAsk = errors.New("Este runner ya no está pidiendo acopio para este run.")
var errNotOffered = errors.New("Ese acopio no ofreció surtir esta carrera.")

// Corredor: pedir acopio para su carrera en camino, cancelar el pedido o elegir uno de los acopios que pueden.
func (a *App) supplyAskRoutes(mux *http.ServeMux, runners *crew) {
	mine := func(w http.ResponseWriter, r *http.Request, apply func(tx *bolt.Tx, n *CampNeed, f map[string]json.RawMessage) error) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		if s.agent {
			fail(w, 403, errAgentNoRuns.Error())
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"cancel": 0, "supplierId": 128})
		if !ok {
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
			if isRepair(n) || n.RunnerID != s.runnerID || n.Status != "IN_PROGRESS" {
				return errNotYours
			}
			if n.SupplierID != "" {
				return errRequestedElsewhere
			}
			if e := apply(tx, &n, f); e != nil {
				return e
			}
			return putNeed(b, &n)
		})
		switch {
		case errors.Is(err, errNotYours), errors.Is(err, errRequestedElsewhere), errors.Is(err, errNoAsk), errors.Is(err, errNotOffered):
			msg := err.Error()
			if errors.Is(err, errRequestedElsewhere) {
				msg = "Esta carrera ya tiene acopio."
			}
			fail(w, 409, msg)
		case serviceError(w, err):
		default:
			response(w, 200, map[string]any{"id": n.ID, "supplyAskedAt": n.SupplyAskedAt, "supplierId": n.SupplierID})
		}
	}
	mux.HandleFunc("POST /api/runner/needs/{id}/supply-ask", func(w http.ResponseWriter, r *http.Request) {
		mine(w, r, func(tx *bolt.Tx, n *CampNeed, f map[string]json.RawMessage) error {
			var cancel bool
			json.Unmarshal(f["cancel"], &cancel)
			if cancel {
				n.SupplyAskedAt, n.SupplyOffers = "", nil
			} else if n.SupplyAskedAt == "" {
				n.SupplyAskedAt = nowStamp()
			}
			return nil
		})
	})
	mux.HandleFunc("POST /api/runner/needs/{id}/supply-choose", func(w http.ResponseWriter, r *http.Request) {
		mine(w, r, func(tx *bolt.Tx, n *CampNeed, f map[string]json.RawMessage) error {
			if n.SupplyAskedAt == "" {
				return errNoAsk
			}
			pick := fieldText(f, "supplierId")
			if o := myOffer(*n, pick); o == nil || !o.Can {
				return errNotOffered
			}
			if u, ok := getUser(tx, pick); !ok || u.Disabled || !u.has(roleSupplier) {
				return errNotOffered
			}
			n.SupplierID = pick
			return nil
		})
	})
}

// alsoRun: otro envío de la misma necesidad (otro runner, quizá de otro acopio), para verlo en las tarjetas.
type alsoRun struct {
	ID         string `json:"id"`
	RunnerName string `json:"runnerName,omitempty"`
	Center     string `json:"center,omitempty"`
	Status     string `json:"status"`
	Dispatched bool   `json:"dispatched,omitempty"`
}

func groupOf(n CampNeed) string {
	if n.SameAs != "" {
		return n.SameAs
	}
	return n.ID
}

// needGroups: envíos activos (abiertos o en camino) por grupo; solo se arma si hay más de uno en algún grupo.
func needGroups(tx *bolt.Tx) (map[string][]CampNeed, error) {
	groups := map[string][]CampNeed{}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if isRepair(n) || (n.Status != "OPEN" && n.Status != "IN_PROGRESS" && n.Status != "COMPLETED") {
			return nil
		}
		groups[groupOf(n)] = append(groups[groupOf(n)], n)
		return nil
	})
	return groups, err
}

// alsoFor: los demás envíos del grupo de n (el original y sus copias), con runner y acopio.
func alsoFor(tx *bolt.Tx, groups map[string][]CampNeed, n CampNeed) []alsoRun {
	var out []alsoRun
	for _, o := range groups[groupOf(n)] {
		if o.ID == n.ID {
			continue
		}
		x := alsoRun{ID: o.ID, Status: o.Status, Dispatched: o.DispatchedAt != ""}
		if u, ok := getUser(tx, o.RunnerID); ok && o.RunnerID != "" {
			x.RunnerName = u.Name
		}
		if sup := supplierOf(tx, o.SupplierID); sup != nil {
			x.Center = sup.Center
			if x.Center == "" {
				x.Center = sup.Name
			}
		}
		out = append(out, x)
	}
	return out
}
