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
	req := httptest.NewRequest("POST", "/api/submissions", strings.NewReader(`{"needs":"Insulina","secondPhone":"6555-1234","gps":"https://maps.apple.com/?q=7.45,-80.35","unexpected":"ignored"}`))
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

func TestPointSubmissionErrorsIdentifyTheProblem(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	call := func(body string) (int, map[string]string) {
		r := httptest.NewRequest("POST", "/api/submissions", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		app.routes().ServeHTTP(w, r)
		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return w.Code, result
	}
	longNeeds, _ := json.Marshal(map[string]string{"code": "AB-5", "area": "Agua Buena", "phone": "6555-1234", "needs": strings.Repeat("Suministros y apoyo de salud. ", 100)})
	if status, result := call(string(longNeeds)); status != 201 {
		t.Fatalf("long needs rejected: %d %v", status, result)
	}
	if status, result := call(`{"people":34}`); status != 400 || result["field"] != "people" || result["error"] == "invalid JSON" {
		t.Fatalf("type error not actionable: %d %v", status, result)
	}
	if status, result := call(`{"coords":[0]}`); status != 400 || result["field"] != "coords" {
		t.Fatalf("coordinate error not actionable: %d %v", status, result)
	}
	for _, body := range []string{`{"needs":"Agua"}`, `{"needs":"Agua","phone":"  ","secondPhone":"12-34"}`} {
		if status, result := call(body); status != 400 || result["field"] != "phone" || !strings.Contains(result["error"], "teléfono") {
			t.Fatalf("missing contact phone accepted: %s → %d %v", body, status, result)
		}
	}
	if status, result := call(`{"needs":"Agua","thirdPhone":"+507 6555 1234"}`); status != 201 {
		t.Fatalf("any one contact phone must be enough: %d %v", status, result)
	}
	if status, result := call(`{"needs":"` + strings.Repeat("a", 40000) + `"}`); status != 413 || !strings.Contains(result["error"], "32 KB") {
		t.Fatalf("oversize error not actionable: %d %v", status, result)
	}
}

func TestPublicCoordinatorWithoutPrivateContacts(t *testing.T) {
	p := publicPoint(Point{Code: "CAMP-1", Coordinator: "Coordinación Agua Buena", Contact: "Contacto privado", Phone: "6670-0000", SecondPhone: "6492-0000"})
	if p.Coordinator != "Coordinación Agua Buena" || p.Contact != "" || p.Phone != "" || p.SecondPhone != "" {
		t.Fatalf("public coordinator projection: %+v", p)
	}
	if got := publicPoint(Point{Coordinator: "Contacto 6670-0000"}); got.Coordinator != "" {
		t.Fatal("phone leaked through coordinator")
	}
}
