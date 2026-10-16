package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// User notes: admin and the team's manager write them; the team's members read them in «Mi equipo»; nobody else sees
// them. Imports link reports to people by phone (or a unique name) and create the ones that don't exist yet.
func TestUserNotesAndImportedUsers(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	mia := user(`{"name":"Mia Manager","telephone":"6999-1111","roles":["runner"],"accessCode":"MIA123"}`)
	leo := user(`{"name":"Leo Runner","telephone":"6999-2222","roles":["runner"],"accessCode":"LEO123","notes":"Tiene pickup 4x4"}`)
	out := user(`{"name":"Otro Equipo","telephone":"6999-3333","roles":["runner"],"accessCode":"OTRO12"}`)
	s.team("Norte", mia, leo)
	s.team("Sur", out)
	s.expect("PATCH", "/api/admin/users/"+mia, `{"roles":["runner","teammanager"]}`, 200)
	// «Mi equipo» is for the team manager (and the Admin role) only: a plain member gets 403 and can't edit notes.
	if w := s.as(s.viewerCookie("LEO123"), "GET", "/api/full/team", ""); w.Code != 403 {
		t.Fatalf("member can't open Mi equipo: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(s.viewerCookie("MIA123"), "GET", "/api/full/team", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Tiene pickup 4x4") || !strings.Contains(w.Body.String(), `"manager":true`) {
		t.Fatalf("manager sees members with notes: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(s.viewerCookie("LEO123"), "PATCH", "/api/full/team/members/"+leo+"/notes", `{"notes":"x"}`); w.Code != 403 {
		t.Fatalf("a member can't edit notes: %d", w.Code)
	}
	// The manager edits a member's notes and their own; not someone on another team.
	mc := s.viewerCookie("MIA123")
	if w := s.as(mc, "PATCH", "/api/full/team/members/"+leo+"/notes", `{"notes":"Pickup 4x4, disponible de noche"}`); w.Code != 200 {
		t.Fatalf("manager edits notes: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(mc, "PATCH", "/api/full/team/members/"+mia+"/notes", `{"notes":"Coordina Norte"}`); w.Code != 200 {
		t.Fatalf("manager edits own notes: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(mc, "PATCH", "/api/full/team/members/"+out+"/notes", `{"notes":"x"}`); w.Code != 404 {
		t.Fatalf("other team's member: %d", w.Code)
	}
	// Another team doesn't see them; admin does.
	if b := s.as(s.viewerCookie("OTRO12"), "GET", "/api/full/team", "").Body.String(); strings.Contains(b, "Pickup 4x4") || strings.Contains(b, "Tiene pickup") {
		t.Fatalf("other team sees notes: %s", b)
	}
	if b := s.call("GET", "/api/admin/users", "", true).Body.String(); !strings.Contains(b, "Pickup 4x4, disponible de noche") || !strings.Contains(b, "Coordina Norte") {
		t.Fatalf("admin sees notes: %s", b)
	}
	// Import (with the team «Norte»): a team member's phone links (any format), a unique name in the team links, a new
	// person is created; someone from another team is never linked, and without a team nobody is.
	imp := func(body string) Point {
		var p Point
		w := s.call("POST", "/api/admin/points?pending=1", body, true)
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatalf("import %s: %d %s", body, w.Code, w.Body.String())
		}
		var subs []Point
		json.Unmarshal(s.call("GET", "/api/admin/submissions", "", true).Body.Bytes(), &subs)
		for _, x := range subs {
			if x.ID == p.ID {
				return x
			}
		}
		t.Fatalf("imported point %s not in Por revisar", p.ID)
		return p
	}
	if p := imp(`{"area":"A","addedByName":"Leo R.","addedByPhone":"+507 6999 2222","addedByTeam":"Norte"}`); p.AddedBy != leo || p.AddedByPhone != "" {
		t.Fatalf("linked by phone: %+v", p)
	}
	n1 := imp(`{"area":"B","addedByName":"Rosa Nueva","addedByPhone":"6111-0000","addedByTeam":"Norte"}`)
	n2 := imp(`{"area":"C","addedByName":"Rosa N.","addedByPhone":"61110000","addedByTeam":"Norte"}`)
	if n1.AddedBy == "" || n1.AddedBy != n2.AddedBy || n1.AddedByName != "Rosa Nueva" {
		t.Fatalf("new person created once and reused: %+v / %+v", n1, n2)
	}
	if p := imp(`{"area":"D","addedByName":"mia manager","addedByTeam":"Norte"}`); p.AddedBy != mia {
		t.Fatalf("linked by unique name: %+v", p)
	}
	if p := imp(`{"area":"E","addedByName":"Otro Equipo","addedByPhone":"6999-3333","addedByTeam":"Norte"}`); p.AddedBy == out || p.AddedBy == "" {
		t.Fatalf("another team's member is never linked: %+v", p)
	}
	if p := imp(`{"area":"F","addedByName":"Leo Runner","addedByPhone":"6999-2222"}`); p.AddedBy == leo || p.AddedBy == "" {
		t.Fatalf("without a team nobody is linked: %+v", p)
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		if u.ID == n1.AddedBy && (u.TeamID == "" || !u.HasAccessCode || !strings.Contains(u.Note, "importar")) {
			t.Fatalf("imported person created in the team (with a code, so they can be in it): %+v", u)
		}
	}
	// Public reports can't pick who added them.
	if w := s.call("POST", "/api/submissions", `{"area":"E","coordinatorPhone":"6555-1234","addedByName":"Mia Manager","addedByPhone":"6999-1111"}`, false); w.Code != 201 || strings.Contains(w.Body.String(), mia) {
		t.Fatalf("public submission: %d %s", w.Code, w.Body.String())
	}
}

// Delivery import: an approved point backdated to the report day, its completed need backdated too, and the runner
// linked or created (runner role; when created with a team, also an access code so they can be in it). Agents can't
// use the import fields; an unknown team is a clear 400.
func TestDeliveryImportBackdatedWithRunners(t *testing.T) {
	s := newServiceHarness(t)
	s.team("Unidos por Azuero")
	day := "2026-10-14T15:00:00Z"
	var p Point
	w := s.call("POST", "/api/admin/points", `{"area":"Tonosí","place":"Río Quema","created":"`+day+`","addedByName":"Yasmine Jamai","addedByPhone":"6821-2299","addedByTeam":"Unidos por Azuero"}`, true)
	if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil || p.Status != "approved" {
		t.Fatalf("approved point: %d %s", w.Code, w.Body.String())
	}
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, x := range pts {
		if x.ID == p.ID && (x.Created != day || x.Decided != day || x.AddedByName != "Yasmine Jamai") {
			t.Fatalf("backdated point with its reporter: %+v", x)
		}
	}
	var n CampNeed
	w = s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Alimentos","urgency":"normal","status":"COMPLETED","runnerName":"Javier Fría","runnerPhone":"6865-8227","runnerTeam":"Unidos por Azuero","at":"`+day+`"}`, true)
	if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &n) != nil || n.RunnerID == "" || n.CreatedAt != day || n.CompletedAt != day {
		t.Fatalf("backdated delivery: %d %s", w.Code, w.Body.String())
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	found := 0
	for _, u := range users {
		if u.ID == n.RunnerID {
			found++
			if u.Name != "Javier Fría" || !slicesContains(u.Roles, "runner") || !u.HasAccessCode || u.TeamID == "" {
				t.Fatalf("created runner in the team with a code: %+v", u)
			}
		}
		if u.Name == "Yasmine Jamai" && (u.TeamID == "" || !u.HasAccessCode) {
			t.Fatalf("created reporter in the team: %+v", u)
		}
	}
	if found != 1 {
		t.Fatal("runner not created")
	}
	// Same runner again (other phone format): linked, not duplicated.
	var n2 CampNeed
	w = s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Agua","urgency":"normal","status":"COMPLETED","runnerName":"Javier","runnerPhone":"+507 6865 8227","runnerTeam":"Unidos por Azuero","at":"`+day+`"}`, true)
	if json.Unmarshal(w.Body.Bytes(), &n2) != nil || n2.RunnerID != n.RunnerID {
		t.Fatalf("same runner linked: %s", w.Body.String())
	}
	// Security: someone OUTSIDE the team with the same phone is never linked; a new person is created in the team.
	outsider := s.expect("POST", "/api/admin/users", `{"name":"Fuera Equipo","telephone":"6111-9999","roles":["runner"]}`, 201)["id"].(string)
	var n3 CampNeed
	w = s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Agua","urgency":"normal","status":"COMPLETED","runnerName":"Fuera Equipo","runnerPhone":"6111-9999","runnerTeam":"Unidos por Azuero","at":"`+day+`"}`, true)
	if json.Unmarshal(w.Body.Bytes(), &n3) != nil || n3.RunnerID == "" || n3.RunnerID == outsider {
		t.Fatalf("outsider must not be linked: %s", w.Body.String())
	}
	// Without a team, nobody is linked: always a new person.
	var n4 CampNeed
	w = s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Agua","urgency":"normal","status":"COMPLETED","runnerName":"Javier Fría","runnerPhone":"6865-8227","at":"`+day+`"}`, true)
	if json.Unmarshal(w.Body.Bytes(), &n4) != nil || n4.RunnerID == "" || n4.RunnerID == n.RunnerID {
		t.Fatalf("no team: never linked: %s", w.Body.String())
	}
	if w := s.call("POST", "/api/admin/camps/"+p.ID+"/needs", `{"need":"Agua","status":"COMPLETED","runnerName":"Ana","runnerTeam":"No Existe"}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "No existe el equipo") {
		t.Fatalf("unknown team: %d %s", w.Code, w.Body.String())
	}
	s.expect("POST", "/api/admin/users", `{"name":"Ada Agente","telephone":"6999-7777","roles":["agent"],"accessCode":"ADA123"}`, 201)
	ac, at := s.agentLogin("ADA123")
	if w := s.agentCall("POST", "/api/agent/camps/"+p.ID+"/needs", `{"need":"Agua","urgency":"normal","runnerName":"X"}`, ac, at); w.Code != 400 {
		t.Fatalf("agents can't use import fields: %d %s", w.Code, w.Body.String())
	}
}
