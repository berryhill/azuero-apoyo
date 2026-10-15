package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Importing reports: admin creates the point pending (?pending=1) and credits the reporter's user (addedBy).
func TestAdminImportPendingWithReporter(t *testing.T) {
	s := newServiceHarness(t)
	user := s.expect("POST", "/api/admin/users", `{"name":"Arianne Prueba","telephone":"6000-2222","roles":["runner"]}`, 201)
	id := user["id"].(string)
	got := s.expect("POST", "/api/admin/points?pending=1", `{"area":"Tonosí","place":"El Espavé","needs":"Agua","note":"Importado de arianne","addedBy":"`+id+`"}`, 201)
	if got["status"] != "pending" {
		t.Fatalf("imported point stays for review: %v", got)
	}
	var list []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &list)
	var p Point
	for _, x := range list {
		if x.ID == got["id"] {
			p = x
		}
	}
	if p.Status != "pending" || p.AddedBy != id || p.AddedByName != "Arianne Prueba" || p.Note != "Importado de arianne" {
		t.Fatalf("pending point credited to the reporter: %+v", p)
	}
	if pub := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pub, "El Espavé") {
		t.Fatalf("pending import is not public: %s", pub)
	}
	if w := s.call("POST", "/api/admin/points", `{"area":"Tonosí","addedBy":"no-such-user"}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "addedBy") {
		t.Fatalf("unknown reporter: %d %s", w.Code, w.Body.String())
	}
	// The public form can't credit anyone.
	pubID := s.expect("POST", "/api/submissions", `{"area":"Nuario","coordinatorPhone":"6000-3333","addedBy":"`+id+`"}`, 201)["id"]
	var fresh []Point // nueva lista: reutilizar la anterior arrastra campos vacíos de otro punto
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &fresh)
	for _, x := range fresh {
		if x.ID == pubID && x.AddedBy != "" {
			t.Fatalf("public submission set addedBy: %+v", x)
		}
	}
}
