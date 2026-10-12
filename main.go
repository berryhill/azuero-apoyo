package main

import (
	"crypto/rand"
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
	"strconv"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

type Point struct {
	ID            string    `json:"id"`
	Source        string    `json:"source,omitempty"`
	Service       string    `json:"service,omitempty"`
	ServiceNotes  string    `json:"serviceNotes,omitempty"`
	Children      string    `json:"children,omitempty"`
	Babies        string    `json:"babies,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	Directions    string    `json:"directions,omitempty"`
	Extra         string    `json:"extra,omitempty"`
	GPS           string    `json:"gps,omitempty"`
	Row           string    `json:"row,omitempty"`
	Ordinal       string    `json:"ordinal,omitempty"`
	Coordinator   string    `json:"coordinator,omitempty"`
	Contact       string    `json:"contact,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Adults        string    `json:"adults,omitempty"`
	OlderAdults   string    `json:"olderAdults,omitempty"`
	Disabled      string    `json:"disabled,omitempty"`
	SecondContact string    `json:"secondContact,omitempty"`
	SecondPhone   string    `json:"secondPhone,omitempty"`
	ThirdContact  string    `json:"thirdContact,omitempty"`
	ThirdPhone    string    `json:"thirdPhone,omitempty"`
	Generator     string    `json:"generator,omitempty"`
	Code          string    `json:"code"`
	Area          string    `json:"area"`
	Place         string    `json:"place"`
	Needs         string    `json:"needs"`
	People        string    `json:"people"`
	Coords        []float64 `json:"coords"`
	Created       string    `json:"created"`
	Status        string    `json:"status"`
	Decided       string    `json:"decided,omitempty"`
}
type session struct {
	csrf    string
	expires time.Time
}
type attempt struct {
	count int
	until time.Time
}
type App struct {
	db       *bolt.DB
	mu       sync.Mutex
	sessions map[string]session
	attempts map[string]attempt
	static   string
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
	a := &App{db: db, static: static, sessions: map[string]session{}, attempts: map[string]attempt{}}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(submissions); e != nil {
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
			return b.Put([]byte(p.ID), data)
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
		if blocked {
			fail(w, 429, "try later")
			return
		}
		var in struct {
			Password string `json:"password"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			fail(w, 400, "invalid JSON")
			return
		}
		if !a.passwordValid(in.Password) {
			a.mu.Lock()
			at = a.attempts[ip]
			if time.Now().After(at.until) {
				at.count = 0
			}
			at.count++
			at.until = time.Now().Add(10 * time.Minute)
			a.attempts[ip] = at
			a.mu.Unlock()
			fail(w, 401, "invalid password")
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
			p.Status = in.Decision
			p.Decided = time.Now().UTC().Format(time.RFC3339)
			data, _ := json.Marshal(p)
			return b.Put([]byte(in.ID), data)
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
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name := ""
		switch r.URL.Path {
		case "/", "/index.html":
			name = "index.html"
		case "/admin.html", "/leaflet.css", "/leaflet.js":
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
