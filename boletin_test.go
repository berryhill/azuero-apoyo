package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBoletinSnapshotIsAdminOnlyAndPublicSafe(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"people":"30 personas","service":"YES","needs":"agua, insulina"}`, 200)
	s.call("POST", "/api/admin/points", `{"code":"C-2","area":"Flores","people":"12","contact":"Otra persona","phone":"6777-0000"}`, true)
	if w := s.call("POST", "/api/admin/boletines", `{}`, false); w.Code != 401 {
		t.Fatalf("generate without login: %d", w.Code)
	}
	if w := s.call("GET", "/api/boletines/latest", "", false); w.Code != 404 {
		t.Fatalf("no boletín yet: %d", w.Code)
	}
	for _, body := range []string{`{}`, `{"accessCode":"PRINT4"}`} {
		if w := s.call("POST", "/api/admin/boletines", body, true); w.Code != 403 || !strings.Contains(w.Body.String(), "Código de acceso incorrecto") {
			t.Fatalf("boletín without valid code %s: %d %s", body, w.Code, w.Body.String())
		}
	}
	if w := s.call("POST", "/api/admin/boletin-access", `{"code":"nope"}`, true); w.Code != 403 {
		t.Fatalf("wrong access code: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/boletin-access", `{"code":" print5 "}`, false); w.Code != 401 {
		t.Fatalf("access check must require admin session: %d", w.Code)
	}
	s.expect("POST", "/api/admin/boletin-access", `{"code":" print5 "}`, 200)
	first := s.expect("POST", "/api/admin/boletines", `{"donor":"No enviar más tanques.","accessCode":"PRINT5"}`, 201)
	if first["n"].(float64) != 1 {
		t.Fatalf("first: %v", first)
	}
	s.expect("POST", "/api/admin/boletines", `{"accessCode":"PRINT5"}`, 201)
	w := s.call("GET", "/api/boletines/1", "", false)
	body := w.Body.String()
	if w.Code != 200 || strings.Contains(body, "6555-1234") || strings.Contains(body, "6777-0000") || strings.Contains(body, "Private camp") || strings.Contains(body, "Otra persona") || !strings.Contains(body, "No enviar más tanques.") {
		t.Fatalf("snapshot: %d %s", w.Code, body)
	}
	var b Boletin
	json.Unmarshal(w.Body.Bytes(), &b)
	if b.Records != 2 || b.People != 42 || b.Attended != 30 || len(b.Points) != 2 {
		t.Fatalf("totals: %+v", b)
	}
	var list []Boletin
	json.Unmarshal(s.call("GET", "/api/boletines", "", false).Body.Bytes(), &list)
	if len(list) != 2 || list[1].N != 2 || list[0].Points != nil || list[0].Donor != "" {
		t.Fatalf("history: %+v", list)
	}
	if latest := s.call("GET", "/api/boletines/latest", "", false).Body.String(); !strings.Contains(latest, `"n":2`) {
		t.Fatalf("latest: %s", latest)
	}
	if page := s.call("GET", "/report/boletin_general.html", "", false); page.Code != 200 {
		t.Fatalf("report page: %d", page.Code)
	}
}
