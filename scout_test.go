package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Scout: every signed-in code user sees all published points with their active needs. People on a team update any
// point's needs directly (report, change, remove), even needs another team updated last; a need on the way can be
// changed but not removed; people without a team can't update. Writes show as page «Scout» in the change history.
func TestScoutAllPointsAndUpdateNeeds(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	ana := user(`{"name":"Ana Norte","telephone":"6999-1111","roles":["runner"],"accessCode":"ANAN77"}`)
	betoID := user(`{"name":"Beto Sur","telephone":"6999-2222","roles":["runner"],"accessCode":"BETO77"}`)
	user(`{"name":"Solo Uno","telephone":"6999-3333","roles":["runner"],"accessCode":"SOLO77"}`)
	s.team("Norte", ana)
	s.team("Sur", betoID)
	if w := s.as(s.viewerCookie("SOLO77"), "POST", "/api/full/scout/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`); w.Code != 400 || !strings.Contains(w.Body.String(), "equipo") {
		t.Fatalf("no team can't update: %d %s", w.Code, w.Body.String())
	}
	beto := s.viewerCookie("BETO77")
	scout := func() []scoutPoint {
		var out []scoutPoint
		w := s.as(beto, "GET", "/api/full/scout", "")
		if w.Code != 200 {
			t.Fatalf("scout: %d %s", w.Code, w.Body.String())
		}
		json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	active := func() []CampNeed {
		for _, p := range scout() {
			if p.ID == s.camp {
				return p.ActiveNeeds
			}
		}
		t.Fatal("published point missing from Scout (not added by me or my team)")
		return nil
	}
	// Contacts (names and phones only) go to people on a team; without a team, none (anyone can make an account).
	if body := s.as(s.viewerCookie("SOLO77"), "GET", "/api/full/scout", "").Body.String(); strings.Contains(body, "6555-1234") || strings.Contains(body, `"contacts"`) {
		t.Fatalf("Scout shows contacts to someone with no team: %s", body)
	}
	if body := s.as(beto, "GET", "/api/full/scout", "").Body.String(); !strings.Contains(body, "6555-1234") || !strings.Contains(body, `"contacts"`) || !strings.Contains(body, `"name":"Private camp"`) || strings.Contains(body, "cedula") {
		t.Fatalf("Scout contacts for a team member (names and phones only): %s", body)
	}
	// Report a new need.
	w := s.as(beto, "POST", "/api/full/scout/camps/"+s.camp+"/needs", `{"need":"Agua, Alimentos","urgency":"urgente"}`)
	if w.Code != 201 {
		t.Fatalf("new need: %d %s", w.Code, w.Body.String())
	}
	var created struct{ ID string }
	json.Unmarshal(w.Body.Bytes(), &created)
	if n := active(); len(n) != 1 || n[0].Need != "Agua, Alimentos" || n[0].Urgency != "urgente" {
		t.Fatalf("active needs: %+v", n)
	}
	if w := s.as(beto, "POST", "/api/full/scout/camps/"+s.camp+"/needs", `{"need":"","urgency":"normal"}`); w.Code != 400 {
		t.Fatalf("empty need: %d", w.Code)
	}
	// Change it; only removal is allowed as a status.
	if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+created.ID, `{"need":"Agua","urgency":"normal"}`); w.Code != 200 {
		t.Fatalf("change: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+created.ID, `{"status":"COMPLETED"}`); w.Code != 400 {
		t.Fatalf("complete from Scout: %d", w.Code)
	}
	if n := active(); len(n) != 1 || n[0].Need != "Agua" || n[0].Urgency != "normal" {
		t.Fatalf("after change: %+v", n)
	}
	// A need Ana (team Norte) updated last: Beto (team Sur) changes it directly, no change request.
	anaNeed := s.as(s.viewerCookie("ANAN77"), "POST", "/api/full/scout/camps/"+s.camp+"/needs", `{"need":"Kit de bebé","urgency":"normal"}`)
	json.Unmarshal(anaNeed.Body.Bytes(), &created)
	if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+created.ID, `{"urgency":"urgente"}`); w.Code != 200 {
		t.Fatalf("other team's need: %d %s", w.Code, w.Body.String())
	}
	if n := active(); len(n) != 2 {
		t.Fatalf("both needs: %+v", n)
	}
	// Remove Beto's own need.
	for _, n := range active() {
		if n.Need == "Agua" {
			if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+n.ID, `{"status":"DELETED"}`); w.Code != 200 {
				t.Fatalf("remove: %d %s", w.Code, w.Body.String())
			}
		}
	}
	if n := active(); len(n) != 1 || n[0].Need != "Kit de bebé" {
		t.Fatalf("after remove: %+v", n)
	}
	// On the way: can be changed, not removed.
	rc, rcsrf := s.sessionLogin("/api/runner/session", "ANAN77", "runner_session")
	if w := s.crewCall("POST", "/api/runner/needs/"+created.ID+"/take", `{}`, rc, rcsrf); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+created.ID, `{"urgency":"normal"}`); w.Code != 200 {
		t.Fatalf("change on the way: %d %s", w.Code, w.Body.String())
	}
	if w := s.as(beto, "PATCH", "/api/full/scout/needs/"+created.ID, `{"status":"DELETED"}`); w.Code != 400 || !strings.Contains(w.Body.String(), "en camino") {
		t.Fatalf("remove on the way: %d %s", w.Code, w.Body.String())
	}
	// The change history names the Scout page.
	if trail := s.call("GET", "/api/admin/camps/"+s.camp+"/trail", "", true).Body.String(); !strings.Contains(trail, `"Scout"`) || !strings.Contains(trail, "Beto Sur") {
		t.Fatalf("history page: %s", trail)
	}
}
