package main

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Centro de comando de agentes (/command.html, con el diseño del panel de corredores): un agente es como un operador
// de centro de llamadas que mantiene los datos de los puntos. Entra como corredores y constructores, con su código de
// acceso y una sesión propia (cookie agent_session en /api/agent), nunca con la contraseña maestra. Solo puede usar los
// endpoints de puntos y necesidades de agentAllowed: /api/agent/<x> llama al mismo manejador que /api/admin/<x>, así se
// comportan igual. Usuarios, equipos (salvo verlos), contraseña, respaldos, ajustes, boletines y recorridos GPS
// siguen siendo solo de administración.
var agentAllowed = []string{
	"GET /api/admin/points",
	"POST /api/admin/points",
	"GET /api/admin/submissions",
	"POST /api/admin/decisions",
	"PATCH /api/admin/camps/{id}",
	"POST /api/admin/points/{id}/merge",
	"POST /api/admin/camps/{id}/checkup",
	"GET /api/admin/camps/{id}/needs",
	"GET /api/admin/needs",
	"POST /api/admin/camps/{id}/needs",
	"PATCH /api/admin/needs/{id}",
	"GET /api/admin/teams",
	"GET /api/admin/packing-items",
	"POST /api/admin/needs/{id}/packing",
	"GET /api/admin/media/{need}/{media}",
	"POST /api/admin/needs/{id}/media",
	"PATCH /api/admin/needs/{id}/media/{media}",
	"DELETE /api/admin/needs/{id}/media/{media}",
}

// agentCtx marca una solicitud que ya pasó por agentSession; a.admin y a.adminWrite la aceptan.
// Solo la pone el proxy de /api/agent/, así que solo llega a los manejadores de agentAllowed.
type agentCtx struct{}

func isAgentRequest(r *http.Request) bool {
	_, ok := r.Context().Value(agentCtx{}).(User)
	return ok
}

// agentBy: el agente que hace la solicitud, o nil si es administración.
func agentBy(r *http.Request) *User {
	if u, ok := r.Context().Value(agentCtx{}).(User); ok {
		return &u
	}
	return nil
}

// actor: quién queda anotado en fotos y listas de empaque: el agente con su nombre, o administración.
func actor(r *http.Request) (string, string) {
	if u := agentBy(r); u != nil {
		return u.ID, u.Name
	}
	return "admin", "Administración"
}

func (a *App) startAgentSession(w http.ResponseWriter, r *http.Request, userID string) runnerSession {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour)}
	a.mu.Lock()
	a.agentSessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "agent_session", Value: token, Path: "/api/agent", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
	return s
}

// agentSession exige una sesión vigente de un usuario activo con rol de agente y código de acceso; las escrituras
// llevan el token CSRF de la sesión.
func (a *App) agentSession(w http.ResponseWriter, r *http.Request) (User, runnerSession, bool) {
	var u User
	c, e := r.Cookie("agent_session")
	if e != nil {
		fail(w, 401, "login required")
		return u, runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := a.agentSessions[c.Value]
	a.mu.Unlock()
	active := false
	if ok && time.Now().Before(s.expires) {
		a.db.View(func(tx *bolt.Tx) error {
			var found bool
			u, found = getUser(tx, s.runnerID)
			active = found && !u.Disabled && u.has(roleAgent) && u.CodeHash != ""
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

func (a *App) agentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/agent/session", func(w http.ResponseWriter, r *http.Request) {
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
		if !u.has(roleAgent) {
			failField(w, 403, "accessCode", "Tu código no tiene acceso al panel de agentes.")
			return
		}
		s := a.startAgentSession(w, r, u.ID)
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/agent/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.agentSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "csrf": s.csrf, "roles": u.Roles, "panels": userPanels(u)})
		}
	})
	mux.HandleFunc("POST /api/agent/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("agent_session"); e == nil {
			a.mu.Lock()
			delete(a.agentSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "agent_session", Value: "", Path: "/api/agent", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	// Para elegir runner o constructor en las necesidades: nombres, roles y equipo, sin actividad. El teléfono va solo
	// de runners y constructores, para que el centro de comando los llame o les escriba por WhatsApp; nunca el de otros agentes.
	mux.HandleFunc("GET /api/agent/users", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := a.agentSession(w, r); !ok {
			return
		}
		list := []adminUser{}
		a.db.View(func(tx *bolt.Tx) error {
			return eachUser(tx, func(u User) error {
				phone := ""
				if u.has(roleRunner) || u.has(roleBuilder) {
					phone = u.Telephone
				}
				list = append(list, adminUser{ID: u.ID, Name: u.Name, Telephone: phone, Roles: u.Roles, Disabled: u.Disabled, TeamID: u.TeamID})
				return nil
			})
		})
		sort.SliceStable(list, func(i, j int) bool { return runnerNameKey(list[i].Name) < runnerNameKey(list[j].Name) })
		response(w, 200, list)
	})
	// Runner nuevo desde la ventana de una necesidad (se escribió un nombre que no está en la lista): solo nombre, teléfono y
	// rol de corredor, sin código de acceso. Si el teléfono ya es de un corredor, devuelve ese; si es de otra persona, no lo toca.
	mux.HandleFunc("POST /api/agent/runners", func(w http.ResponseWriter, r *http.Request) {
		agent, _, ok := a.agentSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"name": 120, "telephone": 40})
		if !ok {
			return
		}
		name := strings.Join(strings.Fields(fieldText(f, "name")), " ")
		phone := strings.Join(strings.Fields(fieldText(f, "telephone")), " ")
		if name == "" {
			failField(w, 400, "name", "Escribe el nombre del runner.")
			return
		}
		if len(phoneDigits(phone)) < 7 {
			failField(w, 400, "telephone", "Escribe el teléfono del runner nuevo (por ejemplo 6000-0000).")
			return
		}
		var u User
		status := 201
		err := a.db.Update(func(tx *bolt.Tx) error {
			var taken *User
			eachUser(tx, func(other User) error {
				if taken == nil && phoneDigits(other.Telephone) == phoneDigits(phone) {
					o := other
					taken = &o
				}
				return nil
			})
			if taken != nil {
				if taken.has(roleRunner) && !taken.Disabled {
					u, status = *taken, 200
					return nil
				}
				return userFieldError{"telephone", "Ese teléfono ya es de " + taken.Name + ". Pide a administración que le dé acceso de corredor."}
			}
			u = User{ID: randomToken(), Name: name, Telephone: phone, Roles: []string{roleRunner}, CreatedAt: nowStamp(), Note: "Creado desde el centro de comando por " + agent.Name + "."}
			u.UpdatedAt = u.CreatedAt
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if userError(w, err) {
			return
		}
		response(w, status, adminUser{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Roles: u.Roles})
	})
	allowed := http.NewServeMux()
	for _, pattern := range agentAllowed {
		allowed.Handle(pattern, mux)
	}
	// Por método, para no chocar con «GET /» (las páginas) en el enrutador.
	proxy := func(w http.ResponseWriter, r *http.Request) {
		u, _, ok := a.agentSession(w, r)
		if !ok {
			return
		}
		r2 := r.Clone(context.WithValue(r.Context(), agentCtx{}, u))
		// Se conserva la ruta codificada: IDs como «seed:s/c-1» llegan como seed%3As%2Fc-1 y deben seguir siendo un solo segmento.
		raw := "/api/admin/" + strings.TrimPrefix(r.URL.EscapedPath(), "/api/agent/")
		path, err := url.PathUnescape(raw)
		if err != nil {
			fail(w, 400, "bad path")
			return
		}
		r2.URL.Path, r2.URL.RawPath = path, raw
		allowed.ServeHTTP(w, r2)
	}
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		mux.HandleFunc(method+" /api/agent/", proxy)
	}
}
