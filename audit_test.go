package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// Every change to a point or a need (runs, packing, photos, supply) is in the point's trail with who (name, team,
// page), when and before → after; agents from elsewhere can't edit, release or delete a team's data.
func TestTrailAndTeamDataGuards(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	fufo := user(`{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`)
	mar := user(`{"name":"Mar","telephone":"6999-2222","roles":["agent"],"accessCode":"MAR123"}`)
	user(`{"name":"Gus","telephone":"6999-3333","roles":["agent"],"accessCode":"GUS123"}`)
	s.team("Norte", fufo, mar)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)

	// Fufo takes it on the runner page and packs it; his photo goes on it from the home page.
	login := s.crewCall("POST", "/api/runner/session", `{"accessCode":"FUFO12"}`, nil, "")
	var runner *http.Cookie
	for _, c := range login.Result().Cookies() {
		if c.Name == "runner_session" {
			runner = c
		}
	}
	var sess struct{ CSRF string }
	json.Unmarshal(login.Body.Bytes(), &sess)
	if w := s.crewCall("POST", "/api/runner/needs/"+need+"/take", `{}`, runner, sess.CSRF); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+need+"/packing", `{"items":[{"item":"Agua 5 gal","quantity":3}]}`, runner, sess.CSRF); w.Code != 200 {
		t.Fatalf("pack: %d %s", w.Code, w.Body.String())
	}
	fufoC := s.viewerCookie("FUFO12")
	w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=pickup", fufoC, "", "image/jpeg", []byte("\xff\xd8\xff x"))
	if w.Code != 201 {
		t.Fatalf("photo: %d %s", w.Code, w.Body.String())
	}
	var media MediaItem
	json.Unmarshal(w.Body.Bytes(), &media)

	trail := func() []AuditEntry {
		var list []AuditEntry
		json.Unmarshal(s.call("GET", "/api/admin/camps/"+s.camp+"/trail", "", true).Body.Bytes(), &list)
		return list
	}
	has := func(list []AuditEntry, by, team, via, label, after string) bool {
		for _, e := range list {
			if e.ByName != by || e.ByTeam != team || e.Via != via {
				continue
			}
			for _, c := range e.Changes {
				if c.Label == label && strings.Contains(c.After, after) {
					return true
				}
			}
		}
		return false
	}
	list := trail()
	if !has(list, "Administración", "", "Admin", "Pedido", "Agua") || !has(list, "Fufo", "Norte", "Runners", "Estado", "En camino") ||
		!has(list, "Fufo", "Norte", "Runners", "Lista de empaque", "Agua 5 gal ×3") || !has(list, "Fufo", "Norte", "Inicio", "Fotos y videos", "1:") {
		t.Fatalf("trail: %+v", list)
	}
	if list[0].At < list[len(list)-1].At {
		t.Fatal("newest first")
	}

	// Gus (global agent): can't delete or move Fufo's photo, can't change the point Mar edited, and his need edit is a request.
	gusC, gusCSRF := s.agentLogin("GUS123")
	if w := s.agentCall("DELETE", "/api/agent/needs/"+need+"/media/"+media.ID, "", gusC, gusCSRF); w.Code != 403 || !strings.Contains(w.Body.String(), "Fufo") {
		t.Fatalf("global agent deletes a team photo: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need+"/media/"+media.ID, `{"stage":"delivery"}`, gusC, gusCSRF); w.Code != 403 {
		t.Fatalf("global agent moves a team photo: %d", w.Code)
	}
	marC, marCSRF := s.agentLogin("MAR123")
	if w := s.agentCall("PATCH", "/api/agent/camps/"+s.camp, `{"people":"40"}`, marC, marCSRF); w.Code != 200 {
		t.Fatalf("Mar edits the point: %d %s", w.Code, w.Body.String())
	}
	w = s.agentCall("PATCH", "/api/agent/camps/"+s.camp, `{"people":"10","directions":"Por el río"}`, gusC, gusCSRF)
	var out struct{ Request ChangeRequest }
	json.Unmarshal(w.Body.Bytes(), &out)
	if w.Code != 202 || out.Request.Kind != "punto" || out.Request.ToName != "Mar" || len(out.Request.Changes) != 2 {
		t.Fatalf("point change request: %d %s", w.Code, w.Body.String())
	}
	var p Point
	pointNow := func() Point {
		var list []Point
		json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &list)
		for _, x := range list {
			if x.ID == s.camp {
				return x
			}
		}
		return Point{}
	}
	if p = pointNow(); p.People != "40" {
		t.Fatalf("point untouched until Mar answers: %+v", p.People)
	}
	// Mar accepts in Command: it applies, stays Mar's, and the trail shows Gus's change.
	if w := s.agentCall("POST", "/api/agent/change-requests/"+out.Request.ID+"/accept", `{}`, marC, marCSRF); w.Code != 200 {
		t.Fatalf("accept point: %d %s", w.Code, w.Body.String())
	}
	if p = pointNow(); p.People != "10" || p.Directions != "Por el río" || p.UpdatedBy != mar {
		t.Fatalf("accepted point: %+v", p)
	}
	if !has(trail(), "Gus", "", "Checkup · aceptado por Mar", "Personas", "10") {
		t.Fatalf("trail of the accepted change: %+v", trail())
	}
	// Mar (Fufo's team) may move the photo.
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need+"/media/"+media.ID, `{"stage":"delivery"}`, marC, marCSRF); w.Code != 200 {
		t.Fatalf("teammate moves the photo: %d %s", w.Code, w.Body.String())
	}
	// The public never sees who updated the point.
	if body := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(body, "updatedBy") {
		t.Fatalf("public: %s", body)
	}
	// The trail is for admin and Command only.
	if w := s.call("GET", "/api/admin/camps/"+s.camp+"/trail", "", false); w.Code != 401 {
		t.Fatalf("trail without a session: %d", w.Code)
	}
	if w := s.agentCall("GET", "/api/agent/camps/"+s.camp+"/trail", "", gusC, ""); w.Code != 200 {
		t.Fatalf("trail in Command: %d", w.Code)
	}
}
