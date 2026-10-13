package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type serviceHarness struct {
	t      *testing.T
	app    *App
	h      http.Handler
	cookie *http.Cookie
	csrf   string
	camp   string
}

func newServiceHarness(t *testing.T) *serviceHarness {
	t.Helper()
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.db.Close() })
	s := &serviceHarness{t: t, app: app, h: app.routes()}
	login := s.call("POST", "/api/admin/login", `{"password":"test-password-only-12345"}`, false)
	if login.Code != 200 {
		t.Fatalf("login: %s", login.Body.String())
	}
	s.cookie = login.Result().Cookies()[0]
	var auth map[string]string
	json.Unmarshal(login.Body.Bytes(), &auth)
	s.csrf = auth["csrf"]
	created := s.call("POST", "/api/admin/points", `{"code":"C-1","area":"Venao","place":"Camp","contact":"Private camp","phone":"6555-1234"}`, true)
	if created.Code != 201 {
		t.Fatalf("camp create: %s", created.Body.String())
	}
	var p map[string]string
	json.Unmarshal(created.Body.Bytes(), &p)
	s.camp = p["id"]
	return s
}
func (s *serviceHarness) call(method, path, body string, auth bool) *httptest.ResponseRecorder {
	s.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if method == "POST" || method == "PATCH" {
		r.Header.Set("Content-Type", "application/json")
	}
	if auth && s.cookie != nil {
		r.AddCookie(s.cookie)
		r.Header.Set("X-CSRF-Token", s.csrf)
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}
func (s *serviceHarness) expect(method, path, body string, status int) map[string]any {
	s.t.Helper()
	w := s.call(method, path, body, true)
	if w.Code != status {
		s.t.Fatalf("%s %s: got %d %s want %d", method, path, w.Code, w.Body.String(), status)
	}
	var v map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		s.t.Fatal(err)
	}
	return v
}
func TestCampServiceRunnerCreateAndList(t *testing.T) {
	s := newServiceHarness(t)
	r := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	if r["name"] != "Ana" || r["telephone"] != "6999-1111" || r["id"] == "" {
		t.Fatalf("runner: %v", r)
	}
	got := s.call("GET", "/api/admin/runners", "", true)
	if got.Code != 200 || !bytes.Contains(got.Body.Bytes(), []byte("6999-1111")) {
		t.Fatalf("list: %d %s", got.Code, got.Body.String())
	}
}
func TestCampServiceNeedsAndPublicPrivacy(t *testing.T) {
	s := newServiceHarness(t)
	runner := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	id := runner["id"].(string)
	n := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Bring potable water","status":"OPEN","urgency":"Today","runnerId":"`+id+`"}`, 201)
	if n["need"] != "Bring potable water" || n["runnerId"] != id {
		t.Fatalf("need: %v", n)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false)
	if pub.Code != 200 || !bytes.Contains(pub.Body.Bytes(), []byte("Bring potable water")) || bytes.Contains(pub.Body.Bytes(), []byte("6999-1111")) || bytes.Contains(pub.Body.Bytes(), []byte("runnerId")) {
		t.Fatalf("public: %d %s", pub.Code, pub.Body.String())
	}
	points := s.call("GET", "/api/points", "", false)
	if !bytes.Contains(points.Body.Bytes(), []byte(`"campStatus":"OPEN"`)) || !bytes.Contains(points.Body.Bytes(), []byte(`"highestUrgency":"Today"`)) || bytes.Contains(points.Body.Bytes(), []byte("6999-1111")) {
		t.Fatalf("aggregate: %s", points.Body.String())
	}
}
func TestCampServiceWriteProtectionAndValidation(t *testing.T) {
	s := newServiceHarness(t)
	for _, path := range []string{"/api/admin/runners", "/api/admin/camps/" + s.camp + "/needs"} {
		if w := s.call("POST", path, `{}`, false); w.Code != 401 {
			t.Fatalf("no login %s: %d", path, w.Code)
		}
		r := httptest.NewRequest("POST", path, strings.NewReader(`{}`))
		r.AddCookie(s.cookie)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("no csrf %s: %d", path, w.Code)
		}
	}
	if w := s.call("POST", "/api/admin/runners", `{"name":"","telephone":"123"}`, true); w.Code != 400 {
		t.Fatalf("invalid runner: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"","urgency":"Now"}`, true); w.Code != 400 {
		t.Fatalf("invalid need: %d %s", w.Code, w.Body.String())
	}
}
func TestCampServicePatchAggregationAndCampEdit(t *testing.T) {
	s := newServiceHarness(t)
	first := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Tomorrow"}`, 201)
	second := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Medicine","urgency":"Urgent"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+first["id"].(string), `{"status":"IN_PROGRESS"}`, 200)
	s.expect("PATCH", "/api/admin/needs/"+second["id"].(string), `{"status":"COMPLETED"}`, 200)
	points := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"IN_PROGRESS"`) || !strings.Contains(points, `"highestUrgency":"Tomorrow"`) {
		t.Fatalf("aggregate: %s", points)
	}
	s.expect("PATCH", "/api/admin/needs/"+first["id"].(string), `{"status":"DELETED"}`, 200)
	points = s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"COMPLETED"`) || strings.Contains(points, "highestUrgency") {
		t.Fatalf("completed: %s", points)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if strings.Contains(pub, "Water") {
		t.Fatalf("deleted exposed: %s", pub)
	}
	edited := s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"note":"Private note","coordinator":"Lead","phone":"6000-2222","place":"New place"}`, 200)
	if edited["note"] != "Private note" || edited["area"] != "Venao" || edited["place"] != "New place" {
		t.Fatalf("partial edit: %v", edited)
	}
	points = s.call("GET", "/api/points", "", false).Body.String()
	if strings.Contains(points, "Private note") || strings.Contains(points, "6000-2222") {
		t.Fatalf("private exposed: %s", points)
	}
}
func TestCampServicePatchAndNoNeeds(t *testing.T) {
	s := newServiceHarness(t)
	points := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"NO_NEEDS"`) {
		t.Fatalf("empty: %s", points)
	}
	r := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	runnerID := r["id"].(string)
	req := httptest.NewRequest("PATCH", "/api/admin/runners/"+runnerID, strings.NewReader(`{"telephone":"6999-2222"}`))
	req.AddCookie(s.cookie)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("PATCH without csrf: %d %s", w.Code, w.Body.String())
	}
	edited := s.expect("PATCH", "/api/admin/runners/"+runnerID, `{"telephone":"6999-2222"}`, 200)
	if edited["name"] != "Ana" || edited["telephone"] != "6999-2222" {
		t.Fatalf("runner edit: %v", edited)
	}
	n := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Today"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+n["id"].(string), `{"status":"DELETED"}`, 200)
	points = s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"NO_NEEDS"`) || strings.Contains(points, "highestUrgency") {
		t.Fatalf("deleted aggregate: %s", points)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Bad"}`, true); w.Code != 400 {
		t.Fatalf("invalid urgency: %d %s", w.Code, w.Body.String())
	}
}
func TestCampServiceCampEditFieldsAndReadOnly(t *testing.T) {
	s := newServiceHarness(t)
	edited := s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"secondContact":"Backup","secondPhone":"6111-1111","thirdContact":"Other","thirdPhone":"6222-2222","gps":"https://maps.google.com/?q=7.45,-80.35","coords":[7.45,-80.35],"note":"operations only"}`, 200)
	if edited["secondPhone"] != "6111-1111" || edited["thirdPhone"] != "6222-2222" || edited["note"] != "operations only" {
		t.Fatalf("edit: %v", edited)
	}
	if w := s.call("PATCH", "/api/admin/camps/"+s.camp, `{"status":"archived"}`, true); w.Code != 400 {
		t.Fatalf("status override: %d %s", w.Code, w.Body.String())
	}
	public := s.call("GET", "/api/points", "", false).Body.String()
	for _, secret := range []string{"operations only", "6111-1111", "6222-2222"} {
		if strings.Contains(public, secret) {
			t.Fatalf("leaked %s: %s", secret, public)
		}
	}
}
func TestCampServiceSeedRefreshPreservesCampNote(t *testing.T) {
	s := newServiceHarness(t)
	seeded := s.call("POST", "/api/admin/seed", `[{"code":"S-2","area":"Original"}]`, true)
	if seeded.Code != 200 {
		t.Fatalf("seed: %s", seeded.Body.String())
	}
	s.expect("PATCH", "/api/admin/camps/seed:s-2", `{"note":"Keep this note"}`, 200)
	refreshed := s.call("POST", "/api/admin/seed", `[{"code":"S-2","area":"Updated"}]`, true)
	if refreshed.Code != 200 {
		t.Fatalf("refresh: %s", refreshed.Body.String())
	}
	all := s.call("GET", "/api/admin/points", "", true).Body.String()
	if !strings.Contains(all, "Keep this note") {
		t.Fatalf("note lost: %s", all)
	}
}
func TestCampServicePersists(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Fuel","urgency":"Eventually"}`, 201)
	path := s.app.db.Path()
	s.app.db.Close()
	reopened, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.db.Close()
	w := httptest.NewRecorder()
	reopened.routes().ServeHTTP(w, httptest.NewRequest("GET", "/api/camps/"+s.camp+"/needs", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Fuel") {
		t.Fatalf("persisted: %d %s", w.Code, w.Body.String())
	}
}
func TestAdminNeedsListAllCamps(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("GET", "/api/admin/needs", "", false); w.Code != 401 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","status":"OPEN","urgency":"Tomorrow"}`, 201)
	gone := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Old","status":"OPEN","urgency":"Today"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+gone["id"].(string), `{"status":"DELETED"}`, 200)
	w := s.call("GET", "/api/admin/needs", "", true)
	var list []CampNeed
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list) != 1 || list[0].CampID != s.camp || list[0].Urgency != "Tomorrow" {
		t.Fatalf("needs: %d %s", w.Code, w.Body.String())
	}
}
func TestPublicPointsListNeedStatuses(t *testing.T) {
	s := newServiceHarness(t)
	if points := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(points, "needStatuses") {
		t.Fatalf("no needs: %s", points)
	}
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","status":"OPEN","urgency":"Today"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Food","status":"COMPLETED","urgency":"Tomorrow"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Old","status":"DELETED","urgency":"Today"}`, 201)
	if points := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(points, `"needStatuses":["OPEN","COMPLETED"]`) {
		t.Fatalf("statuses: %s", points)
	}
	if w := s.call("PATCH", "/api/admin/camps/"+s.camp, `{"needStatuses":["OPEN"]}`, true); w.Code != 400 {
		t.Fatalf("read-only: %d %s", w.Code, w.Body.String())
	}
}
