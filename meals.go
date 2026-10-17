package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Comidas preparadas (cook.html, rol Cocina): alguien de un equipo pide comida caliente para un punto (cuántas personas
// y si es desayuno, almuerzo o cena). Es una necesidad de tipo «meal» (el pedido o ticket). Una cocina lo toma (una sola
// a la vez; las demás ven quién lo tiene), arma su lista de empaque (los platos) y su foto de partida, y pide
// conductores con una hora de recogida: abierto a todos los runners o a runners que elige, que tienen que aceptarlo.
// Cada runner que acepta tiene su propio run («mealrun», ligado al pedido con SameAs) con fotos y «Completé». Cuando el
// último run en camino se completa, el pedido queda completado y el punto registra el servicio: «🍲 200 almuerzos».

const roleCook = "cook"
const kindMeal, kindMealRun = "meal", "mealrun"

var mealTypes = map[string]string{"desayuno": "desayunos", "almuerzo": "almuerzos", "cena": "cenas"}

func isMealKind(n CampNeed) bool { return n.Kind == kindMeal || n.Kind == kindMealRun }

const mealService = "Comidas preparadas"

var errMealTaken = errors.New("Esta comida ya la tomó otra cocina.")
var errNotYourMeal = errors.New("Esta comida no es tuya o ya no está en preparación.")
var errMealClosed = errors.New("Esta comida ya no busca conductores.")

// ServiceEntry: una vez que se dio un servicio en el punto (comidas, fumigación, ronda médica…), con cuánto y quién.
type ServiceEntry struct {
	At        string `json:"at"`
	Quantity  int    `json:"quantity,omitempty"`
	Detail    string `json:"detail,omitempty"`
	ByID      string `json:"byId,omitempty"`
	ByName    string `json:"byName,omitempty"`
	ValidDays int    `json:"validDays,omitempty"`
	NextDueAt string `json:"nextDueAt,omitempty"`
}

// logService anota un servicio dado en el punto, en su necesidad de servicio de ese nombre (la crea si no existe).
// Con validDays o nextDueAt queda programado para esa fecha; si no, completado (o, si se repite, programado solo).
func logService(tx *bolt.Tx, campID, service string, e ServiceEntry) error {
	b := tx.Bucket(needsBucket)
	var n *CampNeed
	if err := b.ForEach(func(_, v []byte) error {
		var x CampNeed
		if json.Unmarshal(v, &x) == nil && x.CampID == campID && isService(x) && x.Status != "DELETED" && foldText(x.Need) == foldText(service) && n == nil {
			n = &x
		}
		return nil
	}); err != nil {
		return err
	}
	if n == nil {
		n = &CampNeed{ID: randomToken(), CampID: campID, Kind: "service", Need: service, Urgency: urgencyNormal, CreatedAt: nowStamp()}
	}
	if e.At == "" {
		e.At = nowStamp()
	}
	at, _ := time.Parse(time.RFC3339, e.At)
	if e.ValidDays > 0 {
		e.NextDueAt = at.Add(time.Duration(e.ValidDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
	}
	n.ServiceLog = append(n.ServiceLog, e)
	if len(n.ServiceLog) > serviceLogMax {
		n.ServiceLog = n.ServiceLog[len(n.ServiceLog)-serviceLogMax:]
	}
	switch {
	case e.ValidDays > 0 || e.NextDueAt != "":
		next := e.NextDueAt
		n.DoneLog = append(n.DoneLog, e.At)
		n.Status, n.NextDueAt, n.RunnerID, n.CompletedAt = "OPEN", next, "", ""
	case n.RepeatDays > 0:
		n.Status, n.CompletedAt = "COMPLETED", e.At // putNeed lo vuelve a programar y anota DoneLog
	default:
		n.DoneLog = append(n.DoneLog, e.At)
		n.Status, n.CompletedAt, n.NextDueAt = "COMPLETED", e.At, ""
	}
	return putNeed(b, n)
}

// parsePickup: «HH:MM» (hoy, hora de Panamá), «AAAA-MM-DDTHH:MM» o RFC3339.
func parsePickup(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if t, err := time.Parse(time.RFC3339, text); err == nil {
		return t.UTC().Format(time.RFC3339), true
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", text, panamaTime()); err == nil {
		return t.UTC().Format(time.RFC3339), true
	}
	if t, err := time.ParseInLocation("15:04", text, panamaTime()); err == nil {
		now := time.Now().In(panamaTime())
		at := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, panamaTime())
		return at.UTC().Format(time.RFC3339), true
	}
	return "", false
}

// mealRuns: los runs (entregas) de un pedido que no se borraron.
func mealRuns(tx *bolt.Tx, ticket string) []CampNeed {
	out := []CampNeed{}
	tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if json.Unmarshal(v, &n) == nil && n.Kind == kindMealRun && n.SameAs == ticket && n.Status != "DELETED" {
			out = append(out, n)
		}
		return nil
	})
	return out
}

// finishMeal cierra el pedido y registra el servicio en el punto (una sola vez).
func finishMeal(tx *bolt.Tx, t *CampNeed, byID, byName string) error {
	t.Status, t.CompletedAt, t.DriversAskedAt, t.Invites = "COMPLETED", nowStamp(), "", nil
	if !t.ServiceLogged {
		t.ServiceLogged = true
		if e := logService(tx, t.CampID, mealService, ServiceEntry{At: t.CompletedAt, Quantity: t.MealPeople, Detail: t.MealType, ByID: byID, ByName: byName}); e != nil {
			return e
		}
	}
	return putNeed(tx.Bucket(needsBucket), t)
}

// mealRunDone: un runner completó su entrega; si era la última en camino, el pedido se completa.
func mealRunDone(tx *bolt.Tx, run CampNeed) error {
	v := tx.Bucket(needsBucket).Get([]byte(run.SameAs))
	var t CampNeed
	if v == nil || json.Unmarshal(v, &t) != nil || t.Kind != kindMeal || t.Status == "COMPLETED" || t.Status == "DELETED" {
		return nil
	}
	for _, o := range mealRuns(tx, t.ID) {
		if o.ID != run.ID && o.Status == "IN_PROGRESS" {
			return nil
		}
	}
	return finishMeal(tx, &t, run.RunnerID, nameOf(tx, run.RunnerID))
}

type mealCamp struct {
	ID     string    `json:"id"`
	Code   string    `json:"code"`
	Area   string    `json:"area"`
	Place  string    `json:"place,omitempty"`
	People string    `json:"people,omitempty"`
	Coords []float64 `json:"coords,omitempty"`
	GPS    string    `json:"gps,omitempty"`
	Dirs   string    `json:"directions,omitempty"`
}

type mealRunView struct {
	ID         string `json:"id"`
	RunnerName string `json:"runnerName"`
	Status     string `json:"status"`
	EnRouteAt  string `json:"enRouteAt,omitempty"`
	DoneAt     string `json:"completedAt,omitempty"`
	Photos     int    `json:"photos,omitempty"`
}

// Cocinas: lista compartida de nombres (clave sin acentos ni mayúsculas), como los centros de acopio.
var kitchensBucket = []byte("kitchens")

func ensureKitchens(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(kitchensBucket)
	return err
}

func kitchens(tx *bolt.Tx) []string {
	seen := map[string]string{}
	if b := tx.Bucket(kitchensBucket); b != nil {
		b.ForEach(func(k, v []byte) error { seen[string(k)] = string(v); return nil })
	}
	eachUser(tx, func(u User) error {
		if k := strings.TrimSpace(u.Kitchen); k != "" {
			if _, ok := seen[foldText(k)]; !ok {
				seen[foldText(k)] = k
			}
		}
		return nil
	})
	out := []string{}
	for _, v := range seen {
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool { return foldText(out[i]) < foldText(out[j]) })
	return out
}

type mealTicket struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	People      int    `json:"people"`
	Meal        string `json:"meal"`
	Need        string `json:"need"`
	CreatedAt   string `json:"createdAt"`
	RequestedBy string `json:"requestedBy,omitempty"`
	CookID      string `json:"cookId,omitempty"`
	CookName    string `json:"cookName,omitempty"`
	CookPhone   string `json:"cookPhone,omitempty"`
	Kitchen     string `json:"kitchen,omitempty"`
	// Team: lo pidió alguien de mi equipo o es un punto que agregó mi equipo («Solo mi equipo»).
	Team           bool          `json:"team,omitempty"`
	TakenAt        string        `json:"takenAt,omitempty"`
	PickupAt       string        `json:"pickupAt,omitempty"`
	DriversAskedAt string        `json:"driversAskedAt,omitempty"`
	Invited        []string      `json:"invited,omitempty"`
	InvitedMe      bool          `json:"invitedMe,omitempty"`
	MyRun          string        `json:"myRun,omitempty"`
	Runs           []mealRunView `json:"runs"`
	Packing        []PackingLine `json:"packing,omitempty"`
	Media          []MediaItem   `json:"media,omitempty"`
	CompletedAt    string        `json:"completedAt,omitempty"`
	Camp           mealCamp      `json:"camp"`
	// FromNeed: es una necesidad común que pide «Comida preparada» (no un pedido de comida todavía).
	FromNeed bool `json:"fromNeed,omitempty"`
}

// asksPreparedFood: una necesidad común (de suministros) que pide comida preparada («Alimentos (Comida preparada…)»).
func asksPreparedFood(n CampNeed) bool {
	return n.Kind == "" && (n.Status == "OPEN" || n.Status == "IN_PROGRESS") && strings.Contains(foldText(n.Need), "comida preparada")
}

func (a *App) mealView(tx *bolt.Tx, t CampNeed, me string) (mealTicket, bool) {
	var p Point
	v := tx.Bucket(submissions).Get([]byte(t.CampID))
	if v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" {
		return mealTicket{}, false
	}
	m := mealTicket{ID: t.ID, Status: t.Status, People: t.MealPeople, Meal: t.MealType, Need: t.Need, CreatedAt: t.CreatedAt, RequestedBy: t.CreatedByName, CookID: t.CookID, TakenAt: t.CookTakenAt, PickupAt: t.PickupAt, DriversAskedAt: t.DriversAskedAt, Packing: t.Packing, CompletedAt: t.CompletedAt, Runs: []mealRunView{},
		Camp: mealCamp{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, People: p.People, Coords: p.Coords, GPS: p.GPS, Dirs: p.Directions}}
	if u, ok := getUser(tx, t.CookID); ok && t.CookID != "" {
		m.CookName, m.CookPhone, m.Kitchen = u.Name, u.Telephone, u.Kitchen
	}
	if mine := teamOf(tx, me); mine != "" && (t.CreatedBy != "" && teamOf(tx, t.CreatedBy) == mine || p.AddedBy != "" && teamOf(tx, p.AddedBy) == mine) {
		m.Team = true
	}
	for _, id := range t.Invites {
		m.Invited = append(m.Invited, nameOf(tx, id))
		if id == me {
			m.InvitedMe = true
		}
	}
	for _, r := range mealRuns(tx, t.ID) {
		m.Runs = append(m.Runs, mealRunView{ID: r.ID, RunnerName: nameOf(tx, r.RunnerID), Status: r.Status, EnRouteAt: r.EnRouteAt, DoneAt: r.CompletedAt, Photos: len(r.Media)})
		if r.RunnerID == me && r.Status == "IN_PROGRESS" {
			m.MyRun = r.ID
		}
	}
	if t.CookID == me {
		m.Media = t.Media
	}
	return m, true
}

// cookSession: sesión vigente de un usuario activo con el rol Cocina y código.
func (a *App) startCookSession(w http.ResponseWriter, r *http.Request, userID string) runnerSession {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour)}
	a.mu.Lock()
	a.cookSessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "cook_session", Value: token, Path: "/api/cook", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
	return s
}

func (a *App) cookSession(w http.ResponseWriter, r *http.Request) (User, runnerSession, bool) {
	var u User
	c, e := r.Cookie("cook_session")
	if e != nil {
		fail(w, 401, "login required")
		return u, runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := a.cookSessions[c.Value]
	a.mu.Unlock()
	active := false
	if ok && time.Now().Before(s.expires) {
		a.db.View(func(tx *bolt.Tx) error {
			var found bool
			u, found = getUser(tx, s.runnerID)
			active = found && !u.Disabled && u.has(roleCook) && u.CodeHash != ""
			return nil
		})
	}
	if !active {
		fail(w, 401, "login required")
		return User{}, runnerSession{}, false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("X-CSRF-Token") != s.csrf {
		fail(w, 403, errCSRF)
		return User{}, runnerSession{}, false
	}
	a.touch(u.ID)
	return u, s, true
}

func mealError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errMealTaken) || errors.Is(err, errNotYourMeal) || errors.Is(err, errMealClosed) || errors.Is(err, errRunnerBusy) {
		fail(w, 409, err.Error())
		return true
	}
	return userError(w, err)
}

// loadMeal: el pedido de comida id (no borrado, de un punto aprobado).
func (a *App) loadMeal(tx *bolt.Tx, id string) (CampNeed, error) {
	var t CampNeed
	v := tx.Bucket(needsBucket).Get([]byte(id))
	if v == nil || json.Unmarshal(v, &t) != nil || t.Kind != kindMeal || t.Status == "DELETED" || !a.campExists(tx, t.CampID) {
		return t, errMissing
	}
	return t, nil
}

func (a *App) mealRoutes(mux *http.ServeMux, runners *crew) {
	// Pedir comida preparada para un punto: solo quien es de un equipo (con código).
	mux.HandleFunc("POST /api/full/camps/{id}/meals", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"people": 0, "meal": 20, "note": 500})
		if !ok {
			return
		}
		var people int
		if json.Unmarshal(f["people"], &people) != nil || people < 1 || people > 5000 {
			failField(w, 400, "people", "Escribe para cuántas personas (de 1 a 5000).")
			return
		}
		meal := fieldText(f, "meal")
		if mealTypes[meal] == "" {
			failField(w, 400, "meal", "Elige desayuno, almuerzo o cena.")
			return
		}
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			if teamOf(tx, me.ID) == "" {
				return userFieldError{"team", "Solo las personas de un equipo pueden pedir comida preparada."}
			}
			if !a.campExists(tx, r.PathValue("id")) {
				return errMissing
			}
			n = CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Kind: kindMeal, Status: "OPEN", Urgency: urgencyNormal, CreatedAt: nowStamp(), MealPeople: people, MealType: meal,
				CreatedBy: me.ID, CreatedByName: me.Name, UpdatedBy: me.ID}
			n.Need = "Comida preparada · " + itoa(people) + " personas · " + meal
			if note := fieldText(f, "note"); note != "" {
				n.Need += " · " + note
			}
			verifyStamp(tx, &me, &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if userError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": n.ID, "need": n.Need, "status": n.Status})
	})

	// Cocina: entrar con código, ver pedidos, tomar uno, armar el empaque, fotos y pedir conductores.
	mux.HandleFunc("POST /api/cook/session", func(w http.ResponseWriter, r *http.Request) {
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
		if !u.has(roleCook) {
			failField(w, 403, "accessCode", "Tu código no tiene acceso a Cocina. Pide el rol «Cocina» a administración o a tu team manager.")
			return
		}
		s := a.startCookSession(w, r, u.ID)
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "panels": userPanels(u), "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/cook/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.cookSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "roles": u.Roles, "panels": userPanels(u), "csrf": s.csrf})
		}
	})
	mux.HandleFunc("POST /api/cook/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("cook_session"); e == nil {
			a.mu.Lock()
			delete(a.cookSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "cook_session", Value: "", Path: "/api/cook", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	mux.HandleFunc("GET /api/cook/tickets", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.cookSession(w, r)
		if !ok {
			return
		}
		open, mine, taken, done := []mealTicket{}, []mealTicket{}, []mealTicket{}, []mealTicket{}
		since := time.Now().AddDate(0, 0, -7).UTC().Format(time.RFC3339)
		a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var t CampNeed
				if json.Unmarshal(v, &t) != nil || t.Kind != kindMeal {
					return nil
				}
				m, ok := a.mealView(tx, t, me.ID)
				if !ok {
					return nil
				}
				switch {
				case t.Status == "OPEN":
					open = append(open, m)
				case t.Status == "IN_PROGRESS" && t.CookID == me.ID:
					mine = append(mine, m)
				case t.Status == "IN_PROGRESS":
					taken = append(taken, m)
				case t.Status == "COMPLETED" && t.CookID == me.ID && t.CompletedAt >= since:
					done = append(done, m)
				}
				return nil
			})
		})
		newest := func(l []mealTicket) {
			sort.SliceStable(l, func(i, j int) bool { return l[i].CreatedAt > l[j].CreatedAt })
		}
		newest(open)
		newest(mine)
		newest(taken)
		sort.SliceStable(done, func(i, j int) bool { return done[i].CompletedAt > done[j].CompletedAt })
		// Necesidades comunes que piden comida preparada y todavía no tienen un pedido de comida activo ligado.
		needs := []mealTicket{}
		var list []string
		a.db.View(func(tx *bolt.Tx) error {
			list = kitchens(tx)
			linked := map[string]bool{}
			var cands []CampNeed
			tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil {
					return nil
				}
				if n.Kind == kindMeal && n.SameAs != "" && n.Status != "DELETED" && n.Status != "COMPLETED" {
					linked[n.SameAs] = true
				}
				if asksPreparedFood(n) {
					cands = append(cands, n)
				}
				return nil
			})
			for _, n := range cands {
				if linked[n.ID] || newTeamView(tx, me.ID).hidden(n) {
					continue
				}
				m, ok := a.mealView(tx, CampNeed{ID: n.ID, CampID: n.CampID, Need: n.Need, Status: "OPEN", CreatedAt: n.CreatedAt, CreatedBy: n.UpdatedBy, CreatedByName: n.RequesterName}, me.ID)
				if !ok {
					continue
				}
				m.FromNeed = true
				m.People = countOf(m.Camp.People)
				needs = append(needs, m)
			}
			return nil
		})
		sort.SliceStable(needs, func(i, j int) bool { return needs[i].CreatedAt > needs[j].CreatedAt })
		response(w, 200, map[string]any{"open": open, "mine": mine, "taken": taken, "done": done, "needs": needs, "kitchen": me.Kitchen, "kitchens": list})
	})
	// Nombre de la cocina de esta persona (como el centro de acopio): de la lista compartida o uno nuevo.
	mux.HandleFunc("POST /api/cook/kitchen", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.cookSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"kitchen": 120})
		if !ok {
			return
		}
		name := strings.Join(strings.Fields(fieldText(f, "kitchen")), " ")
		err := a.write(r, func(tx *bolt.Tx) error {
			u, found := getUser(tx, me.ID)
			if !found {
				return errMissing
			}
			if name != "" {
				b := tx.Bucket(kitchensBucket)
				if v := b.Get([]byte(foldText(name))); v != nil {
					name = string(v)
				} else {
					for _, k := range kitchens(tx) {
						if foldText(k) == foldText(name) {
							name = k
						}
					}
					if e := b.Put([]byte(foldText(name)), []byte(name)); e != nil {
						return e
					}
				}
			}
			if u.Kitchen != name {
				if e := logAudit(tx, AuditEntry{Kind: "acopio", Ref: u.ID, CampID: centersTrail, Label: u.Name, Action: "cambió", Changes: []ChangeLine{{"Su cocina", u.Kitchen, name}}}); e != nil {
					return e
				}
			}
			u.Kitchen, u.UpdatedAt = name, nowStamp()
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]string{"kitchen": name})
	})
	// Runners activos para asignar en persona (nombre y equipo; sin teléfonos).
	mux.HandleFunc("GET /api/cook/runners", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := a.cookSession(w, r); !ok {
			return
		}
		list := []map[string]string{}
		a.db.View(func(tx *bolt.Tx) error {
			return eachUser(tx, func(u User) error {
				if !u.Disabled && u.has(roleRunner) {
					team := ""
					if t, ok := getTeam(tx, teamOf(tx, u.ID)); ok {
						team = t.Name
					}
					list = append(list, map[string]string{"id": u.ID, "name": u.Name, "team": team})
				}
				return nil
			})
		})
		sort.SliceStable(list, func(i, j int) bool { return runnerNameKey(list[i]["name"]) < runnerNameKey(list[j]["name"]) })
		response(w, 200, list)
	})
	// Cambios de la cocina sobre un pedido: tomar, soltar, pedir conductores, invitar, cerrar el pedido, terminar, empaque.
	cookChange := func(w http.ResponseWriter, r *http.Request, apply func(tx *bolt.Tx, me User, t *CampNeed) error) {
		me, _, ok := a.cookSession(w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		var t CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			var e error
			if t, e = a.loadMeal(tx, r.PathValue("id")); e != nil {
				return e
			}
			if e := apply(tx, me, &t); e != nil {
				return e
			}
			if t.Status == "COMPLETED" {
				return nil // finishMeal ya lo guardó
			}
			t.UpdatedBy = me.ID
			return putNeed(tx.Bucket(needsBucket), &t)
		})
		if mealError(w, err) {
			return
		}
		var out mealTicket
		a.db.View(func(tx *bolt.Tx) error { out, _ = a.mealView(tx, t, me.ID); return nil })
		response(w, 200, out)
	}
	mine := func(me User, t *CampNeed) error {
		if t.CookID != me.ID || t.Status != "IN_PROGRESS" {
			return errNotYourMeal
		}
		return nil
	}
	mux.HandleFunc("POST /api/cook/tickets/{id}/take", func(w http.ResponseWriter, r *http.Request) {
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if t.Status != "OPEN" || t.CookID != "" && t.CookID != me.ID {
				return errMealTaken
			}
			t.CookID, t.CookTakenAt, t.Status = me.ID, nowStamp(), "IN_PROGRESS"
			return nil
		})
	})
	// Tomar una necesidad común de comida preparada: la cocina dice para cuántas personas y qué comida, y queda un pedido
	// de comida ya en preparación por esta cocina, ligado a esa necesidad (SameAs).
	mux.HandleFunc("POST /api/cook/needs/{id}/take", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.cookSession(w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"people": 0, "meal": 20})
		if !ok {
			return
		}
		var people int
		if json.Unmarshal(f["people"], &people) != nil || people < 1 || people > 5000 {
			failField(w, 400, "people", "Escribe para cuántas personas (de 1 a 5000).")
			return
		}
		meal := fieldText(f, "meal")
		if mealTypes[meal] == "" {
			failField(w, 400, "meal", "Elige desayuno, almuerzo o cena.")
			return
		}
		var t CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			var n CampNeed
			v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil || !asksPreparedFood(n) || !a.campExists(tx, n.CampID) {
				return errMissing
			}
			var taken error
			tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var o CampNeed
				if json.Unmarshal(v, &o) == nil && o.Kind == kindMeal && o.SameAs == n.ID && o.Status != "DELETED" && o.Status != "COMPLETED" {
					taken = errMealTaken
				}
				return nil
			})
			if taken != nil {
				return taken
			}
			t = CampNeed{ID: randomToken(), CampID: n.CampID, Kind: kindMeal, Status: "IN_PROGRESS", Urgency: n.Urgency, CreatedAt: nowStamp(), MealPeople: people, MealType: meal,
				SameAs: n.ID, CookID: me.ID, CookTakenAt: nowStamp(), CreatedBy: me.ID, CreatedByName: me.Name, UpdatedBy: me.ID}
			t.Need = "Comida preparada · " + itoa(people) + " personas · " + meal
			return putNeed(tx.Bucket(needsBucket), &t)
		})
		if mealError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": t.ID, "status": t.Status})
	})
	mux.HandleFunc("POST /api/cook/tickets/{id}/release", func(w http.ResponseWriter, r *http.Request) {
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if e := mine(me, t); e != nil {
				return e
			}
			for _, run := range mealRuns(tx, t.ID) {
				if run.Status == "IN_PROGRESS" || run.Status == "COMPLETED" {
					return userFieldError{"ticket", "Ya hay runners con esta comida: no se puede soltar."}
				}
			}
			t.CookID, t.CookTakenAt, t.Status, t.PickupAt, t.DriversAskedAt, t.Invites = "", "", "OPEN", "", "", nil
			return nil
		})
	})
	// Pedir conductores con una hora de recogida: a todos los runners (runnerId vacío) o a uno (que debe aceptar).
	mux.HandleFunc("POST /api/cook/tickets/{id}/drivers", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"pickupAt": 40, "runnerId": 128})
		if !ok {
			return
		}
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if e := mine(me, t); e != nil {
				return e
			}
			at, ok := parsePickup(fieldText(f, "pickupAt"))
			if !ok {
				return userFieldError{"pickupAt", "Indica la hora de recogida."}
			}
			t.PickupAt = at
			if id := fieldText(f, "runnerId"); id != "" {
				u, found := getUser(tx, id)
				if !found || u.Disabled || !u.has(roleRunner) {
					return userFieldError{"runnerId", "Elige un runner activo."}
				}
				for _, run := range mealRuns(tx, t.ID) {
					if run.RunnerID == id && run.Status == "IN_PROGRESS" {
						return userFieldError{"runnerId", "Ese runner ya lleva esta comida."}
					}
				}
				if !slices.Contains(t.Invites, id) {
					t.Invites = append(t.Invites, id)
				}
				return nil
			}
			t.DriversAskedAt = nowStamp()
			return nil
		})
	})
	mux.HandleFunc("POST /api/cook/tickets/{id}/drivers/close", func(w http.ResponseWriter, r *http.Request) {
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if e := mine(me, t); e != nil {
				return e
			}
			t.DriversAskedAt, t.Invites = "", nil
			return nil
		})
	})
	mux.HandleFunc("POST /api/cook/tickets/{id}/done", func(w http.ResponseWriter, r *http.Request) {
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if e := mine(me, t); e != nil {
				return e
			}
			for _, run := range mealRuns(tx, t.ID) {
				if run.Status == "IN_PROGRESS" {
					return userFieldError{"ticket", "Hay runners en camino: se termina sola cuando entreguen."}
				}
			}
			return finishMeal(tx, t, me.ID, me.Name)
		})
	})
	mux.HandleFunc("POST /api/cook/tickets/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"items": 0})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		cookChange(w, r, func(tx *bolt.Tx, me User, t *CampNeed) error {
			if e := mine(me, t); e != nil {
				return e
			}
			now := nowStamp()
			for _, l := range lines {
				t.Packing = append(t.Packing, PackingLine{Item: l.Item, Quantity: l.Quantity, AddedBy: me.ID, AddedByName: me.Name, AddedAt: now})
			}
			// Los runs ya aceptados ven lo mismo que se empacó.
			for _, run := range mealRuns(tx, t.ID) {
				if run.Status == "IN_PROGRESS" {
					run.Packing = t.Packing
					if e := putNeed(tx.Bucket(needsBucket), &run); e != nil {
						return e
					}
				}
			}
			return nil
		})
	})
	mux.HandleFunc("POST /api/cook/tickets/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.cookSession(w, r)
		if !ok || !sameOrigin(w, r) {
			return
		}
		a.receiveMedia(w, r, me.ID, me.Name, func(tx *bolt.Tx, n CampNeed) error {
			if n.Kind != kindMeal || n.CookID != me.ID || n.Status != "IN_PROGRESS" {
				return errMissing
			}
			return nil
		})
	})
	mux.HandleFunc("GET /api/cook/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.cookSession(w, r)
		if !ok {
			return
		}
		a.serveMedia(w, r, func(tx *bolt.Tx, n CampNeed) error {
			if n.Kind != kindMeal || n.CookID != me.ID {
				return errMissing
			}
			return nil
		})
	})

	// Runners: «🍲 Comidas» = pedidos con conductores pedidos (o a los que me invitaron) y mis entregas.
	mux.HandleFunc("GET /api/runner/meals", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		list := []mealTicket{}
		a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var t CampNeed
				if json.Unmarshal(v, &t) != nil || t.Kind != kindMeal || t.Status != "IN_PROGRESS" {
					return nil
				}
				if t.DriversAskedAt == "" && !slices.Contains(t.Invites, s.runnerID) {
					return nil
				}
				if m, ok := a.mealView(tx, t, s.runnerID); ok && m.MyRun == "" {
					m.CookPhone = ""
					if u, found := getUser(tx, t.CookID); found {
						m.CookPhone = u.Telephone // para coordinar la recogida
					}
					list = append(list, m)
				}
				return nil
			})
		})
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].InvitedMe != list[j].InvitedMe {
				return list[i].InvitedMe
			}
			return list[i].PickupAt < list[j].PickupAt
		})
		response(w, 200, list)
	})
	// Aceptar: el runner tiene su propio run de esta comida (en «Yo»), con la hora de recogida y lo que se empacó.
	mux.HandleFunc("POST /api/runner/meals/{id}/accept", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		if s.agent {
			fail(w, 403, errAgentNoRuns.Error())
			return
		}
		var run CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			t, e := a.loadMeal(tx, r.PathValue("id"))
			if e != nil {
				return e
			}
			if t.Status != "IN_PROGRESS" || t.DriversAskedAt == "" && !slices.Contains(t.Invites, s.runnerID) {
				return errMealClosed
			}
			for _, o := range mealRuns(tx, t.ID) {
				if o.RunnerID == s.runnerID && o.Status == "IN_PROGRESS" {
					run = o
					return nil
				}
			}
			busy := 0
			tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var o CampNeed
				if json.Unmarshal(v, &o) == nil && o.RunnerID == s.runnerID && o.Status == "IN_PROGRESS" && !isRepair(o) {
					busy++
				}
				return nil
			})
			if busy >= runners.max {
				return errRunnerBusy
			}
			run = CampNeed{ID: randomToken(), CampID: t.CampID, Kind: kindMealRun, Need: t.Need, Urgency: t.Urgency, Status: "IN_PROGRESS", RunnerID: s.runnerID, EnRouteAt: nowStamp(), CreatedAt: nowStamp(),
				SameAs: t.ID, MealPeople: t.MealPeople, MealType: t.MealType, PickupAt: t.PickupAt, CookID: t.CookID, Packing: t.Packing, UpdatedBy: s.runnerID}
			if e := putNeed(tx.Bucket(needsBucket), &run); e != nil {
				return e
			}
			t.Invites = slices.DeleteFunc(t.Invites, func(id string) bool { return id == s.runnerID })
			t.UpdatedBy = t.CookID
			if e := putNeed(tx.Bucket(needsBucket), &t); e != nil {
				return e
			}
			if u, found := getUser(tx, s.runnerID); found {
				u.RunOrder = append(u.RunOrder, run.ID)
				return putJSON(tx.Bucket(usersBucket), u.ID, u)
			}
			return nil
		})
		if mealError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": run.ID, "status": run.Status})
	})
	// «No puedo»: quien fue invitado lo rechaza; la cocina lo ve y puede invitar a otra persona.
	mux.HandleFunc("POST /api/runner/meals/{id}/decline", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		err := a.write(r, func(tx *bolt.Tx) error {
			t, e := a.loadMeal(tx, r.PathValue("id"))
			if e != nil {
				return e
			}
			if !slices.Contains(t.Invites, s.runnerID) {
				return nil
			}
			t.Invites = slices.DeleteFunc(t.Invites, func(id string) bool { return id == s.runnerID })
			return putNeed(tx.Bucket(needsBucket), &t)
		})
		if mealError(w, err) {
			return
		}
		response(w, 200, map[string]string{"status": "ok"})
	})

	// Servicios dados en el punto: público (qué, cuándo y cuánto; sin nombres).
	mux.HandleFunc("GET /api/camps/{id}/services", func(w http.ResponseWriter, r *http.Request) {
		type pub struct {
			Service   string `json:"service"`
			At        string `json:"at"`
			Quantity  int    `json:"quantity,omitempty"`
			Detail    string `json:"detail,omitempty"`
			NextDueAt string `json:"nextDueAt,omitempty"`
		}
		out := []pub{}
		err := a.db.View(func(tx *bolt.Tx) error {
			if !a.campExists(tx, r.PathValue("id")) || campHeldFrom(tx, r.PathValue("id"), "") {
				return errMissing
			}
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil || n.CampID != r.PathValue("id") || !isService(n) || n.Status == "DELETED" {
					return nil
				}
				logged := map[string]bool{}
				for _, e := range n.ServiceLog {
					logged[e.At] = true
					out = append(out, pub{Service: n.Need, At: e.At, Quantity: e.Quantity, Detail: e.Detail, NextDueAt: e.NextDueAt})
				}
				for _, at := range n.DoneLog {
					if !logged[at] {
						out = append(out, pub{Service: n.Need, At: at})
					}
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].At > out[j].At })
		if len(out) > 50 {
			out = out[:50]
		}
		response(w, 200, out)
	})
	// Registrar a mano un servicio dado (fumigación, ronda médica, comidas…) desde administración o Command.
	mux.HandleFunc("POST /api/admin/camps/{id}/services", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"service": 120, "quantity": 0, "detail": 300, "validDays": 0, "nextDueAt": 40})
		if !ok {
			return
		}
		service := strings.TrimSpace(fieldText(f, "service"))
		if service == "" {
			failField(w, 400, "service", "Indica qué servicio se dio.")
			return
		}
		e := ServiceEntry{Detail: fieldText(f, "detail")}
		if raw, has := f["quantity"]; has && json.Unmarshal(raw, &e.Quantity) != nil || e.Quantity < 0 || e.Quantity > 100000 {
			failField(w, 400, "quantity", "Cantidad inválida.")
			return
		}
		if raw, has := f["validDays"]; has && json.Unmarshal(raw, &e.ValidDays) != nil || e.ValidDays < 0 || e.ValidDays > serviceMaxDays {
			failField(w, 400, "validDays", "Válido por: de 0 a 365 días.")
			return
		}
		if text := strings.TrimSpace(fieldText(f, "nextDueAt")); text != "" {
			probe := CampNeed{Kind: "service"}
			if field, msg := applyService(&probe, map[string]json.RawMessage{"nextDueAt": f["nextDueAt"]}); field != "" {
				failField(w, 400, field, msg)
				return
			}
			e.NextDueAt = probe.NextDueAt
		}
		e.ByID, e.ByName = actor(r)
		err := a.write(r, func(tx *bolt.Tx) error {
			if !a.campExists(tx, r.PathValue("id")) {
				return errMissing
			}
			return logService(tx, r.PathValue("id"), service, e)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, e)
	})
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// latestOf: la marca más reciente entre las dadas (RFC3339 se ordena como texto).
func latestOf(stamps ...string) string {
	out := ""
	for _, s := range stamps {
		if s > out {
			out = s
		}
	}
	return out
}
