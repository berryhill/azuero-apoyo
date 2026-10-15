package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «Rechazar» on a pending report archives it unpublished; «Recuperar punto» later publishes it like a first approval.
func TestRejectPendingArchivesIt(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/submissions", `{"area":"Nuario","place":"Rechazo Prueba","needs":"Agua","coordinatorPhone":"6000-4444"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/decisions", `{"id":"`+id+`","decision":"archived"}`, 200)
	point := func() Point {
		var list []Point
		json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &list)
		for _, p := range list {
			if p.ID == id {
				return p
			}
		}
		t.Fatalf("point %s missing", id)
		return Point{}
	}
	if p := point(); p.Status != "archived" || p.LastCheckup != "" {
		t.Fatalf("rejected pending → archived, never approved: %+v", p)
	}
	if strings.Contains(s.call("GET", "/api/points", "", false).Body.String(), "Rechazo Prueba") {
		t.Fatal("rejected report is not public")
	}
	if strings.Contains(s.call("GET", "/api/admin/camps/"+id+"/needs", "", true).Body.String(), "Agua") {
		t.Fatal("no need is created when rejecting")
	}
	s.expect("POST", "/api/admin/decisions", `{"id":"`+id+`","decision":"approved"}`, 200)
	if p := point(); p.Status != "approved" || p.LastCheckup == "" {
		t.Fatalf("recovered: %+v", p)
	}
	if needs := s.call("GET", "/api/admin/camps/"+id+"/needs", "", true).Body.String(); strings.Count(needs, `"need":"Agua"`) != 1 {
		t.Fatalf("recovering creates its need once: %s", needs)
	}
	// An approved point archived and recovered does not get a second need.
	s.expect("POST", "/api/admin/decisions", `{"id":"`+id+`","decision":"archived"}`, 200)
	s.expect("POST", "/api/admin/decisions", `{"id":"`+id+`","decision":"approved"}`, 200)
	if needs := s.call("GET", "/api/admin/camps/"+id+"/needs", "", true).Body.String(); strings.Count(needs, `"need":"Agua"`) != 1 {
		t.Fatalf("no duplicate need: %s", needs)
	}
}
