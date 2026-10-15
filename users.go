package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Usuarios: corredores, constructores y agentes en un solo registro. El panel de administración
// sigue entrando solo con la contraseña maestra; estos usuarios no tienen acceso a él.
var usersBucket = []byte("users")

const (
	roleRunner  = "runner"
	roleBuilder = "builder"
	roleAgent   = "agent"
	// Donaciones: lleva las boletas de donación (donations.html); no ve puntos privados ni carreras.
	roleDonations = "donations"
)

var userRoles = []string{roleRunner, roleBuilder, roleAgent, roleDonations}

type User struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Telephone string   `json:"telephone"`
	Roles     []string `json:"roles"`
	Disabled  bool     `json:"disabled,omitempty"`
	// HMAC del código de acceso normalizado: con él se entra (búsqueda y unicidad).
	CodeHash string `json:"codeHash,omitempty"`
	// El código tal cual, para que administración lo pueda ver (👁) en Usuarios. Nunca va en la lista ni en otras respuestas;
	// solo en GET /api/admin/users/{id}/access-code. Vacío en códigos asignados antes de guardarse así.
	AccessCode string `json:"accessCode,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
	// Aviso para administración, por ejemplo cuando la migración creó al usuario sin datos.
	Note string `json:"note,omitempty"`
	// Orden en que la persona quiere hacer sus carreras en curso (IDs de necesidades).
	RunOrder []string `json:"runOrder,omitempty"`
	// Última actividad con sesión (panel de corredor/constructor o vista completa), guardada cada 2 minutos como máximo.
	LastActiveAt string `json:"lastActiveAt,omitempty"`
	// Equipo (ver teams.go); solo para usuarios con código de acceso.
	TeamID string `json:"teamId,omitempty"`
	// Aceptó el aviso de ubicación en vivo (cuándo y qué versión del texto); sin él no se guarda su ubicación (location.go).
	LocationConsentAt      string `json:"locationConsentAt,omitempty"`
	LocationConsentVersion string `json:"locationConsentVersion,omitempty"`
}

// adminUser es lo que ve el panel: sin el hash del código.
type adminUser struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Telephone     string   `json:"telephone"`
	Roles         []string `json:"roles"`
	Disabled      bool     `json:"disabled"`
	HasAccessCode bool     `json:"hasAccessCode"`
	CreatedAt     string   `json:"createdAt,omitempty"`
	UpdatedAt     string   `json:"updatedAt,omitempty"`
	Note          string   `json:"note,omitempty"`
	LastActiveAt  string   `json:"lastActiveAt,omitempty"`
	Online        bool     `json:"online"`
	TeamID        string   `json:"teamId,omitempty"`
	// Cuándo aceptó el aviso de ubicación vigente; vacío si no lo ha aceptado (o aceptó uno anterior).
	LocationConsentAt string `json:"locationConsentAt,omitempty"`
}

func (u User) admin() adminUser {
	return adminUser{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Roles: u.Roles, Disabled: u.Disabled, HasAccessCode: u.CodeHash != "", CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, Note: u.Note, LastActiveAt: u.LastActiveAt, TeamID: u.TeamID, LocationConsentAt: u.locationConsent()}
}

// En línea: usó su sesión en los últimos 5 minutos (el panel de corredor consulta cada 30 s mientras está abierto).
const (
	onlineWindow = 5 * time.Minute
	persistEvery = 2 * time.Minute
)

// touch anota actividad del usuario: en memoria siempre y en su registro como máximo cada 2 minutos.
func (a *App) touch(id string) {
	if id == "" {
		return
	}
	now := time.Now()
	a.mu.Lock()
	prev := a.lastSeen[id]
	a.lastSeen[id] = now
	a.mu.Unlock()
	if now.Sub(prev) < persistEvery {
		return
	}
	a.db.Update(func(tx *bolt.Tx) error {
		u, ok := getUser(tx, id)
		if !ok {
			return nil
		}
		u.LastActiveAt = now.UTC().Format(time.RFC3339)
		return putJSON(tx.Bucket(usersBucket), id, u)
	})
}

// withPresence completa «en línea» y la última actividad (la de memoria es más reciente que la guardada).
func (a *App) withPresence(u adminUser) adminUser {
	a.mu.Lock()
	seen, ok := a.lastSeen[u.ID]
	a.mu.Unlock()
	if ok {
		u.Online = time.Since(seen) < onlineWindow
		u.LastActiveAt = seen.UTC().Format(time.RFC3339)
	}
	return u
}
func (u User) has(role string) bool { return slices.Contains(u.Roles, role) }
func (u User) runner() Runner {
	return Runner{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Disabled: u.Disabled}
}

func getUser(tx *bolt.Tx, id string) (User, bool) {
	var u User
	v := tx.Bucket(usersBucket).Get([]byte(id))
	return u, v != nil && json.Unmarshal(v, &u) == nil
}
func userHasRole(tx *bolt.Tx, id, role string) bool {
	u, ok := getUser(tx, id)
	return ok && u.has(role)
}
func eachUser(tx *bolt.Tx, fn func(User) error) error {
	return tx.Bucket(usersBucket).ForEach(func(_, v []byte) error {
		var u User
		if e := json.Unmarshal(v, &u); e != nil {
			return e
		}
		return fn(u)
	})
}

// normalizeAccessCode quita espacios y mayúsculas para que «ab 12» y «AB12» sean el mismo código.
func normalizeAccessCode(s string) string {
	return strings.ToUpper(strings.Join(strings.Fields(s), ""))
}
func accessCodeHash(tx *bolt.Tx, code string) string {
	m := hmac.New(sha256.New, tx.Bucket(meta).Get([]byte("access_code_key")))
	m.Write([]byte(normalizeAccessCode(code)))
	return hex.EncodeToString(m.Sum(nil))
}

// Enlaces de la migración: "rol:ID anterior" → ID del usuario. Así cada corredor o constructor anterior
// se procesa una sola vez aunque después se edite su nombre o teléfono.
var legacyLinksBucket = []byte("legacy_user_links")

// legacyUser es un corredor o constructor de los buckets anteriores (runners, builders).
type legacyUser struct {
	Runner
	role string
}

func (l legacyUser) linkKey() string { return l.role + ":" + l.ID }

var roleLabels = map[string]string{roleRunner: "Corredor", roleBuilder: "Constructor", roleAgent: "Agente", roleDonations: "Donaciones"}

// orderRoles deja los roles sin repetir y en el orden fijo de userRoles.
func orderRoles(roles []string) []string {
	out := []string{}
	for _, role := range userRoles {
		if slices.Contains(roles, role) {
			out = append(out, role)
		}
	}
	return out
}

func phoneDigits(s string) string {
	return strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, s)
}

// legacyUsers lee los corredores y luego los constructores anteriores; si un registro no trae ID se usa su clave.
func legacyUsers(tx *bolt.Tx) ([]legacyUser, error) {
	list := []legacyUser{}
	for _, legacy := range []struct {
		bucket []byte
		role   string
	}{{runnersBucket, roleRunner}, {buildersBucket, roleBuilder}} {
		b := tx.Bucket(legacy.bucket)
		if b == nil {
			continue
		}
		if e := b.ForEach(func(k, v []byte) error {
			var r Runner
			if e := json.Unmarshal(v, &r); e != nil {
				return e
			}
			if r.ID == "" {
				r.ID = string(k)
			}
			list = append(list, legacyUser{Runner: r, role: legacy.role})
			return nil
		}); e != nil {
			return nil, e
		}
	}
	return list, nil
}

// findSamePerson busca un usuario con el mismo nombre y el mismo teléfono: la misma persona anotada
// como corredor y como constructor queda en un solo usuario con ambos roles.
func findSamePerson(tx *bolt.Tx, r Runner) (User, bool) {
	name, phone := runnerNameKey(r.Name), phoneDigits(r.Telephone)
	if name == "" || phone == "" {
		return User{}, false
	}
	var match User
	found := false
	err := eachUser(tx, func(u User) error {
		if !found && runnerNameKey(u.Name) == name && phoneDigits(u.Telephone) == phone {
			match, found = u, true
		}
		return nil
	})
	return match, found && err == nil
}

// linkLegacyUser asegura que el corredor o constructor anterior exista como usuario con su rol
// y devuelve el ID al que deben apuntar sus necesidades. Conserva su ID si nadie más lo usa.
func linkLegacyUser(tx *bolt.Tx, l legacyUser) (string, error) {
	links := tx.Bucket(legacyLinksBucket)
	if id := links.Get([]byte(l.linkKey())); id != nil {
		if _, ok := getUser(tx, string(id)); ok {
			return string(id), nil
		}
	}
	u, ok := getUser(tx, l.ID)
	if !ok {
		u, ok = findSamePerson(tx, l.Runner)
	}
	if !ok {
		u = User{ID: l.ID, Name: strings.Join(strings.Fields(l.Name), " "), Telephone: strings.TrimSpace(l.Telephone), CreatedAt: nowStamp()}
		if u.Name == "" {
			u.Name = roleLabels[l.role] + " anterior (sin nombre)"
		}
	}
	if !u.has(l.role) {
		u.Roles = orderRoles(append(u.Roles, l.role))
		u.UpdatedAt = nowStamp()
	}
	if e := putJSON(tx.Bucket(usersBucket), u.ID, u); e != nil {
		return "", e
	}
	return u.ID, links.Put([]byte(l.linkKey()), []byte(u.ID))
}

// placeholderUser registra, desactivado, a alguien a quien una necesidad apunta pero que no está en ningún
// registro, para que el panel muestre a quién estaba asignada en vez de «Asignado» sin nombre.
func placeholderUser(tx *bolt.Tx, id, role string) error {
	if _, ok := getUser(tx, id); ok {
		return nil
	}
	u := User{ID: id, Name: roleLabels[role] + " anterior (sin datos)", Roles: []string{role}, Disabled: true, CreatedAt: nowStamp(),
		Note: "Creado al migrar: una necesidad apuntaba a esta persona, pero no estaba registrada. Completa nombre y teléfono si la conoces."}
	return putJSON(tx.Bucket(usersBucket), id, u)
}

// relinkNeeds cambia en las necesidades (asignado y «cancelada por») los IDs anteriores por el usuario
// actual y registra como usuario a cualquier ID que siga sin existir.
func relinkNeeds(tx *bolt.Tx, links map[string]string) error {
	b := tx.Bucket(needsBucket)
	changed := []CampNeed{}
	missing := map[string]string{}
	if e := b.ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		releasedRole := roleRunner
		if isRepair(n) {
			releasedRole = roleBuilder
		}
		dirty := false
		for _, ref := range []struct {
			id   *string
			role string
		}{{&n.RunnerID, roleRunner}, {&n.BuilderID, roleBuilder}, {&n.ReleasedBy, releasedRole}} {
			if *ref.id == "" {
				continue
			}
			if to, ok := links[ref.role+":"+*ref.id]; ok && to != *ref.id {
				*ref.id, dirty = to, true
			}
			if _, ok := getUser(tx, *ref.id); !ok {
				missing[*ref.id] = ref.role
			}
		}
		if dirty {
			changed = append(changed, n)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, n := range changed {
		if e := putJSON(b, n.ID, n); e != nil {
			return e
		}
	}
	for id, role := range missing {
		if e := placeholderUser(tx, id, role); e != nil {
			return e
		}
	}
	return nil
}

// migrateUsers crea el registro de usuarios y, en cada arranque, pasa a él a los corredores y constructores
// anteriores que falten y corrige las necesidades que apuntaban a ellos. Repetirlo no cambia nada.
func migrateUsers(tx *bolt.Tx) error {
	for _, name := range [][]byte{usersBucket, legacyLinksBucket} {
		if _, e := tx.CreateBucketIfNotExists(name); e != nil {
			return e
		}
	}
	m := tx.Bucket(meta)
	if m.Get([]byte("access_code_key")) == nil {
		if e := m.Put([]byte("access_code_key"), []byte(randomToken())); e != nil {
			return e
		}
	}
	legacy, err := legacyUsers(tx)
	if err != nil {
		return err
	}
	links := map[string]string{}
	for _, l := range legacy {
		id, e := linkLegacyUser(tx, l)
		if e != nil {
			return e
		}
		links[l.linkKey()] = id
	}
	return relinkNeeds(tx, links)
}

// userByAccessCode busca al único usuario con ese código; errMissing si nadie lo tiene.
func userByAccessCode(tx *bolt.Tx, code string) (User, error) {
	hash := accessCodeHash(tx, code)
	var match User
	err := eachUser(tx, func(u User) error {
		if u.CodeHash != "" && hmac.Equal([]byte(u.CodeHash), []byte(hash)) {
			match = u
		}
		return nil
	})
	if err == nil && match.ID == "" {
		err = errMissing
	}
	return match, err
}

// userHasAccessCode: quien entra con código es personal especial; lo que tiene asignado no se detalla en público.
func userHasAccessCode(tx *bolt.Tx, id string) bool {
	u, ok := getUser(tx, id)
	return ok && u.CodeHash != ""
}

// reservedNeed: de una necesidad asignada a un corredor con código el público ve solo el estado (p. ej. en ruta)
// y la urgencia; sin pedido, quién la reportó, asignado ni horas, y sin aviso de que se ocultó algo.
func reservedNeed(tx *bolt.Tx, n CampNeed) bool {
	return n.RunnerID != "" && userHasAccessCode(tx, n.RunnerID)
}

// needItems parte un texto de necesidades en artículos, separados por coma o salto de línea.
func needItems(text string) []string {
	items := []string{}
	for _, part := range strings.FieldsFunc(text, func(c rune) bool { return c == ',' || c == '\n' }) {
		if part = strings.TrimSpace(part); part != "" {
			items = append(items, part)
		}
	}
	return items
}

// reservedNeedItems junta, por punto, los artículos (normalizados) de sus necesidades asignadas a corredores con código
// que no son del equipo de quien mira (t; con me vacío, todas).
func reservedNeedItems(t *teamView) (map[string]map[string]bool, error) {
	tx := t.tx
	items := map[string]map[string]bool{}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.Status == "DELETED" || isRepair(n) || !t.hidden(n) {
			return nil
		}
		if items[n.CampID] == nil {
			items[n.CampID] = map[string]bool{}
		}
		for _, item := range needItems(n.Need) {
			items[n.CampID][foldText(item)] = true
		}
		return nil
	})
	return items, err
}

// withoutReservedItems quita de las necesidades reportadas del punto los artículos de una carrera reservada,
// para que el texto público no repita lo que lleva un corredor con código. Si no quita nada, deja el texto igual.
func withoutReservedItems(text string, reserved map[string]bool) string {
	if len(reserved) == 0 {
		return text
	}
	kept, removed := []string{}, false
	for _, item := range needItems(text) {
		if reserved[foldText(item)] {
			removed = true
		} else {
			kept = append(kept, item)
		}
	}
	if !removed {
		return text
	}
	return strings.Join(kept, ", ")
}

// orderRuns pone primero las carreras en el orden guardado y deja las demás (nuevas) al final, en su orden actual.
func orderRuns(runs []runnerNeed, order []string) []runnerNeed {
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	sort.SliceStable(runs, func(i, j int) bool {
		pi, oki := pos[runs[i].ID]
		pj, okj := pos[runs[j].ID]
		if oki != okj {
			return oki
		}
		return oki && pi < pj
	})
	return runs
}

// startViewerSession abre la vista completa de la página pública (solo lectura) para quien entró con código.
func (a *App) startViewerSession(w http.ResponseWriter, r *http.Request, userID string) {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, expires: time.Now().Add(12 * time.Hour)}
	a.mu.Lock()
	a.viewerSessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "viewer_session", Value: token, Path: "/api/full", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
}

// viewer exige la vista completa: sesión vigente de un usuario activo que todavía tiene código de acceso.
func (a *App) viewer(w http.ResponseWriter, r *http.Request) (User, bool) {
	var u User
	c, e := r.Cookie("viewer_session")
	if e != nil {
		fail(w, 401, "login required")
		return u, false
	}
	a.mu.Lock()
	s, ok := a.viewerSessions[c.Value]
	a.mu.Unlock()
	active := false
	if ok && time.Now().Before(s.expires) {
		a.db.View(func(tx *bolt.Tx) error {
			var found bool
			u, found = getUser(tx, s.runnerID)
			active = found && !u.Disabled && u.CodeHash != ""
			return nil
		})
	}
	if !active {
		fail(w, 401, "login required")
		return User{}, false
	}
	a.touch(u.ID)
	return u, true
}

// Intentos de código: 10 incorrectos en 10 minutos por IP de quien visita bloquean todo intento con código desde esa IP
// (incluso el correcto) hasta que pasen los 10 minutos; si no, se podría seguir adivinando. Los intentos bloqueados no alargan
// el bloqueo. Detrás del ingress la IP real llega en X-Real-IP (nginx la fija; quien visita no puede cambiarla).
const (
	codeMaxWrong = 10
	codeWindow   = 10 * time.Minute
)

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// signInWithCode busca al usuario del código aplicando el límite de intentos; responde el error y devuelve false si no se puede.
func (a *App) signInWithCode(w http.ResponseWriter, r *http.Request, code, wrong string) (User, bool) {
	ip, now := clientIP(r), time.Now()
	a.mu.Lock()
	at := a.codeAttempts[ip]
	if now.After(at.until) {
		at = attempt{}
	}
	blocked := at.count >= codeMaxWrong
	a.mu.Unlock()
	if blocked {
		mins := int(at.until.Sub(now).Minutes()) + 1
		failField(w, 429, "accessCode", fmt.Sprintf("Demasiados códigos incorrectos desde esta conexión. Espera %d min e intenta de nuevo.", mins))
		return User{}, false
	}
	var item User
	err := a.db.View(func(tx *bolt.Tx) error {
		var e error
		item, e = userByAccessCode(tx, code)
		return e
	})
	if errors.Is(err, errMissing) {
		a.mu.Lock()
		at = a.codeAttempts[ip]
		if now.After(at.until) {
			at = attempt{}
		}
		at.count++
		at.until = now.Add(codeWindow)
		a.codeAttempts[ip] = at
		a.mu.Unlock()
		failField(w, 401, "accessCode", wrong)
		return User{}, false
	}
	if serviceError(w, err) {
		return User{}, false
	}
	a.mu.Lock()
	delete(a.codeAttempts, ip)
	a.mu.Unlock()
	if item.Disabled {
		fail(w, 403, errUserDisabled.Error())
		return User{}, false
	}
	return item, true
}

// secureCookie: la cookie solo viaja por HTTPS, también cuando el TLS termina en el ingress (X-Forwarded-Proto).
func secureCookie(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// dropCrewSessions cierra las sesiones de un usuario (al desactivarlo o quitarle roles) y quita su ubicación en vivo.
func (a *App) dropCrewSessions(id string) {
	a.mu.Lock()
	for _, sessions := range []map[string]runnerSession{a.runnerSessions, a.builderSessions, a.viewerSessions, a.agentSessions, a.deskSessions} {
		for token, s := range sessions {
			if s.runnerID == id {
				delete(sessions, token)
			}
		}
	}
	a.mu.Unlock()
	// Sin sesión deja de compartir su ubicación en vivo.
	a.dropLocation(id)
}

type userFieldError struct{ field, msg string }

func (e userFieldError) Error() string { return e.msg }

var errCodeTaken = userFieldError{"accessCode", "Ese código de acceso ya lo usa otra persona. Elige otro."}

// applyUserFields valida y aplica nombre, teléfono, roles, código y estado. El código solo cambia si viene con texto.
func applyUserFields(tx *bolt.Tx, u *User, f map[string]json.RawMessage) error {
	if _, ok := f["name"]; ok {
		u.Name = strings.Join(strings.Fields(fieldText(f, "name")), " ")
	}
	if _, ok := f["telephone"]; ok {
		u.Telephone = strings.Join(strings.Fields(fieldText(f, "telephone")), " ")
	}
	if raw, ok := f["roles"]; ok {
		var roles []string
		if json.Unmarshal(raw, &roles) != nil {
			return userFieldError{"roles", "Roles inválidos."}
		}
		u.Roles = orderRoles(roles)
		for _, role := range roles {
			if !slices.Contains(userRoles, role) {
				return userFieldError{"roles", "Rol desconocido: " + role}
			}
		}
	}
	if raw, ok := f["disabled"]; ok {
		if json.Unmarshal(raw, &u.Disabled) != nil {
			return userFieldError{"disabled", "Estado inválido."}
		}
	}
	if u.Name == "" {
		return userFieldError{"name", "Escribe el nombre y apellido."}
	}
	if u.Telephone == "" {
		return userFieldError{"telephone", "Escribe el teléfono."}
	}
	if len(u.Roles) == 0 {
		return userFieldError{"roles", "Elige al menos un acceso."}
	}
	if code := normalizeAccessCode(fieldText(f, "accessCode")); code != "" {
		if len(code) < 4 {
			return userFieldError{"accessCode", "El código de acceso debe tener al menos 4 caracteres."}
		}
		hash := accessCodeHash(tx, code)
		if e := eachUser(tx, func(other User) error {
			if other.ID != u.ID && other.CodeHash == hash {
				return errCodeTaken
			}
			return nil
		}); e != nil {
			return e
		}
		u.CodeHash, u.AccessCode = hash, code
	}
	if _, ok := f["teamId"]; ok {
		u.TeamID = fieldText(f, "teamId")
		if _, found := getTeam(tx, u.TeamID); u.TeamID != "" && !found {
			return userFieldError{"teamId", "Ese equipo ya no existe. Recarga la página."}
		}
	}
	if u.TeamID != "" && u.CodeHash == "" {
		return userFieldError{"teamId", "Solo los usuarios con código de acceso pueden estar en un equipo. Asígnale un código."}
	}
	return nil
}

func userError(w http.ResponseWriter, err error) bool {
	var fe userFieldError
	if errors.As(err, &fe) {
		status := 400
		if fe == errCodeTaken {
			status = 409
		}
		failField(w, status, fe.field, fe.msg)
		return true
	}
	return serviceError(w, err)
}

func (a *App) userRoutes(mux *http.ServeMux) {
	fields := map[string]int{"name": 120, "telephone": 40, "accessCode": 64, "roles": 0, "disabled": 0, "teamId": 120}
	mux.HandleFunc("GET /api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		list := []adminUser{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return eachUser(tx, func(u User) error { list = append(list, a.withPresence(u.admin())); return nil })
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(list, func(i, j int) bool { return runnerNameKey(list[i].Name) < runnerNameKey(list[j].Name) })
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, fields)
		if !ok {
			return
		}
		u := User{ID: randomToken(), CreatedAt: nowStamp()}
		u.UpdatedAt = u.CreatedAt
		err := a.db.Update(func(tx *bolt.Tx) error {
			if e := applyUserFields(tx, &u, f); e != nil {
				return e
			}
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if userError(w, err) {
			return
		}
		response(w, 201, u.admin())
	})
	// Ver el código de un usuario (🔑 en Usuarios lo copia); la respuesta no se guarda en caché.
	mux.HandleFunc("GET /api/admin/users/{id}/access-code", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		var u User
		var found bool
		a.db.View(func(tx *bolt.Tx) error { u, found = getUser(tx, r.PathValue("id")); return nil })
		switch {
		case !found:
			fail(w, 404, "not found")
		case u.CodeHash == "":
			fail(w, 404, "Este usuario no tiene código de acceso.")
		case u.AccessCode == "":
			fail(w, 404, "Este código se asignó antes de poder verse. Asigna uno nuevo para poder mostrarlo.")
		default:
			response(w, 200, map[string]string{"accessCode": u.AccessCode})
		}
	})
	mux.HandleFunc("PATCH /api/admin/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, fields)
		if !ok {
			return
		}
		var u User
		var before []string
		err := a.db.Update(func(tx *bolt.Tx) error {
			var found bool
			if u, found = getUser(tx, r.PathValue("id")); !found {
				return errMissing
			}
			before = slices.Clone(u.Roles)
			if e := applyUserFields(tx, &u, f); e != nil {
				return e
			}
			u.UpdatedAt = nowStamp()
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if userError(w, err) {
			return
		}
		if u.Disabled || slices.ContainsFunc(before, func(role string) bool { return !u.has(role) }) {
			a.dropCrewSessions(u.ID)
		}
		response(w, 200, u.admin())
	})
}
