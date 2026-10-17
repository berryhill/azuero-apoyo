package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «＋ Nuevo run» (Runners): the runner starts a run to an existing need, to an existing point with its own load, or to a
// new place; the packing list is required; it lands in «Yo» like any other run.
func TestRunnerNewRun(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","status":"OPEN"}`, 201)
	water := s.needsOf(s.camp)[0].ID
	rc := s.runnerLogin("Rui Corre")
	// Destinations: published points with their open needs (takeable), no contacts.
	w := rc.call("GET", "/api/runner/destinations", "")
	var dests []runDestination
	json.Unmarshal(w.Body.Bytes(), &dests)
	var found *runDestination
	for i := range dests {
		if dests[i].ID == s.camp {
			found = &dests[i]
		}
	}
	if w.Code != 200 || found == nil || len(found.Needs) != 1 || !found.Needs[0].Takeable || strings.Contains(w.Body.String(), `"contacts"`) {
		t.Fatalf("destinations: %d %s", w.Code, w.Body.String())
	}
	// Packing list required.
	if w := rc.call("POST", "/api/runner/runs", `{"pointId":"`+s.camp+`","needId":"`+water+`","items":[]}`); w.Code != 400 {
		t.Fatalf("no items = %d, want 400", w.Code)
	}
	// Existing need: it becomes the runner's, en camino, with the load.
	w = rc.call("POST", "/api/runner/runs", `{"pointId":"`+s.camp+`","needId":"`+water+`","items":[{"item":"Agua","quantity":3}]}`)
	if w.Code != 201 {
		t.Fatalf("take existing need: %d %s", w.Code, w.Body.String())
	}
	if n := s.needsOf(s.camp)[0]; n.Status != "IN_PROGRESS" || n.RunnerID == "" || len(n.Packing) != 1 {
		t.Fatalf("existing need taken with its packing: %+v", n)
	}
	if w := s.runnerLogin("Ana Corre").call("POST", "/api/runner/runs", `{"pointId":"`+s.camp+`","needId":"`+water+`","items":[{"item":"Agua","quantity":1}]}`); w.Code != 400 {
		t.Fatalf("someone else already carries it: %d", w.Code)
	}
	// Own run to an existing point, then to a new place.
	if w := rc.call("POST", "/api/runner/runs", `{"pointId":"`+s.camp+`","items":[{"item":"Arroz","quantity":2}]}`); w.Code != 201 {
		t.Fatalf("own run to an existing point: %d %s", w.Code, w.Body.String())
	}
	w = rc.call("POST", "/api/runner/runs", `{"point":{"area":"Agua Buena","place":"Casa de Elia"},"items":[{"item":"Agua","quantity":1}],"need":"Agua"}`)
	var res struct{ ID, PointID string }
	json.Unmarshal(w.Body.Bytes(), &res)
	if w.Code != 201 || res.PointID == "" || res.PointID == s.camp {
		t.Fatalf("new place: %d %s", w.Code, w.Body.String())
	}
	if b := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(b, "Casa de Elia") {
		t.Fatal("the new place is published")
	}
	if _, _, raw := rc.needs(); !strings.Contains(raw, res.ID) || !strings.Contains(raw, "Arroz") {
		t.Fatalf("all three runs are in «Yo»: %s", raw)
	}
	if w := rc.call("POST", "/api/runner/runs", `{"items":[{"item":"Agua","quantity":1}]}`); w.Code != 400 {
		t.Fatalf("destination required: %d", w.Code)
	}
	if w := rc.call("GET", "/api/runner/maplink?url=", ""); w.Code != 400 {
		t.Fatalf("map link required: %d", w.Code)
	}
}
