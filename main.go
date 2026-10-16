package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

type Point struct {
	ID           string `json:"id"`
	Source       string `json:"source,omitempty"`
	Service      string `json:"service,omitempty"`
	ServiceNotes string `json:"serviceNotes,omitempty"`
	Children     string `json:"children,omitempty"`
	Babies       string `json:"babies,omitempty"`
	Notes        string `json:"notes,omitempty"`
	Directions   string `json:"directions,omitempty"`
	Extra        string `json:"extra,omitempty"`
	GPS          string `json:"gps,omitempty"`
	Row          string `json:"row,omitempty"`
	Ordinal      string `json:"ordinal,omitempty"`
	Coordinator  string `json:"coordinator,omitempty"`
	// Teléfono del coordinador: privado como los demás contactos.
	CoordinatorPhone string `json:"coordinatorPhone,omitempty"`
	Contact          string `json:"contact,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Adults           string `json:"adults,omitempty"`
	OlderAdults      string `json:"olderAdults,omitempty"`
	Disabled         string `json:"disabled,omitempty"`
	// Familias (grupos familiares) en el punto; conteo como los demás.
	Families string `json:"families,omitempty"`
	// «Nuestro equipo lo abastece»: al agregarlo alguien de un equipo puede retener la publicación unas horas
	// (holdHours). Mientras tanto no lo ven el público ni otros equipos; luego se publica solo. HoldTeam es solo la
	// casilla del formulario (no se guarda); HeldBy es el equipo y HeldUntil hasta cuándo.
	HoldTeam       bool      `json:"holdTeam,omitempty"`
	HeldBy         string    `json:"heldBy,omitempty"`
	HeldUntil      string    `json:"heldUntil,omitempty"`
	SecondContact  string    `json:"secondContact,omitempty"`
	SecondPhone    string    `json:"secondPhone,omitempty"`
	ThirdContact   string    `json:"thirdContact,omitempty"`
	ThirdPhone     string    `json:"thirdPhone,omitempty"`
	Generator      string    `json:"generator,omitempty"`
	Code           string    `json:"code"`
	Area           string    `json:"area"`
	Place          string    `json:"place"`
	Needs          string    `json:"needs"`
	People         string    `json:"people"`
	Coords         []float64 `json:"coords"`
	Created        string    `json:"created"`
	Status         string    `json:"status"`
	Decided        string    `json:"decided,omitempty"`
	Note           string    `json:"note,omitempty"`
	CampStatus     string    `json:"campStatus,omitempty"`
	HighestUrgency string    `json:"highestUrgency,omitempty"`
	NeedStatuses   []string  `json:"needStatuses,omitempty"`
	// Personas del grupo; privadas (cédula y edad nunca salen del panel).
	Members []Member `json:"members,omitempty"`
	// Punto aprobado en el que se fusionó este reporte duplicado; no se puede volver a aprobar.
	MergedInto string `json:"mergedInto,omitempty"`
	// Última revisión del punto por administración (aprobar, editar, fusionar o «Marcar revisado»); es pública.
	LastCheckup string `json:"lastCheckup,omitempty"`
	// Valores anteriores de «Personas» y de la lista de personas, del más antiguo al más reciente; privado y de solo lectura.
	History []FieldChange `json:"history,omitempty"`
	// Cada clic en «Checkup» del panel, del más antiguo al más reciente; privado y de solo lectura.
	Checkups []string `json:"checkups,omitempty"`
	// Usuario con código de acceso que publicó el punto (ID y nombre completo); lo fija el servidor y solo lo ve administración.
	AddedBy     string `json:"addedBy,omitempty"`
	AddedByName string `json:"addedByName,omitempty"`
	// Identificador que pone el navegador al crear el punto (cola sin conexión): reenviar el mismo no crea un duplicado.
	ClientRef string `json:"clientRef,omitempty"`
	// «Verificado por equipo»: la última vez que alguien de un equipo creó o actualizó esto (revisión más a fondo). Equipo
	// y persona solo los ve quien tiene acceso; en público solo queda la hora (insignia sin nombres). Lo pone el servidor.
	VerifiedTeam string `json:"verifiedTeam,omitempty"`
	VerifiedBy   string `json:"verifiedBy,omitempty"`
	VerifiedAt   string `json:"verifiedAt,omitempty"`
	// «Solo Nuestro Equipo»: lo agregó alguien del equipo de quien mira o alguna necesidad la lleva alguien de su equipo.
	// Solo sale en la vista completa (/api/full/points); nunca se guarda.
	Team bool `json:"team,omitempty"`
}

// FieldChange guarda el valor que tenía un campo hasta la edición que lo cambió (Until).
type FieldChange struct {
	Field   string   `json:"field"` // "people" o "members"
	People  string   `json:"people,omitempty"`
	Members []Member `json:"members,omitempty"`
	Until   string   `json:"until"`
}

// Máximo de cambios guardados por punto; se descartan los más antiguos.
const historyLimit = 100

// Máximo de checkups guardados por punto; se descartan los más antiguos.
const checkupLimit = 200

// recordHistory agrega a p.History los valores de before que cambiaron en p (Personas y lista de personas).
// Los contactos sueltos de puntos anteriores cuentan como la lista previa, así que pasarlos a personas no es un cambio.
func recordHistory(before Point, p *Point, until string) {
	oldMembers := before.Members
	if len(oldMembers) == 0 {
		for _, c := range [][2]string{{before.Contact, before.Phone}, {before.SecondContact, before.SecondPhone}, {before.ThirdContact, before.ThirdPhone}} {
			if c[0] != "" || c[1] != "" {
				oldMembers = append(oldMembers, Member{Name: c[0], Phone: c[1]})
			}
		}
	}
	if strings.TrimSpace(before.People) != strings.TrimSpace(p.People) {
		p.History = append(p.History, FieldChange{Field: "people", People: before.People, Until: until})
	}
	if p.Members != nil && !slices.Equal(oldMembers, p.Members) {
		p.History = append(p.History, FieldChange{Field: "members", Members: oldMembers, Until: until})
	}
	if len(p.History) > historyLimit {
		p.History = p.History[len(p.History)-historyLimit:]
	}
}

type Member struct {
	Name   string `json:"name,omitempty"`
	Cedula string `json:"cedula,omitempty"`
	Age    string `json:"age,omitempty"`
	Phone  string `json:"phone,omitempty"`
}

// normalizeMembers limpia la lista de personas y, si se envió, copia las tres primeras a los
// campos de contacto que usan el panel, los corredores y la validación de teléfono.
func normalizeMembers(p *Point) error {
	if p.Members == nil {
		return nil
	}
	list := []Member{}
	for _, m := range p.Members {
		m = Member{Name: strings.Join(strings.Fields(m.Name), " "), Cedula: strings.TrimSpace(m.Cedula), Age: strings.TrimSpace(m.Age), Phone: strings.TrimSpace(m.Phone)}
		if m == (Member{}) {
			continue
		}
		if len(m.Name) > 120 || len(m.Cedula) > 30 || len(m.Age) > 10 || len(m.Phone) > 40 {
			return errors.New("Datos de persona demasiado largos.")
		}
		list = append(list, m)
	}
	if len(list) > 50 {
		return errors.New("Máximo 50 personas por punto.")
	}
	p.Members = list
	pairs := [][2]*string{{&p.Contact, &p.Phone}, {&p.SecondContact, &p.SecondPhone}, {&p.ThirdContact, &p.ThirdPhone}}
	for i, pair := range pairs {
		*pair[0], *pair[1] = "", ""
		if i < len(list) {
			*pair[0], *pair[1] = list[i].Name, list[i].Phone
		}
	}
	return nil
}

var countPattern = regexp.MustCompile(`^\d{0,3}$`)
var phonePattern = regexp.MustCompile(`^[\d+\-\s()]*$`)

// publicFormatError valida lo que el formulario público ya restringe: conteos y edades de 0 a 999 y teléfonos solo con números.
func publicFormatError(p Point) (string, string) {
	for field, v := range map[string]string{"people": p.People, "adults": p.Adults, "children": p.Children, "babies": p.Babies, "olderAdults": p.OlderAdults, "disabled": p.Disabled, "families": p.Families} {
		if !countPattern.MatchString(strings.TrimSpace(v)) {
			return field, "Escribe solo números del 0 al 999."
		}
	}
	if !phonePattern.MatchString(p.CoordinatorPhone) {
		return "coordinatorPhone", "El teléfono solo puede tener números."
	}
	for _, m := range p.Members {
		if !countPattern.MatchString(m.Age) {
			return "members", "La edad debe ser un número."
		}
		if !phonePattern.MatchString(m.Phone) {
			return "members", "El teléfono solo puede tener números."
		}
	}
	return "", ""
}
func hasValidPhone(p Point) bool {
	if validPhone(p.CoordinatorPhone) || validPhone(p.Phone) || validPhone(p.SecondPhone) || validPhone(p.ThirdPhone) {
		return true
	}
	for _, m := range p.Members {
		if validPhone(m.Phone) {
			return true
		}
	}
	return false
}

type Runner struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Disabled  bool   `json:"disabled,omitempty"`
}
type CampNeed struct {
	ID       string `json:"id"`
	CampID   string `json:"campId"`
	Need     string `json:"need"`
	Status   string `json:"status"`
	Urgency  string `json:"urgency"`
	RunnerID string `json:"runnerId,omitempty"`
	// Tipo: vacío = suministros (corredores); "repair" = reparación de casa (constructores, nunca pública).
	Kind      string `json:"kind,omitempty"`
	BuilderID string `json:"builderId,omitempty"`
	// Quien reportó la necesidad desde la página pública; visible para todos por decisión del equipo.
	RequesterName  string `json:"requesterName,omitempty"`
	RequesterPhone string `json:"requesterPhone,omitempty"`
	// Marcas de tiempo UTC fijadas por el servidor; el navegador las muestra en hora local.
	CreatedAt   string `json:"createdAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
	// Cuándo pasó a En curso (en ruta) por última vez; se borra si vuelve a Abierta.
	EnRouteAt  string `json:"enRouteAt,omitempty"`
	ReleasedAt string `json:"releasedAt,omitempty"`
	ReleasedBy string `json:"releasedBy,omitempty"`
	// Última escritura de la necesidad (crear, tomar, soltar, editar, completar); ordena la actividad reciente en el panel.
	UpdatedAt string `json:"updatedAt,omitempty"`
	// «Verificado por equipo»: la última vez que alguien de un equipo creó o actualizó esto (revisión más a fondo). Equipo
	// y persona solo los ve quien tiene acceso; en público solo queda la hora (insignia sin nombres). Lo pone el servidor.
	VerifiedTeam string `json:"verifiedTeam,omitempty"`
	VerifiedBy   string `json:"verifiedBy,omitempty"`
	VerifiedAt   string `json:"verifiedAt,omitempty"`
	// Cuándo cambió el pedido (texto) después de crearse, por ejemplo al fusionar un duplicado; se muestra como «Actualizado».
	RevisedAt string `json:"revisedAt,omitempty"`
	// Lista de empaque: lo que corredores y agentes preparan para cumplir la necesidad. Privada (no sale en público).
	Packing []PackingLine `json:"packing,omitempty"`
	// Fotos y videos de partida y entrega; privados (corredor asignado y administración).
	Media []MediaItem `json:"media,omitempty"`
	// Reparaciones (constructores): qué falta para terminar, materiales necesarios (para pedir donaciones) y avance 0–100.
	Scope     string        `json:"scope,omitempty"`
	Materials []PackingLine `json:"materials,omitempty"`
	// Evaluaciones de la obra antes de tomarla (cualquier constructor o explorador; varias por obra).
	Assessments []Assessment `json:"assessments,omitempty"`
	Progress    int          `json:"progress,omitempty"`
	ProgressAt  string       `json:"progressAt,omitempty"`
	// Servicios del punto (Kind "service"): se repiten cada RepeatDays días. Al completarse quedan programados para
	// NextDueAt y no cuentan como necesidad abierta hasta esa fecha. DoneLog: cuándo se hizo cada vez.
	RepeatDays int      `json:"repeatDays,omitempty"`
	NextDueAt  string   `json:"nextDueAt,omitempty"`
	DoneLog    []string `json:"doneLog,omitempty"`
	// Despachada: acopio la marcó como salida con su carga (supplier.go). Es un paso dentro de En camino; se borra si
	// la carrera vuelve a Abierta. Quién la despachó no sale en público.
	DispatchedAt     string `json:"dispatchedAt,omitempty"`
	DispatchedBy     string `json:"dispatchedBy,omitempty"`
	DispatchedByName string `json:"dispatchedByName,omitempty"`
	// Acopio que lleva esta carrera (supplier.go): la pidió a los corredores (SupplierRequestedAt, mientras siga Abierta)
	// o se la asignó en persona a un corredor. Solo esa persona de acopio la ve en «Por despachar».
	SupplierID          string `json:"supplierId,omitempty"`
	SupplierRequestedAt string `json:"supplierRequestedAt,omitempty"`
	// Al revés: el corredor con la carrera pide a los acopios que la surtan (SupplyAskedAt); cada acopio responde si
	// puede o no (SupplyOffers) y el corredor elige uno, que queda como SupplierID. Se borra si deja de estar en camino.
	SupplyAskedAt string        `json:"supplyAskedAt,omitempty"`
	SupplyOffers  []SupplyOffer `json:"supplyOffers,omitempty"`
	// «¡Yo abastezco!»: acopios que ofrecen surtir esta necesidad Abierta (varios a la vez). El runner elige uno al
	// tomarla (queda como SupplierID, con SupplierRequestedAt = cuándo lo ofreció). Se borran al dejar de estar Abierta.
	OpenOffers []SupplyOffer `json:"openOffers,omitempty"`
	// Otro runner a lo mismo: copia de la necesidad (SameAs = la original) que un acopio manda con otro runner; cada
	// una tiene su runner, lista de empaque, fotos, despacho y «Completé» (así queda documentado cada envío).
	SameAs string `json:"sameAs,omitempty"`
	// Quien agregó esta necesidad al cambiar lo que pedía la carrera de un runner (runs.go); privado.
	CreatedBy     string `json:"createdBy,omitempty"`
	CreatedByName string `json:"createdByName,omitempty"`
	// Solo en la respuesta de la ficha para quien entró con código (nombre del conductor asignado, sin su ID); nunca se guarda.
	RunnerName string `json:"runnerName,omitempty"`
}

func nowStamp() string { return time.Now().UTC().Format(time.RFC3339) }

// Buckets anteriores de corredores y constructores; solo se leen una vez para migrarlos a usersBucket.
var runnersBucket = []byte("runners")
var buildersBucket = []byte("builders")
var needsBucket = []byte("camp_needs")
var boletinesBucket = []byte("boletines")

type session struct {
	csrf    string
	expires time.Time
}
type attempt struct {
	count int
	until time.Time
}
type App struct {
	db             *bolt.DB
	mu             sync.Mutex
	sessions       map[string]session
	runnerSessions map[string]runnerSession
	// Constructores: sesión aparte (cookie en /api/builder), igual de liviana que la de corredores.
	builderSessions map[string]runnerSession
	// Vista completa de la página pública para quien entró con código («!código»); cookie en /api/full.
	viewerSessions map[string]runnerSession
	// Centro de comando de agentes (agent.go, command.html): sesión con código, cookie en /api/agent.
	agentSessions    map[string]runnerSession
	deskSessions     map[string]runnerSession
	supplierSessions map[string]runnerSession
	attempts         map[string]attempt
	// Códigos de acceso incorrectos por IP de quien visita (ver signInWithCode).
	codeAttempts map[string]attempt
	// Última vez que cada usuario hizo algo con su sesión (en memoria; se guarda en el usuario cada 2 minutos).
	lastSeen map[string]time.Time
	// Última ubicación en vivo de cada usuario que la comparte (copia en memoria de la base; ver location.go).
	locations map[string]liveLocation
	// Última vez que se borraron los puntos vencidos del recorrido GPS (track.go).
	trackPruned time.Time
	static      string
	router      *router
	media       mediaStore
	// mediaOn: se pueden subir fotos y videos (almacenamiento de objetos, o MEDIA_LOCAL_DISK=1 para desarrollo local).
	mediaOn bool
}

var submissions = []byte("submissions")
var meta = []byte("meta")

var latLon = regexp.MustCompile(`(-?\d{1,2}\.\d+),\s*(-?\d{1,3}\.\d+)`)
var mapEmbed = regexp.MustCompile(`!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)`)

func mapHost(host string) bool {
	switch strings.ToLower(host) {
	case "maps.app.goo.gl", "maps.google.com", "www.google.com", "google.com":
		return true
	}
	return false
}
func parseMapLink(link string) ([]float64, error) {
	if link == "" {
		return nil, nil
	}
	if len(link) > 1000 {
		return nil, errors.New("map link too long")
	}
	u, e := url.Parse(link)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !mapHost(u.Hostname()) {
		return nil, errors.New("use an HTTPS Google Maps link")
	}
	if u.Hostname() == "maps.app.goo.gl" {
		client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 || req.URL.Scheme != "https" || !mapHost(req.URL.Hostname()) {
				return http.ErrUseLastResponse
			}
			return nil
		}}
		if resp, err := client.Get(link); err == nil {
			u = resp.Request.URL
			resp.Body.Close()
		}
	}
	text, _ := url.QueryUnescape(u.String())
	match := mapEmbed.FindStringSubmatch(text)
	if match == nil {
		match = latLon.FindStringSubmatch(text)
	}
	if match == nil {
		return nil, nil
	} // Valid map link, but no coordinates: retain link for manual review.
	lat, e1 := strconv.ParseFloat(match[1], 64)
	lon, e2 := strconv.ParseFloat(match[2], 64)
	if e1 != nil || e2 != nil || lat < 6 || lat > 10 || lon < -83 || lon > -78 {
		return nil, errors.New("map location outside Azuero")
	}
	return []float64{lat, lon}, nil
}

// holdHours: cuánto se retiene la publicación de un punto que un equipo va a abastecer.
const holdHours = 3

// heldFrom: el punto está retenido para su equipo y quien mira (team, vacío para el público) no es de ese equipo.
func heldFrom(p Point, team string) bool {
	return p.HeldBy != "" && p.HeldUntil > nowStamp() && team != p.HeldBy
}

// campHeldFrom: como heldFrom, buscando el punto por su ID.
func campHeldFrom(tx *bolt.Tx, id, team string) bool {
	var p Point
	v := tx.Bucket(submissions).Get([]byte(id))
	return v != nil && json.Unmarshal(v, &p) == nil && heldFrom(p, team)
}

func publicPoint(p Point) Point {
	p.ClientRef = ""
	p.HeldBy = ""
	p.Needs, _ = splitRepair(p.Needs)
	p.AddedBy, p.AddedByName = "", ""
	p.VerifiedTeam, p.VerifiedBy = "", ""
	p.Members = nil
	p.History = nil
	p.Checkups = nil
	p.CoordinatorPhone = ""
	p.Note = ""
	p.Contact = ""
	p.Phone = ""
	p.SecondContact = ""
	p.SecondPhone = ""
	p.ThirdContact = ""
	p.ThirdPhone = ""
	containsPhone := func(text string) bool {
		digits := 0
		for _, c := range text {
			if c >= '0' && c <= '9' {
				digits++
			}
		}
		return digits >= 7
	}
	if containsPhone(p.Coordinator) {
		p.Coordinator = ""
	}
	if containsPhone(p.Generator) {
		p.Generator = ""
	}
	return p
}
func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func response(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int, msg string) {
	response(w, status, map[string]string{"error": msg})
}
func failField(w http.ResponseWriter, status int, field, msg string) {
	response(w, status, map[string]string{"field": field, "error": msg})
}
func openApp(path, static string) (*App, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}
	a := &App{db: db, static: static, sessions: map[string]session{}, runnerSessions: map[string]runnerSession{}, builderSessions: map[string]runnerSession{}, viewerSessions: map[string]runnerSession{}, agentSessions: map[string]runnerSession{}, deskSessions: map[string]runnerSession{}, supplierSessions: map[string]runnerSession{}, attempts: map[string]attempt{}, codeAttempts: map[string]attempt{}, lastSeen: map[string]time.Time{}, locations: map[string]liveLocation{}, router: newRouter(), media: newMediaStore(path)}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(submissions); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(needsBucket); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(boletinesBucket); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(donationsBucket); e != nil {
			return e
		}
		if e := ensureSupplyCenters(tx); e != nil {
			return e
		}
		if e := ensureTeamInvites(tx); e != nil {
			return e
		}
		b, e := tx.CreateBucketIfNotExists(meta)
		if e != nil {
			return e
		}
		if b.Get([]byte("password_hash")) != nil {
			return nil
		}
		password := os.Getenv("ADMIN_BOOTSTRAP_PASSWORD")
		if len(password) < 12 {
			return errors.New("new database: set ADMIN_BOOTSTRAP_PASSWORD (at least 12 characters); it is hashed in BoltDB")
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if e != nil {
			return e
		}
		return b.Put([]byte("password_hash"), hash)
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	_, objectStorage := a.media.(*s3Store)
	a.mediaOn = objectStorage || os.Getenv("MEDIA_LOCAL_DISK") == "1"
	if err := a.migrate(); err != nil {
		return nil, err
	}
	return a, nil
}

// migrate pone al día los datos guardados (al arrancar y después de importar un respaldo, que puede ser más viejo).
func (a *App) migrate() error {
	for _, step := range []func(*bolt.Tx) error{migrateUsers, migrateTeams, seedPackingItems, splitLegacyRepairs, migrateOpenRequests, migrateNeedUrgency, ensureDonations, ensureSupplyCenters, ensureTeamInvites} {
		if err := a.db.Update(step); err != nil {
			return err
		}
	}
	return a.loadLocations()
}
func (a *App) passwordValid(p string) bool {
	var hash []byte
	_ = a.db.View(func(tx *bolt.Tx) error {
		hash = append([]byte(nil), tx.Bucket(meta).Get([]byte("password_hash"))...)
		return nil
	})
	return bcrypt.CompareHashAndPassword(hash, []byte(p)) == nil
}
func (a *App) guard(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			fail(w, 415, "JSON required")
			return false
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, e := url.Parse(origin)
			if e != nil || u.Host != r.Host {
				fail(w, 403, "origin mismatch")
				return false
			}
		}
	}
	return true
}
func (a *App) admin(w http.ResponseWriter, r *http.Request) bool {
	// Un agente llega aquí solo por el proxy de /api/agent/ (agent.go), que ya revisó su sesión y CSRF.
	if isAgentRequest(r) {
		return true
	}
	c, e := r.Cookie("relief_session")
	if e != nil {
		fail(w, 401, "login required")
		return false
	}
	a.mu.Lock()
	s, ok := a.sessions[c.Value]
	a.mu.Unlock()
	if !ok || time.Now().After(s.expires) {
		fail(w, 401, "login required")
		return false
	}
	if r.Method == http.MethodPost && r.Header.Get("X-CSRF-Token") != s.csrf {
		fail(w, 403, "CSRF token required")
		return false
	}
	return true
}
func decodeFields(w http.ResponseWriter, r *http.Request, allowed map[string]int) (map[string]json.RawMessage, bool) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		fail(w, 415, "JSON required")
		return nil, false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, e := url.Parse(origin)
		if e != nil || u.Host != r.Host {
			fail(w, 403, "origin mismatch")
			return nil, false
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	d := json.NewDecoder(r.Body)
	var fields map[string]json.RawMessage
	if err := d.Decode(&fields); err != nil || fields == nil {
		fail(w, 400, "invalid JSON object")
		return nil, false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		fail(w, 400, "one JSON object required")
		return nil, false
	}
	if key, msg := checkFields(fields, allowed); key != "" {
		failField(w, 400, key, msg)
		return nil, false
	}
	return fields, true
}

// checkFields devuelve el primer campo desconocido o demasiado largo y su error.
func checkFields(fields map[string]json.RawMessage, allowed map[string]int) (string, string) {
	for key, v := range fields {
		limit, ok := allowed[key]
		if !ok {
			return key, "unknown field"
		}
		if limit > 0 {
			var text string
			if json.Unmarshal(v, &text) != nil || len(strings.TrimSpace(text)) > limit {
				return key, "invalid or too long"
			}
		}
	}
	return "", ""
}

// validPhone acepta cualquier formato con al menos 7 dígitos (números de Panamá).
func validPhone(s string) bool {
	digits := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	return digits >= 7
}
func fieldText(fields map[string]json.RawMessage, key string) string {
	var s string
	json.Unmarshal(fields[key], &s)
	return strings.TrimSpace(s)
}
func validStatus(s string) bool {
	return s == "OPEN" || s == "IN_PROGRESS" || s == "COMPLETED" || s == "DELETED"
}

// Prioridad de una necesidad: solo «normal» o «urgente».
const (
	urgencyNormal  = "normal"
	urgencyUrgente = "urgente"
)

// legacyUrgency traduce las prioridades antiguas (Today/Urgent/Tomorrow/Eventually y sus etiquetas en español)
// para que datos ya guardados y páginas viejas abiertas en teléfonos sigan funcionando.
var legacyUrgency = map[string]string{
	"normal": urgencyNormal, "urgente": urgencyUrgente,
	"today": urgencyUrgente, "urgent": urgencyUrgente, "hoy": urgencyUrgente,
	"tomorrow": urgencyNormal, "eventually": urgencyNormal, "manana": urgencyNormal, "eventualmente": urgencyNormal,
}

// normalizeUrgency devuelve la prioridad canónica; «normal» si viene vacía y "" si no se reconoce.
func normalizeUrgency(s string) string {
	s = foldText(s)
	if s == "" {
		return urgencyNormal
	}
	return legacyUrgency[s]
}
func validUrgency(s string) bool { return s == urgencyNormal || s == urgencyUrgente }
func urgencyRank(s string) int {
	switch legacyUrgency[foldText(s)] {
	case urgencyUrgente:
		return 2
	case urgencyNormal:
		return 1
	}
	return 0
}

// migrateNeedUrgency reescribe las prioridades antiguas guardadas como «normal» o «urgente»
// (Today/Urgent → urgente; Tomorrow/Eventually/vacía → normal). Es idempotente, corre al abrir la base y no toca
// UpdatedAt para no reordenar la actividad reciente. Un valor no reconocido queda «normal» para que se pueda editar.
func migrateNeedUrgency(tx *bolt.Tx) error {
	b := tx.Bucket(needsBucket)
	var changed []CampNeed
	if e := b.ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		u := normalizeUrgency(n.Urgency)
		if u == "" {
			u = urgencyNormal
		}
		if u != n.Urgency {
			n.Urgency = u
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
	if len(changed) > 0 {
		log.Printf("migrated urgency of %d needs to normal/urgente", len(changed))
	}
	return nil
}
func putNeed(b *bolt.Bucket, n *CampNeed) error {
	n.UpdatedAt, n.RunnerName = nowStamp(), ""
	if n.Status != "IN_PROGRESS" && n.Status != "COMPLETED" {
		n.DispatchedAt, n.DispatchedBy, n.DispatchedByName = "", "", ""
	}
	// Vuelve a Abierta: si venía de una oferta de acopio (o un pedido de antes), ese acopio sigue ofreciendo; si se la
	// habían asignado en persona, queda libre. Abierta nunca tiene acopio fijo: solo ofertas.
	if n.Status == "OPEN" && n.SupplierRequestedAt != "" && n.SupplierID != "" && offerIndex(n.OpenOffers, n.SupplierID) < 0 {
		n.OpenOffers = append(n.OpenOffers, SupplyOffer{SupplierID: n.SupplierID, Can: true, At: n.SupplierRequestedAt})
	}
	if n.Status == "OPEN" || n.Status == "DELETED" {
		n.SupplierID, n.SupplierRequestedAt = "", ""
	}
	if n.Status != "OPEN" {
		n.OpenOffers = nil
	}
	if n.Status != "IN_PROGRESS" || n.SupplierID != "" {
		n.SupplyAskedAt, n.SupplyOffers = "", nil
	}
	rescheduleService(n)
	return putJSON(b, n.ID, n)
}

func putJSON(b *bolt.Bucket, key string, value any) error {
	data, e := json.Marshal(value)
	if e != nil {
		return e
	}
	return b.Put([]byte(key), data)
}

// needFromPoint convierte el texto «Necesidades» de un punto recién aprobado en una necesidad abierta de suministros
// y, si pidió «Reparación de casa», en otra de reparación aparte. Los contactos del punto son privados, por eso no se copian como solicitante.
func needFromPoint(tx *bolt.Tx, p Point) error {
	supplies, repair := splitRepair(p.Needs)
	if supplies != "" {
		n := CampNeed{ID: randomToken(), CampID: p.ID, Need: supplies, Status: "OPEN", Urgency: urgencyNormal, CreatedAt: nowStamp()}
		if e := putNeed(tx.Bucket(needsBucket), &n); e != nil {
			return e
		}
	}
	if repair {
		n := CampNeed{ID: randomToken(), CampID: p.ID, Need: repairText(p.Needs), Kind: "repair", Status: "OPEN", Urgency: urgencyNormal, CreatedAt: nowStamp()}
		return putNeed(tx.Bucket(needsBucket), &n)
	}
	return nil
}

// Reparación de casa: la atienden constructores y no sale en páginas ni datos públicos.
const repairLabel = "Reparación de casa"

var accentFold = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n", "Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ñ", "n")

func foldText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(accentFold.Replace(s))), " ")
}

// splitRepair separa «Reparación de casa» del texto de necesidades (lista separada por comas).
func splitRepair(text string) (string, bool) {
	rest, repair := []string{}, false
	for _, part := range strings.Split(text, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if isRepairPart(part) {
			repair = true
		} else {
			rest = append(rest, part)
		}
	}
	if !repair {
		return strings.TrimSpace(text), false
	}
	return strings.Join(rest, ", "), true
}
func isRepair(n CampNeed) bool { return n.Kind == "repair" }

// isRepairPart: «Reparación de casa» sola o con materiales: «Reparación de casa (Varilla / Hojas de zinc)».
func isRepairPart(part string) bool {
	f := foldText(part)
	return f == foldText(repairLabel) || strings.HasPrefix(f, foldText(repairLabel)+" (")
}

// repairText: la parte de reparación del texto (con sus materiales), o «Reparación de casa» si no la hay.
func repairText(text string) string {
	for _, part := range strings.Split(text, ",") {
		if part = strings.TrimSpace(part); isRepairPart(part) {
			return part
		}
	}
	return repairLabel
}

// putNeedSplit guarda una necesidad nueva de suministros; si el texto pide «Reparación de casa», esa parte va en una
// necesidad de reparación aparte (constructores, nunca pública). Devuelve la de suministros, o la de reparación si no
// quedaba otra cosa.
func putNeedSplit(tx *bolt.Tx, n CampNeed) (CampNeed, error) {
	rest, repair := splitRepair(n.Need)
	if !repair || n.Kind != "" {
		return n, putNeed(tx.Bucket(needsBucket), &n)
	}
	fix := n
	fix.ID, fix.Kind, fix.Need, fix.RunnerID = randomToken(), "repair", repairText(n.Need), ""
	if e := putNeed(tx.Bucket(needsBucket), &fix); e != nil {
		return n, e
	}
	if rest == "" {
		return fix, nil
	}
	n.Need = rest
	return n, putNeed(tx.Bucket(needsBucket), &n)
}

// splitLegacyRepairs separa «Reparación de casa» de necesidades de suministros creadas antes de que las reparaciones
// fueran aparte: se quita del texto y pasa a una necesidad de reparación abierta (o la necesidad entera se vuelve
// reparación si no quedaba otra cosa). Es idempotente y corre al abrir la base.
func splitLegacyRepairs(tx *bolt.Tx) error {
	b := tx.Bucket(needsBucket)
	var legacy []CampNeed
	activeRepair := map[string]bool{}
	if e := b.ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if isRepair(n) {
			activeRepair[n.CampID] = activeRepair[n.CampID] || n.Status == "OPEN" || n.Status == "IN_PROGRESS"
		} else if _, repair := splitRepair(n.Need); repair && n.Status != "DELETED" {
			legacy = append(legacy, n)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, n := range legacy {
		rest, _ := splitRepair(n.Need)
		if rest == "" {
			// Solo pedía la reparación: la necesidad entera pasa a constructores, sin el corredor que tuviera.
			n.Kind, n.Need, n.RunnerID = "repair", repairText(n.Need), ""
			if n.Status == "IN_PROGRESS" {
				n.Status, n.EnRouteAt = "OPEN", ""
			}
			activeRepair[n.CampID] = true
			if e := putNeed(b, &n); e != nil {
				return e
			}
			continue
		}
		n.Need, n.RevisedAt = rest, nowStamp()
		if e := putNeed(b, &n); e != nil {
			return e
		}
		if n.Status == "COMPLETED" || activeRepair[n.CampID] {
			continue
		}
		r := CampNeed{ID: randomToken(), CampID: n.CampID, Kind: "repair", Need: repairLabel, Status: "OPEN", Urgency: n.Urgency, CreatedAt: n.CreatedAt}
		if r.CreatedAt == "" {
			r.CreatedAt = nowStamp()
		}
		activeRepair[n.CampID] = true
		if e := putNeed(b, &r); e != nil {
			return e
		}
	}
	return nil
}
func validKind(s string) bool { return s == "" || s == "supply" || s == "repair" || s == "service" }
func storedKind(s string) string {
	if s == "repair" || s == "service" {
		return s
	}
	return ""
}

// validAssignee: suministros solo con corredores registrados y reparaciones solo con constructores registrados.
func validAssignee(tx *bolt.Tx, n CampNeed) bool {
	if isRepair(n) {
		return n.RunnerID == "" && (n.BuilderID == "" || userHasRole(tx, n.BuilderID, roleBuilder))
	}
	return n.BuilderID == "" && (n.RunnerID == "" || userHasRole(tx, n.RunnerID, roleRunner))
}

// enRouteOnAssign ajusta el estado cuando administración cambia el corredor: asignar uno a una necesidad abierta
// la pone en ruta desde ahora; quitarlo («Sin asignar») la devuelve a Abierta; cambiarlo por otro reinicia la hora en ruta.
func enRouteOnAssign(n *CampNeed, runnerBefore string) {
	switch {
	case n.RunnerID == runnerBefore:
	case n.RunnerID == "" && n.Status == "IN_PROGRESS":
		n.Status, n.EnRouteAt = "OPEN", ""
	case n.RunnerID != "" && n.Status == "OPEN":
		n.Status, n.EnRouteAt = "IN_PROGRESS", nowStamp()
	case n.RunnerID != "" && n.Status == "IN_PROGRESS":
		n.EnRouteAt = nowStamp()
	}
}
func (a *App) campExists(tx *bolt.Tx, id string) bool {
	v := tx.Bucket(submissions).Get([]byte(id))
	if v == nil {
		return false
	}
	var p Point
	return json.Unmarshal(v, &p) == nil && p.Status == "approved"
}
func (a *App) adminWrite(w http.ResponseWriter, r *http.Request) bool {
	if !a.admin(w, r) {
		return false
	}
	if isAgentRequest(r) {
		return true
	}
	if r.Method == http.MethodPatch || r.Method == http.MethodDelete {
		c, _ := r.Cookie("relief_session")
		a.mu.Lock()
		s := a.sessions[c.Value]
		a.mu.Unlock()
		if r.Header.Get("X-CSRF-Token") != s.csrf {
			fail(w, 403, "CSRF token required")
			return false
		}
	}
	return true
}

// campHasNeedRecords: si el punto tiene alguna necesidad guardada, en cualquier estado (también eliminadas).
func campHasNeedRecords(tx *bolt.Tx, id string) (bool, error) {
	found := errors.New("found")
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID == id {
			return found
		}
		return nil
	})
	if err == found {
		return true, nil
	}
	return false, err
}

// campNeedStatuses lists which of OPEN, IN_PROGRESS and COMPLETED a camp's saved needs include.
func campNeedStatuses(tx *bolt.Tx, id string) ([]string, error) {
	seen := map[string]bool{}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID == id && !isRepair(n) && !serviceScheduled(n, time.Now()) {
			seen[n.Status] = true
		}
		return nil
	})
	statuses := []string{}
	for _, s := range []string{"OPEN", "IN_PROGRESS", "COMPLETED"} {
		if seen[s] {
			statuses = append(statuses, s)
		}
	}
	return statuses, err
}
func aggregateCamp(tx *bolt.Tx, id string) (string, string, error) {
	status := "NO_NEEDS"
	urgency := ""
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID != id || n.Status == "DELETED" || isRepair(n) || serviceScheduled(n, time.Now()) {
			return nil
		}
		switch n.Status {
		case "OPEN":
			status = "OPEN"
		case "IN_PROGRESS":
			if status != "OPEN" {
				status = "IN_PROGRESS"
			}
		case "COMPLETED":
			if status == "NO_NEEDS" {
				status = "COMPLETED"
			}
		}
		if n.Status == "OPEN" || n.Status == "IN_PROGRESS" {
			if urgencyRank(n.Urgency) > urgencyRank(urgency) {
				urgency = normalizeUrgency(n.Urgency)
			}
		}
		return nil
	})
	return status, urgency, err
}

// publicPoints: puntos aprobados para el público; viewer (quien entró con código) conserva en «necesidades»
// los artículos de las carreras reservadas de su equipo y ve qué puntos son de su equipo.
func (a *App) publicPoints(w http.ResponseWriter, viewer string) {
	points := []Point{}
	err := a.db.View(func(tx *bolt.Tx) error {
		tv := newTeamView(tx, viewer)
		reserved, e := reservedNeedItems(tv)
		if e != nil {
			return e
		}
		team, e := teamCamps(tv)
		if e != nil {
			return e
		}
		return tx.Bucket(submissions).ForEach(func(k, v []byte) error {
			var p Point
			if e := json.Unmarshal(v, &p); e != nil {
				return e
			}
			if p.Status == "approved" && !heldFrom(p, tv.team) {
				status, urgency, e := aggregateCamp(tx, p.ID)
				if e != nil {
					return e
				}
				p.CampStatus = status
				p.HighestUrgency = urgency
				if p.NeedStatuses, e = campNeedStatuses(tx, p.ID); e != nil {
					return e
				}
				p.Needs = withoutReservedItems(p.Needs, reserved[p.ID])
				p.Team = team[p.ID]
				points = append(points, publicPoint(p))
			}
			return nil
		})
	})
	if err != nil {
		fail(w, 500, "database error")
		return
	}
	response(w, 200, points)
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { response(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /api/points", func(w http.ResponseWriter, r *http.Request) { a.publicPoints(w, "") })
	// reviewed: administración (campos internos, código propio). publish: queda aprobado al guardar
	// (administración, o un usuario que entró con su código de acceso).
	// Una solicitud pública también se publica si administración apagó la revisión (ver settings.go).
	// by: usuario con código que lo agrega (queda como «Agregado por»); nil en solicitudes públicas y de administración.
	createPoint := func(w http.ResponseWriter, r *http.Request, reviewed, publish bool, by *User) {
		if !a.guard(w, r) || (reviewed && !a.admin(w, r)) {
			return
		}
		var p Point
		r.Body = http.MaxBytesReader(w, r.Body, 32768)
		d := json.NewDecoder(r.Body)
		if err := d.Decode(&p); err != nil {
			var max *http.MaxBytesError
			var field *json.UnmarshalTypeError
			switch {
			case errors.As(err, &max):
				fail(w, 413, "El reporte supera 32 KB. Divide el texto en notas más cortas.")
			case errors.As(err, &field):
				failField(w, 400, field.Field, "Formato incorrecto para «"+field.Field+"». Envíalo como texto.")
			case errors.Is(err, io.EOF):
				fail(w, 400, "El reporte está vacío. Agrega al menos un dato.")
			default:
				fail(w, 400, "El reporte no es JSON válido. Revisa el formato y vuelve a intentar.")
			}
			return
		}
		var rest any
		if err := d.Decode(&rest); err != io.EOF {
			fail(w, 400, "Envía un solo objeto JSON sin contenido adicional.")
			return
		}
		p.Code = strings.TrimSpace(p.Code)
		p.Area = strings.TrimSpace(p.Area)
		p.Place = strings.TrimSpace(p.Place)
		p.Needs = strings.TrimSpace(p.Needs)
		p.GPS = strings.TrimSpace(p.GPS)
		if p.GPS != "" {
			// Keep unrecognized map links for review; never fetch untrusted hosts.
			if coords, err := parseMapLink(p.GPS); err == nil {
				p.Coords = coords
			} else {
				p.Coords = nil
			}
		}
		if len(p.Coords) != 0 && (len(p.Coords) != 2 || p.Coords[0] < 6 || p.Coords[0] > 10 || p.Coords[1] < -83 || p.Coords[1] > -78) {
			failField(w, 400, "coords", "Coordenadas fuera de Azuero o incompletas. Omite coords si no tienes ubicación.")
			return
		}
		if err := normalizeMembers(&p); err != nil {
			failField(w, 400, "members", err.Error())
			return
		}
		p.History, p.Checkups, p.Team = nil, nil, false
		p.ClientRef = strings.TrimSpace(p.ClientRef)
		if len(p.ClientRef) > 64 {
			failField(w, 400, "clientRef", "Identificador del envío demasiado largo.")
			return
		}
		// Importación de administración: «addedBy» con el ID de un usuario existente deja anotado quién reportó el punto.
		askedBy := strings.TrimSpace(p.AddedBy)
		p.AddedBy, p.AddedByName = "", ""
		p.VerifiedTeam, p.VerifiedBy, p.VerifiedAt = "", "", ""
		hold := p.HoldTeam
		p.HoldTeam, p.HeldBy, p.HeldUntil = false, "", ""
		if by != nil {
			p.AddedBy, p.AddedByName = by.ID, by.Name
		}
		if !reviewed {
			if field, msg := publicFormatError(p); field != "" {
				failField(w, 400, field, msg)
				return
			}
			// Campos internos: solo administración los llena.
			p.Code, p.Row, p.Ordinal, p.Service, p.ServiceNotes, p.Notes, p.Note = "", "", "", "", "", "", ""
		}
		if !reviewed && !hasValidPhone(p) {
			failField(w, 400, "phone", "Agrega al menos un teléfono de contacto (7 dígitos o más).")
			return
		}
		p.ID = randomToken()
		if p.Code == "" {
			p.Code = "P-" + p.ID[:12]
		}
		if p.Area == "" {
			p.Area = "Sin ubicación"
		}
		p.Source = "submission"
		p.Created = time.Now().UTC().Format(time.RFC3339)
		p.Status = "pending"
		p.Decided, p.MergedInto, p.LastCheckup = "", "", ""
		if publish {
			// Un punto público publicado sin revisión conserva source "submission".
			if by != nil {
				p.Source = "user"
			}
			if reviewed {
				p.Source = "admin"
			}
			p.Status = "approved"
			p.Decided, p.LastCheckup = p.Created, p.Created
		}
		errUnknownUser := errors.New("unknown user")
		var already *Point
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			// Reenvío desde la cola sin conexión: si ya llegó, se responde con el punto guardado.
			if p.ClientRef != "" {
				errFound := errors.New("found")
				if e := b.ForEach(func(_, v []byte) error {
					var existing Point
					if err := json.Unmarshal(v, &existing); err != nil {
						return err
					}
					if existing.ClientRef == p.ClientRef {
						already = &existing
						return errFound
					}
					return nil
				}); e == errFound {
					return nil
				} else if e != nil {
					return e
				}
			}
			if reviewed && by == nil && askedBy != "" {
				u, ok := getUser(tx, askedBy)
				if !ok {
					return errUnknownUser
				}
				p.AddedBy, p.AddedByName = u.ID, u.Name
			}
			verifyStamp(tx, by, &p.VerifiedTeam, &p.VerifiedBy, &p.VerifiedAt)
			if hold && by != nil && publish {
				if t := teamOf(tx, by.ID); t != "" {
					p.HeldBy, p.HeldUntil = t, time.Now().Add(holdHours*time.Hour).UTC().Format(time.RFC3339)
				}
			}
			data, _ := json.Marshal(p)
			if reviewed {
				errDuplicate := errors.New("duplicate code")
				if err := b.ForEach(func(_, v []byte) error {
					var existing Point
					if err := json.Unmarshal(v, &existing); err != nil {
						return err
					}
					if existing.Status != "rejected" && strings.EqualFold(existing.Code, p.Code) {
						return errDuplicate
					}
					return nil
				}); err != nil {
					return err
				}
			}
			if e := b.Put([]byte(p.ID), data); e != nil || !publish {
				return e
			}
			return needFromPoint(tx, p)
		})
		if err != nil {
			if err == errUnknownUser {
				failField(w, 400, "addedBy", "No hay un usuario con ese ID.")
				return
			}
			if err.Error() == "duplicate code" {
				failField(w, 409, "code", "Este código ya existe. Usa otro o deja el campo vacío para asignar uno automáticamente.")
				return
			}
			fail(w, 500, "database error")
			return
		}
		if already != nil {
			response(w, 200, map[string]any{"id": already.ID, "status": already.Status, "duplicate": true})
			return
		}
		response(w, 201, map[string]string{"id": p.ID, "status": p.Status})
	}
	// Público: queda por revisar, salvo que administración haya apagado «Requerir revisión de puntos».
	mux.HandleFunc("POST /api/submissions", func(w http.ResponseWriter, r *http.Request) {
		createPoint(w, r, false, !a.requireReviews(), nil)
	})
	// ?pending=1: administración lo crea sin publicar (queda en «Puntos por revisar»), p. ej. al importar reportes.
	mux.HandleFunc("POST /api/admin/points", func(w http.ResponseWriter, r *http.Request) {
		createPoint(w, r, true, r.URL.Query().Get("pending") != "1", agentBy(r))
	})
	// Usuario con código («!código» en la página pública o en su panel): su punto se publica sin revisión.
	mux.HandleFunc("POST /api/full/points", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			createPoint(w, r, false, true, &u)
		}
	})
	mux.HandleFunc("POST /api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) {
			return
		}
		ip := r.RemoteAddr
		if h, _, e := strings.Cut(ip, ":"); e {
			ip = h
		}
		a.mu.Lock()
		at := a.attempts[ip]
		blocked := at.count >= 5 && time.Now().Before(at.until)
		a.mu.Unlock()
		var in struct {
			Password string `json:"password"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			fail(w, 400, "invalid JSON")
			return
		}
		valid := a.passwordValid(in.Password)
		if blocked && !valid {
			fail(w, 429, "Demasiados intentos incorrectos desde esta red. La contraseña correcta todavía permite entrar.")
			return
		}
		if !valid {
			a.mu.Lock()
			at = a.attempts[ip]
			if time.Now().After(at.until) {
				at.count = 0
			}
			at.count++
			at.until = time.Now().Add(10 * time.Minute)
			a.attempts[ip] = at
			a.mu.Unlock()
			fail(w, 401, "Contraseña incorrecta. Vuelve a intentarlo.")
			return
		}
		a.mu.Lock()
		delete(a.attempts, ip)
		token := randomToken()
		s := session{csrf: randomToken(), expires: time.Now().Add(8 * time.Hour)}
		a.sessions[token] = s
		a.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "relief_session", Value: token, Path: "/api/admin", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
		response(w, 200, map[string]string{"csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/admin/session", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		c, _ := r.Cookie("relief_session")
		a.mu.Lock()
		s := a.sessions[c.Value]
		a.mu.Unlock()
		response(w, 200, map[string]string{"csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/admin/points", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		points := []Point{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				points = append(points, p)
				return nil
			})
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, points)
	})
	mux.HandleFunc("GET /api/admin/submissions", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		points := []Point{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(submissions).ForEach(func(k, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status == "pending" {
					points = append(points, p)
				}
				return nil
			})
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, points)
	})
	mux.HandleFunc("POST /api/admin/seed", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) || !a.admin(w, r) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var rows []Point
		d := json.NewDecoder(r.Body)
		if err := d.Decode(&rows); err != nil || len(rows) == 0 || len(rows) > 500 {
			fail(w, 400, "expected 1–500 points")
			return
		}
		seen := map[string]bool{}
		for _, p := range rows {
			p.Code = strings.TrimSpace(p.Code)
			if p.Code == "" || len(p.Code) > 45 || len(p.Area) > 80 || seen[strings.ToLower(p.Code)] || (len(p.Coords) != 0 && (len(p.Coords) != 2 || p.Coords[0] < 6 || p.Coords[0] > 10 || p.Coords[1] < -83 || p.Coords[1] > -78)) {
				fail(w, 400, "invalid or duplicate checkpoint")
				return
			}
			seen[strings.ToLower(p.Code)] = true
		}
		added := 0
		updated := 0
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			for _, p := range rows {
				key := []byte("seed:" + strings.ToLower(strings.TrimSpace(p.Code)))
				if old := b.Get(key); old != nil {
					var previous Point
					if e := json.Unmarshal(old, &previous); e != nil {
						return e
					}
					if previous.Source != "sheet" {
						continue
					}
					p.ID = previous.ID
					p.Source = "sheet"
					p.Status = previous.Status
					p.Created = previous.Created
					p.Decided = previous.Decided
					p.Note = previous.Note
					p.History = previous.History
					p.Checkups = previous.Checkups
					data, e := json.Marshal(p)
					if e != nil {
						return e
					}
					if string(data) != string(old) {
						if e = b.Put(key, data); e != nil {
							return e
						}
						updated++
					}
					continue
				}
				p.ID = string(key)
				p.Source = "sheet"
				p.Status = "approved"
				p.Created = time.Now().UTC().Format(time.RFC3339)
				p.Decided = p.Created
				data, e := json.Marshal(p)
				if e != nil {
					return e
				}
				if e = b.Put(key, data); e != nil {
					return e
				}
				added++
			}
			return nil
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, map[string]int{"added": added, "updated": updated, "skipped": len(rows) - added - updated})
	})
	mux.HandleFunc("POST /api/admin/decisions", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) || !a.admin(w, r) {
			return
		}
		var in struct {
			ID       string `json:"id"`
			Decision string `json:"decision"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.ID) == 0 || len(in.ID) > 120 || (in.Decision != "approved" && in.Decision != "rejected" && in.Decision != "archived") {
			fail(w, 400, "invalid decision")
			return
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			v := b.Get([]byte(in.ID))
			if v == nil {
				return errors.New("not found")
			}
			var p Point
			if e := json.Unmarshal(v, &p); e != nil {
				return e
			}
			if p.MergedInto != "" {
				return errors.New("merged report")
			}
			// Pendiente → archivado: «Rechazar» en Puntos por revisar.
			if (p.Status == "pending" && in.Decision != "approved" && in.Decision != "rejected" && in.Decision != "archived") ||
				(p.Status == "approved" && in.Decision != "archived") ||
				(p.Status == "archived" && in.Decision != "approved") ||
				(p.Status == "rejected" && in.Decision != "approved") ||
				(p.Status != "pending" && p.Status != "approved" && p.Status != "archived" && p.Status != "rejected") {
				return errors.New("already decided")
			}
			// Rejected points were never approved, so approving one creates its need like a pending approval.
			firstApproval := (p.Status == "pending" || p.Status == "rejected") && in.Decision == "approved"
			// Un pendiente rechazado (archivado sin aprobarse nunca) no tiene necesidades guardadas: recuperarlo es su primera aprobación.
			if p.Status == "archived" && in.Decision == "approved" && p.LastCheckup == "" {
				had, e := campHasNeedRecords(tx, p.ID)
				if e != nil {
					return e
				}
				firstApproval = !had
			}
			p.Status = in.Decision
			p.Decided = time.Now().UTC().Format(time.RFC3339)
			if firstApproval {
				p.LastCheckup = p.Decided
			}
			data, _ := json.Marshal(p)
			if e := b.Put([]byte(in.ID), data); e != nil || !firstApproval {
				return e
			}
			return needFromPoint(tx, p)
		})
		if err != nil {
			fail(w, 409, err.Error())
			return
		}
		response(w, 200, map[string]string{"id": in.ID, "status": in.Decision})
	})
	mux.HandleFunc("POST /api/admin/password", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) || !a.admin(w, r) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		// Para cambiarla hay que escribir la contraseña actual, además de tener la sesión abierta.
		var in struct {
			Current  string `json:"current"`
			Password string `json:"password"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.Password) < 12 || len(in.Password) > 72 {
			failField(w, 400, "password", "La nueva contraseña debe tener entre 12 y 72 caracteres.")
			return
		}
		if !a.passwordValid(in.Current) {
			failField(w, 403, "current", "La contraseña actual no es correcta.")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			fail(w, 500, "hash error")
			return
		}
		if err = a.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(meta).Put([]byte("password_hash"), hash) }); err != nil {
			fail(w, 500, "database error")
			return
		}
		a.mu.Lock()
		a.sessions = map[string]session{}
		a.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "relief_session", Value: "", Path: "/api/admin", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "password changed; sign in again"})
	})
	mux.HandleFunc("POST /api/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) || !a.admin(w, r) {
			return
		}
		c, _ := r.Cookie("relief_session")
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "relief_session", Value: "", Path: "/api/admin", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	a.campRoutes(mux)
	a.runnerRoutes(mux)
	a.mergeRoutes(mux)
	a.boletinRoutes(mux)
	a.userRoutes(mux)
	a.signupRoutes(mux)
	a.statsRoutes(mux)
	a.donationRoutes(mux)
	a.supplierRoutes(mux)
	a.teamInviteRoutes(mux)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name := ""
		switch r.URL.Path {
		case "/", "/index.html":
			name = "index.html"
		case "/admin.html", "/donations.html", "/supplier.html", "/signup.html", "/team.html", "/camp.html", "/runner.html", "/builder.html", "/command.html", "/pointform.js", "/nav.js", "/leaflet.css", "/leaflet.js", "/report/boletin_general.html", "/report/logo.png":
			name = strings.TrimPrefix(r.URL.Path, "/")
		default:
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(a.static, name))
	})
	return mux
}
func (a *App) campRoutes(mux *http.ServeMux) {
	runnerFields := map[string]int{"name": 120, "telephone": 40}
	needFields := map[string]int{"need": 8192, "status": 32, "urgency": 32, "runnerId": 128, "kind": 16, "builderId": 128, "repeatDays": 0, "nextDueAt": 40}
	// Corredores y constructores son usuarios con ese rol: nombre y teléfono privados.
	for path, role := range map[string]string{"runners": roleRunner, "builders": roleBuilder} {
		mux.HandleFunc("GET /api/admin/"+path, func(w http.ResponseWriter, r *http.Request) {
			if !a.admin(w, r) {
				return
			}
			list := []Runner{}
			err := a.db.View(func(tx *bolt.Tx) error {
				return eachUser(tx, func(u User) error {
					if u.has(role) {
						list = append(list, u.runner())
					}
					return nil
				})
			})
			if err != nil {
				fail(w, 500, "database error")
				return
			}
			response(w, 200, list)
		})
		mux.HandleFunc("POST /api/admin/"+path, func(w http.ResponseWriter, r *http.Request) {
			if !a.adminWrite(w, r) {
				return
			}
			f, ok := decodeFields(w, r, runnerFields)
			if !ok {
				return
			}
			name, telephone := fieldText(f, "name"), fieldText(f, "telephone")
			if name == "" || telephone == "" {
				fail(w, 400, "name and telephone required")
				return
			}
			u := User{ID: randomToken(), Name: name, Telephone: telephone, Roles: []string{role}, CreatedAt: nowStamp()}
			if a.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(usersBucket), u.ID, u) }) != nil {
				fail(w, 500, "database error")
				return
			}
			response(w, 201, u.runner())
		})
		mux.HandleFunc("PATCH /api/admin/"+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
			if !a.adminWrite(w, r) {
				return
			}
			f, ok := decodeFields(w, r, runnerFields)
			if !ok {
				return
			}
			id := r.PathValue("id")
			var item User
			err := a.db.Update(func(tx *bolt.Tx) error {
				var found bool
				if item, found = getUser(tx, id); !found || !item.has(role) {
					return errMissing
				}
				if _, yes := f["name"]; yes {
					item.Name = fieldText(f, "name")
				}
				if _, yes := f["telephone"]; yes {
					item.Telephone = fieldText(f, "telephone")
				}
				if item.Name == "" || item.Telephone == "" {
					return errInvalid
				}
				item.UpdatedAt = nowStamp()
				return putJSON(tx.Bucket(usersBucket), id, item)
			})
			if serviceError(w, err) {
				return
			}
			response(w, 200, item.runner())
		})
	}
	mux.HandleFunc("GET /api/admin/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		a.listNeeds(w, r.PathValue("id"), false, "")
	})
	mux.HandleFunc("GET /api/admin/needs", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		list := []CampNeed{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				if n.Status != "DELETED" {
					list = append(list, n)
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("GET /api/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) { a.listNeeds(w, r.PathValue("id"), true, "") })
	mux.HandleFunc("POST /api/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 32, "requesterName": 120, "requesterPhone": 40})
		if !ok {
			return
		}
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: fieldText(f, "urgency"), CreatedAt: nowStamp(),
			RequesterName: strings.Join(strings.Fields(fieldText(f, "requesterName")), " "), RequesterPhone: strings.Join(strings.Fields(fieldText(f, "requesterPhone")), " ")}
		n.Urgency = normalizeUrgency(n.Urgency)
		// «Reparación de casa» va aparte, para constructores (putNeedSplit).
		switch {
		case n.Need == "":
			failField(w, 400, "need", "Describe la necesidad.")
			return
		case !validUrgency(n.Urgency):
			failField(w, 400, "urgency", "Elige una urgencia válida.")
			return
		case n.RequesterName == "":
			failField(w, 400, "requesterName", "Escribe tu nombre.")
			return
		case !validPhone(n.RequesterPhone):
			failField(w, 400, "requesterPhone", "Escribe tu número de teléfono (7 dígitos o más).")
			return
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			if !a.campExists(tx, n.CampID) {
				return errMissing
			}
			var e error
			n, e = putNeedSplit(tx, n)
			return e
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, map[string]string{"id": n.ID, "status": n.Status})
	})
	mux.HandleFunc("POST /api/admin/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, needFields)
		if !ok {
			return
		}
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: fieldText(f, "urgency"), RunnerID: fieldText(f, "runnerId"), BuilderID: fieldText(f, "builderId"), Kind: storedKind(fieldText(f, "kind")), CreatedAt: nowStamp()}
		if !validKind(fieldText(f, "kind")) {
			failField(w, 400, "kind", "invalid kind")
			return
		}
		if _, yes := f["status"]; yes {
			n.Status = fieldText(f, "status")
		}
		n.Urgency = normalizeUrgency(n.Urgency)
		if n.Need == "" || !validStatus(n.Status) || !validUrgency(n.Urgency) {
			fail(w, 400, "invalid need, status or urgency")
			return
		}
		if field, msg := applyService(&n, f); field != "" {
			failField(w, 400, field, msg)
			return
		}
		if n.Status == "COMPLETED" {
			n.CompletedAt = n.CreatedAt
		}
		if n.Status == "IN_PROGRESS" {
			n.EnRouteAt = n.CreatedAt
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			if !a.campExists(tx, n.CampID) {
				return errMissing
			}
			if !validAssignee(tx, n) {
				return errInvalid
			}
			enRouteOnAssign(&n, "")
			verifyStamp(tx, agentBy(r), &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, n)
	})
	mux.HandleFunc("PATCH /api/admin/needs/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, needFields)
		if !ok {
			return
		}
		var n CampNeed
		id := r.PathValue("id")
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(id))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			runnerBefore, orig := n.RunnerID, n
			if _, yes := f["need"]; yes {
				if text := fieldText(f, "need"); text != n.Need {
					n.Need, n.RevisedAt = text, nowStamp()
				}
			}
			removed := false
			if _, yes := f["status"]; yes {
				status := fieldText(f, "status")
				removed = status == "DELETED" && n.Status != "DELETED"
				if status == "COMPLETED" && n.Status != "COMPLETED" {
					n.CompletedAt = nowStamp()
				} else if status != "COMPLETED" {
					n.CompletedAt = ""
				}
				if status == "IN_PROGRESS" && n.Status != "IN_PROGRESS" {
					n.EnRouteAt = nowStamp()
				} else if status == "OPEN" {
					n.EnRouteAt = ""
				}
				n.Status = status
			}
			if _, yes := f["urgency"]; yes {
				n.Urgency = fieldText(f, "urgency")
			}
			n.Urgency = normalizeUrgency(n.Urgency)
			if _, yes := f["kind"]; yes {
				if !validKind(fieldText(f, "kind")) {
					return errInvalid
				}
				// Al cambiar de tipo se suelta la asignación del otro equipo.
				if kind := storedKind(fieldText(f, "kind")); kind != n.Kind {
					n.Kind, n.RunnerID, n.BuilderID = kind, "", ""
				}
			}
			if _, yes := f["runnerId"]; yes {
				n.RunnerID = fieldText(f, "runnerId")
			}
			if _, yes := f["builderId"]; yes {
				n.BuilderID = fieldText(f, "builderId")
			}
			if field, _ := applyService(&n, f); field != "" {
				return errInvalid
			}
			if n.Need == "" || !validStatus(n.Status) || !validUrgency(n.Urgency) || !validAssignee(tx, n) {
				return errInvalid
			}
			// Otro runner en una necesidad que ya lleva alguien: el run de ese runner no se toca; lo pedido va al run
			// propio del otro (el que ya tenga abierto o en camino en este grupo, o uno nuevo).
			if n.Kind == orig.Kind && otherRunner(orig, n.RunnerID) && n.Status != "DELETED" {
				run, has, e := runOf(tx, orig, n.RunnerID)
				if e != nil {
					return e
				}
				if !has {
					if run, e = newRun(tx, orig, n.RunnerID, n.Status); e != nil {
						return e
					}
				} else {
					before := run.Status
					run.Status = n.Status
					if run.Status == "OPEN" {
						run.Status = "IN_PROGRESS"
					}
					if run.Status == "COMPLETED" && before != "COMPLETED" {
						run.CompletedAt = nowStamp()
					}
				}
				if run.Need != n.Need {
					run.Need, run.RevisedAt = n.Need, nowStamp()
				}
				run.Urgency = n.Urgency
				n, removed = run, false
				runnerBefore = n.RunnerID
			}
			// Quien no es el runner cambia lo que pide su carrera: el run sigue igual y lo agregado queda como otra necesidad.
			var extra *CampNeed
			if n.ID == orig.ID && n.Need != orig.Need && orig.Kind == "" && orig.RunnerID != "" && n.RunnerID == orig.RunnerID && (orig.Status == "IN_PROGRESS" || orig.Status == "COMPLETED") {
				if by := agentBy(r); by == nil || by.ID != orig.RunnerID {
					byID, byName := actor(r)
					var e error
					if extra, e = editOfRun(orig, n.Need, byID, byName); e != nil {
						return e
					}
					n.Need, n.RevisedAt = orig.Need, orig.RevisedAt
					verifyStamp(tx, agentBy(r), &extra.VerifiedTeam, &extra.VerifiedBy, &extra.VerifiedAt)
					if e := putNeed(b, extra); e != nil {
						return e
					}
				}
			}
			enRouteOnAssign(&n, runnerBefore)
			// «Ya no necesita nada» (quitar la necesidad) confirma el estado del punto: cuenta como un checkup.
			if removed {
				if e := markCheckup(tx, n.CampID, true); e != nil {
					return e
				}
			}
			verifyStamp(tx, agentBy(r), &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			return putNeed(b, &n)
		})
		if runError(w, err) {
			return
		}
		response(w, 200, n)
	})
	mux.HandleFunc("PATCH /api/admin/camps/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, campFields)
		if !ok {
			return
		}
		if key, msg := campFieldError(f); key != "" {
			failField(w, 400, key, msg)
			return
		}
		id := r.PathValue("id")
		var p Point
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			v := b.Get([]byte(id))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &p); e != nil {
				return e
			}
			// También los por revisar, para corregirlos antes de aprobarlos; siguen ocultos al público hasta la aprobación.
			if p.Status != "approved" && p.Status != "pending" {
				return errMissing
			}
			if e := applyCampFields(&p, f); e != nil {
				return e
			}
			if p.Status == "approved" {
				p.LastCheckup = nowStamp()
			}
			verifyStamp(tx, agentBy(r), &p.VerifiedTeam, &p.VerifiedBy, &p.VerifiedAt)
			return putJSON(b, id, p)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, p)
	})
}

// Campos editables de un punto aprobado y su largo máximo (0 = sin límite de texto o de solo lectura).
var campFields = map[string]int{"code": 45, "area": 80, "place": 200, "needs": 8192, "people": 200, "note": noteLimit, "source": 0, "status": 0, "id": 0, "created": 0, "decided": 0, "campStatus": 0, "highestUrgency": 0, "needStatuses": 0, "service": 200, "serviceNotes": 8192, "children": 100, "babies": 100, "notes": 8192, "directions": 8192, "extra": 8192, "gps": 1000, "coords": 0, "row": 100, "ordinal": 100, "coordinator": 200, "contact": 200, "phone": 40, "adults": 100, "olderAdults": 100, "disabled": 100, "families": 100, "secondContact": 200, "secondPhone": 40, "thirdContact": 200, "thirdPhone": 40, "generator": 200, "members": 0, "coordinatorPhone": 40, "mergedInto": 0, "lastCheckup": 0, "history": 0, "checkups": 0, "clientRef": 0}

// campFieldError rechaza campos de solo lectura y coordenadas fuera de Panamá.
func campFieldError(f map[string]json.RawMessage) (string, string) {
	for _, key := range []string{"source", "status", "id", "created", "decided", "campStatus", "highestUrgency", "needStatuses", "mergedInto", "lastCheckup", "history", "checkups", "clientRef"} {
		if _, yes := f[key]; yes {
			return key, "read-only field"
		}
	}
	if raw, yes := f["coords"]; yes {
		var coords []float64
		if json.Unmarshal(raw, &coords) != nil || len(coords) != 0 && (len(coords) != 2 || coords[0] < 6 || coords[0] > 10 || coords[1] < -83 || coords[1] > -78) {
			return "coords", "invalid coordinates"
		}
	}
	return "", ""
}

// applyCampFields superpone los campos enviados sobre el punto y normaliza código, área, personas y GPS.
// También registra en el historial los valores anteriores de Personas y de la lista de personas.
func applyCampFields(p *Point, f map[string]json.RawMessage) error {
	before := *p
	// json.Unmarshal reutiliza el arreglo de p.Members; sin copia, before vería la lista nueva.
	before.Members = slices.Clone(p.Members)
	data, e := json.Marshal(p)
	if e != nil {
		return e
	}
	var obj map[string]json.RawMessage
	if e = json.Unmarshal(data, &obj); e != nil {
		return e
	}
	for key, val := range f {
		obj[key] = val
	}
	data, e = json.Marshal(obj)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(data, p); e != nil {
		return errInvalid
	}
	p.Code = strings.TrimSpace(p.Code)
	p.Area = strings.TrimSpace(p.Area)
	if p.Code == "" || p.Area == "" || normalizeMembers(p) != nil {
		return errInvalid
	}
	if _, yes := f["gps"]; yes {
		if _, explicit := f["coords"]; !explicit {
			if p.GPS == "" {
				p.Coords = nil
			} else {
				p.Coords, _ = parseMapLink(p.GPS)
			}
		}
	}
	recordHistory(before, p, nowStamp())
	return nil
}

// La nota privada guarda el registro de fusiones, por eso admite más texto que las demás notas.
const noteLimit = 20000

// mergeNeeds agrega los pedidos a la necesidad Abierta más reciente del mismo tipo (queda «Actualizado»)
// o, si no hay ninguna abierta, crea una sola con todos. «Reparación de casa» va siempre a una necesidad de reparación aparte.
// Las necesidades En curso no se tocan.
func mergeNeeds(tx *bolt.Tx, campID string, items []string) error {
	supplies, repairs := []string{}, []string{}
	for _, it := range items {
		if it = strings.TrimSpace(it); it == "" {
			continue
		}
		if foldText(it) == foldText(repairLabel) {
			repairs = append(repairs, repairLabel)
		} else {
			supplies = append(supplies, it)
		}
	}
	if e := mergeNeedKind(tx, campID, "", supplies); e != nil {
		return e
	}
	if len(repairs) > 0 {
		return mergeNeedKind(tx, campID, "repair", repairs[:1])
	}
	return nil
}
func mergeNeedKind(tx *bolt.Tx, campID, kind string, items []string) error {
	if len(items) == 0 {
		return nil
	}
	b := tx.Bucket(needsBucket)
	var open *CampNeed
	if e := b.ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID == campID && n.Kind == kind && n.Status == "OPEN" && (open == nil || needActivity(n) > needActivity(*open)) {
			open = &n
		}
		return nil
	}); e != nil {
		return e
	}
	if open == nil {
		n := CampNeed{ID: randomToken(), CampID: campID, Kind: kind, Need: strings.Join(items, ", "), Status: "OPEN", Urgency: urgencyNormal, CreatedAt: nowStamp()}
		return putNeed(b, &n)
	}
	open.Need = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(open.Need), ",") + ", " + strings.Join(items, ", "))
	open.RevisedAt = nowStamp()
	return putNeed(b, open)
}

// needActivity es la marca más reciente de una necesidad, igual que el orden del panel.
func needActivity(n CampNeed) string {
	latest := ""
	for _, t := range []string{n.UpdatedAt, n.CompletedAt, n.ReleasedAt, n.CreatedAt, n.RevisedAt} {
		if t > latest {
			latest = t
		}
	}
	return latest
}

// mergeRoutes fusiona un reporte pendiente en un punto aprobado: el panel envía los valores elegidos,
// las necesidades a crear y el registro de lo descartado, que se agrega a la nota privada del punto.
func (a *App) mergeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/admin/points/{id}/merge", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			fail(w, 415, "JSON required")
			return
		}
		var in struct {
			TargetID string                     `json:"targetId"`
			Fields   map[string]json.RawMessage `json:"fields"`
			Needs    []string                   `json:"needs"`
			Log      string                     `json:"log"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 131072)
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.TargetID == "" || len(in.TargetID) > 120 {
			fail(w, 400, "Solicitud de fusión inválida.")
			return
		}
		if in.Fields == nil {
			in.Fields = map[string]json.RawMessage{}
		}
		if _, yes := in.Fields["note"]; yes {
			failField(w, 400, "note", "read-only field")
			return
		}
		if key, msg := checkFields(in.Fields, campFields); key != "" {
			failField(w, 400, key, msg)
			return
		}
		if key, msg := campFieldError(in.Fields); key != "" {
			failField(w, 400, key, msg)
			return
		}
		if len(in.Needs) > 50 {
			fail(w, 400, "Máximo 50 necesidades por fusión.")
			return
		}
		for _, n := range in.Needs {
			if strings.TrimSpace(n) == "" || len(n) > 8192 {
				fail(w, 400, "Cada necesidad debe tener texto (máximo 8192 caracteres).")
				return
			}
		}
		id := r.PathValue("id")
		if id == in.TargetID {
			fail(w, 400, "Elige otro punto: no se puede fusionar consigo mismo.")
			return
		}
		var target Point
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			var src Point
			v := b.Get([]byte(id))
			if v == nil || json.Unmarshal(v, &src) != nil || src.Status != "pending" {
				return fmt.Errorf("%w: el reporte ya no está pendiente", errMissing)
			}
			v = b.Get([]byte(in.TargetID))
			if v == nil || json.Unmarshal(v, &target) != nil || target.Status != "approved" {
				return fmt.Errorf("%w: el punto destino no está aprobado", errMissing)
			}
			if e := applyCampFields(&target, in.Fields); e != nil {
				return e
			}
			now := time.Now().UTC()
			if log := strings.TrimSpace(in.Log); log != "" {
				header := fmt.Sprintf("— Fusión del reporte duplicado %s (%s UTC):", src.Code, now.Format("2006-01-02 15:04"))
				target.Note = strings.TrimSpace(strings.TrimSpace(target.Note) + "\n\n" + header + "\n" + log)
			}
			if len(target.Note) > noteLimit {
				return fmt.Errorf("%w: la nota privada quedaría demasiado larga; recórtala antes de fusionar", errInvalid)
			}
			target.LastCheckup = now.Format(time.RFC3339)
			if e := putJSON(b, target.ID, target); e != nil {
				return e
			}
			if e := mergeNeeds(tx, target.ID, in.Needs); e != nil {
				return e
			}
			src.Status, src.Decided, src.MergedInto = "rejected", now.Format(time.RFC3339), target.ID
			return putJSON(b, src.ID, src)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, target)
	})
	// «Publicar ahora»: termina la retención de un punto que un equipo iba a abastecer (si no, se publica solo).
	mux.HandleFunc("POST /api/admin/camps/{id}/publish", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		var p Point
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &p) != nil {
				return errMissing
			}
			p.HeldBy, p.HeldUntil = "", ""
			return putJSON(b, p.ID, p)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]string{"id": p.ID})
	})
	mux.HandleFunc("POST /api/admin/camps/{id}/checkup", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		var p Point
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" {
				return errMissing
			}
			if e := markCheckup(tx, p.ID, false); e != nil {
				return e
			}
			json.Unmarshal(b.Get([]byte(p.ID)), &p)
			return nil
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, p)
	})
}

// markCheckup anota ahora como último checkup del punto (y en su registro de checkups); sin punto aprobado no hace nada.
func markCheckup(tx *bolt.Tx, campID string, merge bool) error {
	b := tx.Bucket(submissions)
	var p Point
	v := b.Get([]byte(campID))
	if v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" {
		return nil
	}
	// merge: quitar necesidades justo después de un checkup (p. ej. «Sí, están bien») no agrega otro al registro, lo actualiza.
	// Cada clic en «Checkup» sí queda anotado.
	now := nowStamp()
	if last, e := time.Parse(time.RFC3339, p.LastCheckup); merge && e == nil && time.Since(last) < time.Minute && len(p.Checkups) > 0 {
		p.LastCheckup = now
		p.Checkups[len(p.Checkups)-1] = now
		return putJSON(b, p.ID, p)
	}
	p.LastCheckup = now
	p.Checkups = append(p.Checkups, p.LastCheckup)
	if len(p.Checkups) > checkupLimit {
		p.Checkups = p.Checkups[len(p.Checkups)-checkupLimit:]
	}
	return putJSON(b, p.ID, p)
}

// Corredores: sesión solo con nombre, separada del panel. Solo ven datos públicos del punto.
type runnerSession struct {
	runnerID string
	csrf     string
	expires  time.Time
	// agent: sesión de un agente en el panel de corredores; arma listas de empaque pero no toma carreras.
	agent bool
}
type runnerCamp struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Area   string `json:"area"`
	Place  string `json:"place,omitempty"`
	People string `json:"people,omitempty"`
	// Desglose de personas (para «60 personas · 30 adultos · 20 niños…» en las tarjetas).
	Adults      string    `json:"adults,omitempty"`
	Children    string    `json:"children,omitempty"`
	Babies      string    `json:"babies,omitempty"`
	OlderAdults string    `json:"olderAdults,omitempty"`
	Disabled    string    `json:"disabled,omitempty"`
	Families    string    `json:"families,omitempty"`
	Directions  string    `json:"directions,omitempty"`
	Coordinator string    `json:"coordinator,omitempty"`
	GPS         string    `json:"gps,omitempty"`
	Coords      []float64 `json:"coords,omitempty"`
	// Contactos del punto para que el corredor llame antes de tomar la carrera; nunca en páginas públicas.
	Contacts []runnerContact `json:"contacts,omitempty"`
	// team: lo agregó alguien del equipo de quien mira o una carrera la lleva alguien de su equipo (filtro «Solo Nuestro Equipo»).
	Team bool `json:"team,omitempty"`
	// «Verificado por equipo» del punto (privado: solo paneles con sesión).
	VerifiedTeam string `json:"verifiedTeam,omitempty"`
	VerifiedBy   string `json:"verifiedBy,omitempty"`
	VerifiedAt   string `json:"verifiedAt,omitempty"`
}
type runnerContact struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone"`
}

func campContacts(p Point) []runnerContact {
	list := []runnerContact{}
	if phone := strings.TrimSpace(p.CoordinatorPhone); phone != "" {
		name := "Coordinador"
		if p.Coordinator != "" {
			name = p.Coordinator + " (coordinador)"
		}
		list = append(list, runnerContact{Name: name, Phone: phone})
	}
	if len(p.Members) > 0 {
		for _, m := range p.Members {
			if phone := strings.TrimSpace(m.Phone); phone != "" {
				list = append(list, runnerContact{Name: m.Name, Phone: phone})
			}
		}
		return list
	}
	for _, c := range [][2]string{{p.Contact, p.Phone}, {p.SecondContact, p.SecondPhone}, {p.ThirdContact, p.ThirdPhone}} {
		if phone := strings.TrimSpace(c[1]); phone != "" {
			list = append(list, runnerContact{Name: strings.TrimSpace(c[0]), Phone: phone})
		}
	}
	return list
}

type runnerNeed struct {
	ID             string `json:"id"`
	Need           string `json:"need"`
	Status         string `json:"status"`
	Urgency        string `json:"urgency"`
	RequesterName  string `json:"requesterName,omitempty"`
	RequesterPhone string `json:"requesterPhone,omitempty"`
	CreatedAt      string `json:"createdAt,omitempty"`
	RevisedAt      string `json:"revisedAt,omitempty"`
	EnRouteAt      string `json:"enRouteAt,omitempty"`
	// Despachada desde acopio (sigue en camino).
	DispatchedAt string `json:"dispatchedAt,omitempty"`
	// Pedido de acopio (Abierta) o acopio donde se recoge (en camino): quién, dónde y su teléfono para llamarle.
	RequestedAt string `json:"requestedAt,omitempty"`
	// «Pedidos»: acopios que ofrecen surtir esta necesidad Abierta; el runner elige de cuál la recoge.
	OpenOffers []supplyOffer `json:"openOffers,omitempty"`
	Supplier   *supplierInfo `json:"supplier,omitempty"`
	// Pedido del corredor a los acopios (en su carrera): cuándo lo pidió y qué acopios dijeron que pueden surtirla.
	SupplyAskedAt string        `json:"supplyAskedAt,omitempty"`
	SupplyOffers  []supplyOffer `json:"supplyOffers,omitempty"`
	// Otros runners que llevan lo mismo al mismo punto (envíos hermanos).
	Also        []alsoRun     `json:"also,omitempty"`
	Packing     []PackingLine `json:"packing,omitempty"`
	Media       []MediaItem   `json:"media,omitempty"`
	Scope       string        `json:"scope,omitempty"`
	Materials   []PackingLine `json:"materials,omitempty"`
	Progress    int           `json:"progress,omitempty"`
	Kind        string        `json:"kind,omitempty"`
	RepeatDays  int           `json:"repeatDays,omitempty"`
	Assessments []Assessment  `json:"assessments,omitempty"`
	// assigned: la necesidad está asignada a quien consulta (puede armar su lista de empaque).
	Assigned bool       `json:"assigned,omitempty"`
	Camp     runnerCamp `json:"camp"`
	// «Verificado por equipo» de la necesidad (paneles con sesión).
	VerifiedTeam string `json:"verifiedTeam,omitempty"`
	VerifiedBy   string `json:"verifiedBy,omitempty"`
	VerifiedAt   string `json:"verifiedAt,omitempty"`
}

var errBusy = errors.New("Ya tienes 10 carreras en camino. Completa o deja alguna antes de tomar otra.")
var errTaken = errors.New("Esta necesidad ya no está disponible. Otra persona la tomó o cambió de estado.")
var errNotYours = errors.New("Esta carrera ya no está asignada a ti.")
var errBuilderBusy = errors.New("Ya tienes 20 reparaciones en curso. Termina o deja alguna antes de tomar otra.")
var errRepairTaken = errors.New("Esta reparación ya no está disponible. Otra persona la tomó o cambió de estado.")
var errRepairNotYours = errors.New("Esta reparación ya no está asignada a ti.")
var errUserDisabled = errors.New("Tu acceso está desactivado. Habla con administración.")
var errAgentNoRuns = errors.New("Como agente puedes armar listas de empaque, pero no tomar carreras.")
var errUseAccessCode = errors.New("Tienes un código de acceso: entra con tu código en vez de nombre y teléfono.")

func runnerNameKey(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

// crew agrupa a quienes atienden necesidades sin entrar al panel: corredores (suministros, una carrera a la vez)
// y constructores (reparaciones de casa, hasta 20 a la vez). Cada uno entra solo con nombre y teléfono.
type crew struct {
	role  string
	label string
	title string
	page  string
	// agents: este panel también acepta agentes (con código), en modo solo listas de empaque.
	agents   bool
	cookie   string
	path     string
	repair   bool
	max      int
	sessions map[string]runnerSession
	assignee func(n *CampNeed) *string
	busy     error
	taken    error
	notYours error
}

func (a *App) crews() (*crew, *crew) {
	runners := &crew{role: roleRunner, label: "corredores", title: "Corredores", page: "/runner.html", agents: true, cookie: "runner_session", path: "/api/runner", max: 10, sessions: a.runnerSessions,
		assignee: func(n *CampNeed) *string { return &n.RunnerID }, busy: errBusy, taken: errTaken, notYours: errNotYours}
	builders := &crew{role: roleBuilder, label: "constructores", title: "Constructores", page: "/builder.html", cookie: "builder_session", path: "/api/builder", repair: true, max: 20, sessions: a.builderSessions,
		assignee: func(n *CampNeed) *string { return &n.BuilderID }, busy: errBuilderBusy, taken: errRepairTaken, notYours: errRepairNotYours}
	return runners, builders
}
func (c *crew) fail(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errAgentNoRuns) {
		fail(w, 403, err.Error())
		return true
	}
	if errors.Is(err, c.busy) || errors.Is(err, c.taken) || errors.Is(err, c.notYours) || errors.Is(err, errNotOffered) {
		fail(w, 409, err.Error())
		return true
	}
	return serviceError(w, err)
}
func (a *App) crewSession(c *crew, w http.ResponseWriter, r *http.Request) (runnerSession, bool) {
	cookie, e := r.Cookie(c.cookie)
	if e != nil {
		fail(w, 401, "Escribe tu nombre para continuar.")
		return runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := c.sessions[cookie.Value]
	a.mu.Unlock()
	if !ok || time.Now().After(s.expires) {
		fail(w, 401, "Escribe tu nombre para continuar.")
		return runnerSession{}, false
	}
	if r.Method == http.MethodPost && r.Header.Get("X-CSRF-Token") != s.csrf {
		fail(w, 403, "CSRF token required")
		return runnerSession{}, false
	}
	a.touch(s.runnerID)
	return s, true
}

// userPanels: las páginas que el usuario puede abrir ahora con su código (las mismas que abre la entrada con código).
// Las páginas lo consultan seguido (nav.js) para mostrar las nuevas en el menú y salir de las que ya no tiene.
func userPanels(u User) []map[string]string {
	panels := []map[string]string{}
	if u.Disabled || u.CodeHash == "" {
		return panels
	}
	switch {
	case u.has(roleRunner):
		panels = append(panels, map[string]string{"label": "Corredores", "url": "/runner.html"})
	case u.has(roleAgent):
		panels = append(panels, map[string]string{"label": "Listas de empaque", "url": "/runner.html"})
	}
	if u.has(roleBuilder) {
		panels = append(panels, map[string]string{"label": "Constructores", "url": "/builder.html"})
	}
	if u.has(roleAgent) {
		panels = append(panels, map[string]string{"label": "Command", "url": "/command.html"})
	}
	if u.has(roleSupplier) {
		panels = append(panels, map[string]string{"label": "Supply", "url": "/supplier.html"})
	}
	if u.has(roleDonations) {
		panels = append(panels, map[string]string{"label": "Donaciones", "url": "/donations.html"})
	}
	if u.has(roleTeamManager) && u.TeamID != "" {
		panels = append(panels, map[string]string{"label": "Mi equipo", "url": "/team.html"})
	}
	return panels
}

// accessOf: roles y páginas actuales de un usuario, para las respuestas de sesión.
func (a *App) accessOf(id string) (roles []string, panels []map[string]string) {
	a.db.View(func(tx *bolt.Tx) error {
		if u, ok := getUser(tx, id); ok {
			roles, panels = u.Roles, userPanels(u)
		}
		return nil
	})
	if roles == nil {
		roles = []string{}
	}
	return roles, panels
}

func (a *App) runnerRoutes(mux *http.ServeMux) {
	runners, builders := a.crews()
	a.crewRoutes(mux, runners, runners, builders)
	a.crewRoutes(mux, builders, runners, builders)
	// Entrada oculta desde la página pública («!código» en Buscar): abre sesión en cada panel que tiene el usuario.
	mux.HandleFunc("POST /api/session/code", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"accessCode": 64})
		if !ok {
			return
		}
		code := normalizeAccessCode(fieldText(f, "accessCode"))
		if code == "" {
			failField(w, 400, "accessCode", "Escribe tu código de acceso.")
			return
		}
		item, ok := a.signInWithCode(w, r, code, "Código de acceso incorrecto.")
		if !ok {
			return
		}
		panels := []map[string]string{}
		for _, c := range []*crew{runners, builders} {
			agentOnly := !item.has(c.role)
			if agentOnly && !(c.agents && item.has(roleAgent)) {
				continue
			}
			a.startCrewSession(w, r, c, item.ID, agentOnly)
			label := c.title
			if agentOnly {
				label = "Listas de empaque"
			}
			panels = append(panels, map[string]string{"label": label, "url": c.page})
		}
		if item.has(roleAgent) {
			a.startAgentSession(w, r, item.ID)
			panels = append(panels, map[string]string{"label": "Command", "url": "/command.html"})
		}
		if item.has(roleSupplier) {
			a.startSupplierSession(w, r, item.ID)
			panels = append(panels, map[string]string{"label": "Supply", "url": "/supplier.html"})
		}
		if item.has(roleDonations) {
			a.startDeskSession(w, r, item.ID)
			panels = append(panels, map[string]string{"label": "Donaciones", "url": "/donations.html"})
		}
		if item.has(roleTeamManager) && item.TeamID != "" {
			panels = append(panels, map[string]string{"label": "Mi equipo", "url": "/team.html"})
		}
		if len(panels) == 0 && len(item.Roles) == 0 {
			// Alta desde la página pública todavía sin acceso (signup.go).
			failField(w, 403, "accessCode", "Tu cuenta está pendiente de aprobación. Administración te dará acceso pronto.")
			return
		}
		if len(panels) == 0 {
			failField(w, 403, "accessCode", "Tu código no tiene acceso a ningún panel.")
			return
		}
		a.startViewerSession(w, r, item.ID)
		response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "telephone": item.Telephone, "roles": item.Roles, "panels": panels})
	})
	// Vista completa: los mismos datos públicos, pero con el detalle de las carreras reservadas.
	mux.HandleFunc("GET /api/full/points", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			a.publicPoints(w, u.ID)
		}
	})
	mux.HandleFunc("GET /api/full/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			a.listNeeds(w, r.PathValue("id"), true, u.ID)
		}
	})
	// Tu propio ID (tocar tu nombre en la página pública lo copia), para sesiones guardadas antes de que se incluyera.
	mux.HandleFunc("GET /api/full/me", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			roles, panels := a.accessOf(u.ID)
			// team: nombre del equipo (vacío sin equipo); el formulario de punto pone a quien entra como coordinador.
			team := ""
			a.db.View(func(tx *bolt.Tx) error {
				if t, ok := getTeam(tx, teamOf(tx, u.ID)); ok {
					team = t.Name
				}
				return nil
			})
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "telephone": u.Telephone, "team": team, "locationConsent": u.locationConsent() != "", "roles": roles, "panels": panels})
		}
	})
	mux.HandleFunc("POST /api/full/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("viewer_session"); e == nil {
			a.mu.Lock()
			s, ok := a.viewerSessions[c.Value]
			delete(a.viewerSessions, c.Value)
			a.mu.Unlock()
			if ok {
				a.dropLocation(s.runnerID)
			}
		}
		http.SetCookie(w, &http.Cookie{Name: "viewer_session", Value: "", Path: "/api/full", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	a.packingRoutes(mux, runners)
	a.routingRoutes(mux, runners)
	a.mediaRoutes(mux, runners)
	a.repairRoutes(mux, builders)
	a.supplyAskRoutes(mux, runners)
	a.locationRoutes(mux, runners)
	a.settingsRoutes(mux)
	a.backupRoutes(mux)
	a.fullNeedRoutes(mux)
	a.teamRoutes(mux)
	a.trackRoutes(mux)
	a.agentRoutes(mux)
	mux.HandleFunc("POST /api/runner/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		sess, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 32})
		if !ok {
			return
		}
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: fieldText(f, "urgency"), CreatedAt: nowStamp()}
		n.Urgency = normalizeUrgency(n.Urgency)
		if n.Need == "" || !validUrgency(n.Urgency) {
			fail(w, 400, "Describe la necesidad y elige una urgencia válida.")
			return
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			if !a.campExists(tx, n.CampID) {
				return errMissing
			}
			if u, found := getUser(tx, sess.runnerID); found {
				verifyStamp(tx, &u, &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			}
			var e error
			n, e = putNeedSplit(tx, n)
			return e
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, map[string]string{"id": n.ID, "status": n.Status})
	})
}

// startCrewSession abre una sesión de ese equipo para el usuario y deja su cookie.
func (a *App) startCrewSession(w http.ResponseWriter, r *http.Request, c *crew, userID string, agent bool) runnerSession {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour), agent: agent}
	a.mu.Lock()
	c.sessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: c.cookie, Value: token, Path: c.path, HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
	return s
}

// crewRoutes: all son todos los equipos; quien entra con código queda con sesión en cada panel que su usuario tiene.
func (a *App) crewRoutes(mux *http.ServeMux, c *crew, all ...*crew) {
	mux.HandleFunc("POST "+c.path+"/session", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"name": 120, "telephone": 40, "accessCode": 64})
		if !ok {
			return
		}
		// Con código de acceso: el código identifica al usuario, sin nombre ni teléfono.
		if code := normalizeAccessCode(fieldText(f, "accessCode")); code != "" {
			item, ok := a.signInWithCode(w, r, code, "Código de acceso incorrecto. Revísalo o pide uno nuevo a administración.")
			if !ok {
				return
			}
			canUse := func(c *crew) bool { return item.has(c.role) || c.agents && item.has(roleAgent) }
			if !canUse(c) {
				failField(w, 403, "accessCode", "Tu código no tiene acceso al panel de "+c.label+".")
				return
			}
			var s runnerSession
			for _, other := range all {
				if canUse(other) {
					if started := a.startCrewSession(w, r, other, item.ID, !item.has(other.role)); other == c {
						s = started
					}
				}
			}
			response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "telephone": item.Telephone, "csrf": s.csrf, "roles": item.Roles, "agent": s.agent, "locationConsent": item.locationConsent() != ""})
			return
		}
		name := strings.Join(strings.Fields(fieldText(f, "name")), " ")
		if name == "" {
			failField(w, 400, "name", "Escribe tu nombre.")
			return
		}
		telephone := strings.Join(strings.Fields(fieldText(f, "telephone")), " ")
		if !validPhone(telephone) {
			failField(w, 400, "telephone", "Escribe tu número de teléfono.")
			return
		}
		var item User
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(usersBucket)
			key := runnerNameKey(name)
			// Se prefiere un usuario activo con este rol y nombre; uno desactivado no puede volver a registrarse.
			if e := eachUser(tx, func(existing User) error {
				if existing.has(c.role) && runnerNameKey(existing.Name) == key && (item.ID == "" || item.Disabled && !existing.Disabled) {
					item = existing
				}
				return nil
			}); e != nil {
				return e
			}
			if item.Disabled {
				return errUserDisabled
			}
			if item.CodeHash != "" {
				return errUseAccessCode
			}
			// Un teléfono guardado por curación no se sobrescribe; solo se completa si falta.
			if item.ID != "" {
				if item.Telephone != "" {
					return nil
				}
				item.Telephone, item.UpdatedAt = telephone, nowStamp()
				return putJSON(b, item.ID, item)
			}
			item = User{ID: randomToken(), Name: name, Telephone: telephone, Roles: []string{c.role}, CreatedAt: nowStamp()}
			return putJSON(b, item.ID, item)
		})
		if errors.Is(err, errUserDisabled) {
			fail(w, 403, err.Error())
			return
		}
		if errors.Is(err, errUseAccessCode) {
			failField(w, 403, "accessCode", err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		s := a.startCrewSession(w, r, c, item.ID, false)
		response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "csrf": s.csrf, "locationConsent": item.locationConsent() != ""})
	})
	mux.HandleFunc("GET "+c.path+"/session", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(c, w, r)
		if !ok {
			return
		}
		var item User
		err := a.db.View(func(tx *bolt.Tx) error {
			var found bool
			if item, found = getUser(tx, s.runnerID); !found || item.Disabled || !item.has(c.role) && !(s.agent && item.has(roleAgent)) {
				return errMissing
			}
			return nil
		})
		if errors.Is(err, errMissing) {
			fail(w, 401, "Escribe tu nombre para continuar.")
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "csrf": s.csrf, "agent": s.agent, "locationConsent": item.locationConsent() != "", "roles": item.Roles, "panels": userPanels(item)})
	})
	mux.HandleFunc("POST "+c.path+"/logout", func(w http.ResponseWriter, r *http.Request) {
		if cookie, e := r.Cookie(c.cookie); e == nil {
			a.mu.Lock()
			s, ok := c.sessions[cookie.Value]
			delete(c.sessions, cookie.Value)
			a.mu.Unlock()
			if ok && !c.repair {
				a.dropLocation(s.runnerID)
			}
		}
		http.SetCookie(w, &http.Cookie{Name: c.cookie, Value: "", Path: c.path, MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	// Corredores ven solo suministros y constructores solo reparaciones; «mine» son los trabajos en curso propios.
	mux.HandleFunc("GET "+c.path+"/needs", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(c, w, r)
		if !ok {
			return
		}
		mine, available := []runnerNeed{}, []runnerNeed{}
		err := a.db.View(func(tx *bolt.Tx) error {
			tv := newTeamView(tx, s.runnerID)
			team, e := teamCamps(tv)
			if e != nil {
				return e
			}
			camps := map[string]runnerCamp{}
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				// Retenido para otro equipo: sus runners todavía no lo ven.
				if p.Status == "approved" && !heldFrom(p, tv.team) {
					contacts := campContacts(p)
					// Paneles con sesión: el equipo y quién verificó sí se muestran (publicPoint los quita).
					vTeam, vBy := p.VerifiedTeam, p.VerifiedBy
					p = publicPoint(p)
					p.VerifiedTeam, p.VerifiedBy = vTeam, vBy
					camps[p.ID] = runnerCamp{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, People: p.People, Adults: p.Adults, Children: p.Children, Babies: p.Babies, OlderAdults: p.OlderAdults, Disabled: p.Disabled, Families: p.Families, Directions: p.Directions, Coordinator: p.Coordinator, GPS: p.GPS, Coords: p.Coords, Contacts: contacts, Team: team[p.ID], VerifiedTeam: p.VerifiedTeam, VerifiedBy: p.VerifiedBy, VerifiedAt: p.VerifiedAt}
				}
				return nil
			}); e != nil {
				return e
			}
			groups, e := needGroups(tx)
			if e != nil {
				return e
			}
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				camp, ok := camps[n.CampID]
				if !ok || isRepair(n) != c.repair {
					return nil
				}
				if !c.repair {
					if n.Need, _ = splitRepair(n.Need); n.Need == "" {
						return nil
					}
				}
				who := *c.assignee(&n)
				item := runnerNeed{VerifiedTeam: n.VerifiedTeam, VerifiedBy: n.VerifiedBy, VerifiedAt: n.VerifiedAt, ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, DispatchedAt: n.DispatchedAt, RequestedAt: openRequest(n), OpenOffers: openOffersFor(tx, n), Supplier: supplierOf(tx, n.SupplierID), SupplyAskedAt: n.SupplyAskedAt, SupplyOffers: offersFor(tx, n), Also: alsoFor(tx, groups, n), Packing: n.Packing, Scope: n.Scope, Materials: n.Materials, Progress: n.Progress, Kind: n.Kind, RepeatDays: n.RepeatDays, Assigned: who == s.runnerID, Camp: camp}
				if item.Assigned {
					item.Media = n.Media
				} else if c.repair {
					// Las fotos de evaluación las ven todos los constructores, para decidir si toman la obra.
					item.Media = inspectMedia(n.Media)
				}
				if c.repair {
					item.Assessments = n.Assessments
				}
				if s.agent {
					// Los agentes ven lo abierto o en ruta para armar listas de empaque, salvo las carreras de otro equipo.
					if (n.Status == "OPEN" || n.Status == "IN_PROGRESS") && !tv.hidden(n) {
						available = append(available, item)
					}
				} else if who == s.runnerID && n.Status == "IN_PROGRESS" {
					mine = append(mine, item)
				} else if n.Status == "OPEN" && (who == "" || who == s.runnerID) && !serviceScheduled(n, time.Now()) {
					available = append(available, item)
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(available, func(i, j int) bool {
			if ri, rj := urgencyRank(available[i].Urgency), urgencyRank(available[j].Urgency); ri != rj {
				return ri > rj
			}
			return available[i].Camp.Code < available[j].Camp.Code
		})
		sort.SliceStable(mine, func(i, j int) bool { return mine[i].EnRouteAt < mine[j].EnRouteAt })
		// El orden que eligió la persona (arrastrando sus carreras) va primero; las nuevas quedan al final.
		var order []string
		a.db.View(func(tx *bolt.Tx) error {
			if u, ok := getUser(tx, s.runnerID); ok {
				order = u.RunOrder
			}
			return nil
		})
		mine = orderRuns(mine, order)
		var current *runnerNeed
		if len(mine) > 0 {
			current = &mine[0]
		}
		response(w, 200, map[string]any{"current": current, "mine": mine, "available": available, "max": c.max})
	})
	// change aplica una transición solo si la necesidad sigue en el estado esperado y es del tipo de este equipo.
	change := func(w http.ResponseWriter, r *http.Request, apply func(tx *bolt.Tx, s runnerSession, n *CampNeed) error) {
		s, ok := a.crewSession(c, w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		var n CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return c.taken
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) || isRepair(n) != c.repair {
				return c.taken
			}
			if e := apply(tx, s, &n); e != nil {
				return e
			}
			return putNeed(b, &n)
		})
		if c.fail(w, err) {
			return
		}
		response(w, 200, map[string]string{"id": n.ID, "status": n.Status})
	}
	mux.HandleFunc("POST "+c.path+"/needs/{id}/take", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(tx *bolt.Tx, s runnerSession, n *CampNeed) error {
			if s.agent {
				return errAgentNoRuns
			}
			who := c.assignee(n)
			if n.Status != "OPEN" || *who != "" && *who != s.runnerID {
				return c.taken
			}
			active := map[string]bool{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var other CampNeed
				if e := json.Unmarshal(v, &other); e != nil {
					return e
				}
				if *c.assignee(&other) == s.runnerID && other.Status == "IN_PROGRESS" && isRepair(other) == c.repair {
					active[other.ID] = true
				}
				return nil
			}); e != nil {
				return e
			}
			if len(active) >= c.max {
				return c.busy
			}
			// Desde «Pedidos»: el runner elige de qué acopio la recoge (uno de los que ofrecieron); queda ligada a ese acopio.
			if sup := r.URL.Query().Get("supplier"); sup != "" && !c.repair {
				i := offerIndex(n.OpenOffers, sup)
				if i < 0 {
					return errNotOffered
				}
				n.SupplierID, n.SupplierRequestedAt = sup, n.OpenOffers[i].At
			}
			*who, n.Status, n.EnRouteAt = s.runnerID, "IN_PROGRESS", nowStamp()
			// La carrera nueva va al final del orden de la persona; se descartan las que ya no están en curso.
			if u, ok := getUser(tx, s.runnerID); ok {
				order := []string{}
				for _, id := range u.RunOrder {
					if active[id] {
						order = append(order, id)
					}
				}
				u.RunOrder = append(order, n.ID)
				return putJSON(tx.Bucket(usersBucket), u.ID, u)
			}
			return nil
		})
	})
	// Orden de las carreras en curso, como la persona las arrastró en su lista.
	mux.HandleFunc("POST "+c.path+"/order", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(c, w, r)
		if !ok {
			return
		}
		if s.agent {
			fail(w, 403, errAgentNoRuns.Error())
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"ids": 0})
		if !ok {
			return
		}
		var ids []string
		if json.Unmarshal(f["ids"], &ids) != nil || len(ids) > 200 {
			failField(w, 400, "ids", "Orden inválido.")
			return
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			u, found := getUser(tx, s.runnerID)
			if !found {
				return errMissing
			}
			// Solo se guardan carreras en curso de esta persona, sin repetir.
			own := map[string]bool{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				if *c.assignee(&n) == s.runnerID && n.Status == "IN_PROGRESS" {
					own[n.ID] = true
				}
				return nil
			}); e != nil {
				return e
			}
			u.RunOrder = []string{}
			for _, id := range ids {
				if own[id] {
					u.RunOrder = append(u.RunOrder, id)
					delete(own, id)
				}
			}
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"ids": ids})
	})
	mux.HandleFunc("POST "+c.path+"/needs/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(_ *bolt.Tx, s runnerSession, n *CampNeed) error {
			// Reenvío desde la pantalla sin conexión: si ya la completó esta misma persona, no es un error.
			if *c.assignee(n) == s.runnerID && n.Status == "COMPLETED" {
				return nil
			}
			if *c.assignee(n) != s.runnerID || n.Status != "IN_PROGRESS" {
				return c.notYours
			}
			n.Status, n.CompletedAt = "COMPLETED", nowStamp()
			if isRepair(*n) {
				n.Progress, n.ProgressAt = 100, n.CompletedAt
			}
			return nil
		})
	})
	mux.HandleFunc("POST "+c.path+"/needs/{id}/release", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(_ *bolt.Tx, s runnerSession, n *CampNeed) error {
			if *c.assignee(n) != s.runnerID || n.Status != "IN_PROGRESS" {
				return c.notYours
			}
			*c.assignee(n), n.Status, n.EnRouteAt, n.ReleasedAt, n.ReleasedBy = "", "OPEN", "", nowStamp(), s.runnerID
			return nil
		})
	})
}

// Boletín: cada corte guarda una copia pública (sin contactos ni teléfonos) de los puntos aprobados.
type BoletinPoint struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Area        string    `json:"area"`
	People      string    `json:"people,omitempty"`
	Adults      string    `json:"adults,omitempty"`
	Children    string    `json:"children,omitempty"`
	Babies      string    `json:"babies,omitempty"`
	OlderAdults string    `json:"olderAdults,omitempty"`
	Disabled    string    `json:"disabled,omitempty"`
	Needs       string    `json:"needs,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Service     string    `json:"service,omitempty"`
	GPS         string    `json:"gps,omitempty"`
	Coords      []float64 `json:"coords,omitempty"`
}
type Boletin struct {
	N        int            `json:"n"`
	Cut      string         `json:"cut"`
	Records  int            `json:"records"`
	People   int            `json:"people"`
	Attended int            `json:"attended"`
	Donor    string         `json:"donor,omitempty"`
	Points   []BoletinPoint `json:"points,omitempty"`
}

var firstNumber = regexp.MustCompile(`\d+`)

// countOf toma el primer número del texto reportado; sin número no se suma.
func countOf(text string) int {
	n, _ := strconv.Atoi(firstNumber.FindString(text))
	return n
}
func attendedService(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "yes", "sí", "si":
		return true
	}
	return false
}

// Código que el panel pide en cada acción de boletín (ver o generar), además de la sesión de administrador.
const boletinAccessCode = "PRINT5"

func validBoletinCode(code string) bool {
	return subtle.ConstantTimeCompare([]byte(strings.ToUpper(strings.TrimSpace(code))), []byte(boletinAccessCode)) == 1
}
func (a *App) boletinRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/admin/boletin-access", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"code": 40})
		if !ok {
			return
		}
		if !validBoletinCode(fieldText(f, "code")) {
			failField(w, 403, "code", "Código de acceso incorrecto.")
			return
		}
		response(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/admin/boletines", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"donor": 2000, "accessCode": 40})
		if !ok {
			return
		}
		if !validBoletinCode(fieldText(f, "accessCode")) {
			failField(w, 403, "accessCode", "Código de acceso incorrecto.")
			return
		}
		b := Boletin{Cut: nowStamp(), Donor: fieldText(f, "donor"), Points: []BoletinPoint{}}
		err := a.db.Update(func(tx *bolt.Tx) error {
			reserved, e := reservedNeedItems(newTeamView(tx, ""))
			if e != nil {
				return e
			}
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status != "approved" || heldFrom(p, "") {
					return nil
				}
				p.Needs = withoutReservedItems(p.Needs, reserved[p.ID])
				p = publicPoint(p)
				b.Points = append(b.Points, BoletinPoint{ID: p.ID, Code: p.Code, Area: p.Area, People: p.People, Adults: p.Adults, Children: p.Children, Babies: p.Babies, OlderAdults: p.OlderAdults, Disabled: p.Disabled, Needs: p.Needs, Notes: p.Notes, Service: p.Service, GPS: p.GPS, Coords: p.Coords})
				b.Records++
				b.People += countOf(p.People)
				if attendedService(p.Service) {
					b.Attended += countOf(p.People)
				}
				return nil
			}); e != nil {
				return e
			}
			sort.Slice(b.Points, func(i, j int) bool { return b.Points[i].Code < b.Points[j].Code })
			bucket := tx.Bucket(boletinesBucket)
			seq, e := bucket.NextSequence()
			if e != nil {
				return e
			}
			b.N = int(seq)
			return putJSON(bucket, fmt.Sprintf("%08d", b.N), b)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, map[string]any{"n": b.N, "cut": b.Cut})
	})
	mux.HandleFunc("GET /api/boletines", func(w http.ResponseWriter, r *http.Request) {
		list := []Boletin{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(boletinesBucket).ForEach(func(_, v []byte) error {
				var b Boletin
				if e := json.Unmarshal(v, &b); e != nil {
					return e
				}
				b.Points, b.Donor = nil, ""
				list = append(list, b)
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("GET /api/boletines/{n}", func(w http.ResponseWriter, r *http.Request) {
		var raw []byte
		err := a.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(boletinesBucket)
			if r.PathValue("n") == "latest" {
				_, raw = bucket.Cursor().Last()
			} else if n, e := strconv.Atoi(r.PathValue("n")); e == nil {
				raw = bucket.Get([]byte(fmt.Sprintf("%08d", n)))
			}
			if raw == nil {
				return errMissing
			}
			raw = append([]byte(nil), raw...)
			return nil
		})
		if errors.Is(err, errMissing) {
			fail(w, 404, "Todavía no hay un boletín con ese número.")
			return
		}
		if serviceError(w, err) {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(raw)
	})
}

var errMissing = errors.New("not found")
var errInvalid = errors.New("invalid fields or runner")

func serviceError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errMissing) {
		fail(w, 404, err.Error())
	} else if errors.Is(err, errInvalid) {
		fail(w, 400, err.Error())
	} else {
		fail(w, 500, "database error")
	}
	return true
}

// listNeeds: public quita datos privados; viewer (quien entró con código) además ve conductor, empaque y fotos de partida,
// salvo en las carreras que lleva un corredor con código de otro equipo, que ve como el público.
func (a *App) listNeeds(w http.ResponseWriter, id string, public bool, viewer string) {
	list := []CampNeed{}
	err := a.db.View(func(tx *bolt.Tx) error {
		tv := newTeamView(tx, viewer)
		if !a.campExists(tx, id) || public && campHeldFrom(tx, id, tv.team) {
			return errMissing
		}
		return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
			var n CampNeed
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if n.CampID == id && n.Status != "DELETED" && !(public && (isRepair(n) || serviceScheduled(n, time.Now()))) {
				if public {
					reserved := tv.hidden(n)
					reveal := viewer != "" && !reserved
					// Quien entró con código ve el nombre del conductor asignado y la lista de empaque (para asignar y empacar desde la ficha).
					if reveal && n.RunnerID != "" {
						if u, found := getUser(tx, n.RunnerID); found {
							n.RunnerName = u.Name
						}
					}
					n.RunnerID, n.ReleasedBy, n.DispatchedBy, n.DispatchedByName, n.SupplierID, n.SupplierRequestedAt = "", "", "", "", "", ""
					n.SupplyAskedAt, n.SupplyOffers, n.OpenOffers = "", nil, nil
					n.VerifiedTeam, n.VerifiedBy, n.CreatedBy, n.CreatedByName = "", "", "", ""
					if !reveal {
						n.Packing, n.Media = nil, nil
					} else {
						// Solo las fotos de partida, sin la clave de almacenamiento.
						pickups := []MediaItem{}
						for _, m := range n.Media {
							if m.Stage == stagePickup {
								m.StoredKey, m.AddedBy = "", ""
								pickups = append(pickups, m)
							}
						}
						n.Media = pickups
					}
					if n.Need, _ = splitRepair(n.Need); n.Need == "" {
						return nil
					}
					if reserved {
						n = CampNeed{ID: n.ID, CampID: n.CampID, Status: n.Status, Urgency: n.Urgency}
					}
				}
				list = append(list, n)
			}
			return nil
		})
	})
	if serviceError(w, err) {
		return
	}
	response(w, 200, list)
}
func main() {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "./data/relief.db"
	}
	static := os.Getenv("STATIC_DIR")
	if static == "" {
		static = "."
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	a, err := openApp(path, static)
	if err != nil {
		log.Fatal(err)
	}
	defer a.db.Close()
	fmt.Printf("listening on %s, database %s\n", addr, path)
	server := &http.Server{Addr: addr, Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
