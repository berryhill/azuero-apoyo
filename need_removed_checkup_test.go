package main

import (
	"encoding/json"
	"testing"
)

// «Ya no necesita nada» removes the need (DELETED) and records a checkup on the point, like the Checkup button.
func TestRemovingANeedCountsAsCheckup(t *testing.T) {
	s := newServiceHarness(t)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)
	var p Point
	pointOf := func() Point {
		var list []Point
		json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &list)
		for _, x := range list {
			if x.ID == s.camp {
				return x
			}
		}
		t.Fatalf("camp %s not found", s.camp)
		return Point{}
	}
	p = pointOf()
	checkups := len(p.Checkups)
	s.expect("PATCH", "/api/admin/needs/"+need["id"].(string), `{"status":"DELETED"}`, 200)
	p = pointOf()
	if p.LastCheckup == "" || len(p.Checkups) != checkups+1 || p.Checkups[len(p.Checkups)-1] != p.LastCheckup {
		t.Fatalf("removing the need records a checkup: %+v", p)
	}
	// Removing it again does not add another one.
	s.expect("PATCH", "/api/admin/needs/"+need["id"].(string), `{"status":"DELETED"}`, 200)
	if again := pointOf(); len(again.Checkups) != len(p.Checkups) {
		t.Fatalf("only the change to DELETED counts: %d → %d", len(p.Checkups), len(again.Checkups))
	}
}

// «Sí, están bien»: the Checkup click is logged, and removing the open needs right after does not add more entries.
func TestCheckupThenRemovingNeedsLogsOnce(t *testing.T) {
	s := newServiceHarness(t)
	a := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)
	b := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Ropa","urgency":"normal"}`, 201)
	after := s.expect("POST", "/api/admin/camps/"+s.camp+"/checkup", `{}`, 200)
	n := len(after["checkups"].([]any))
	s.expect("PATCH", "/api/admin/needs/"+a["id"].(string), `{"status":"DELETED"}`, 200)
	s.expect("PATCH", "/api/admin/needs/"+b["id"].(string), `{"status":"DELETED"}`, 200)
	var list []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &list)
	for _, p := range list {
		if p.ID == s.camp && len(p.Checkups) != n {
			t.Fatalf("one answer, one checkup: %d → %d", n, len(p.Checkups))
		}
	}
}
