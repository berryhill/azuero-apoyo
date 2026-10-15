package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestRunnerLiveLocationVisibleToCodeUsersOnly(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	rosaCookie, rosaCSRF := s.crewLogin("/api/runner", "Rosa")
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":7.43,"lon":-80.19,"accuracy":12}`, rosaCookie, "wrong"); w.Code != 403 {
		t.Fatalf("location needs CSRF: %d", w.Code)
	}
	s.crewCall("POST", "/api/runner/location/consent", `{"accept":true}`, rosaCookie, rosaCSRF)
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
	s.crewCall("POST", "/api/runner/location/consent", `{"accept":true}`, fufoRunner, fufoCSRF)
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

func TestCodeUserSharesLocationFromPublicMap(t *testing.T) {
	s := newServiceHarness(t)
	s.team("Norte",
		s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string),
		s.expect("POST", "/api/admin/users", `{"name":"Bea","telephone":"6999-2222","roles":["builder"],"accessCode":"BEA123"}`, 201)["id"].(string))
	viewer := func(code string) *http.Cookie {
		for _, c := range s.call("POST", "/api/session/code", `{"accessCode":"`+code+`"}`, false).Result().Cookies() {
			if c.Name == "viewer_session" {
				return c
			}
		}
		t.Fatalf("no viewer session for %s", code)
		return nil
	}
	fufo, bea := viewer("FUFO12"), viewer("BEA123")
	call := func(method, path, body string, c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	if w := call("POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`, nil); w.Code != 401 {
		t.Fatalf("sharing needs a code session: %d", w.Code)
	}
	call("POST", "/api/full/location/consent", `{"accept":true}`, bea)
	if w := call("POST", "/api/full/location", `{"lat":7.5,"lon":-200}`, bea); w.Code != 400 {
		t.Fatalf("bad location: %d", w.Code)
	}
	// A builder with a code shares too, and a code runner sees her.
	if w := call("POST", "/api/full/location", `{"lat":7.5,"lon":-80.3,"accuracy":9}`, bea); w.Code != 200 {
		t.Fatalf("share: %d %s", w.Code, w.Body.String())
	}
	if body := call("GET", "/api/full/runners", "", fufo).Body.String(); !strings.Contains(body, `"name":"Bea"`) {
		t.Fatalf("teammates see each other: %s", body)
	}
	if body := call("GET", "/api/full/runners", "", bea).Body.String(); strings.Contains(body, "Bea") {
		t.Fatalf("not herself: %s", body)
	}
	call("POST", "/api/full/location/stop", `{}`, bea)
	if body := call("GET", "/api/full/runners", "", fufo).Body.String(); strings.Contains(body, "Bea") {
		t.Fatalf("stop removes her: %s", body)
	}
	call("POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`, bea)
	call("POST", "/api/full/logout", `{}`, bea)
	if body := call("GET", "/api/full/runners", "", fufo).Body.String(); strings.Contains(body, "Bea") {
		t.Fatalf("logout removes her: %s", body)
	}
}

func TestCodeSignInReturnsOwnID(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	login := s.call("POST", "/api/session/code", `{"accessCode":"FUFO12"}`, false)
	if !strings.Contains(login.Body.String(), `"id":"`+id+`"`) {
		t.Fatalf("sign-in returns the user ID: %s", login.Body.String())
	}
	r := httptest.NewRequest("GET", "/api/full/me", nil)
	none := httptest.NewRecorder()
	s.h.ServeHTTP(none, r)
	if none.Code != 401 {
		t.Fatalf("me needs a code session: %d", none.Code)
	}
	for _, c := range login.Result().Cookies() {
		if c.Name == "viewer_session" {
			r.AddCookie(c)
		}
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"id":"`+id+`"`) {
		t.Fatalf("me: %d %s", w.Code, w.Body.String())
	}
}

func TestLiveLocationStoredInDatabaseAndSurvivesRestart(t *testing.T) {
	s := newServiceHarness(t)
	s.team("Norte",
		s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string),
		s.expect("POST", "/api/admin/users", `{"name":"Bea","telephone":"6999-2222","roles":["builder"],"accessCode":"BEA123"}`, 201)["id"].(string))
	viewer := func(h http.Handler, code string) *http.Cookie {
		r := httptest.NewRequest("POST", "/api/session/code", strings.NewReader(`{"accessCode":"`+code+`"}`))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		for _, c := range w.Result().Cookies() {
			if c.Name == "viewer_session" {
				return c
			}
		}
		t.Fatalf("no viewer session for %s: %d %s", code, w.Code, w.Body.String())
		return nil
	}
	call := func(h http.Handler, method, path, body string, c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.AddCookie(c)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	stored := func(app *App) int {
		n := 0
		app.db.View(func(tx *bolt.Tx) error { n = tx.Bucket(locationsBucket).Stats().KeyN; return nil })
		return n
	}
	beaFirst := viewer(s.h, "BEA123")
	call(s.h, "POST", "/api/full/location/consent", `{"accept":true}`, beaFirst)
	if w := call(s.h, "POST", "/api/full/location", `{"lat":7.5,"lon":-80.3,"accuracy":9}`, beaFirst); w.Code != 200 {
		t.Fatalf("share: %d %s", w.Code, w.Body.String())
	}
	if n := stored(s.app); n != 1 {
		t.Fatalf("position saved in the database: %d", n)
	}
	// Restart (deploy): sessions are gone but the position is still there for the others.
	path := s.app.db.Path()
	s.app.db.Close()
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.db.Close() })
	h := app.routes()
	fufo := viewer(h, "FUFO12")
	if body := call(h, "GET", "/api/full/runners", "", fufo).Body.String(); !strings.Contains(body, `"name":"Bea"`) || !strings.Contains(body, `"lat":7.5`) {
		t.Fatalf("position survives restart: %s", body)
	}
	bea := viewer(h, "BEA123")
	call(h, "POST", "/api/full/location/stop", `{}`, bea)
	if n := stored(app); n != 0 {
		t.Fatalf("stop removes it from the database: %d", n)
	}
	// An old stored position is dropped on start.
	app.db.Update(func(tx *bolt.Tx) error {
		return putJSON(tx.Bucket(locationsBucket), "old", liveLocation{Lat: 7, Lon: -80, At: time.Now().Add(-5 * time.Minute)})
	})
	app.db.Close()
	again, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer again.db.Close()
	if n := stored(again); n != 0 || len(again.locations) != 0 {
		t.Fatalf("stale position pruned on start: %d %v", n, again.locations)
	}
}

// Nobody's location is stored until they accept the location notice; the acceptance is recorded with the notice version.
func TestLocationNeedsAcceptedNotice(t *testing.T) {
	s := newServiceHarness(t)
	fufo := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	viewer := s.viewerCookie("FUFO12")
	if w := s.as(viewer, "POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`); w.Code != 428 || !strings.Contains(w.Body.String(), `"field":"locationConsent"`) {
		t.Fatalf("no location without the notice: %d %s", w.Code, w.Body.String())
	}
	if body := s.as(viewer, "GET", "/api/full/me", "").Body.String(); !strings.Contains(body, `"locationConsent":false`) {
		t.Fatalf("session says it's missing: %s", body)
	}
	if w := s.as(viewer, "POST", "/api/full/location/consent", `{"accept":false}`); w.Code != 400 {
		t.Fatalf("must tick Acepto: %d", w.Code)
	}
	if w := s.as(nil, "POST", "/api/full/location/consent", `{"accept":true}`); w.Code != 401 {
		t.Fatalf("needs a session: %d", w.Code)
	}
	if w := s.as(viewer, "POST", "/api/full/location/consent", `{"accept":true}`); w.Code != 200 {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(viewer, "POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`); w.Code != 200 {
		t.Fatalf("shares after accepting: %d", w.Code)
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	if len(users) != 1 || users[0].LocationConsentAt == "" {
		t.Fatalf("admin sees when it was accepted: %+v", users)
	}
	// A new notice version asks everyone again.
	s.app.db.Update(func(tx *bolt.Tx) error {
		u, _ := getUser(tx, fufo)
		u.LocationConsentVersion = "old"
		return putJSON(tx.Bucket(usersBucket), u.ID, u)
	})
	if w := s.as(viewer, "POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`); w.Code != 428 {
		t.Fatalf("older notice: %d", w.Code)
	}
	// Runner panel: same rule, and its session tells the page.
	cookie, csrf := s.crewLogin("/api/runner", "Rosa")
	if body := s.crewCall("GET", "/api/runner/session", "", cookie, "").Body.String(); !strings.Contains(body, `"locationConsent":false`) {
		t.Fatalf("runner session: %s", body)
	}
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":7.4,"lon":-80.2}`, cookie, csrf); w.Code != 428 {
		t.Fatalf("runner without notice: %d", w.Code)
	}
	s.crewCall("POST", "/api/runner/location/consent", `{"accept":true}`, cookie, csrf)
	if w := s.crewCall("POST", "/api/runner/location", `{"lat":7.4,"lon":-80.2}`, cookie, csrf); w.Code != 200 {
		t.Fatalf("runner after notice: %d", w.Code)
	}
}
