package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func pledgeOf(t *testing.T, s *serviceHarness, key string) statPledge {
	t.Helper()
	var st publicStats
	json.Unmarshal(s.call("GET", "/api/stats", "", false).Body.Bytes(), &st)
	for _, p := range st.Pledges {
		if p.Key == key {
			return p
		}
	}
	return statPledge{}
}

// A public offer is potential until coordination confirms it; only then it counts as committed. The donor's contact
// stays private; the public stats show totals only.
func TestDonationOfferConfirmAndTrack(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/donations", `{"name":" Marta  Díaz ","phone":"6777-1234","email":"marta@example.com","when":"El sábado en Las Tablas","items":[{"key":"water","quantity":500},{"key":"mat:Zinc","item":"Zinc","quantity":20},{"key":"other","item":"Linternas","unit":"cajas","quantity":3}]}`, false)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"ref":"D-0001"`) || strings.Contains(w.Body.String(), "6777") {
		t.Fatalf("offer: %d %s", w.Code, w.Body.String())
	}
	if p := pledgeOf(t, s, "water"); p.Potential != 500 || p.Committed != 0 || p.Unit != "L" {
		t.Fatalf("offered water is potential only: %+v", p)
	}
	if p := pledgeOf(t, s, "mat:zinc"); p.Potential != 20 {
		t.Fatalf("materials keyed by folded name: %+v", p)
	}
	if stats := s.call("GET", "/api/stats", "", false).Body.String(); strings.Contains(stats, "Marta") || strings.Contains(stats, "6777") || strings.Contains(stats, "example.com") {
		t.Fatalf("donor contact leaked to public stats: %s", stats)
	}
	if w := s.call("GET", "/api/admin/donations", "", false); w.Code != 401 {
		t.Fatalf("list is admin-only: %d", w.Code)
	}
	var list []Donation
	json.Unmarshal(s.call("GET", "/api/admin/donations", "", true).Body.Bytes(), &list)
	if len(list) != 1 || list[0].DonorName != "Marta Díaz" || list[0].DonorPhone != "6777-1234" || list[0].Status != donationOffered || len(list[0].Items) != 3 {
		t.Fatalf("admin list: %+v", list)
	}
	id := list[0].ID
	// Confirmed with a smaller quantity, a collection point and a destination point.
	s.expect("PATCH", "/api/admin/donations/"+id, `{"status":"confirmed","event":"Llamé, confirma 400 L","items":[{"key":"water","quantity":400}],"dropPlace":"Acopio Las Tablas","dropAt":"2026-10-18T09:00","campId":"`+s.camp+`","dispatchAt":"2026-10-19"}`, 200)
	if p := pledgeOf(t, s, "water"); p.Potential != 0 || p.Committed != 400 {
		t.Fatalf("confirmed water is committed: %+v", p)
	}
	if p := pledgeOf(t, s, "mat:zinc"); p.Offers != 0 {
		t.Fatalf("items removed at confirmation no longer count: %+v", p)
	}
	s.expect("PATCH", "/api/admin/donations/"+id, `{"status":"delivered"}`, 200)
	if p := pledgeOf(t, s, "water"); p.Committed != 0 || p.Delivered != 400 {
		t.Fatalf("delivered is reported apart: %+v", p)
	}
	json.Unmarshal(s.call("GET", "/api/admin/donations", "", true).Body.Bytes(), &list)
	d := list[0]
	if d.CampName == "" || d.DropPlace != "Acopio Las Tablas" || len(d.History) != 3 || d.History[1].Note != "Llamé, confirma 400 L" || d.History[1].By != "Administración" {
		t.Fatalf("logistics and history: %+v", d)
	}
	for body, field := range map[string]string{
		`{"campId":"nope"}`:                        "campId",
		`{"status":"lost"}`:                        "status",
		`{"dropAt":"mañana"}`:                      "dropAt",
		`{"items":[{"key":"water","quantity":0}]}`: "items",
	} {
		if w := s.call("PATCH", "/api/admin/donations/"+id, body, true); w.Code != 400 || !strings.Contains(w.Body.String(), field) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	// Public validation: phone and at least one known item.
	for body, field := range map[string]string{
		`{"name":"Ana","phone":"12","items":[{"key":"food","quantity":5}]}`:        "phone",
		`{"name":"Ana","phone":"6000-0000"}`:                                       "items",
		`{"name":"Ana","phone":"6000-0000","items":[{"key":"gold","quantity":1}]}`: "items",
	} {
		if w := s.call("POST", "/api/donations", body, false); w.Code != 400 || !strings.Contains(w.Body.String(), field) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	// Admin-only fields are not accepted from the public form.
	if w := s.call("POST", "/api/donations", `{"name":"Ana","phone":"6000-0000","status":"confirmed","items":[{"key":"food","quantity":5}]}`, false); w.Code != 400 {
		t.Fatalf("public status: %d", w.Code)
	}
}

// Only admin and users with the «Donaciones» role (own code session) see and manage tickets; agents do not.
func TestDonationsDeskRoleAndPhoneIntake(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Dora Donaciones","telephone":"6999-1111","roles":["donations"],"accessCode":"DORA12"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-2222","roles":["agent"],"accessCode":"ANA123"}`, 201)
	if w := s.agentCall("POST", "/api/desk/session", `{"accessCode":"ANA123"}`, nil, ""); w.Code != 403 {
		t.Fatalf("an agent code does not open Donaciones: %d %s", w.Code, w.Body.String())
	}
	agent, agentCsrf := s.agentLogin("ANA123")
	if w := s.agentCall("GET", "/api/agent/donations", "", agent, agentCsrf); w.Code == 200 {
		t.Fatalf("agents have no donations route: %d", w.Code)
	}
	if w := s.agentCall("GET", "/api/desk/donations", "", nil, ""); w.Code != 401 {
		t.Fatalf("desk needs a session: %d", w.Code)
	}
	w := s.agentCall("POST", "/api/desk/session", `{"accessCode":"dora12"}`, nil, "")
	if w.Code != 200 {
		t.Fatalf("desk login: %d %s", w.Code, w.Body.String())
	}
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	csrf := out["csrf"].(string)
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "desk_session" {
			cookie = c
		}
	}
	if cookie == nil || cookie.Path != "/api/desk" {
		t.Fatalf("desk cookie: %+v", w.Result().Cookies())
	}
	body := `{"name":"Pedro","phone":"6111-2222","status":"confirmed","items":[{"key":"food","quantity":90}],"transporter":"Camión de Luis","transporterPhone":"6222-3333"}`
	if w := s.agentCall("POST", "/api/desk/donations", body, cookie, "wrong"); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	w = s.agentCall("POST", "/api/desk/donations", body, cookie, csrf)
	if w.Code != 201 || !strings.Contains(w.Body.String(), "Dora Donaciones") {
		t.Fatalf("phone intake by the donations role: %d %s", w.Code, w.Body.String())
	}
	if p := pledgeOf(t, s, "food"); p.Committed != 90 {
		t.Fatalf("intake confirmed directly: %+v", p)
	}
	if w := s.agentCall("GET", "/api/desk/donations", "", cookie, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Camión de Luis") {
		t.Fatalf("desk list: %d %s", w.Code, w.Body.String())
	}
	// The unified code login lists the Donaciones panel.
	if w := s.call("POST", "/api/session/code", `{"accessCode":"DORA12"}`, false); w.Code != 200 || !strings.Contains(w.Body.String(), "/donations.html") {
		t.Fatalf("code login panels: %d %s", w.Code, w.Body.String())
	}
	// Removing the role closes the session.
	var users []map[string]any
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		if u["name"] == "Dora Donaciones" {
			s.expect("PATCH", "/api/admin/users/"+u["id"].(string), `{"roles":["runner"]}`, 200)
		}
	}
	if w := s.agentCall("GET", "/api/desk/donations", "", cookie, ""); w.Code != 401 {
		t.Fatalf("role removed: %d", w.Code)
	}
}
