package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Donaciones como boletas: alguien ofrece desde la página pública («Quiero donar») cuánto de qué puede dar y deja su
// contacto. La oferta queda «por confirmar» (potencial) hasta que coordinación habla con la persona y la confirma; solo
// entonces resta de lo que falta. Después se sigue la logística: a qué centro de acopio llega y cuándo, quién la
// transporta, a qué punto va (con su GPS) y cuándo sale hacia allá. El contacto del donante es privado (administración y
// agentes); lo público son solo los totales por artículo.
var donationsBucket = []byte("donations")

const (
	donationOffered    = "offered"    // por confirmar: aún no cuenta
	donationConfirmed  = "confirmed"  // comprometida
	donationReceived   = "received"   // llegó al centro de acopio
	donationDispatched = "dispatched" // va en camino al punto
	donationDelivered  = "delivered"  // entregada en el punto
	donationCancelled  = "cancelled"
)

var donationStatuses = []string{donationOffered, donationConfirmed, donationReceived, donationDispatched, donationDelivered, donationCancelled}

// donationKinds: artículos con unidad fija que se comparan con lo que falta en «¿Qué donar?». Materiales van como «mat:<nombre>»
// y lo demás como «other» con su propio nombre y unidad.
var donationKinds = map[string][2]string{
	"water":      {"Agua", "L"},
	"food":       {"Comida", "comidas"},
	"shelter":    {"Carpas", "carpas"},
	"sanitation": {"Baños o letrinas portátiles", "baños"},
	"baby":       {"Pañales", "pañales"},
	"fuel":       {"Combustible", "gal"},
}

type DonationItem struct {
	Key      string `json:"key"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

type DonationEvent struct {
	At     string `json:"at"`
	Status string `json:"status"`
	By     string `json:"by,omitempty"`
	Note   string `json:"note,omitempty"`
}

type Donation struct {
	ID         string         `json:"id"`
	Ref        string         `json:"ref"`
	Status     string         `json:"status"`
	DonorName  string         `json:"donorName"`
	DonorPhone string         `json:"donorPhone"`
	DonorEmail string         `json:"donorEmail,omitempty"`
	DonorWhen  string         `json:"donorWhen,omitempty"` // dónde y cuándo puede entregar, en palabras del donante
	DonorNote  string         `json:"donorNote,omitempty"`
	Items      []DonationItem `json:"items"`
	// Logística: centro de acopio y llegada prevista, transporte, punto de destino y salida prevista hacia allá.
	DropPlace        string          `json:"dropPlace,omitempty"`
	DropAt           string          `json:"dropAt,omitempty"`
	Transporter      string          `json:"transporter,omitempty"`
	TransporterPhone string          `json:"transporterPhone,omitempty"`
	CampID           string          `json:"campId,omitempty"`
	CampName         string          `json:"campName,omitempty"` // se arma al listar
	Destination      string          `json:"destination,omitempty"`
	Coords           []float64       `json:"coords,omitempty"`
	DispatchAt       string          `json:"dispatchAt,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	Source           string          `json:"source,omitempty"` // public | admin
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	History          []DonationEvent `json:"history,omitempty"`
}

// donationCounts: committed = confirmada, en acopio o en camino; lo entregado ya se ve cuando el punto actualiza sus necesidades.
func donationCounts(status string) (potential, committed, delivered bool) {
	switch status {
	case donationOffered:
		return true, false, false
	case donationConfirmed, donationReceived, donationDispatched:
		return false, true, false
	case donationDelivered:
		return false, false, true
	}
	return false, false, false
}

// ensureDonations: un respaldo importado de antes de las donaciones no trae el registro.
func ensureDonations(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(donationsBucket)
	return err
}

func validDonationStatus(s string) bool {
	for _, x := range donationStatuses {
		if x == s {
			return true
		}
	}
	return false
}

// parseDonationItems revisa la lista: claves conocidas, cantidades 1–1.000.000, hasta 20 artículos.
func parseDonationItems(raw json.RawMessage) ([]DonationItem, string) {
	var in []DonationItem
	if json.Unmarshal(raw, &in) != nil {
		return nil, "Lista de artículos inválida."
	}
	if len(in) == 0 {
		return nil, "Agrega al menos una cosa que puedas donar."
	}
	if len(in) > 20 {
		return nil, "Máximo 20 artículos por oferta."
	}
	out := []DonationItem{}
	for _, it := range in {
		it.Key, it.Item, it.Unit = strings.TrimSpace(it.Key), strings.Join(strings.Fields(it.Item), " "), strings.Join(strings.Fields(it.Unit), " ")
		if it.Quantity < 1 || it.Quantity > 1000000 {
			return nil, "Escribe una cantidad entre 1 y 1,000,000 para cada artículo."
		}
		if k, ok := donationKinds[it.Key]; ok {
			it.Item, it.Unit = k[0], k[1]
		} else if strings.HasPrefix(it.Key, "mat:") || it.Key == "other" {
			if it.Item == "" && strings.HasPrefix(it.Key, "mat:") {
				it.Item = strings.TrimSpace(strings.TrimPrefix(it.Key, "mat:"))
			}
			if it.Item == "" || len(it.Item) > 80 || len(it.Unit) > 24 {
				return nil, "Escribe qué vas a donar (hasta 80 letras) y su unidad."
			}
			if strings.HasPrefix(it.Key, "mat:") {
				it.Key = "mat:" + strings.ToLower(foldText(it.Item))
			}
			if it.Unit == "" {
				it.Unit = "unidades"
			}
		} else {
			return nil, "Artículo desconocido: " + it.Key
		}
		out = append(out, it)
	}
	return out, ""
}

var donationText = map[string]int{"name": 120, "phone": 40, "email": 120, "when": 600, "note": 1000, "items": 0}
var donationAdminText = map[string]int{"name": 120, "phone": 40, "email": 120, "when": 600, "note": 1000, "items": 0,
	"status": 16, "dropPlace": 200, "dropAt": 40, "transporter": 120, "transporterPhone": 40, "campId": 128,
	"destination": 500, "dispatchAt": 40, "notes": 4000, "event": 300}

func newDonationRef(tx *bolt.Tx) (string, error) {
	seq, err := tx.Bucket(donationsBucket).NextSequence()
	return fmt.Sprintf("D-%04d", seq), err
}

// applyDonationFields copia los campos enviados (solo los presentes) y valida contacto y destino.
func applyDonationFields(tx *bolt.Tx, d *Donation, f map[string]json.RawMessage) (string, string) {
	set := func(key string, dst *string) {
		if _, has := f[key]; has {
			*dst = fieldText(f, key)
		}
	}
	set("name", &d.DonorName)
	set("phone", &d.DonorPhone)
	set("email", &d.DonorEmail)
	set("when", &d.DonorWhen)
	set("note", &d.DonorNote)
	set("dropPlace", &d.DropPlace)
	set("transporter", &d.Transporter)
	set("transporterPhone", &d.TransporterPhone)
	set("notes", &d.Notes)
	d.DonorName = strings.Join(strings.Fields(d.DonorName), " ")
	if raw, has := f["items"]; has {
		items, msg := parseDonationItems(raw)
		if msg != "" {
			return "items", msg
		}
		d.Items = items
	}
	if d.DonorName == "" {
		return "name", "Escribe tu nombre."
	}
	if len(phoneDigits(d.DonorPhone)) < 7 {
		return "phone", "Escribe un teléfono válido (por ejemplo 6000-0000) para coordinar contigo."
	}
	if d.TransporterPhone != "" && len(phoneDigits(d.TransporterPhone)) < 7 {
		return "transporterPhone", "Teléfono del transporte incompleto."
	}
	if d.DonorEmail != "" && (!strings.Contains(d.DonorEmail, "@") || strings.ContainsAny(d.DonorEmail, " <>")) {
		return "email", "Correo inválido."
	}
	for _, k := range []struct {
		key string
		v   *string
	}{{"dropAt", &d.DropAt}, {"dispatchAt", &d.DispatchAt}} {
		if _, has := f[k.key]; has {
			*k.v = fieldText(f, k.key)
		}
		if *k.v != "" {
			if _, e := time.Parse("2006-01-02T15:04", *k.v); e != nil {
				if _, e := time.Parse("2006-01-02", *k.v); e != nil {
					return k.key, "Fecha inválida."
				}
			}
		}
	}
	if _, has := f["campId"]; has {
		d.CampID = fieldText(f, "campId")
		if d.CampID != "" {
			v := tx.Bucket(submissions).Get([]byte(d.CampID))
			var p Point
			if v == nil || json.Unmarshal(v, &p) != nil || p.Status != "approved" {
				return "campId", "Ese punto no existe o no está publicado."
			}
		}
	}
	if _, has := f["destination"]; has {
		d.Destination = fieldText(f, "destination")
		d.Coords = nil
		if d.Destination != "" {
			if c, err := parseMapLink(d.Destination); err == nil {
				d.Coords = c
			}
		}
	}
	return "", ""
}

func (a *App) donationRoutes(mux *http.ServeMux) {
	// Oferta pública: como las altas, 10 cada 10 minutos por conexión.
	mux.HandleFunc("POST /api/donations", func(w http.ResponseWriter, r *http.Request) {
		key, now := "donate|"+clientIP(r), time.Now()
		a.mu.Lock()
		at := a.codeAttempts[key]
		if now.After(at.until) {
			at = attempt{}
		}
		a.mu.Unlock()
		if at.count >= codeMaxWrong {
			fail(w, 429, "Demasiadas ofertas desde esta conexión. Espera unos minutos o llámanos.")
			return
		}
		f, ok := decodeFields(w, r, donationText)
		if !ok {
			return
		}
		if _, has := f["items"]; !has {
			failField(w, 400, "items", "Agrega al menos una cosa que puedas donar.")
			return
		}
		d := Donation{ID: randomToken(), Status: donationOffered, Source: "public", CreatedAt: nowStamp()}
		d.UpdatedAt = d.CreatedAt
		var field, msg string
		err := a.db.Update(func(tx *bolt.Tx) error {
			if field, msg = applyDonationFields(tx, &d, f); field != "" {
				return nil
			}
			ref, e := newDonationRef(tx)
			if e != nil {
				return e
			}
			d.Ref = ref
			d.History = []DonationEvent{{At: d.CreatedAt, Status: donationOffered, By: "Donante", Note: "Oferta desde la página pública"}}
			return putJSON(tx.Bucket(donationsBucket), d.ID, d)
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		if field != "" {
			failField(w, 400, field, msg)
			return
		}
		a.mu.Lock()
		at = a.codeAttempts[key]
		if now.After(at.until) {
			at = attempt{}
		}
		at.count++
		at.until = now.Add(codeWindow)
		a.codeAttempts[key] = at
		a.mu.Unlock()
		response(w, 201, map[string]string{"ref": d.Ref, "status": d.Status})
	})
	// Dos puertas a las mismas boletas: administración (/api/admin) y quien tiene el rol «Donaciones» con su código
	// (/api/desk, sesión desk_session). Los agentes no entran salvo que también tengan ese rol.
	for _, door := range []struct {
		prefix string
		who    func(http.ResponseWriter, *http.Request) (string, bool)
	}{
		{"/api/admin", func(w http.ResponseWriter, r *http.Request) (string, bool) {
			if r.Method == http.MethodGet {
				return "Administración", a.admin(w, r)
			}
			return "Administración", a.adminWrite(w, r)
		}},
		{"/api/desk", func(w http.ResponseWriter, r *http.Request) (string, bool) {
			u, _, ok := a.deskSession(w, r)
			return u.Name, ok
		}},
	} {
		a.donationDesk(mux, door.prefix, door.who)
	}
	mux.HandleFunc("POST /api/desk/session", func(w http.ResponseWriter, r *http.Request) {
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
		if !u.has(roleDonations) {
			failField(w, 403, "accessCode", "Tu código no tiene acceso a Donaciones. Pide a administración el rol «Donaciones».")
			return
		}
		s := a.startDeskSession(w, r, u.ID)
		response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "csrf": s.csrf})
	})
	mux.HandleFunc("GET /api/desk/session", func(w http.ResponseWriter, r *http.Request) {
		if u, s, ok := a.deskSession(w, r); ok {
			response(w, 200, map[string]any{"id": u.ID, "name": u.Name, "csrf": s.csrf, "roles": u.Roles, "panels": userPanels(u)})
		}
	})
	mux.HandleFunc("POST /api/desk/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("desk_session"); e == nil {
			a.mu.Lock()
			delete(a.deskSessions, c.Value)
			a.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: "desk_session", Value: "", Path: "/api/desk", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		response(w, 200, map[string]string{"status": "logged out"})
	})
}

func (a *App) startDeskSession(w http.ResponseWriter, r *http.Request, userID string) runnerSession {
	a.touch(userID)
	token := randomToken()
	s := runnerSession{runnerID: userID, csrf: randomToken(), expires: time.Now().Add(12 * time.Hour)}
	a.mu.Lock()
	a.deskSessions[token] = s
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "desk_session", Value: token, Path: "/api/desk", HttpOnly: true, Secure: secureCookie(r), SameSite: http.SameSiteStrictMode, Expires: s.expires})
	return s
}

// deskSession: sesión vigente de un usuario activo con el rol «Donaciones» y código; las escrituras llevan el token CSRF.
func (a *App) deskSession(w http.ResponseWriter, r *http.Request) (User, runnerSession, bool) {
	var u User
	c, e := r.Cookie("desk_session")
	if e != nil {
		fail(w, 401, "login required")
		return u, runnerSession{}, false
	}
	a.mu.Lock()
	s, ok := a.deskSessions[c.Value]
	a.mu.Unlock()
	active := false
	if ok && time.Now().Before(s.expires) {
		a.db.View(func(tx *bolt.Tx) error {
			var found bool
			u, found = getUser(tx, s.runnerID)
			active = found && !u.Disabled && u.has(roleDonations) && u.CodeHash != ""
			return nil
		})
	}
	if !active {
		fail(w, 401, "login required")
		return User{}, runnerSession{}, false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			fail(w, 415, "JSON required")
			return User{}, runnerSession{}, false
		}
		if r.Header.Get("X-CSRF-Token") != s.csrf {
			fail(w, 403, "CSRF token required")
			return User{}, runnerSession{}, false
		}
	}
	a.touch(u.ID)
	return u, s, true
}

func (a *App) donationDesk(mux *http.ServeMux, prefix string, who func(http.ResponseWriter, *http.Request) (string, bool)) {
	mux.HandleFunc("GET "+prefix+"/donations", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := who(w, r); !ok {
			return
		}
		list := []Donation{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(donationsBucket).ForEach(func(_, v []byte) error {
				var d Donation
				if e := json.Unmarshal(v, &d); e != nil {
					return e
				}
				if d.CampID != "" {
					var p Point
					if raw := tx.Bucket(submissions).Get([]byte(d.CampID)); raw != nil && json.Unmarshal(raw, &p) == nil {
						d.CampName = strings.Trim(strings.TrimSpace(p.Area)+" · "+strings.TrimSpace(p.Place), " ·")
						if len(d.Coords) == 0 {
							d.Coords = p.Coords
						}
					}
				}
				list = append(list, d)
				return nil
			})
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].CreatedAt > list[j].CreatedAt })
		response(w, 200, list)
	})
	// Administración o quien lleva Donaciones registra una oferta recibida por teléfono; puede quedar ya confirmada.
	mux.HandleFunc("POST "+prefix+"/donations", func(w http.ResponseWriter, r *http.Request) {
		by, ok := who(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, donationAdminText)
		if !ok {
			return
		}
		if _, has := f["items"]; !has {
			failField(w, 400, "items", "Agrega al menos un artículo.")
			return
		}
		d := Donation{ID: randomToken(), Status: donationOffered, Source: "admin", CreatedAt: nowStamp()}
		d.UpdatedAt = d.CreatedAt
		if s := fieldText(f, "status"); s != "" {
			if !validDonationStatus(s) {
				failField(w, 400, "status", "Estado inválido.")
				return
			}
			d.Status = s
		}
		var field, msg string
		err := a.db.Update(func(tx *bolt.Tx) error {
			if field, msg = applyDonationFields(tx, &d, f); field != "" {
				return nil
			}
			ref, e := newDonationRef(tx)
			if e != nil {
				return e
			}
			d.Ref = ref
			d.History = []DonationEvent{{At: d.CreatedAt, Status: d.Status, By: by, Note: fieldText(f, "event")}}
			return putJSON(tx.Bucket(donationsBucket), d.ID, d)
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		if field != "" {
			failField(w, 400, field, msg)
			return
		}
		response(w, 201, d)
	})
	mux.HandleFunc("PATCH "+prefix+"/donations/{id}", func(w http.ResponseWriter, r *http.Request) {
		by, ok := who(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, donationAdminText)
		if !ok {
			return
		}
		var d Donation
		var field, msg string
		status := 200
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(donationsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				status = 404
				return nil
			}
			if e := json.Unmarshal(v, &d); e != nil {
				return e
			}
			if field, msg = applyDonationFields(tx, &d, f); field != "" {
				return nil
			}
			now := nowStamp()
			if s := fieldText(f, "status"); s != "" && s != d.Status {
				if !validDonationStatus(s) {
					field, msg = "status", "Estado inválido."
					return nil
				}
				d.Status = s
				d.History = append(d.History, DonationEvent{At: now, Status: s, By: by, Note: fieldText(f, "event")})
			} else if note := fieldText(f, "event"); note != "" {
				d.History = append(d.History, DonationEvent{At: now, Status: d.Status, By: by, Note: note})
			}
			d.UpdatedAt, d.CampName = now, ""
			return putJSON(b, d.ID, d)
		})
		switch {
		case err != nil:
			fail(w, 500, "database error")
		case status == 404:
			fail(w, 404, "Donación no encontrada.")
		case field != "":
			failField(w, 400, field, msg)
		default:
			response(w, 200, d)
		}
	})
}

// statPledge: totales públicos por artículo (sin nombres ni teléfonos).
type statPledge struct {
	Key       string `json:"key"`
	Item      string `json:"item"`
	Unit      string `json:"unit"`
	Potential int    `json:"potential"`
	Committed int    `json:"committed"`
	Delivered int    `json:"delivered"`
	Offers    int    `json:"offers"`
}

func donationPledges(tx *bolt.Tx) ([]statPledge, error) {
	by, order := map[string]*statPledge{}, []string{}
	b := tx.Bucket(donationsBucket)
	if b == nil {
		return []statPledge{}, nil
	}
	err := b.ForEach(func(_, v []byte) error {
		var d Donation
		if e := json.Unmarshal(v, &d); e != nil {
			return e
		}
		pot, com, del := donationCounts(d.Status)
		if !pot && !com && !del {
			return nil
		}
		for _, it := range d.Items {
			key := it.Key
			if key == "other" {
				key = "other:" + strings.ToLower(foldText(it.Item))
			}
			p := by[key]
			if p == nil {
				p = &statPledge{Key: key, Item: it.Item, Unit: it.Unit}
				by[key] = p
				order = append(order, key)
			}
			p.Offers++
			switch {
			case pot:
				p.Potential += it.Quantity
			case com:
				p.Committed += it.Quantity
			default:
				p.Delivered += it.Quantity
			}
		}
		return nil
	})
	out := []statPledge{}
	for _, k := range order {
		out = append(out, *by[k])
	}
	return out, err
}
