package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPackingListAddReviewAndCatalog(t *testing.T) {
	s := newServiceHarness(t)
	open := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua, Alimentos","urgency":"normal"}`, 201)["id"].(string)
	other := s.expect("POST", "/api/admin/users", `{"name":"Otro Corredor","telephone":"6999-0000","roles":["runner"]}`, 201)["id"].(string)
	theirs := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Refugio","urgency":"normal","runnerId":"`+other+`"}`, 201)["id"].(string)
	done := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Kit de bebé","urgency":"normal"}`, 201)["id"].(string)
	s.expect("PATCH", "/api/admin/needs/"+done, `{"status":"COMPLETED"}`, 200)

	if w := s.call("GET", "/api/runner/packing-items", "", false); w.Code != 401 {
		t.Fatalf("catalog requires a session: %d", w.Code)
	}
	cookie, csrf := s.crewLogin("/api/runner", "Rosa")
	var items []PackingItem
	json.Unmarshal(s.crewCall("GET", "/api/runner/packing-items", "", cookie, "").Body.Bytes(), &items)
	if len(items) != len(defaultPackingItems) {
		t.Fatalf("catalog starts with the default items: %v", items)
	}

	for body, status := range map[string]int{
		`{"items":[]}`: 400,
		`{"items":[{"item":"Linternas","quantity":0}]}`:   400,
		`{"items":[{"item":"  ","quantity":2}]}`:          400,
		`{"items":[{"item":"Linternas","quantity":2.5}]}`: 400,
		`{"items":"Linternas"}`:                           400,
	} {
		if w := s.crewCall("POST", "/api/runner/needs/"+open+"/packing", body, cookie, csrf); w.Code != status {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	// A runner packs only for needs they took or were assigned.
	if w := s.crewCall("POST", "/api/runner/needs/"+open+"/packing", `{"items":[{"item":"Linternas","quantity":1}]}`, cookie, csrf); w.Code != 404 {
		t.Fatalf("packing before taking the need: %d", w.Code)
	}
	if body := s.crewCall("GET", "/api/runner/needs", "", cookie, "").Body.String(); strings.Contains(body, `"assigned":true`) {
		t.Fatalf("open needs are not marked as assigned: %s", body)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+open+"/take", "{}", cookie, csrf); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	w := s.crewCall("POST", "/api/runner/needs/"+open+"/packing", `{"items":[{"item":"Linternas","quantity":5},{"item":"Agua  embotellada","quantity":10},{"item":"linternas","quantity":2}]}`, cookie, csrf)
	if w.Code != 200 {
		t.Fatalf("add: %d %s", w.Code, w.Body.String())
	}
	var saved struct{ Packing []PackingLine }
	json.Unmarshal(w.Body.Bytes(), &saved)
	if len(saved.Packing) != 2 || saved.Packing[0].Item != "Linternas" || saved.Packing[0].Quantity != 7 || saved.Packing[1].Item != "Agua embotellada" || saved.Packing[0].AddedByName != "Rosa" {
		t.Fatalf("same item in one add is summed; new item kept: %+v", saved.Packing)
	}
	json.Unmarshal(s.crewCall("GET", "/api/runner/packing-items", "", cookie, "").Body.Bytes(), &items)
	if len(items) != len(defaultPackingItems)+1 {
		t.Fatalf("a typed item is saved for next time: %v", items)
	}
	// A second add appends; the runner list shows the packing list.
	s.crewCall("POST", "/api/runner/needs/"+open+"/packing", `{"items":[{"item":"Toallas","quantity":3}]}`, cookie, csrf)
	if body := s.crewCall("GET", "/api/runner/needs", "", cookie, "").Body.String(); !strings.Contains(body, `"item":"Toallas"`) || !strings.Contains(body, `"item":"Linternas"`) {
		t.Fatalf("runner list includes the packing list: %s", body)
	}

	if w := s.crewCall("POST", "/api/runner/needs/"+theirs+"/packing", `{"items":[{"item":"Sillas","quantity":1}]}`, cookie, csrf); w.Code != 404 {
		t.Fatalf("runner can't add to another runner's need: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+done+"/packing", `{"items":[{"item":"Sillas","quantity":1}]}`, cookie, csrf); w.Code != 404 {
		t.Fatalf("completed need that was never theirs: %d", w.Code)
	}
	// Their own run, once completed, no longer takes packing.
	finished := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Toallas","urgency":"normal"}`, 201)["id"].(string)
	s.crewCall("POST", "/api/runner/needs/"+finished+"/take", "{}", cookie, csrf)
	s.crewCall("POST", "/api/runner/needs/"+finished+"/complete", "{}", cookie, csrf)
	if w := s.crewCall("POST", "/api/runner/needs/"+finished+"/packing", `{"items":[{"item":"Sillas","quantity":1}]}`, cookie, csrf); w.Code != 409 {
		t.Fatalf("own completed run: %d %s", w.Code, w.Body.String())
	}

	// Agents: packing on anything open or en ruta, but no runs.
	s.expect("POST", "/api/admin/users", `{"name":"Agente Uno","telephone":"6999-3333","roles":["agent"],"accessCode":"AGENTE1"}`, 201)
	login := s.call("POST", "/api/runner/session", `{"accessCode":"AGENTE1"}`, false)
	var auth map[string]any
	json.Unmarshal(login.Body.Bytes(), &auth)
	agent, agentCSRF := login.Result().Cookies()[0], auth["csrf"].(string)
	if body := s.crewCall("GET", "/api/runner/needs", "", agent, "").Body.String(); !strings.Contains(body, theirs) || !strings.Contains(body, open) {
		t.Fatalf("agent sees open and assigned needs: %s", body)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+theirs+"/packing", `{"items":[{"item":"Sillas","quantity":4}]}`, agent, agentCSRF); w.Code != 200 || !strings.Contains(w.Body.String(), "Agente Uno") {
		t.Fatalf("agent packing: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+open+"/take", "{}", agent, agentCSRF); w.Code != 403 {
		t.Fatalf("agents don't take runs: %d %s", w.Code, w.Body.String())
	}

	// Admin sees the packing list; the public never does.
	if body := s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true).Body.String(); !strings.Contains(body, `"item":"Sillas"`) || !strings.Contains(body, `"item":"Linternas"`) {
		t.Fatalf("admin packing: %s", body)
	}
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, "packing") || strings.Contains(body, "Linternas") {
		t.Fatalf("public needs must not include packing lists: %s", body)
	}
}

func TestAdminAddsToPackingList(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal","status":"COMPLETED"}`, 201)["id"].(string)
	repair := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","kind":"repair","urgency":"normal"}`, 201)["id"].(string)
	if w := s.call("GET", "/api/admin/packing-items", "", false); w.Code != 401 {
		t.Fatalf("catalog requires admin: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/needs/"+id+"/packing", `{"items":[{"item":"Sillas","quantity":2}]}`, false); w.Code != 401 {
		t.Fatalf("packing requires admin: %d", w.Code)
	}
	var items []PackingItem
	json.Unmarshal(s.call("GET", "/api/admin/packing-items", "", true).Body.Bytes(), &items)
	if len(items) != len(defaultPackingItems) {
		t.Fatalf("admin catalog: %v", items)
	}
	// Admin may record packing even on a completed need.
	w := s.call("POST", "/api/admin/needs/"+id+"/packing", `{"items":[{"item":"Sillas","quantity":2},{"item":"Garrafones de agua","quantity":5}]}`, true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"addedByName":"Administración"`) {
		t.Fatalf("admin packing: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(s.call("GET", "/api/admin/packing-items", "", true).Body.Bytes(), &items)
	if len(items) != len(defaultPackingItems)+1 {
		t.Fatalf("new item saved to the catalog: %v", items)
	}
	if w := s.call("POST", "/api/admin/needs/"+repair+"/packing", `{"items":[{"item":"Sillas","quantity":1}]}`, true); w.Code != 404 {
		t.Fatalf("repairs have no packing list: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/needs/"+id+"/packing", `{"items":[{"item":"Sillas","quantity":-1}]}`, true); w.Code != 400 {
		t.Fatalf("bad quantity: %d", w.Code)
	}
}
