package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// runnerCodeLogin enters Runners with an access code (a user with the runner role).
func (s *serviceHarness) runnerCodeLogin(code string) *runnerClient {
	s.t.Helper()
	w := s.call("POST", "/api/runner/session", `{"accessCode":"`+code+`"}`, false)
	if w.Code != 200 {
		s.t.Fatalf("runner code login: %d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	return &runnerClient{s: s, cookie: w.Result().Cookies()[0].Value, csrf: body["csrf"].(string)}
}

// A team manager sees the team's runs en camino in Runners («Mi equipo») and can finish a run for a teammate, e.g. a
// runner Supply added while dispatching (no code, cannot sign in); runners who are not team managers cannot.
func TestTeamManagerCompletesTeamRuns(t *testing.T) {
	s := newServiceHarness(t)
	julie := s.expect("POST", "/api/admin/users", `{"name":"Julie Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)["id"].(string)
	plain := s.expect("POST", "/api/admin/users", `{"name":"Rui Corre","telephone":"6999-3030","roles":["runner"],"accessCode":"RUI777"}`, 201)["id"].(string)
	ours := s.team("Unidos por Azuero", julie, plain)
	theirs := s.team("Otro")
	lead := s.expect("POST", "/api/admin/users", `{"name":"Lía Líder","telephone":"6999-2020","roles":["runner","teammanager"],"accessCode":"LIDER77","teamId":"`+ours+`"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/users", `{"name":"Otro Líder","telephone":"6999-4040","roles":["runner","teammanager"],"accessCode":"OTRO77","teamId":"`+theirs+`"}`, 201)
	jc, jt := s.supplierLogin("JULIE77")
	// Supply dispatches to a new runner: created in the center's team.
	w := s.agentCall("POST", "/api/supplier/dispatches", `{"pointId":"`+s.camp+`","runner":{"name":"Pucha"},"items":[{"item":"Agua","quantity":2}]}`, jc, jt)
	var res struct{ ID, RunnerID string }
	json.Unmarshal(w.Body.Bytes(), &res)
	if w.Code != 201 {
		t.Fatalf("dispatch: %d %s", w.Code, w.Body.String())
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	var leadTeam, puchaTeam string
	for _, u := range users {
		switch u.ID {
		case lead:
			leadTeam = u.TeamID
		case res.RunnerID:
			puchaTeam = u.TeamID
		}
	}
	if puchaTeam == "" || puchaTeam != leadTeam {
		t.Fatalf("the new runner is in the dispatching center's team: %q vs %q", puchaTeam, leadTeam)
	}
	// The team manager sees it under «teamRuns» with the runner's name; a plain runner and another team's manager do not.
	lc := s.runnerCodeLogin("LIDER77")
	if body := lc.call("GET", "/api/runner/needs", "").Body.String(); !strings.Contains(body, `"teamManager":true`) || !strings.Contains(body, `"runnerName":"Pucha"`) || !strings.Contains(body, res.ID) {
		t.Fatalf("team manager sees the team's runs: %s", body)
	}
	rc := s.runnerCodeLogin("RUI777")
	if body := rc.call("GET", "/api/runner/needs", "").Body.String(); strings.Contains(body, "teamRuns") || strings.Contains(body, "teamManager") {
		t.Fatalf("a plain runner gets no team runs: %s", body)
	}
	oc := s.runnerCodeLogin("OTRO77")
	if body := oc.call("GET", "/api/runner/needs", "").Body.String(); strings.Contains(body, res.ID) {
		t.Fatal("another team's manager does not see it")
	}
	if w := rc.call("POST", "/api/runner/needs/"+res.ID+"/complete", ``); w.Code != 409 {
		t.Fatalf("a plain runner cannot complete a teammate's run: %d", w.Code)
	}
	if w := oc.call("POST", "/api/runner/needs/"+res.ID+"/complete", ``); w.Code != 409 {
		t.Fatalf("another team's manager cannot complete it: %d", w.Code)
	}
	if w := lc.call("POST", "/api/runner/needs/"+res.ID+"/complete", ``); w.Code != 200 {
		t.Fatalf("team manager completes it: %d %s", w.Code, w.Body.String())
	}
	for _, n := range s.needsOf(s.camp) {
		if n.ID == res.ID && (n.Status != "COMPLETED" || n.UpdatedBy != lead) {
			t.Fatalf("completed by the team manager (recorded): %+v", n)
		}
	}
}
