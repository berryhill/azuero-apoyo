package main

import (
	"bytes"
	"encoding/json"
	"image/png"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// Vista previa al compartir: cada página lleva la necesidad más reciente; /runner/<código> la tarjeta del punto. Solo datos
// públicos: nada de contactos, ni reparaciones, ni puntos «por ubicar».
func TestShareCardsLatestNeedAndShortPointLink(t *testing.T) {
	s := newServiceHarness(t)
	camp := s.expect("POST", "/api/admin/points", `{"code":"AB-7","area":"Agua buena","place":"Al final de la calle","coordinator":"Chicho","coordinatorPhone":"6111-2222","people":"5","adults":"3","children":"2","contact":"Private camp","phone":"6555-1234"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/camps/"+camp+"/needs", `{"need":"Agua (Para beber), Alimentos","urgency":"urgente"}`, 201)
	s.expect("POST", "/api/admin/camps/"+camp+"/needs", `{"need":"Reparación de casa (Hojas de zinc)","urgency":"normal","kind":"repair"}`, 201)
	hidden := s.call("POST", "/api/admin/points", `{"area":"Por ubicar","place":"Escondido","toLocate":true}`, true)
	var hp map[string]any
	if hidden.Code != 201 || json.Unmarshal(hidden.Body.Bytes(), &hp) != nil {
		t.Fatalf("hidden point: %s", hidden.Body.String())
	}
	s.expect("POST", "/api/admin/camps/"+hp["id"].(string)+"/needs", `{"need":"Secreto","urgency":"normal"}`, 201)

	home := s.call("GET", "/", "", false).Body.String()
	for _, want := range []string{`<meta property="og:title" content="Mapa de apoyo · Azuero">`, `og:image" content="http://example.com/og/latest.png?v=`, `URGENTE`, `Agua buena`, `Agua (Para beber), Alimentos`, `twitter:card" content="summary_large_image"`} {
		if !strings.Contains(home, want) {
			t.Fatalf("home tags missing %q", want)
		}
	}
	for _, leak := range []string{"Private camp", "6555-1234", "Secreto", "Escondido", "Hojas de zinc"} {
		if strings.Contains(home, leak) {
			t.Fatalf("home tags leak %q", leak)
		}
	}

	// /runner/<código>: la página de Runners con la vista previa del punto (runner.html pone la búsqueda desde la dirección).
	card := s.call("GET", "/runner/ab-7", "", false)
	body := card.Body.String()
	if card.Code != 200 || !strings.Contains(body, `<script src="/nav.js">`) || !strings.Contains(body, `<title>Punto AB-7 · Agua buena · Azuero</title>`) || !strings.Contains(body, "3 Adultos · 2 Niños") || !strings.Contains(body, `/og/p/`+camp+`.png?v=`) || !strings.Contains(body, `og:url" content="http://example.com/runner/ab-7"`) {
		t.Fatalf("share link: %d %s", card.Code, body[:min(len(body), 600)])
	}
	if strings.Contains(body, "Private camp") || strings.Contains(body, "6555-1234") || strings.Contains(body, "6111-2222") || strings.Contains(body, "Hojas de zinc") {
		t.Fatal("share link leaks private data")
	}
	if w := s.call("GET", "/runner/"+hp["id"].(string), "", false); w.Code != 200 || strings.Contains(w.Body.String(), "Escondido") || strings.Contains(w.Body.String(), "/og/p/") {
		t.Fatalf("hidden point has a card: %d", w.Code)
	}
	if w := s.call("GET", "/p/ab-7", "", false); w.Code != 301 || w.Header().Get("Location") != "/runner/ab-7" {
		t.Fatalf("old /p/ links: %d %s", w.Code, w.Header().Get("Location"))
	}
	// /supply/<código>: el enlace corto de Supply: la página de Supply con la misma vista previa (sin datos privados).
	sup := s.call("GET", "/supply/ab-7", "", false)
	if b := sup.Body.String(); sup.Code != 200 || !strings.Contains(b, `<title>Punto AB-7 · Agua buena · Azuero</title>`) || !strings.Contains(b, `id="tabFind"`) || !strings.Contains(b, `<script src="/nav.js">`) || strings.Contains(b, "6555-1234") || strings.Contains(b, "Private camp") {
		t.Fatalf("Supply share link: %d", sup.Code)
	}
	// La tarjeta del punto (Go): el código no automático, quién reporta y sus fichas; nunca contactos.
	var pc ogNeed
	s.app.db.View(func(tx *bolt.Tx) error {
		p, _ := findPublicPoint(tx, "AB-7")
		pc = pointCard(tx, p)
		return nil
	})
	if pc.Coordinator != "Chicho" || len(pc.Counts) != 2 || len(pc.Items) != 2 {
		t.Fatalf("point card: %+v", pc)
	}
	if c := ogChipOf("Gas de cocina (Tropigas / Panagas)"); c.icon != "gas" || len(c.badges) != 2 || c.badges[1].bg != "#272974" {
		t.Fatalf("gas chip: %+v", c)
	}
	if c := ogChipOf("Algo escrito a mano"); !c.text || c.icon != "" {
		t.Fatalf("text chip: %+v", c)
	}

	for _, path := range []string{"/og/latest.png", "/og/p/" + camp + ".png"} {
		w := s.call("GET", path, "", false)
		img, err := png.Decode(bytes.NewReader(w.Body.Bytes()))
		if w.Code != 200 || err != nil || img.Bounds().Dx() != ogW || img.Bounds().Dy() != ogH {
			t.Fatalf("%s: %d %v", path, w.Code, err)
		}
	}
}

// La imagen de un punto nunca queda vieja: se dibuja con los datos de ese momento y su dirección (?v=) cambia en cuanto
// cambian sus necesidades o su información, así WhatsApp y Facebook piden la nueva. Ocultarlo la retira de inmediato.
func TestShareCardFollowsPointChanges(t *testing.T) {
	s := newServiceHarness(t)
	camp := s.expect("POST", "/api/admin/points", `{"code":"CV-1","area":"Cambutal","people":"4"}`, 201)["id"].(string)
	imageURL := func() string {
		body := s.call("GET", "/runner/cv-1", "", false).Body.String()
		i := strings.Index(body, `og:image" content="`)
		if i < 0 {
			return ""
		}
		return body[i : i+strings.Index(body[i:], `">`)]
	}
	before := imageURL()
	need := s.expect("POST", "/api/admin/camps/"+camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	withNeed := imageURL()
	s.expect("PATCH", "/api/admin/needs/"+need, `{"need":"Agua, Alimentos","urgency":"urgente"}`, 200)
	edited := imageURL()
	if before == "" || before == withNeed || withNeed == edited {
		t.Fatalf("image URL must change with the needs:\n%s\n%s\n%s", before, withNeed, edited)
	}
	w := s.call("GET", "/og/p/"+camp+".png", "", false)
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("image: %d", w.Code)
	}
}

func BenchmarkShareCardRender(b *testing.B) {
	n := ogNeed{ID: "x", Area: "Agua buena", Place: "Al final de la calle", Coordinator: "Chicho", People: "3", Counts: [][3]string{{"adult", "3", "adultos"}}, CreatedAt: "2026-10-16T22:13:02Z", Point: true, GPS: true,
		Items: []string{"Agua (Para beber / Para cocinar)", "Alimentos (Comida preparada / Ingredientes para cocinar)", "Kit de cocina", "Gas de cocina (Tropigas / Panagas)", "Refugio (carpa / lona / toldo)", "Tarjeta de datos prepago (Tigo / +Móvil)"}}
	for i := 0; i < b.N; i++ {
		if _, err := ogImage(n); err != nil {
			b.Fatal(err)
		}
	}
}
