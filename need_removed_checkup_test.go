package main

import (
	"encoding/json"
	"strings"
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

// Checkup answers are kept in the point's private log («Nadie contestó · no necesita nada» included) and follow the
// checkup when needs are removed right after it; public points never carry the log.
func TestCheckupAnswerLog(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","status":"OPEN"}`, 201)
	var p Point
	w := s.call("POST", "/api/admin/camps/"+s.camp+"/checkup", `{"result":"none"}`, true)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil || len(p.CheckupLog) != 1 || p.CheckupLog[0].Result != "Nadie contestó · no necesita nada" || p.CheckupLog[0].At != p.LastCheckup {
		t.Fatalf("checkup answer: %d %s", w.Code, w.Body.String())
	}
	for _, n := range s.needsOf(s.camp) {
		s.expect("PATCH", "/api/admin/needs/"+n.ID, `{"status":"DELETED"}`, 200)
	}
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, x := range pts {
		if x.ID == s.camp && (len(x.CheckupLog) != 1 || x.CheckupLog[0].At != x.Checkups[len(x.Checkups)-1]) {
			t.Fatalf("answer follows the checkup: %+v / %v", x.CheckupLog, x.Checkups)
		}
	}
	if pub := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pub, "checkupLog") || strings.Contains(pub, "Nadie contestó") {
		t.Fatalf("public points carry no checkup log: %s", pub)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/checkup", `{}`, true); w.Code != 200 || strings.Count(w.Body.String(), `"result"`) != 1 {
		t.Fatalf("a checkup without an answer adds no note: %s", w.Body.String())
	}
}
