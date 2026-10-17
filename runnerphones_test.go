package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// A runner never sees the phone of a runner from another team. A team member who reports a point with themselves as
// «Reportado por» puts their own phone in the point's contacts: only their team sees that phone (the name stays).
func TestRunnerPhonesOnlyForTheTeam(t *testing.T) {
	s := newServiceHarness(t)
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana Equipo A","telephone":"6111-2233","roles":["runner"],"accessCode":"ANAAA1"}`, 201)["id"].(string)
	beto := s.expect("POST", "/api/admin/users", `{"name":"Beto Equipo A","telephone":"6111-4455","roles":["runner"],"accessCode":"BETOA1"}`, 201)["id"].(string)
	caro := s.expect("POST", "/api/admin/users", `{"name":"Caro Equipo B","telephone":"6999-8877","roles":["runner"],"accessCode":"CAROB1"}`, 201)["id"].(string)
	s.team("Equipo A", ana, beto)
	s.team("Equipo B", caro)
	// Ana reports a point with herself as the reporter (her phone, another format) and a family contact.
	var p Point
	if w := s.call("POST", "/api/admin/points", `{"area":"Sigua","place":"Casa","coordinator":"Ana Equipo A","coordinatorPhone":"+507 6111 2233","members":[{"name":"Familia Pérez","phone":"6222-3344"}],"addedBy":"`+ana+`"}`, true); json.Unmarshal(w.Body.Bytes(), &p) != nil || p.ID == "" {
		t.Fatalf("point: %d %s", w.Code, w.Body.String())
	}
	s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Agua","urgency":"normal"}`, true)
	needs := func(code string) string {
		login := s.crewCall("POST", "/api/runner/session", `{"accessCode":"`+code+`"}`, nil, "")
		if login.Code != 200 {
			t.Fatalf("login %s: %d %s", code, login.Code, login.Body.String())
		}
		return s.crewCall("GET", "/api/runner/needs", "", login.Result().Cookies()[0], "").Body.String()
	}
	if body := needs("CAROB1"); strings.Contains(body, "6111") || !strings.Contains(body, "6222-3344") {
		t.Fatalf("another team's runner: no phone of Ana, but the family's phone stays: %s", body)
	}
	if body := needs("BETOA1"); !strings.Contains(body, "6111 2233") {
		t.Fatalf("Ana's teammate sees her phone: %s", body)
	}
	// Scout (team members of any team): same rule, the name stays.
	scout := func(code string) string { return s.as(s.viewerCookie(code), "GET", "/api/full/scout", "").Body.String() }
	if body := scout("CAROB1"); strings.Contains(body, "6111") || !strings.Contains(body, "Ana Equipo A") {
		t.Fatalf("scout, other team: name without phone: %s", body)
	}
	if body := scout("BETOA1"); !strings.Contains(body, "6111 2233") {
		t.Fatalf("scout, same team: %s", body)
	}
}
