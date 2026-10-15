package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// A pending report can be corrected before approval; it stays pending, hidden from the public, and gets no checkup stamp.
func TestEditPendingPointBeforeApproval(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("POST", "/api/submissions", `{"area":"La Colorada","needs":"Agua","coordinator":"Ana","phone":"6555-1234"}`, false); w.Code/100 != 2 {
		t.Fatalf("submit: %d %s", w.Code, w.Body.String())
	}
	var points []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &points)
	id := ""
	for _, p := range points {
		if p.Area == "La Colorada" && p.Status == "pending" {
			id = p.ID
		}
	}
	if id == "" {
		t.Fatalf("pending point not found: %+v", points)
	}
	got := s.expect("PATCH", "/api/admin/camps/"+id, `{"area":"La Colorada","place":"Escuela vieja","people":"17"}`, 200)
	if got["status"] != "pending" || got["place"] != "Escuela vieja" || got["lastCheckup"] != nil {
		t.Fatalf("edited pending point: %v", got)
	}
	if public := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(public, "Escuela vieja") {
		t.Fatalf("pending edits must stay private: %s", public)
	}
}
