package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// agentCall: solicitud del panel de agentes con su cookie y token CSRF.
func (s *serviceHarness) agentCall(method, path, body string, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	s.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}

func (s *serviceHarness) agentLogin(code string) (*http.Cookie, string) {
	s.t.Helper()
	w := s.agentCall("POST", "/api/agent/session", `{"accessCode":"`+code+`"}`, nil, "")
	if w.Code != 200 {
		s.t.Fatalf("agent login %s: %d %s", code, w.Code, w.Body.String())
	}
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	for _, c := range w.Result().Cookies() {
		if c.Name == "agent_session" {
			return c, out["csrf"].(string)
		}
	}
	s.t.Fatal("no agent_session cookie")
	return nil, ""
}

// An agent signs in with an access code and works points and needs through /api/agent, credited by name;
// everything else stays admin-only.
func TestAgentPanelUsesPointAndNeedEndpointsOnly(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-1111","roles":["agent"],"accessCode":"ANA123"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Rui","telephone":"6999-2222","roles":["runner"],"accessCode":"RUI123"}`, 201)

	if w := s.agentCall("POST", "/api/agent/session", `{"accessCode":"RUI123"}`, nil, ""); w.Code != 403 {
		t.Fatalf("a runner code is not an agent: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/agent/session", `{"accessCode":"NOPE99"}`, nil, ""); w.Code == 200 {
		t.Fatalf("wrong code: %d", w.Code)
	}
	if w := s.agentCall("GET", "/api/agent/points", "", nil, ""); w.Code != 401 {
		t.Fatalf("points need an agent session: %d", w.Code)
	}
	cookie, csrf := s.agentLogin("ANA123")
	if w := s.agentCall("GET", "/api/agent/session", "", cookie, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Ana Agente") {
		t.Fatalf("session: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("GET", "/api/agent/points", "", cookie, ""); w.Code != 200 || !strings.Contains(w.Body.String(), s.camp) {
		t.Fatalf("points: %d %s", w.Code, w.Body.String())
	}
	// Writes need the session's CSRF token.
	if w := s.agentCall("PATCH", "/api/agent/camps/"+s.camp, `{"place":"Otro lugar"}`, cookie, "wrong"); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	if w := s.agentCall("PATCH", "/api/agent/camps/"+s.camp, `{"place":"Otro lugar"}`, cookie, csrf); w.Code != 200 {
		t.Fatalf("edit point: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/agent/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, cookie, csrf); w.Code != 201 {
		t.Fatalf("create need: %d %s", w.Code, w.Body.String())
	}
	if needs := s.needsOf(s.camp); len(needs) != 1 || needs[0].Need != "Agua" {
		t.Fatalf("need saved: %+v", needs)
	}
	w := s.agentCall("POST", "/api/agent/points", `{"code":"C-2","area":"Venao","place":"Escuela","contact":"Privado","phone":"6555-0000"}`, cookie, csrf)
	if w.Code != 201 {
		t.Fatalf("add point: %d %s", w.Code, w.Body.String())
	}
	var added Point
	json.Unmarshal(w.Body.Bytes(), &added)
	var points []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &points)
	credited := false
	for _, p := range points {
		credited = credited || p.ID == added.ID && p.AddedByName == "Ana Agente" && p.Status == "approved"
	}
	if !credited {
		t.Fatalf("point published and credited to the agent: %+v", points)
	}
	// Assignment list: names and roles; runner and builder phones so the command center can call them, never other agents'.
	if body := s.agentCall("GET", "/api/agent/users", "", cookie, "").Body.String(); !strings.Contains(body, "Rui") || !strings.Contains(body, "6999-2222") || strings.Contains(body, "6999-1111") {
		t.Fatalf("users for assignment: %s", body)
	}
	for _, c := range []struct{ method, path, body string }{
		{"POST", "/api/agent/users", `{"name":"X","telephone":"6999-3333","roles":["agent"],"accessCode":"XXX123"}`},
		{"GET", "/api/agent/settings", ""},
		{"POST", "/api/agent/settings", `{"requireReviews":false}`},
		{"GET", "/api/agent/backup", ""},
		{"POST", "/api/agent/password", `{"current":"x","password":"y"}`},
		{"POST", "/api/agent/teams", `{"name":"Nuevo"}`},
		{"GET", "/api/agent/boletines", ""},
	} {
		if w := s.agentCall(c.method, c.path, c.body, cookie, csrf); w.Code != 404 && w.Code != 405 {
			t.Fatalf("%s %s must stay admin-only: %d %s", c.method, c.path, w.Code, w.Body.String())
		}
	}
	// The admin cookie is not an agent session and the agent cookie is not an admin session.
	if w := s.call("GET", "/api/agent/points", "", true); w.Code != 401 {
		t.Fatalf("admin cookie on /api/agent: %d", w.Code)
	}
	if w := s.agentCall("GET", "/api/admin/users", "", cookie, ""); w.Code != 401 {
		t.Fatalf("agent cookie on /api/admin: %d", w.Code)
	}
	s.agentCall("POST", "/api/agent/logout", "{}", cookie, csrf)
	if w := s.agentCall("GET", "/api/agent/points", "", cookie, ""); w.Code != 401 {
		t.Fatalf("after logout: %d", w.Code)
	}
}
