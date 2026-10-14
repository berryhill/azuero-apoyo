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
	CoordinatorPhone string    `json:"coordinatorPhone,omitempty"`
	Contact          string    `json:"contact,omitempty"`
	Phone            string    `json:"phone,omitempty"`
	Adults           string    `json:"adults,omitempty"`
	OlderAdults      string    `json:"olderAdults,omitempty"`
	Disabled         string    `json:"disabled,omitempty"`
	SecondContact    string    `json:"secondContact,omitempty"`
	SecondPhone      string    `json:"secondPhone,omitempty"`
	ThirdContact     string    `json:"thirdContact,omitempty"`
	ThirdPhone       string    `json:"thirdPhone,omitempty"`
	Generator        string    `json:"generator,omitempty"`
	Code             string    `json:"code"`
	Area             string    `json:"area"`
	Place            string    `json:"place"`
	Needs            string    `json:"needs"`
	People           string    `json:"people"`
	Coords           []float64 `json:"coords"`
	Created          string    `json:"created"`
	Status           string    `json:"status"`
	Decided          string    `json:"decided,omitempty"`
	Note             string    `json:"note,omitempty"`
	CampStatus       string    `json:"campStatus,omitempty"`
	HighestUrgency   string    `json:"highestUrgency,omitempty"`
	NeedStatuses     []string  `json:"needStatuses,omitempty"`
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
	for field, v := range map[string]string{"people": p.People, "adults": p.Adults, "children": p.Children, "babies": p.Babies, "olderAdults": p.OlderAdults, "disabled": p.Disabled} {
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
	// Cuándo cambió el pedido (texto) después de crearse, por ejemplo al fusionar un duplicado; se muestra como «Actualizado».
	RevisedAt string `json:"revisedAt,omitempty"`
	// Lista de empaque: lo que corredores y agentes preparan para cumplir la necesidad. Privada (no sale en público).
	Packing []PackingLine `json:"packing,omitempty"`
	// Fotos y videos de partida y entrega; privados (corredor asignado y administración).
	Media []MediaItem `json:"media,omitempty"`
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
	attempts       map[string]attempt
	// Códigos de acceso incorrectos por IP de quien visita (ver signInWithCode).
	codeAttempts map[string]attempt
	// Última vez que cada usuario hizo algo con su sesión (en memoria; se guarda en el usuario cada 2 minutos).
	lastSeen map[string]time.Time
	// Última ubicación en vivo de cada corredor que la comparte (solo memoria; ver location.go).
	locations map[string]liveLocation
	static    string
	router    *router
	media     mediaStore
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
func publicPoint(p Point) Point {
	p.Needs, _ = splitRepair(p.Needs)
	p.AddedBy, p.AddedByName = "", ""
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
	a := &App{db: db, static: static, sessions: map[string]session{}, runnerSessions: map[string]runnerSession{}, builderSessions: map[string]runnerSession{}, viewerSessions: map[string]runnerSession{}, attempts: map[string]attempt{}, codeAttempts: map[string]attempt{}, lastSeen: map[string]time.Time{}, locations: map[string]liveLocation{}, router: newRouter(), media: newMediaStore(path)}
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
	if err := db.Update(migrateUsers); err != nil {
		return nil, err
	}
	if err := db.Update(seedPackingItems); err != nil {
		return nil, err
	}
	if err := db.Update(splitLegacyRepairs); err != nil {
		return nil, err
	}
	if err := db.Update(migrateNeedUrgency); err != nil {
		return nil, err
	}
	return a, nil
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
	n.UpdatedAt = nowStamp()
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
		n := CampNeed{ID: randomToken(), CampID: p.ID, Need: repairLabel, Kind: "repair", Status: "OPEN", Urgency: urgencyNormal, CreatedAt: nowStamp()}
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
		if foldText(part) == foldText(repairLabel) {
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
			n.Kind, n.Need, n.RunnerID = "repair", repairLabel, ""
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
func validKind(s string) bool { return s == "" || s == "supply" || s == "repair" }
func storedKind(s string) string {
	if s == "repair" {
		return "repair"
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

// campNeedStatuses lists which of OPEN, IN_PROGRESS and COMPLETED a camp's saved needs include.
func campNeedStatuses(tx *bolt.Tx, id string) ([]string, error) {
	seen := map[string]bool{}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID == id && !isRepair(n) {
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
		if n.CampID != id || n.Status == "DELETED" || isRepair(n) {
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

// publicPoints: puntos aprobados para el público; reveal (quien entró con código) conserva en «necesidades»
// los artículos de las carreras reservadas.
func (a *App) publicPoints(w http.ResponseWriter, reveal bool) {
	points := []Point{}
	err := a.db.View(func(tx *bolt.Tx) error {
		reserved, e := reservedNeedItems(tx)
		if e != nil {
			return e
		}
		return tx.Bucket(submissions).ForEach(func(k, v []byte) error {
			var p Point
			if e := json.Unmarshal(v, &p); e != nil {
				return e
			}
			if p.Status == "approved" {
				status, urgency, e := aggregateCamp(tx, p.ID)
				if e != nil {
					return e
				}
				p.CampStatus = status
				p.HighestUrgency = urgency
				if p.NeedStatuses, e = campNeedStatuses(tx, p.ID); e != nil {
					return e
				}
				if !reveal {
					p.Needs = withoutReservedItems(p.Needs, reserved[p.ID])
				}
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
	mux.HandleFunc("GET /api/points", func(w http.ResponseWriter, r *http.Request) { a.publicPoints(w, false) })
	// reviewed: administración (campos internos, código propio). publish: queda aprobado al guardar
	// (administración, o un usuario que entró con su código de acceso).
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
		p.History, p.Checkups = nil, nil
		p.AddedBy, p.AddedByName = "", ""
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
			p.Source = "user"
			if reviewed {
				p.Source = "admin"
			}
			p.Status = "approved"
			p.Decided, p.LastCheckup = p.Created, p.Created
		}
		data, _ := json.Marshal(p)
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(submissions)
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
			if err.Error() == "duplicate code" {
				failField(w, 409, "code", "Este código ya existe. Usa otro o deja el campo vacío para asignar uno automáticamente.")
				return
			}
			fail(w, 500, "database error")
			return
		}
		response(w, 201, map[string]string{"id": p.ID, "status": p.Status})
	}
	mux.HandleFunc("POST /api/submissions", func(w http.ResponseWriter, r *http.Request) { createPoint(w, r, false, false, nil) })
	mux.HandleFunc("POST /api/admin/points", func(w http.ResponseWriter, r *http.Request) { createPoint(w, r, true, true, nil) })
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
			if (p.Status == "pending" && in.Decision != "approved" && in.Decision != "rejected") ||
				(p.Status == "approved" && in.Decision != "archived") ||
				(p.Status == "archived" && in.Decision != "approved") ||
				(p.Status == "rejected" && in.Decision != "approved") ||
				(p.Status != "pending" && p.Status != "approved" && p.Status != "archived" && p.Status != "rejected") {
				return errors.New("already decided")
			}
			// Rejected points were never approved, so approving one creates its need like a pending approval.
			firstApproval := (p.Status == "pending" || p.Status == "rejected") && in.Decision == "approved"
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
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name := ""
		switch r.URL.Path {
		case "/", "/index.html":
			name = "index.html"
		case "/admin.html", "/camp.html", "/runner.html", "/builder.html", "/pointform.js", "/leaflet.css", "/leaflet.js", "/report/boletin_general.html", "/report/logo.png":
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
	needFields := map[string]int{"need": 8192, "status": 32, "urgency": 32, "runnerId": 128, "kind": 16, "builderId": 128}
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
		a.listNeeds(w, r.PathValue("id"), false, false)
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
	mux.HandleFunc("GET /api/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) { a.listNeeds(w, r.PathValue("id"), true, false) })
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
			return putNeed(tx.Bucket(needsBucket), &n)
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
			runnerBefore := n.RunnerID
			if _, yes := f["need"]; yes {
				if text := fieldText(f, "need"); text != n.Need {
					n.Need, n.RevisedAt = text, nowStamp()
				}
			}
			if _, yes := f["status"]; yes {
				status := fieldText(f, "status")
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
			if n.Need == "" || !validStatus(n.Status) || !validUrgency(n.Urgency) || !validAssignee(tx, n) {
				return errInvalid
			}
			enRouteOnAssign(&n, runnerBefore)
			return putNeed(b, &n)
		})
		if serviceError(w, err) {
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
			if p.Status != "approved" {
				return errMissing
			}
			if e := applyCampFields(&p, f); e != nil {
				return e
			}
			p.LastCheckup = nowStamp()
			return putJSON(b, id, p)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, p)
	})
}

// Campos editables de un punto aprobado y su largo máximo (0 = sin límite de texto o de solo lectura).
var campFields = map[string]int{"code": 45, "area": 80, "place": 200, "needs": 8192, "people": 200, "note": noteLimit, "source": 0, "status": 0, "id": 0, "created": 0, "decided": 0, "campStatus": 0, "highestUrgency": 0, "needStatuses": 0, "service": 200, "serviceNotes": 8192, "children": 100, "babies": 100, "notes": 8192, "directions": 8192, "extra": 8192, "gps": 1000, "coords": 0, "row": 100, "ordinal": 100, "coordinator": 200, "contact": 200, "phone": 40, "adults": 100, "olderAdults": 100, "disabled": 100, "secondContact": 200, "secondPhone": 40, "thirdContact": 200, "thirdPhone": 40, "generator": 200, "members": 0, "coordinatorPhone": 40, "mergedInto": 0, "lastCheckup": 0, "history": 0, "checkups": 0}

// campFieldError rechaza campos de solo lectura y coordenadas fuera de Panamá.
func campFieldError(f map[string]json.RawMessage) (string, string) {
	for _, key := range []string{"source", "status", "id", "created", "decided", "campStatus", "highestUrgency", "needStatuses", "mergedInto", "lastCheckup", "history", "checkups"} {
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
			p.LastCheckup = nowStamp()
			p.Checkups = append(p.Checkups, p.LastCheckup)
			if len(p.Checkups) > checkupLimit {
				p.Checkups = p.Checkups[len(p.Checkups)-checkupLimit:]
			}
			return putJSON(b, p.ID, p)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, p)
	})
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
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Area        string    `json:"area"`
	Place       string    `json:"place,omitempty"`
	People      string    `json:"people,omitempty"`
	Directions  string    `json:"directions,omitempty"`
	Coordinator string    `json:"coordinator,omitempty"`
	GPS         string    `json:"gps,omitempty"`
	Coords      []float64 `json:"coords,omitempty"`
	// Contactos del punto para que el corredor llame antes de tomar la carrera; nunca en páginas públicas.
	Contacts []runnerContact `json:"contacts,omitempty"`
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
	ID             string        `json:"id"`
	Need           string        `json:"need"`
	Status         string        `json:"status"`
	Urgency        string        `json:"urgency"`
	RequesterName  string        `json:"requesterName,omitempty"`
	RequesterPhone string        `json:"requesterPhone,omitempty"`
	CreatedAt      string        `json:"createdAt,omitempty"`
	RevisedAt      string        `json:"revisedAt,omitempty"`
	EnRouteAt      string        `json:"enRouteAt,omitempty"`
	Packing        []PackingLine `json:"packing,omitempty"`
	Media          []MediaItem   `json:"media,omitempty"`
	// assigned: la necesidad está asignada a quien consulta (puede armar su lista de empaque).
	Assigned bool       `json:"assigned,omitempty"`
	Camp     runnerCamp `json:"camp"`
}

var errBusy = errors.New("Ya tienes 10 carreras en curso. Completa o deja alguna antes de tomar otra.")
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
	if errors.Is(err, c.busy) || errors.Is(err, c.taken) || errors.Is(err, c.notYours) {
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
		if len(panels) == 0 {
			failField(w, 403, "accessCode", "Tu código no tiene acceso a ningún panel.")
			return
		}
		a.startViewerSession(w, r, item.ID)
		response(w, 200, map[string]any{"name": item.Name, "telephone": item.Telephone, "roles": item.Roles, "panels": panels})
	})
	// Vista completa: los mismos datos públicos, pero con el detalle de las carreras reservadas.
	mux.HandleFunc("GET /api/full/points", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); ok {
			a.publicPoints(w, true)
		}
	})
	mux.HandleFunc("GET /api/full/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); ok {
			a.listNeeds(w, r.PathValue("id"), true, true)
		}
	})
	mux.HandleFunc("POST /api/full/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("viewer_session"); e == nil {
			a.mu.Lock()
			if s, ok := a.viewerSessions[c.Value]; ok {
				delete(a.locations, s.runnerID)
			}
			delete(a.viewerSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "viewer_session", Value: "", Path: "/api/full", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	a.packingRoutes(mux, runners)
	a.routingRoutes(mux, runners)
	a.mediaRoutes(mux, runners)
	a.locationRoutes(mux, runners)
	mux.HandleFunc("POST /api/runner/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.crewSession(runners, w, r); !ok {
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
			return putNeed(tx.Bucket(needsBucket), &n)
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
			response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "telephone": item.Telephone, "csrf": s.csrf, "roles": item.Roles, "agent": s.agent})
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
		response(w, 200, map[string]string{"id": item.ID, "name": item.Name, "csrf": s.csrf})
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
		response(w, 200, map[string]any{"id": item.ID, "name": item.Name, "csrf": s.csrf, "agent": s.agent})
	})
	mux.HandleFunc("POST "+c.path+"/logout", func(w http.ResponseWriter, r *http.Request) {
		if cookie, e := r.Cookie(c.cookie); e == nil {
			a.mu.Lock()
			if s, ok := c.sessions[cookie.Value]; ok && !c.repair {
				delete(a.locations, s.runnerID)
			}
			delete(c.sessions, cookie.Value)
			a.mu.Unlock()
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
				if !ok || isRepair(n) != c.repair {
					return nil
				}
				if !c.repair {
					if n.Need, _ = splitRepair(n.Need); n.Need == "" {
						return nil
					}
				}
				who := *c.assignee(&n)
				item := runnerNeed{ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, Packing: n.Packing, Assigned: who == s.runnerID, Camp: camp}
				if item.Assigned {
					item.Media = n.Media
				}
				if s.agent {
					// Los agentes ven todo lo abierto o en ruta para armar listas de empaque.
					if n.Status == "OPEN" || n.Status == "IN_PROGRESS" {
						available = append(available, item)
					}
				} else if who == s.runnerID && n.Status == "IN_PROGRESS" {
					mine = append(mine, item)
				} else if n.Status == "OPEN" && (who == "" || who == s.runnerID) {
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
			if *c.assignee(n) != s.runnerID || n.Status != "IN_PROGRESS" {
				return c.notYours
			}
			n.Status, n.CompletedAt = "COMPLETED", nowStamp()
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
			reserved, e := reservedNeedItems(tx)
			if e != nil {
				return e
			}
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status != "approved" {
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

// listNeeds: public quita datos privados; reveal (quien entró con código) además muestra el detalle de las carreras reservadas.
func (a *App) listNeeds(w http.ResponseWriter, id string, public, reveal bool) {
	list := []CampNeed{}
	err := a.db.View(func(tx *bolt.Tx) error {
		if !a.campExists(tx, id) {
			return errMissing
		}
		return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
			var n CampNeed
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if n.CampID == id && n.Status != "DELETED" && !(public && isRepair(n)) {
				if public {
					reserved := reservedNeed(tx, n)
					n.RunnerID, n.ReleasedBy, n.Packing, n.Media = "", "", nil, nil
					if n.Need, _ = splitRepair(n.Need); n.Need == "" {
						return nil
					}
					if reserved && !reveal {
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
