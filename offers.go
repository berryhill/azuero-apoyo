package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Ofrecimientos de ayuda («🙌 ¿Quieres ayudar?» en el inicio, sin cuenta): un formulario de contacto que llega a la
// bandeja «📥 Ofrecimientos» de quien tiene el rol «Donaciones» (y administración). Allí se les da seguimiento como en
// un CRM: estado (nuevo, contactado, en curso, cerrado), notas con quién y cuándo, y los datos de la donación si ofrece
// donar (cómo, cuánto, qué). No crea usuarios ni se publica.
var offersBucket = []byte("help_offers")

// Orden del formulario: reportar necesidades, llevar donaciones en el sitio, organizar acopio y despachos, cocinar,
// centro de llamadas (verificar necesidades), reparar casas y donar. Reportar, llevar donaciones, acopio, cocinar y
// reparar casas ponen su nombre y teléfono a la vista de la red: para esas también se pide aceptar que se compartan
// (como al crear cuenta); el centro de llamadas y donar no lo piden.
var offerHelp = []string{"report", "deliver", "supply", "cook", "calls", "repair", "donate"}

var offerSharesContact = []string{"report", "deliver", "supply", "cook", "repair"}

func offerUsesApp(help []string) bool {
	for _, h := range help {
		if slices.Contains(offerSharesContact, h) {
			return true
		}
	}
	return false
}

var offerKinds = []string{"money", "items", "service"}
var offerStatuses = []string{"new", "contacted", "progress", "closed"}

type OfferNote struct {
	At     string `json:"at"`
	ByName string `json:"byName"`
	Text   string `json:"text,omitempty"`
	Status string `json:"status,omitempty"`
}

type HelpOffer struct {
	ID        string   `json:"id"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	Name      string   `json:"name"`
	Telephone string   `json:"telephone"`
	Email     string   `json:"email,omitempty"`
	HelpWith  []string `json:"helpWith"`
	Message   string   `json:"message,omitempty"`
	// Si ofrece donar: money (dinero) | items (artículos) | service (servicio o transporte); cuánto (texto libre, p. ej.
	// «$200» o «20 bolsas») y qué.
	DonationKind   string `json:"donationKind,omitempty"`
	DonationAmount string `json:"donationAmount,omitempty"`
	DonationWhat   string `json:"donationWhat,omitempty"`
	// Aceptó que su nombre y teléfono se compartan con quienes ayudan (obligatorio si ofrece una tarea en la app).
	ShareConsentAt string      `json:"shareConsentAt,omitempty"`
	Status         string      `json:"status"`
	Notes          []OfferNote `json:"notes,omitempty"`
}

func ensureOffers(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(offersBucket)
	return err
}

func (a *App) offerRoutes(mux *http.ServeMux) {
	// Público: guarda el ofrecimiento. Como las altas, 10 cada 10 minutos por conexión.
	mux.HandleFunc("POST /api/help-offers", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) {
			return
		}
		key, now := "offer|"+clientIP(r), time.Now()
		a.mu.Lock()
		at := a.codeAttempts[key]
		if now.After(at.until) {
			at = attempt{}
		}
		if at.count >= codeMaxWrong {
			a.mu.Unlock()
			fail(w, 429, "Demasiados envíos desde esta conexión. Espera unos minutos e intenta de nuevo.")
			return
		}
		a.mu.Unlock()
		f, ok := decodeFields(w, r, map[string]int{"name": 120, "telephone": 40, "email": 160, "helpWith": 0, "message": 2000, "donationKind": 20, "donationAmount": 80, "donationWhat": 1000, "consent": 0, "shareConsent": 0})
		if !ok {
			return
		}
		o := HelpOffer{ID: randomToken()[:20], CreatedAt: nowStamp(), Status: "new", HelpWith: []string{}}
		o.Name = strings.Join(strings.Fields(fieldText(f, "name")), " ")
		o.Telephone = strings.Join(strings.Fields(fieldText(f, "telephone")), " ")
		o.Email = normalizeEmail(fieldText(f, "email"))
		o.Message = strings.TrimSpace(fieldText(f, "message"))
		var help []string
		if raw, has := f["helpWith"]; has && json.Unmarshal(raw, &help) != nil {
			failField(w, 400, "helpWith", "Opciones de ayuda inválidas.")
			return
		}
		for _, h := range help {
			if slices.Contains(offerHelp, h) && !slices.Contains(o.HelpWith, h) {
				o.HelpWith = append(o.HelpWith, h)
			}
		}
		// Nombre obligatorio; teléfono o correo (al menos uno); el teléfono es obligatorio si ofrece una tarea que comparte
		// su contacto con la red (reportar, llevar donaciones, acopio, cocinar, reparar casas).
		hasPhone := len(phoneDigits(o.Telephone)) >= 7
		switch {
		case len([]rune(o.Name)) < 2:
			failField(w, 400, "name", "Escribe tu nombre.")
			return
		case o.Telephone != "" && !hasPhone:
			failField(w, 400, "telephone", "Escribe un teléfono válido (por ejemplo 6000-0000).")
			return
		case !hasPhone && o.Email == "":
			failField(w, 400, "telephone", "Escribe tu teléfono o tu correo para poder contactarte.")
			return
		case !hasPhone && offerUsesApp(o.HelpWith):
			failField(w, 400, "telephone", "Para esta forma de ayudar necesitamos tu teléfono.")
			return
		case o.Email != "" && !validEmail(o.Email):
			failField(w, 400, "email", "Correo inválido (por ejemplo nombre@gmail.com).")
			return
		case len(o.HelpWith) == 0:
			failField(w, 400, "helpWith", "Marca al menos una forma de ayudar.")
			return
		case strings.TrimSpace(string(f["consent"])) != "true":
			failField(w, 400, "consent", "Marca la casilla para aceptar que te contacten.")
			return
		case offerUsesApp(o.HelpWith) && strings.TrimSpace(string(f["shareConsent"])) != "true":
			failField(w, 400, "shareConsent", "Para ayudar con la app, marca la casilla para aceptar que tu nombre y teléfono se compartan.")
			return
		}
		if strings.TrimSpace(string(f["shareConsent"])) == "true" {
			o.ShareConsentAt = o.CreatedAt
		}
		if slices.Contains(o.HelpWith, "donate") {
			o.DonationKind = fieldText(f, "donationKind")
			if o.DonationKind != "" && !slices.Contains(offerKinds, o.DonationKind) {
				failField(w, 400, "donationKind", "Elige cómo quieres donar.")
				return
			}
			o.DonationAmount = strings.TrimSpace(fieldText(f, "donationAmount"))
			o.DonationWhat = strings.TrimSpace(fieldText(f, "donationWhat"))
		}
		o.UpdatedAt = o.CreatedAt
		if serviceError(w, a.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(offersBucket), o.ID, o) })) {
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
		response(w, 201, map[string]string{"id": o.ID, "status": o.Status})
	})
}

// offerDesk: la bandeja, por las mismas dos puertas que las boletas de donación (administración y el rol Donaciones).
// Las dos ven todos los ofrecimientos, marquen o no «Donar», para que ninguno se pierda.
func (a *App) offerDesk(mux *http.ServeMux, prefix string, who func(http.ResponseWriter, *http.Request) (string, bool)) {
	visible := func(HelpOffer) bool { return true }
	mux.HandleFunc("GET "+prefix+"/offers", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := who(w, r); !ok {
			return
		}
		list := []HelpOffer{}
		err := a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(offersBucket).ForEach(func(_, v []byte) error {
				var o HelpOffer
				if json.Unmarshal(v, &o) == nil && visible(o) {
					list = append(list, o)
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt > list[j].CreatedAt })
		response(w, 200, list)
	})
	// Seguimiento: cambia el estado y/o agrega una nota; cada cambio queda en el historial con quién y cuándo.
	mux.HandleFunc("PATCH "+prefix+"/offers/{id}", func(w http.ResponseWriter, r *http.Request) {
		by, ok := who(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"status": 20, "note": 2000})
		if !ok {
			return
		}
		status, note := fieldText(f, "status"), strings.TrimSpace(fieldText(f, "note"))
		if status != "" && !slices.Contains(offerStatuses, status) {
			failField(w, 400, "status", "Estado inválido.")
			return
		}
		if status == "" && note == "" {
			failField(w, 400, "note", "Escribe una nota o cambia el estado.")
			return
		}
		var o HelpOffer
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(offersBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &o); e != nil {
				return e
			}
			if !visible(o) {
				return errMissing
			}
			n := OfferNote{At: nowStamp(), ByName: by, Text: note}
			if status != "" && status != o.Status {
				o.Status, n.Status = status, status
			}
			if n.Text == "" && n.Status == "" {
				return nil
			}
			o.Notes = append(o.Notes, n)
			o.UpdatedAt = n.At
			return putJSON(b, o.ID, o)
		})
		if err == errMissing {
			fail(w, 404, "Ese ofrecimiento ya no existe.")
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, o)
	})
}

// countNewOffers: cuántos ofrecimientos siguen en «Nuevo» (nadie los atendió todavía).
func countNewOffers(tx *bolt.Tx) int {
	n := 0
	if b := tx.Bucket(offersBucket); b != nil {
		b.ForEach(func(_, v []byte) error {
			var o HelpOffer
			if json.Unmarshal(v, &o) == nil && o.Status == "new" {
				n++
			}
			return nil
		})
	}
	return n
}
