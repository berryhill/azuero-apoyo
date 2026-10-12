package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMapLinks(t *testing.T) {
	for _, link := range []string{"http://maps.google.com/?q=7.45,-80.35", "https://evil.example/?q=7.45,-80.35", "https://maps.google.com.evil.example/?q=7.45,-80.35"} {
		if _, err := parseMapLink(link); err == nil {
			t.Fatalf("accepted unsafe link %s", link)
		}
	}
	coords, err := parseMapLink("https://www.google.com/maps?q=7.454881%2C-80.3531883&z=17")
	if err != nil || len(coords) != 2 || coords[0] != 7.454881 || coords[1] != -80.3531883 {
		t.Fatalf("coordinates %v, %v", coords, err)
	}
}

func TestApprovalFlow(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	path := filepath.Join(t.TempDir(), "relief.db")
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	h := app.routes()
	call := func(method, path, body string, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		if method == "POST" {
			r.Header.Set("Content-Type", "application/json")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if csrf != "" {
			r.Header.Set("X-CSRF-Token", csrf)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	p := `{"code":"TEST-1","area":"Cañas","place":"Escuela","needs":"Agua","people":"12","gps":"https://maps.google.com/?q=7.45,-80.35","contact":"Private Person","phone":"0000","adults":"6","children":"4","babies":"2","olderAdults":"1","disabled":"1","secondContact":"Backup","serviceNotes":"Revisar"}`
	sub := call("POST", "/api/submissions", p, nil, "")
	if sub.Code != 201 {
		t.Fatalf("submit %d %s", sub.Code, sub.Body.String())
	}
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	if got := call("GET", "/api/points", "", nil, ""); got.Body.String() != "[]\n" {
		t.Fatalf("premature publish: %s", got.Body.String())
	}
	if got := call("GET", "/api/admin/submissions", "", nil, ""); got.Code != 401 {
		t.Fatalf("unauthenticated queue: %d", got.Code)
	}
	login := call("POST", "/api/admin/login", `{"password":"test-password-only-12345"}`, nil, "")
	if login.Code != 200 {
		t.Fatalf("login: %d %s", login.Code, login.Body.String())
	}
	var auth map[string]string
	json.Unmarshal(login.Body.Bytes(), &auth)
	cookie := login.Result().Cookies()[0]
	full := call("GET", "/api/admin/points", "", cookie, "")
	var all []Point
	if err := json.Unmarshal(full.Body.Bytes(), &all); err != nil || len(all) != 1 || all[0].Contact != "Private Person" || len(all[0].Coords) != 2 || all[0].Coords[0] != 7.45 {
		t.Fatalf("admin data: %s %v", full.Body.String(), err)
	}
	if got := call("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, cookie, ""); got.Code != 403 {
		t.Fatalf("missing CSRF: %d", got.Code)
	}
	decision := call("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, cookie, auth["csrf"])
	if decision.Code != 200 {
		t.Fatalf("decision %d %s", decision.Code, decision.Body.String())
	}
	points := call("GET", "/api/points", "", nil, "")
	var list []Point
	if err := json.Unmarshal(points.Body.Bytes(), &list); err != nil || len(list) != 1 || list[0].Status != "approved" || list[0].Contact != "" || list[0].Phone != "" || list[0].SecondContact != "" || list[0].GPS == "" {
		t.Fatalf("published: %s %v", points.Body.String(), err)
	}
	rotated := call("POST", "/api/admin/password", `{"password":"rotated-test-only-45678"}`, cookie, auth["csrf"])
	if rotated.Code != 200 {
		t.Fatalf("rotate: %d %s", rotated.Code, rotated.Body.String())
	}
	if got := call("GET", "/api/admin/points", "", cookie, ""); got.Code != 401 {
		t.Fatalf("session survived rotation: %d", got.Code)
	}
	if got := call("POST", "/api/admin/login", `{"password":"test-password-only-12345"}`, nil, ""); got.Code != 401 {
		t.Fatalf("old password accepted: %d", got.Code)
	}
	if got := call("POST", "/api/admin/login", `{"password":"rotated-test-only-45678"}`, nil, ""); got.Code != 200 {
		t.Fatalf("new password rejected: %d", got.Code)
	}
	if err := app.db.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "")
	again, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer again.db.Close()
	persisted := httptest.NewRecorder()
	again.routes().ServeHTTP(persisted, httptest.NewRequest("GET", "/api/points", nil))
	var check []Point
	json.Unmarshal(persisted.Body.Bytes(), &check)
	if len(check) != 1 {
		t.Fatalf("not persisted: %s", persisted.Body.String())
	}
}
