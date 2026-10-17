package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// «🙌 ¿Quieres ayudar?»: a public contact form (no account) that lands in the Donaciones inbox; follow-up with status
// and notes, each with who and when. Only the donations role and admin see it.
func TestHelpOffersInbox(t *testing.T) {
	s := newServiceHarness(t)
	ok := `{"name":"Ana Ayuda","telephone":"6555-1212","email":"ana@gmail.com","helpWith":["donate","repair","hack"],"message":"Albañil en Las Tablas","donationKind":"money","donationAmount":"$200","donationWhat":"Para techos","consent":true,"shareConsent":true}`
	for body, field := range map[string]string{
		strings.Replace(ok, `"consent":true`, `"consent":false`, 1):           "consent",
		strings.Replace(ok, `["donate","repair","hack"]`, `[]`, 1):            "helpWith",
		strings.Replace(ok, `"Ana Ayuda"`, `" "`, 1):                          "name",
		strings.Replace(ok, `"money"`, `"bitcoin"`, 1):                        "donationKind",
		strings.Replace(ok, `"shareConsent":true`, `"shareConsent":false`, 1): "shareConsent",
	} {
		if w := s.call("POST", "/api/help-offers", body, false); w.Code != 400 || !strings.Contains(w.Body.String(), field) {
			t.Fatalf("%s: %d %s", field, w.Code, w.Body.String())
		}
	}
	// Phone or email: either is enough for donating / the call center; phone is required for tasks that share contact.
	if w := s.call("POST", "/api/help-offers", `{"name":"Sin Contacto","helpWith":["donate"],"consent":true}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), "telephone") {
		t.Fatalf("needs phone or email: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/help-offers", `{"name":"Solo Correo","email":"solo@gmail.com","helpWith":["donate"],"consent":true}`, false); w.Code != 201 {
		t.Fatalf("email only is fine for donating: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/help-offers", `{"name":"Correo Reporta","email":"r@gmail.com","helpWith":["report"],"consent":true,"shareConsent":true}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), "telephone") {
		t.Fatalf("reporting needs a phone: %d %s", w.Code, w.Body.String())
	}
	// Donating or the call center: no sharing consent needed.
	if w := s.call("POST", "/api/help-offers", `{"name":"Don Ante","telephone":"6555-9999","helpWith":["donate","calls"],"consent":true}`, false); w.Code != 201 {
		t.Fatalf("donate/calls only: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/help-offers", ok, false); w.Code != 201 {
		t.Fatalf("offer: %d %s", w.Code, w.Body.String())
	}
	// No account is created.
	if body := s.call("GET", "/api/admin/users", "", true).Body.String(); strings.Contains(body, "Ana Ayuda") {
		t.Fatalf("a help offer is not a sign-up")
	}
	// Admin sees it; a runner's viewer session can't.
	var list []HelpOffer
	json.Unmarshal(s.call("GET", "/api/admin/offers", "", true).Body.Bytes(), &list)
	// Ana first, Don second (by name), whatever the order.
	byName := map[string]HelpOffer{}
	for _, o := range list {
		byName[o.Name] = o
	}
	if ana, don := byName["Ana Ayuda"], byName["Don Ante"]; ana.ID != "" && don.ID != "" {
		list = []HelpOffer{ana, don}
	}
	if len(list) != 2 || list[0].ShareConsentAt == "" || list[1].ShareConsentAt != "" || list[0].Status != "new" || strings.Join(list[0].HelpWith, ",") != "donate,repair" || list[0].DonationAmount != "$200" || list[0].Message != "Albañil en Las Tablas" {
		t.Fatalf("inbox: %+v", list)
	}
	if w := s.agentCall("GET", "/api/desk/offers", "", nil, ""); w.Code != 401 {
		t.Fatalf("desk needs a session: %d", w.Code)
	}
	// The donations role follows up: status + note, logged with their name.
	s.expect("POST", "/api/admin/users", `{"name":"Dora Donaciones","telephone":"6999-8888","roles":["donations"],"accessCode":"dora12"}`, 201)
	w := s.agentCall("POST", "/api/desk/session", `{"accessCode":"dora12"}`, nil, "")
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	csrf, _ := out["csrf"].(string)
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "desk_session" {
			cookie = c
		}
	}
	// Every offer reaches the donations desk too, «Donar» or not, so none is lost.
	if w := s.call("POST", "/api/help-offers", `{"name":"Vale Voluntaria","telephone":"6555-7777","helpWith":["calls"],"consent":true}`, false); w.Code != 201 {
		t.Fatalf("call-center offer (no sharing consent needed): %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/help-offers", `{"name":"Coco Cocina","telephone":"6555-6666","helpWith":["cook"],"consent":true}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), "shareConsent") {
		t.Fatalf("cooking asks for the sharing consent: %d %s", w.Code, w.Body.String())
	}
	var desk []HelpOffer
	json.Unmarshal(s.agentCall("GET", "/api/desk/offers", "", cookie, csrf).Body.Bytes(), &desk)
	if len(desk) != 4 {
		t.Fatalf("donations desk sees all 4 offers: %+v", desk)
	}
	// The header shows the donations role how many are still «Nuevo»; others get 0.
	if body := s.as(s.viewerCookie("dora12"), "GET", "/api/full/me", "").Body.String(); !strings.Contains(body, `"newOffers":4`) {
		t.Fatalf("donations role sees new offers in /api/full/me: %s", body)
	}
	w = s.agentCall("PATCH", "/api/desk/offers/"+list[0].ID, `{"status":"contacted","note":"Llamé: dona el lunes"}`, cookie, csrf)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"contacted"`) || !strings.Contains(w.Body.String(), "Dora Donaciones") || !strings.Contains(w.Body.String(), "Llamé: dona el lunes") {
		t.Fatalf("follow-up: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("PATCH", "/api/desk/offers/"+list[0].ID, `{"status":"lost"}`, cookie, csrf); w.Code != 400 {
		t.Fatalf("bad status: %d", w.Code)
	}
}
