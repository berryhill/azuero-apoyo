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
		t.Fatalf("not on a team: %d", w.Code)
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

// Admin invite for one person: a one-use link to a team, valid 24 h, that only that user ID can accept (anyone else
// signed in gets 403 and the link is not spent).
func TestAdminInviteForOneUser(t *testing.T) {
	s := newServiceHarness(t)
	team := s.expect("POST", "/api/admin/teams", `{"name":"Oro"}`, 201)["id"].(string)
	mia := s.expect("POST", "/api/admin/users", `{"name":"Mia","telephone":"6999-4444","roles":["runner"],"accessCode":"MIA123"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/users", `{"name":"Leo","telephone":"6999-5555","roles":["runner"],"accessCode":"LEO123"}`, 201)
	noCode := s.expect("POST", "/api/admin/users", `{"name":"Sin Código","telephone":"6999-6666","roles":["runner"]}`, 201)["id"].(string)
	if w := s.call("POST", "/api/admin/teams/"+team+"/invites", `{"userId":"`+noCode+`"}`, true); w.Code != 400 {
		t.Fatalf("a user without a code can't accept: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/admin/teams/"+team+"/invites", `{"userId":"nope"}`, true); w.Code != 400 {
		t.Fatalf("unknown user: %d", w.Code)
	}
	inv := s.expect("POST", "/api/admin/teams/"+team+"/invites", `{"userId":"`+mia+`"}`, 201)
	token, exp := inv["token"].(string), inv["expiresAt"].(string)
	if at, _ := time.Parse(time.RFC3339, exp); at.Sub(time.Now()) > 24*time.Hour || at.Sub(time.Now()) < 23*time.Hour || inv["user"] != "Mia" {
		t.Fatalf("24 h, for Mia: %v", inv)
	}
	body := `{"token":"` + token + `"}`
	if w := s.as(s.viewerCookie("LEO123"), "POST", "/api/full/invites/accept", body); w.Code != 403 || !strings.Contains(w.Body.String(), "otra persona") {
		t.Fatalf("someone else can't use it: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(s.viewerCookie("MIA123"), "POST", "/api/full/invites/accept", body); w.Code != 200 || !strings.Contains(w.Body.String(), "Oro") {
		t.Fatalf("Mia joins (the link was not spent by Leo): %d %s", w.Code, w.Body.String())
	}
	if w := s.as(s.viewerCookie("LEO123"), "POST", "/api/full/invites/accept", body); w.Code != 410 {
		t.Fatalf("used once: %d", w.Code)
	}
	// Admin monitors every invite (status, team, for whom, used by) and can revoke a pending one.
	other := s.expect("POST", "/api/admin/teams/"+team+"/invites", `{"userId":"`+mia+`"}`, 201)
	list := s.call("GET", "/api/admin/invites", "", true).Body.String()
	if !strings.Contains(list, `"status":"used"`) || !strings.Contains(list, `"usedByName":"Mia"`) || !strings.Contains(list, `"forUserName":"Mia"`) || !strings.Contains(list, `"teamName":"Oro"`) || !strings.Contains(list, `"status":"pending"`) || strings.Contains(list, "hash") || strings.Contains(list, token) {
		t.Fatalf("admin invite list: %s", list)
	}
	s.expect("POST", "/api/admin/invites/"+other["id"].(string)+"/revoke", `{}`, 200)
	if list := s.call("GET", "/api/admin/invites", "", true).Body.String(); !strings.Contains(list, `"status":"revoked"`) {
		t.Fatalf("revoked: %s", list)
	}
	if w := s.as(s.viewerCookie("MIA123"), "POST", "/api/full/invites/accept", `{"token":"`+other["token"].(string)+`"}`); w.Code != 410 {
		t.Fatalf("a revoked link no longer works: %d", w.Code)
	}
}
