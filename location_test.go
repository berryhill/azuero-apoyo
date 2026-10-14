package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunnerLiveLocationVisibleToCodeUsersOnly(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	rosaCookie, rosaCSRF := s.crewLogin("/api/runner", "Rosa")
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":7.43,"lon":-80.19,"accuracy":12}`, rosaCookie, "wrong"); w.Code != 403 {
		t.Fatalf("location needs CSRF: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":95,"lon":-80.19}`, rosaCookie, rosaCSRF); w.Code != 400 {
		t.Fatalf("bad location: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":7.43,"lon":-80.19,"accuracy":12}`, rosaCookie, rosaCSRF); w.Code != 200 {
		t.Fatalf("share location: %d %s", w.Code, w.Body.String())
	}
	// Fufo signs in with his code (full view) and also shares from his runner panel.
	login := s.call("POST", "/api/session/code", `{"accessCode":"FUFO12"}`, false)
	var viewer, fufoRunner *http.Cookie
	for _, c := range login.Result().Cookies() {
		switch c.Name {
		case "viewer_session":
			viewer = c
		case "runner_session":
			fufoRunner = c
		}
	}
	var sess map[string]any
	json.Unmarshal(s.crewCall("GET", "/api/runner/session", "", fufoRunner, "").Body.Bytes(), &sess)
	fufoCSRF, _ := sess["csrf"].(string)
	s.crewCall("POST", "/api/runner/location", `{"lat":7.5,"lon":-80.3}`, fufoRunner, fufoCSRF)
	runners := func(c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/api/full/runners", nil)
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	if w := runners(nil); w.Code != 401 {
		t.Fatalf("public can't see runners: %d", w.Code)
	}
	body := runners(viewer).Body.String()
	if !strings.Contains(body, `"name":"Rosa"`) || !strings.Contains(body, `"lat":7.43`) || strings.Contains(body, "Fufo") {
		t.Fatalf("code user sees other runners but not himself: %s", body)
	}
	// On the runner panel, a code user sees the team too; a name-and-phone runner doesn't.
	if w := s.crewCall("GET", "/api/runner/team", "", fufoRunner, ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Rosa"`) || strings.Contains(w.Body.String(), "Fufo") {
		t.Fatalf("code runner sees the others on the runner map: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("GET", "/api/runner/team", "", rosaCookie, ""); w.Code != 403 {
		t.Fatalf("runner without a code can't see the team: %d", w.Code)
	}
	if w := s.crewCall("GET", "/api/runner/team", "", nil, ""); w.Code != 401 {
		t.Fatalf("team needs a session: %d", w.Code)
	}
	// Stopping, or an old position, removes the runner from the map.
	rosa := rosaIDFrom(t, s, rosaCookie)
	s.app.mu.Lock()
	loc := s.app.locations[rosa]
	loc.At = time.Now().Add(-3 * time.Minute)
	s.app.locations[rosa] = loc
	s.app.mu.Unlock()
	if body := runners(viewer).Body.String(); strings.Contains(body, "Rosa") {
		t.Fatalf("stale position must disappear: %s", body)
	}
	s.crewCall("POST", "/api/runner/location", `{"lat":7.43,"lon":-80.19}`, rosaCookie, rosaCSRF)
	s.crewCall("POST", "/api/runner/location/stop", `{}`, rosaCookie, rosaCSRF)
	if body := runners(viewer).Body.String(); strings.Contains(body, "Rosa") {
		t.Fatalf("stop must remove the position: %s", body)
	}
	s.crewCall("POST", "/api/runner/location", `{"lat":7.43,"lon":-80.19}`, rosaCookie, rosaCSRF)
	s.crewCall("POST", "/api/runner/logout", `{}`, rosaCookie, rosaCSRF)
	if body := runners(viewer).Body.String(); strings.Contains(body, "Rosa") {
		t.Fatalf("logout must remove the position: %s", body)
	}
}

func rosaIDFrom(t *testing.T, s *serviceHarness, c *http.Cookie) string {
	t.Helper()
	var sess map[string]any
	json.Unmarshal(s.crewCall("GET", "/api/runner/session", "", c, "").Body.Bytes(), &sess)
	id, _ := sess["id"].(string)
	return id
}
