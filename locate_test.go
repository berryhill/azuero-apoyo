package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «📍 Por ubicar»: a delivery imported with no destination is kept (hidden from everything public) until someone
// involved — the runner, a teammate — or admin gives it a place or merges it into an existing point.
func TestToLocateDeliveries(t *testing.T) {
	s := newServiceHarness(t)
	s.team("Unidos por Azuero")
	outsider := s.expect("POST", "/api/admin/users", `{"name":"Otra Persona","telephone":"6999-4444","roles":["runner"],"accessCode":"OTRA12"}`, 201)["id"].(string)
	s.team("Otro", outsider)
	newLost := func(beneficiary string) (Point, CampNeed) {
		var p Point
		w := s.call("POST", "/api/admin/points", `{"area":"Por ubicar","place":"`+beneficiary+`","toLocate":true,"members":[{"name":"`+beneficiary+`"}]}`, true)
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatalf("to-locate point: %d %s", w.Code, w.Body.String())
		}
		var n CampNeed
		w = s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Alimentos","urgency":"normal","status":"COMPLETED","runnerName":"Pucha","runnerTeam":"Unidos por Azuero","at":"2026-10-14T12:00:00-05:00"}`, true)
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &n) != nil {
			t.Fatalf("delivery: %d %s", w.Code, w.Body.String())
		}
		return p, n
	}
	p, n := newLost("Familia Abdiel Nieto")
	// Hidden from everything public.
	if b := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(b, p.ID) {
		t.Fatal("a to-locate point is not public")
	}
	if w := s.call("GET", "/api/camps/"+p.ID+"/needs", "", false); w.Code != 404 {
		t.Fatalf("public needs of a to-locate point: %d", w.Code)
	}
	// The runner (created in the team by the import) can locate it; someone from another team can't.
	code := s.call("GET", "/api/admin/users/"+n.RunnerID+"/access-code", "", true)
	var c struct{ AccessCode string }
	json.Unmarshal(code.Body.Bytes(), &c)
	if b := s.as(s.viewerCookie(c.AccessCode), "GET", "/api/full/tolocate", "").Body.String(); !strings.Contains(b, p.ID) || !strings.Contains(b, "Familia Abdiel Nieto") {
		t.Fatalf("the runner sees it in «por ubicar»: %s", b)
	}
	if b := strings.TrimSpace(s.as(s.viewerCookie("OTRA12"), "GET", "/api/full/tolocate", "").Body.String()); b != "[]" {
		t.Fatalf("someone not involved sees nothing: %s", b)
	}
	if w := s.as(s.viewerCookie("OTRA12"), "POST", "/api/full/camps/"+p.ID+"/locate", `{"area":"Río Viejo"}`); w.Code != 403 {
		t.Fatalf("not involved: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(s.viewerCookie(c.AccessCode), "POST", "/api/full/camps/"+p.ID+"/locate", `{"area":"Río Viejo","place":"Casa de los Nieto"}`); w.Code != 200 {
		t.Fatalf("runner locates it: %d %s", w.Code, w.Body.String())
	}
	if b := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(b, p.ID) || !strings.Contains(b, "Casa de los Nieto") {
		t.Fatal("once located it's a normal public point")
	}
	if w := s.as(s.viewerCookie(c.AccessCode), "POST", "/api/full/camps/"+p.ID+"/locate", `{"area":"X"}`); w.Code != 404 {
		t.Fatalf("already located: %d", w.Code)
	}
	// A teammate merges another one into an existing point: its delivery and contacts move, it's archived.
	q, qn := newLost("Omaira")
	mate := s.expect("POST", "/api/admin/users", `{"name":"Compa Equipo","telephone":"6999-5555","roles":["runner"],"accessCode":"COMPA1"}`, 201)["id"].(string)
	s.expect("PATCH", "/api/admin/users/"+mate, `{"teamId":"`+teamIDByName(t, s, "Unidos por Azuero")+`"}`, 200)
	if w := s.as(s.viewerCookie("COMPA1"), "POST", "/api/full/camps/"+q.ID+"/locate", `{"targetId":"`+s.camp+`"}`); w.Code != 200 {
		t.Fatalf("teammate merges it: %d %s", w.Code, w.Body.String())
	}
	moved := false
	for _, x := range s.needsOf(s.camp) {
		if x.ID == qn.ID {
			moved = true
		}
	}
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, x := range pts {
		if x.ID == q.ID && (x.Status != "archived" || x.MergedInto != s.camp) {
			t.Fatalf("merged point archived: %+v", x)
		}
		if x.ID == s.camp && !strings.Contains(func() string { b, _ := json.Marshal(x.Members); return string(b) }(), "Omaira") {
			t.Fatalf("contacts merged: %+v", x.Members)
		}
	}
	if !moved {
		t.Fatal("delivery moved to the existing point")
	}
}

func teamIDByName(t *testing.T, s *serviceHarness, name string) string {
	var teams []Team
	json.Unmarshal(s.call("GET", "/api/admin/teams", "", true).Body.Bytes(), &teams)
	for _, x := range teams {
		if x.Name == name {
			return x.ID
		}
	}
	t.Fatal("no team " + name)
	return ""
}
