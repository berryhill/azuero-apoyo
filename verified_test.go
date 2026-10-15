package main

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// «Verificado por equipo»: un punto o una necesidad que crea o actualiza alguien de un equipo queda marcado (equipo,
// persona, hora). En público solo sale la hora, sin nombres. Si luego lo cambia alguien sin equipo, se quita;
// administración (contraseña maestra) no lo cambia.
func TestTeamWritesMarkPointsAndNeedsVerified(t *testing.T) {
	s := newServiceHarness(t)
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-1111","roles":["agent"],"accessCode":"ANA123"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/users", `{"name":"Leo Libre","telephone":"6999-3333","roles":["agent"],"accessCode":"LEO123"}`, 201)
	s.team("Equipo Norte", ana)
	anaC, anaT := s.agentLogin("ANA123")
	leoC, leoT := s.agentLogin("LEO123")

	n := map[string]any{}
	w := s.agentCall("POST", "/api/agent/camps/"+s.camp+"/needs", `{"need":"Agua (para beber)","urgency":"normal"}`, anaC, anaT)
	if w.Code != 201 {
		t.Fatalf("create need: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &n)
	id := n["id"].(string)
	if n["verifiedTeam"] != "Equipo Norte" || n["verifiedBy"] != "Ana Agente" || n["verifiedAt"] == nil {
		t.Fatalf("a team agent's need is verified: %v", n)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if !strings.Contains(pub, `"verifiedAt"`) || strings.Contains(pub, "Equipo Norte") || strings.Contains(pub, "Ana Agente") {
		t.Fatalf("public needs show the badge only, no names: %s", pub)
	}
	// Point edited by the team agent: verified; public points keep only the time.
	if w := s.agentCall("PATCH", "/api/agent/camps/"+s.camp, `{"place":"Escuela"}`, anaC, anaT); w.Code != 200 || !strings.Contains(w.Body.String(), `"verifiedTeam":"Equipo Norte"`) {
		t.Fatalf("point edit: %d %s", w.Code, w.Body.String())
	}
	if pts := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(pts, `"verifiedAt"`) || strings.Contains(pts, "Equipo Norte") || strings.Contains(pts, "Ana Agente") {
		t.Fatalf("public points: badge only: %s", pts)
	}
	// Runners see the point's and the need's badge with names (private panel).
	rc := s.runnerLogin("Rui")
	if _, _, b := rc.needs(); !strings.Contains(b, `"verifiedTeam":"Equipo Norte"`) {
		t.Fatalf("runner needs carry the badge: %s", b)
	}
	// Admin edits keep it; an agent without a team clears it.
	s.expect("PATCH", "/api/admin/needs/"+id, `{"urgency":"urgente"}`, 200)
	if b := s.call("GET", "/api/admin/needs", "", true).Body.String(); !strings.Contains(b, `"verifiedTeam":"Equipo Norte"`) {
		t.Fatalf("admin edit keeps it: %s", b)
	}
	if w := s.agentCall("PATCH", "/api/agent/needs/"+id, `{"urgency":"normal"}`, leoC, leoT); w.Code != 200 || strings.Contains(w.Body.String(), "verifiedTeam") {
		t.Fatalf("a non-team update clears it: %d %s", w.Code, w.Body.String())
	}
	// Code user in a team adds a point: verified. Public submissions never are, even if they send the fields.
	mia := s.expect("POST", "/api/admin/users", `{"name":"Mia","telephone":"6999-4444","roles":["runner"],"accessCode":"MIA123"}`, 201)["id"].(string)
	s.team("Equipo Sur", mia)
	c := s.viewerCookie("MIA123")
	if w := s.as(c, "POST", "/api/full/points", `{"area":"Sigua","coordinatorPhone":"6555-1234","needs":"Agua"}`); w.Code != 201 {
		t.Fatalf("team code user point: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/submissions", `{"area":"Otra Zona","coordinatorPhone":"6555-1235","needs":"Agua","verifiedTeam":"Falso","verifiedAt":"2026-01-01"}`, false); w.Code != 201 {
		t.Fatalf("public submission: %d %s", w.Code, w.Body.String())
	}
	var pts, subs []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	json.Unmarshal(s.call("GET", "/api/admin/submissions", "", true).Body.Bytes(), &subs)
	seen := 0
	for _, p := range append(pts, subs...) {
		if p.Area == "Sigua" || p.Area == "Otra Zona" {
			seen++
		}
		if p.Area == "Sigua" && (p.VerifiedTeam != "Equipo Sur" || p.VerifiedBy != "Mia" || p.VerifiedAt == "") {
			t.Fatalf("team code user point is verified: %+v", p)
		}
		if p.Area == "Otra Zona" && (p.VerifiedTeam != "" || p.VerifiedAt != "") {
			t.Fatalf("public submission is never verified: %+v", p)
		}
	}
	if seen < 2 {
		t.Fatalf("both points listed: %d", seen)
	}
}

// «Nuestro equipo lo abastece»: a team member adding a point can hold it from the public (and other teams) for
// holdHours; their team sees it at once; it publishes by itself after that, or sooner with «Publicar ahora».
func TestTeamHoldDelaysPublicationForThreeHours(t *testing.T) {
	s := newServiceHarness(t)
	mia := s.expect("POST", "/api/admin/users", `{"name":"Mia","telephone":"6999-4444","roles":["runner"],"accessCode":"MIA123"}`, 201)["id"].(string)
	leo := s.expect("POST", "/api/admin/users", `{"name":"Leo","telephone":"6999-5555","roles":["runner"],"accessCode":"LEO123"}`, 201)["id"].(string)
	otto := s.expect("POST", "/api/admin/users", `{"name":"Otto","telephone":"6999-6666","roles":["runner"],"accessCode":"OTTO12"}`, 201)["id"].(string)
	s.team("Norte", mia, leo)
	s.team("Sur", otto)
	mc, lc, oc := s.viewerCookie("MIA123"), s.viewerCookie("LEO123"), s.viewerCookie("OTTO12")
	if w := s.as(mc, "POST", "/api/full/points", `{"area":"Retenido","coordinatorPhone":"6555-1234","needs":"Agua","holdTeam":true}`); w.Code != 201 {
		t.Fatalf("add: %d %s", w.Code, w.Body.String())
	}
	var id string
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, p := range pts {
		if p.Area == "Retenido" {
			id = p.ID
			if p.HeldBy == "" || p.HeldUntil <= nowStamp() || p.HoldTeam {
				t.Fatalf("held for the team, a few hours: %+v", p)
			}
		}
	}
	has := func(c *http.Cookie) bool {
		path := "/api/points"
		if c != nil {
			path = "/api/full/points"
		}
		return strings.Contains(s.as(c, "GET", path, "").Body.String(), "Retenido")
	}
	if has(nil) || has(oc) || !has(lc) {
		t.Fatalf("public %v, other team %v, own team %v", has(nil), has(oc), has(lc))
	}
	if w := s.call("GET", "/api/camps/"+id+"/needs", "", false); w.Code != 404 {
		t.Fatalf("its needs are hidden from the public too: %d", w.Code)
	}
	// A point added without the toggle, or by someone without a team, is public right away.
	s.as(mc, "POST", "/api/full/points", `{"area":"Abierto","coordinatorPhone":"6555-1235","needs":"Agua"}`)
	if !strings.Contains(s.call("GET", "/api/points", "", false).Body.String(), "Abierto") {
		t.Fatal("no toggle: public at once")
	}
	// After the hold it publishes by itself.
	s.app.db.Update(func(tx *bolt.Tx) error {
		var p Point
		json.Unmarshal(tx.Bucket(submissions).Get([]byte(id)), &p)
		p.HeldUntil = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
		return putJSON(tx.Bucket(submissions), id, p)
	})
	if !has(nil) || !has(oc) {
		t.Fatal("published by itself after the hold")
	}
	// «Publicar ahora» (admin or agent) ends a hold early.
	s.app.db.Update(func(tx *bolt.Tx) error {
		var p Point
		json.Unmarshal(tx.Bucket(submissions).Get([]byte(id)), &p)
		p.HeldUntil = time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		return putJSON(tx.Bucket(submissions), id, p)
	})
	if has(nil) {
		t.Fatal("held again")
	}
	s.expect("POST", "/api/admin/camps/"+id+"/publish", `{}`, 200)
	if !has(nil) {
		t.Fatal("published now")
	}
}
