package main

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestSparsePointSubmission(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	h := app.routes()
	req := httptest.NewRequest("POST", "/api/submissions", strings.NewReader(`{"needs":"Insulina","gps":"https://maps.apple.com/?q=7.45,-80.35","unexpected":"ignored"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("sparse submission rejected: %d %s", w.Code, w.Body.String())
	}
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result["status"] != "pending" {
		t.Fatalf("result: %s %v", w.Body.String(), err)
	}
	login := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{"password":"test-password-only-12345"}`))
	login.Header.Set("Content-Type", "application/json")
	auth := httptest.NewRecorder()
	h.ServeHTTP(auth, login)
	if auth.Code != 200 {
		t.Fatal(auth.Body.String())
	}
	queue := httptest.NewRequest("GET", "/api/admin/submissions", nil)
	queue.AddCookie(auth.Result().Cookies()[0])
	records := httptest.NewRecorder()
	h.ServeHTTP(records, queue)
	var rows []Point
	if err := json.Unmarshal(records.Body.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].Code == "" || rows[0].Area == "" || rows[0].Needs != "Insulina" || rows[0].GPS == "" {
		t.Fatalf("sparse data lost: %s %v", records.Body.String(), err)
	}
}
