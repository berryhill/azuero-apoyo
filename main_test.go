package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

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
	p := `{"code":"TEST-1","area":"Cañas","place":"Escuela","needs":"Agua","people":"12","coords":[7.45,-80.35]}`
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
	if got := call("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, cookie, ""); got.Code != 403 {
		t.Fatalf("missing CSRF: %d", got.Code)
	}
	decision := call("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, cookie, auth["csrf"])
	if decision.Code != 200 {
		t.Fatalf("decision %d %s", decision.Code, decision.Body.String())
	}
	points := call("GET", "/api/points", "", nil, "")
	var list []Point
	if err := json.Unmarshal(points.Body.Bytes(), &list); err != nil || len(list) != 1 || list[0].Status != "approved" {
		t.Fatalf("published: %s %v", points.Body.String(), err)
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
