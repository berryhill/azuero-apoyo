package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// «Mi equipo» for every member: activity board, team map (live or last position), team name edited by the team
// manager only, and invites that ask before joining (who invited, who's in it, no phone numbers).
func TestTeamBoardMapRenameAndInvitePreview(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	mia := user(`{"name":"Mia Manager","telephone":"6999-1111","roles":["runner"],"accessCode":"MIA123"}`)
	leo := user(`{"name":"Leo Runner","telephone":"6999-2222","roles":["runner"],"accessCode":"LEO123"}`)
	user(`{"name":"Lu Nueva","telephone":"6999-3333","roles":["runner"],"accessCode":"LU1234"}`)
	s.team("Norte", mia, leo)
	s.expect("PATCH", "/api/admin/users/"+mia, `{"roles":["runner","teammanager"]}`, 200)
	mc, lc, uc := s.viewerCookie("MIA123"), s.viewerCookie("LEO123"), s.viewerCookie("LU1234")

	if w := s.as(lc, "GET", "/api/full/team", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"board":{"members":2`) {
		t.Fatalf("member sees the board: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"LEO123"}`, false); !strings.Contains(w.Body.String(), "/team.html") {
		t.Fatalf("any member gets Mi equipo in their pages: %s", w.Body.String())
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"LU1234"}`, false); strings.Contains(w.Body.String(), "/team.html") {
		t.Fatalf("no team, no Mi equipo: %s", w.Body.String())
	}

	// Team name: only the team manager.
	if w := s.as(lc, "PATCH", "/api/full/team", `{"name":"Hackeado"}`); w.Code != 403 {
		t.Fatalf("a member can't rename: %d", w.Code)
	}
	if w := s.as(mc, "PATCH", "/api/full/team", `{"name":"  "}`); w.Code != 400 {
		t.Fatalf("empty name: %d", w.Code)
	}
	if w := s.as(mc, "PATCH", "/api/full/team", `{"name":"Norte  Azuero"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Norte Azuero"`) {
		t.Fatalf("rename: %d %s", w.Code, w.Body.String())
	}

	// Map: Leo shares live; Mia sees him as live; someone with no team gets 403.
	s.as(lc, "POST", "/api/full/location/consent", `{"accept":true}`)
	if w := s.as(lc, "POST", "/api/full/location", `{"lat":7.9,"lon":-80.4,"accuracy":12}`); w.Code != 200 {
		t.Fatalf("location: %d %s", w.Code, w.Body.String())
	}
	var pins []teamMapPin
	json.Unmarshal(s.as(mc, "GET", "/api/full/team/map", "").Body.Bytes(), &pins)
	if len(pins) != 1 || pins[0].Name != "Leo Runner" || !pins[0].Live {
		t.Fatalf("team map: %+v", pins)
	}
	if w := s.as(uc, "GET", "/api/full/team/map", ""); w.Code != 403 {
		t.Fatalf("no team, no map: %d", w.Code)
	}
	// Stopped sharing: no longer live, but the last position (track) stays for 24 h.
	s.as(lc, "POST", "/api/full/location/stop", `{}`)
	pins = nil
	json.Unmarshal(s.as(mc, "GET", "/api/full/team/map", "").Body.Bytes(), &pins)
	if len(pins) != 1 || pins[0].Live {
		t.Fatalf("last position: %+v", pins)
	}

	// Time-lapse: a teammate's track for a day (last 7 days only); not for someone outside the team.
	day := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	end := time.Now().UTC().Add(time.Minute).Format(time.RFC3339)
	var track []trackPoint
	if w := s.as(mc, "GET", "/api/full/team/members/"+leo+"/track?from="+day+"&to="+end, ""); w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &track) != nil || len(track) == 0 {
		t.Fatalf("teammate track: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(uc, "GET", "/api/full/team/members/"+leo+"/track?from="+day+"&to="+end, ""); w.Code != 404 {
		t.Fatalf("outsider can't see the track: %d", w.Code)
	}
	if w := s.as(mc, "GET", "/api/full/team/members/"+leo+"/track?from="+day+"&to="+time.Now().UTC().Add(72*time.Hour).Format(time.RFC3339), ""); w.Code != 400 {
		t.Fatalf("one day at a time: %d", w.Code)
	}

	// Invite: preview first (no phones, who invited), not in the team until accepted.
	var inv map[string]any
	json.Unmarshal(s.as(mc, "POST", "/api/full/team/invites", `{"hours":24}`).Body.Bytes(), &inv)
	tok := inv["token"].(string)
	// The manager can show a pending link's QR again (token in the list); gone once used.
	if body := s.as(mc, "GET", "/api/full/team", "").Body.String(); !strings.Contains(body, `"token":"`+tok+`"`) {
		t.Fatalf("pending invite keeps its link for the QR: %s", body)
	}
	if body := s.as(lc, "GET", "/api/full/team", "").Body.String(); strings.Contains(body, tok) {
		t.Fatalf("a plain member never sees invite links")
	}
	w := s.as(uc, "POST", "/api/full/invites/preview", `{"token":"`+tok+`"}`)
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, `"invitedBy":"Mia Manager"`) || !strings.Contains(body, "Leo Runner") || strings.Contains(body, "6999") || !strings.Contains(body, `"team":"Norte Azuero"`) {
		t.Fatalf("preview: %d %s", w.Code, body)
	}
	if w := s.as(uc, "GET", "/api/full/team", ""); w.Code != 403 {
		t.Fatalf("preview doesn't join: %d", w.Code)
	}
	if w := s.call("POST", "/api/full/invites/preview", `{"token":"`+tok+`"}`, false); w.Code != 401 {
		t.Fatalf("preview needs a session: %d", w.Code)
	}
	if w := s.as(uc, "POST", "/api/full/invites/accept", `{"token":"`+tok+`"}`); w.Code != 200 {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	if body := s.as(mc, "GET", "/api/full/team", "").Body.String(); strings.Contains(body, tok) {
		t.Fatalf("used invite: link no longer shown")
	}
	if w := s.as(uc, "POST", "/api/full/invites/preview", `{"token":"`+tok+`"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
		t.Fatalf("after joining: %d %s", w.Code, w.Body.String())
	}
}
