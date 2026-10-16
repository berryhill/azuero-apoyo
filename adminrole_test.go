package main

import (
	"strings"
	"testing"
)

// The «admin» role is given and removed only from the admin panel (master password): a team manager can't grant
// it, keeps it untouched when changing a member's other roles, and agents have no user endpoints at all.
func TestAdminRoleOnlyFromAdminPanel(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	mia := user(`{"name":"Mia Manager","telephone":"6999-1111","roles":["runner"],"accessCode":"MIA777"}`)
	leo := user(`{"name":"Leo Runner","telephone":"6999-2222","roles":["runner"],"accessCode":"LEO777"}`)
	ana := user(`{"name":"Ana Admin","telephone":"6999-3333","roles":["runner","admin"],"accessCode":"ANA777"}`)
	user(`{"name":"Gus Agente","telephone":"6999-4444","roles":["agent"],"accessCode":"GUS777"}`)
	s.team("Norte", mia, leo, ana)
	s.expect("PATCH", "/api/admin/users/"+mia, `{"roles":["runner","teammanager"]}`, 200)
	miaC := s.viewerCookie("MIA777")
	if w := s.as(miaC, "PATCH", "/api/full/team/members/"+leo, `{"roles":["runner","admin"]}`); w.Code != 400 || !strings.Contains(w.Body.String(), "solo lo asigna administración") {
		t.Fatalf("team manager grants admin: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(miaC, "PATCH", "/api/full/team/members/"+ana, `{"roles":["builder"]}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"admin"`) {
		t.Fatalf("team manager edits other roles; admin stays: %d %s", w.Code, w.Body.String())
	}
	gusC, csrf := s.agentLogin("GUS777")
	if w := s.agentCall("PATCH", "/api/agent/users/"+leo, `{"roles":["runner","admin"]}`, gusC, csrf); w.Code == 200 {
		t.Fatalf("an agent changed roles: %d", w.Code)
	}
	if body := s.expect("PATCH", "/api/admin/users/"+leo, `{"roles":["runner","admin"]}`, 200); !strings.Contains(strings.Join(anyStrings(body["roles"]), ","), "admin") {
		t.Fatalf("admin panel grants admin: %+v", body)
	}
	if w := s.as(s.viewerCookie("ANA777"), "GET", "/api/full/me", ""); !strings.Contains(w.Body.String(), `"admin"`) {
		t.Fatalf("the role reaches the page (nav link): %s", w.Body.String())
	}
}

func anyStrings(v any) []string {
	out := []string{}
	for _, x := range v.([]any) {
		out = append(out, x.(string))
	}
	return out
}
