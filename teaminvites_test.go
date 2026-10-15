package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

// A team manager (who must be in a team) creates one-time, expiring invitation links; a signed-in user who opens
// one joins that team.
func TestTeamManagerInvites(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("POST", "/api/admin/users", `{"name":"Mia Manager","telephone":"6999-5050","roles":["teammanager"],"accessCode":"MIA7777"}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "equipo") {
		t.Fatalf("a team manager needs a team: %d %s", w.Code, w.Body.String())
	}
	s.expect("POST", "/api/admin/users", `{"name":"Mia Manager","telephone":"6999-5050","roles":["runner"],"accessCode":"MIA7777"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Leo Nuevo","telephone":"6999-6060","roles":["runner"],"accessCode":"LEO7777"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Lu Tercera","telephone":"6999-7070","roles":["runner"],"accessCode":"LU77777"}`, 201)
	ids := map[string]string{}
	var users []map[string]any
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		ids[u["name"].(string)] = u["id"].(string)
	}
	team := s.team("Equipo Mia", ids["Mia Manager"])
	s.expect("PATCH", "/api/admin/users/"+ids["Mia Manager"], `{"roles":["runner","teammanager"]}`, 200)
	if w := s.call("PATCH", "/api/admin/users/"+ids["Mia Manager"], `{"teamId":""}`, true); w.Code != 400 {
		t.Fatalf("cannot leave the team while team manager: %d", w.Code)
	}
	mia, leo, lu := s.viewerCookie("MIA7777"), s.viewerCookie("LEO7777"), s.viewerCookie("LU77777")
	if w := s.as(leo, "GET", "/api/full/team", ""); w.Code != 403 {
		t.Fatalf("only team managers: %d", w.Code)
	}
	if w := s.as(mia, "GET", "/api/full/team", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Equipo Mia") || !strings.Contains(w.Body.String(), "Mia Manager") {
		t.Fatalf("team page: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"MIA7777"}`, false); !strings.Contains(w.Body.String(), "/team.html") {
		t.Fatalf("panel: %s", w.Body.String())
	}
	invite := func(body string) string {
		w := s.as(mia, "POST", "/api/full/team/invites", body)
		if w.Code != 201 {
			t.Fatalf("invite: %d %s", w.Code, w.Body.String())
		}
		var out map[string]string
		json.Unmarshal(w.Body.Bytes(), &out)
		return out["token"]
	}
	if w := s.as(mia, "POST", "/api/full/team/invites", `{"hours":5}`); w.Code != 400 {
		t.Fatalf("only 24/72/168 h: %d", w.Code)
	}
	tok := invite(`{"hours":24}`)
	// Signed-in user only.
	if w := s.call("POST", "/api/full/invites/accept", `{"token":"`+tok+`"}`, false); w.Code != 401 {
		t.Fatalf("needs a session: %d", w.Code)
	}
	if w := s.as(leo, "POST", "/api/full/invites/accept", `{"token":"`+tok+`"}`); w.Code != 200 || !strings.Contains(w.Body.String(), "Equipo Mia") {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		if u["name"] == "Leo Nuevo" && u["teamId"] != team {
			t.Fatalf("Leo joined: %v", u)
		}
	}
	if w := s.as(lu, "POST", "/api/full/invites/accept", `{"token":"`+tok+`"}`); w.Code != 410 {
		t.Fatalf("one use only: %d", w.Code)
	}
	if w := s.as(leo, "POST", "/api/full/invites/accept", `{"token":"`+tok+`"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
		t.Fatalf("opening it again yourself is fine: %d %s", w.Code, w.Body.String())
	}
	// Expired and revoked links are refused.
	old := invite(`{}`)
	s.app.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(teamInvitesBucket)
		var inv TeamInvite
		json.Unmarshal(b.Get([]byte(inviteHash(old))), &inv)
		inv.ExpiresAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		return putJSON(b, inv.Hash, inv)
	})
	if w := s.as(lu, "POST", "/api/full/invites/accept", `{"token":"`+old+`"}`); w.Code != 410 {
		t.Fatalf("expired: %d", w.Code)
	}
	rev := invite(`{"hours":168}`)
	var view struct {
		Invites []teamInviteView
		Members []map[string]any
	}
	json.Unmarshal(s.as(mia, "GET", "/api/full/team", "").Body.Bytes(), &view)
	statuses := map[string]int{}
	var revID string
	for _, inv := range view.Invites {
		statuses[inv.Status]++
		if inv.Status == "pending" {
			revID = inv.ID
		}
	}
	if statuses["used"] != 1 || statuses["expired"] != 1 || statuses["pending"] != 1 || len(view.Members) != 2 {
		t.Fatalf("invites %v, members %d", statuses, len(view.Members))
	}
	if w := s.as(mia, "POST", "/api/full/team/invites/"+revID+"/revoke", `{}`); w.Code != 200 {
		t.Fatalf("revoke: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(lu, "POST", "/api/full/invites/accept", `{"token":"`+rev+`"}`); w.Code != 410 {
		t.Fatalf("revoked: %d", w.Code)
	}
	if body := s.as(mia, "GET", "/api/full/team", "").Body.String(); strings.Contains(body, rev) || strings.Contains(body, inviteHash(rev)) {
		t.Fatalf("tokens never listed")
	}
	// Manage members: change roles (only team-level roles), not yourself; remove from the team.
	leoID := ids["Leo Nuevo"]
	if w := s.as(mia, "PATCH", "/api/full/team/members/"+leoID, `{"roles":["runner","builder"]}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"builder"`) {
		t.Fatalf("roles: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(mia, "PATCH", "/api/full/team/members/"+leoID, `{"roles":["agent"]}`); w.Code != 400 {
		t.Fatalf("agent is admin-only: %d", w.Code)
	}
	if w := s.as(mia, "PATCH", "/api/full/team/members/"+ids["Mia Manager"], `{"roles":["runner"]}`); w.Code != 400 {
		t.Fatalf("not yourself: %d", w.Code)
	}
	if w := s.as(mia, "PATCH", "/api/full/team/members/"+ids["Lu Tercera"], `{"roles":["runner"]}`); w.Code != 404 {
		t.Fatalf("not in my team: %d", w.Code)
	}
	if w := s.as(mia, "POST", "/api/full/team/members/"+leoID+"/remove", `{}`); w.Code != 200 {
		t.Fatalf("remove: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(s.as(mia, "GET", "/api/full/team", "").Body.Bytes(), &view)
	if len(view.Members) != 1 {
		t.Fatalf("Leo left the team: %+v", view.Members)
	}
}
