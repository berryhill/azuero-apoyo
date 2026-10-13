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
}
type CampNeed struct {
	ID       string `json:"id"`
	CampID   string `json:"campId"`
	Need     string `json:"need"`
	Status   string `json:"status"`
	Urgency  string `json:"urgency"`
	RunnerID string `json:"runnerId,omitempty"`
	// Quien reportó la necesidad desde la página pública; visible para todos por decisión del equipo.
	RequesterName  string `json:"requesterName,omitempty"`
	RequesterPhone string `json:"requesterPhone,omitempty"`
	// Marcas de tiempo UTC fijadas por el servidor; el navegador las muestra en hora local.
	CreatedAt   string `json:"createdAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
	ReleasedAt  string `json:"releasedAt,omitempty"`
	ReleasedBy  string `json:"releasedBy,omitempty"`
}

func nowStamp() string { return time.Now().UTC().Format(time.RFC3339) }

var runnersBucket = []byte("runners")
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
	attempts       map[string]attempt
	static         string
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
	p.Members = nil
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
	a := &App{db: db, static: static, sessions: map[string]session{}, runnerSessions: map[string]runnerSession{}, attempts: map[string]attempt{}}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(submissions); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(runnersBucket); e != nil {
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
	r.Body = http.MaxBytesReader(w, r.Body, 32768)
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
	for key, v := range fields {
		limit, ok := allowed[key]
		if !ok {
			failField(w, 400, key, "unknown field")
			return nil, false
		}
		if limit > 0 {
			var text string
			if json.Unmarshal(v, &text) != nil || len(strings.TrimSpace(text)) > limit {
				failField(w, 400, key, "invalid or too long")
				return nil, false
			}
		}
	}
	return fields, true
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
func validUrgency(s string) bool {
	return s == "Today" || s == "Urgent" || s == "Tomorrow" || s == "Eventually"
}
func putJSON(b *bolt.Bucket, key string, value any) error {
	data, e := json.Marshal(value)
	if e != nil {
		return e
	}
	return b.Put([]byte(key), data)
}

// needFromPoint convierte el texto «Necesidades» de un punto recién aprobado en una necesidad abierta.
// Los contactos del punto son privados, por eso no se copian como solicitante.
func needFromPoint(tx *bolt.Tx, p Point) error {
	if strings.TrimSpace(p.Needs) == "" {
		return nil
	}
	n := CampNeed{ID: randomToken(), CampID: p.ID, Need: strings.TrimSpace(p.Needs), Status: "OPEN", Urgency: "Eventually", CreatedAt: nowStamp()}
	return putJSON(tx.Bucket(needsBucket), n.ID, n)
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
	if r.Method == http.MethodPatch {
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
		if n.CampID == id {
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
	rank := map[string]int{"Eventually": 1, "Tomorrow": 2, "Urgent": 3, "Today": 4}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if n.CampID != id || n.Status == "DELETED" {
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
			if rank[n.Urgency] > rank[urgency] {
				urgency = n.Urgency
			}
		}
		return nil
	})
	return status, urgency, err
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { response(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /api/points", func(w http.ResponseWriter, r *http.Request) {
		points := []Point{}
		err := a.db.View(func(tx *bolt.Tx) error {
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
					points = append(points, publicPoint(p))
				} else if p.Status == "archived" && p.Source == "sheet" {
					// Hide bundled snapshot rows that were archived after deployment.
					points = append(points, Point{Code: p.Code, Source: "sheet", Status: "archived"})
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
	createPoint := func(w http.ResponseWriter, r *http.Request, reviewed bool) {
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
		p.Decided = ""
		if reviewed {
			p.Source = "admin"
			p.Status = "approved"
			p.Decided = p.Created
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
			if e := b.Put([]byte(p.ID), data); e != nil || !reviewed {
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
	mux.HandleFunc("POST /api/submissions", func(w http.ResponseWriter, r *http.Request) { createPoint(w, r, false) })
	mux.HandleFunc("POST /api/admin/points", func(w http.ResponseWriter, r *http.Request) { createPoint(w, r, true) })
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
		http.SetCookie(w, &http.Cookie{Name: "relief_session", Value: token, Path: "/api/admin", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: s.expires})
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
			if (p.Status == "pending" && in.Decision != "approved" && in.Decision != "rejected") ||
				(p.Status == "approved" && in.Decision != "archived") ||
				(p.Status == "archived" && in.Decision != "approved") ||
				(p.Status != "pending" && p.Status != "approved" && p.Status != "archived") {
				return errors.New("already decided")
			}
			firstApproval := p.Status == "pending" && in.Decision == "approved"
			p.Status = in.Decision
			p.Decided = time.Now().UTC().Format(time.RFC3339)
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
		var in struct {
			Password string `json:"password"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.Password) < 12 || len(in.Password) > 72 {
			fail(w, 400, "password must be 12–72 characters")
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
	a.boletinRoutes(mux)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name := ""
		switch r.URL.Path {
		case "/", "/index.html":
			name = "index.html"
		case "/admin.html", "/camp.html", "/runner.html", "/pointform.js", "/leaflet.css", "/leaflet.js", "/report/boletin_general.html", "/report/logo.png":
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
	needFields := map[string]int{"need": 8192, "status": 32, "urgency": 32, "runnerId": 128}
	mux.HandleFunc("GET /api/admin/runners", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		list := []Runner{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(runnersBucket).ForEach(func(_, v []byte) error {
				var item Runner
				if e := json.Unmarshal(v, &item); e != nil {
					return e
				}
				list = append(list, item)
				return nil
			})
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/admin/runners", func(w http.ResponseWriter, r *http.Request) {
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
		item := Runner{ID: randomToken(), Name: name, Telephone: telephone}
		if a.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(runnersBucket), item.ID, item) }) != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 201, item)
	})
	mux.HandleFunc("PATCH /api/admin/runners/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, runnerFields)
		if !ok {
			return
		}
		id := r.PathValue("id")
		var item Runner
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(runnersBucket)
			v := b.Get([]byte(id))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &item); e != nil {
				return e
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
			return putJSON(b, id, item)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, item)
	})
	mux.HandleFunc("GET /api/admin/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		a.listNeeds(w, r.PathValue("id"), false)
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
	mux.HandleFunc("GET /api/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) { a.listNeeds(w, r.PathValue("id"), true) })
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
		if n.Urgency == "" {
			n.Urgency = "Eventually"
		}
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
			return putJSON(tx.Bucket(needsBucket), n.ID, n)
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
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: fieldText(f, "urgency"), RunnerID: fieldText(f, "runnerId"), CreatedAt: nowStamp()}
		if _, yes := f["status"]; yes {
			n.Status = fieldText(f, "status")
		}
		if n.Urgency == "" {
			n.Urgency = "Eventually"
		}
		if n.Need == "" || !validStatus(n.Status) || !validUrgency(n.Urgency) {
			fail(w, 400, "invalid need, status or urgency")
			return
		}
		if n.Status == "COMPLETED" {
			n.CompletedAt = n.CreatedAt
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			if !a.campExists(tx, n.CampID) {
				return errMissing
			}
			if n.RunnerID != "" && tx.Bucket(runnersBucket).Get([]byte(n.RunnerID)) == nil {
				return errInvalid
			}
			return putJSON(tx.Bucket(needsBucket), n.ID, n)
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
			if _, yes := f["need"]; yes {
				n.Need = fieldText(f, "need")
			}
			if _, yes := f["status"]; yes {
				status := fieldText(f, "status")
				if status == "COMPLETED" && n.Status != "COMPLETED" {
					n.CompletedAt = nowStamp()
				} else if status != "COMPLETED" {
					n.CompletedAt = ""
				}
				n.Status = status
			}
			if _, yes := f["urgency"]; yes {
				n.Urgency = fieldText(f, "urgency")
			}
			if _, yes := f["runnerId"]; yes {
				n.RunnerID = fieldText(f, "runnerId")
			}
			if n.Need == "" || !validStatus(n.Status) || !validUrgency(n.Urgency) || n.RunnerID != "" && tx.Bucket(runnersBucket).Get([]byte(n.RunnerID)) == nil {
				return errInvalid
			}
			return putJSON(b, id, n)
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
		allowed := map[string]int{"code": 45, "area": 80, "place": 200, "needs": 8192, "people": 200, "note": 8192, "source": 0, "status": 0, "id": 0, "created": 0, "decided": 0, "campStatus": 0, "highestUrgency": 0, "needStatuses": 0, "service": 200, "serviceNotes": 8192, "children": 100, "babies": 100, "notes": 8192, "directions": 8192, "extra": 8192, "gps": 1000, "coords": 0, "row": 100, "ordinal": 100, "coordinator": 200, "contact": 200, "phone": 40, "adults": 100, "olderAdults": 100, "disabled": 100, "secondContact": 200, "secondPhone": 40, "thirdContact": 200, "thirdPhone": 40, "generator": 200, "members": 0, "coordinatorPhone": 40}
		f, ok := decodeFields(w, r, allowed)
		if !ok {
			return
		}
		for _, key := range []string{"source", "status", "id", "created", "decided", "campStatus", "highestUrgency", "needStatuses"} {
			if _, yes := f[key]; yes {
				failField(w, 400, key, "read-only field")
				return
			}
		}
		var coords []float64
		if raw, yes := f["coords"]; yes {
			if json.Unmarshal(raw, &coords) != nil || len(coords) != 0 && (len(coords) != 2 || coords[0] < 6 || coords[0] > 10 || coords[1] < -83 || coords[1] > -78) {
				failField(w, 400, "coords", "invalid coordinates")
				return
			}
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
			if e = json.Unmarshal(data, &p); e != nil {
				return errInvalid
			}
			p.Code = strings.TrimSpace(p.Code)
			p.Area = strings.TrimSpace(p.Area)
			if p.Code == "" || p.Area == "" || normalizeMembers(&p) != nil {
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
			return putJSON(b, id, p)
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
	ID             string     `json:"id"`
	Need           string     `json:"need"`
	Status         string     `json:"status"`
	Urgency        string     `json:"urgency"`
	RequesterName  string     `json:"requesterName,omitempty"`
	RequesterPhone string     `json:"requesterPhone,omitempty"`
	CreatedAt      string     `json:"createdAt,omitempty"`
	Camp           runnerCamp `json:"camp"`
}

var errBusy = errors.New("Ya tienes una carrera en curso. Complétala o déjala antes de tomar otra.")
var errTaken = errors.New("Esta necesidad ya no está disponible. Otra persona la tomó o cambió de estado.")
var errNotYours = errors.New("Esta carrera ya no está asignada a ti.")

func runnerNameKey(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}
func runnerError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errBusy) || errors.Is(err, errTaken) || errors.Is(err, errNotYours) {
		fail(w, 409, err.Error())
		return true
	}
	return serviceError(w, err)
}
func (a *App) runner(w http.ResponseWriter, r *http.Request) (runnerSession, bool) {
	c, e := r.Cookie("runner_session")
	if e != nil {
		fail(w, 401, "Escribe tu nombre para continuar.")
		return runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := a.runnerSessions[c.Value]
	a.mu.Unlock()
	if !ok || time.Now().After(s.expires) {
		fail(w, 401, "Escribe tu nombre para continuar.")
		return runnerSession{}, false
	}
	if r.Method == http.MethodPost && r.Header.Get("X-CSRF-Token") != s.csrf {
		fail(w, 403, "CSRF token required")
		return runnerSession{}, false
	}
	return s, true
}
func (a *App) runnerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/runner/session", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"name": 120, "telephone": 40})
		if !ok {
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
		var item Runner
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(runnersBucket)
			key := runnerNameKey(name)
			if e := b.ForEach(func(_, v []byte) error {
				var existing Runner
				if e := json.Unmarshal(v, &existing); e != nil {
					return e
				}
				if item.ID == "" && runnerNameKey(existing.Name) == key {
					item = existing
				}
				return nil
			}); e != nil {
				return e
			}
			// Un teléfono guardado por curación no se sobrescribe; solo se completa si falta.
			if item.ID != "" {
				if item.Telephone != "" {
					return nil
				}
				item.Telephone = telephone
				return putJSON(b, item.ID, item)
			}
			item = Runner{ID: randomToken(), Name: name, Telephone: telephone}
			return putJSON(b, item.ID, item)
		})
		if serviceError(w, err) {
			return
		}
		token := randomToken()
		s := runnerSession{runnerID: item.ID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour)}
		a.mu.Lock()
		a.runnerSessions[token] = s
		a.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "runner_session", Value: token, Path: "/api/runner", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: s.expires})
		response(w, 200, map[string]string{"name": item.Name, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/runner/session", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.runner(w, r)
		if !ok {
			return
		}
		var item Runner
		err := a.db.View(func(tx *bolt.Tx) error {
			v := tx.Bucket(runnersBucket).Get([]byte(s.runnerID))
			if v == nil {
				return errMissing
			}
			return json.Unmarshal(v, &item)
		})
		if errors.Is(err, errMissing) {
			fail(w, 401, "Escribe tu nombre para continuar.")
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]string{"name": item.Name, "csrf": s.csrf})
	})
	mux.HandleFunc("POST /api/runner/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("runner_session"); e == nil {
			a.mu.Lock()
			delete(a.runnerSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "runner_session", Value: "", Path: "/api/runner", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
	mux.HandleFunc("GET /api/runner/needs", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.runner(w, r)
		if !ok {
			return
		}
		var current *runnerNeed
		available := []runnerNeed{}
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
				if !ok {
					return nil
				}
				item := runnerNeed{ID: n.ID, Need: n.Need, Status: n.Status, Urgency: n.Urgency, RequesterName: n.RequesterName, RequesterPhone: n.RequesterPhone, CreatedAt: n.CreatedAt, Camp: camp}
				if n.RunnerID == s.runnerID && n.Status == "IN_PROGRESS" {
					current = &item
				} else if n.Status == "OPEN" && (n.RunnerID == "" || n.RunnerID == s.runnerID) {
					available = append(available, item)
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		rank := map[string]int{"Today": 0, "Urgent": 1, "Tomorrow": 2, "Eventually": 3}
		sort.SliceStable(available, func(i, j int) bool {
			if rank[available[i].Urgency] != rank[available[j].Urgency] {
				return rank[available[i].Urgency] < rank[available[j].Urgency]
			}
			return available[i].Camp.Code < available[j].Camp.Code
		})
		response(w, 200, map[string]any{"current": current, "available": available})
	})
	// change aplica una transición solo si la necesidad sigue en el estado esperado.
	change := func(w http.ResponseWriter, r *http.Request, apply func(tx *bolt.Tx, s runnerSession, n *CampNeed) error) {
		s, ok := a.runner(w, r)
		if !ok || !a.guard(w, r) {
			return
		}
		var n CampNeed
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errTaken
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !a.campExists(tx, n.CampID) {
				return errTaken
			}
			if e := apply(tx, s, &n); e != nil {
				return e
			}
			return putJSON(b, n.ID, n)
		})
		if runnerError(w, err) {
			return
		}
		response(w, 200, map[string]string{"id": n.ID, "status": n.Status})
	}
	mux.HandleFunc("POST /api/runner/needs/{id}/take", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(tx *bolt.Tx, s runnerSession, n *CampNeed) error {
			if n.Status != "OPEN" || n.RunnerID != "" && n.RunnerID != s.runnerID {
				return errTaken
			}
			busy := false
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var other CampNeed
				if e := json.Unmarshal(v, &other); e != nil {
					return e
				}
				busy = busy || other.RunnerID == s.runnerID && other.Status == "IN_PROGRESS"
				return nil
			}); e != nil {
				return e
			}
			if busy {
				return errBusy
			}
			n.RunnerID, n.Status = s.runnerID, "IN_PROGRESS"
			return nil
		})
	})
	mux.HandleFunc("POST /api/runner/needs/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(_ *bolt.Tx, s runnerSession, n *CampNeed) error {
			if n.RunnerID != s.runnerID || n.Status != "IN_PROGRESS" {
				return errNotYours
			}
			n.Status, n.CompletedAt = "COMPLETED", nowStamp()
			return nil
		})
	})
	mux.HandleFunc("POST /api/runner/needs/{id}/release", func(w http.ResponseWriter, r *http.Request) {
		change(w, r, func(_ *bolt.Tx, s runnerSession, n *CampNeed) error {
			if n.RunnerID != s.runnerID || n.Status != "IN_PROGRESS" {
				return errNotYours
			}
			n.RunnerID, n.Status, n.ReleasedAt, n.ReleasedBy = "", "OPEN", nowStamp(), s.runnerID
			return nil
		})
	})
	mux.HandleFunc("POST /api/runner/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.runner(w, r); !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 32})
		if !ok {
			return
		}
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: fieldText(f, "urgency"), CreatedAt: nowStamp()}
		if n.Urgency == "" {
			n.Urgency = "Eventually"
		}
		if n.Need == "" || !validUrgency(n.Urgency) {
			fail(w, 400, "Describe la necesidad y elige una urgencia válida.")
			return
		}
		err := a.db.Update(func(tx *bolt.Tx) error {
			if !a.campExists(tx, n.CampID) {
				return errMissing
			}
			return putJSON(tx.Bucket(needsBucket), n.ID, n)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 201, map[string]string{"id": n.ID, "status": n.Status})
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
			if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if e := json.Unmarshal(v, &p); e != nil {
					return e
				}
				if p.Status != "approved" {
					return nil
				}
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
func (a *App) listNeeds(w http.ResponseWriter, id string, public bool) {
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
			if n.CampID == id && n.Status != "DELETED" {
				if public {
					n.RunnerID, n.ReleasedBy = "", ""
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
