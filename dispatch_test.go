package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «＋ Nuevo despacho»: Supply dispatches to a new or existing destination with a new or existing runner, in one step.
func TestSupplyNewDispatch(t *testing.T) {
	s := newServiceHarness(t)
	julie := s.expect("POST", "/api/admin/users", `{"name":"Julie Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)["id"].(string)
	other := s.expect("POST", "/api/admin/users", `{"name":"Fuera Equipo","telephone":"6555-0000","roles":["runner"],"accessCode":"FUERA1"}`, 201)["id"].(string)
	s.team("Unidos por Azuero", julie)
	s.team("Otro", other)
	jc, jt := s.supplierLogin("JULIE77")
	// New destination (no phone needed) + new runner whose phone matches someone OUTSIDE the team: not linked.
	w := s.agentCall("POST", "/api/supplier/dispatches", `{"point":{"area":"Río Viejo","place":"Casa de Omaira","directions":"después del puente"},"runner":{"name":"Pucha","phone":"6555-0000","note":"Tiene moto"},"need":"Alimentos, Agua","items":[{"item":"Bolsas alimentación","quantity":2},{"item":"Agua","quantity":3}]}`, jc, jt)
	if w.Code != 201 {
		t.Fatalf("new dispatch: %d %s", w.Code, w.Body.String())
	}
	var res struct{ ID, PointID, RunnerID string }
	json.Unmarshal(w.Body.Bytes(), &res)
	if res.RunnerID == other {
		t.Fatal("a person from another team must not be linked")
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		if u.ID == res.RunnerID && (u.Name != "Pucha" || !slicesContains(u.Roles, "runner") || u.TeamID == "" || !u.HasAccessCode || u.Notes != "Tiene moto" || !strings.Contains(u.Note, "Supply")) {
			t.Fatalf("new runner in the supplier's team: %+v", u)
		}
	}
	if b := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(b, "Casa de Omaira") {
		t.Fatal("the new destination is published")
	}
	v := s.supplierView(jc)
	run := findRun(v.Runs, res.ID)
	if run == nil || run.Status != "IN_PROGRESS" || run.DispatchedAt == "" || len(run.Packing) != 2 || run.RunnerName != "Pucha" {
		t.Fatalf("dispatched run with its packing: %+v", run)
	}
	// Existing point + existing runner (same runner again: linked by phone within the team).
	w = s.agentCall("POST", "/api/supplier/dispatches", `{"pointId":"`+s.camp+`","runner":{"name":"Pucha R.","phone":"+507 6555 0000"},"items":[{"item":"Agua","quantity":1}]}`, jc, jt)
	var res2 struct{ ID, PointID, RunnerID string }
	json.Unmarshal(w.Body.Bytes(), &res2)
	if w.Code != 201 || res2.PointID != s.camp || res2.RunnerID != res.RunnerID {
		t.Fatalf("existing point, same runner linked within the team: %d %s", w.Code, w.Body.String())
	}
	// Generic destination (only the zone): kept «📍 Por ubicar», not public; the dispatching center can locate it.
	if w := s.agentCall("POST", "/api/supplier/dispatches", `{"pointId":"`+s.camp+`","runner":{"name":"Pucha"},"items":[]}`, jc, jt); w.Code != 400 {
		t.Fatalf("dispatch without items = %d, want 400", w.Code)
	}
	w = s.agentCall("POST", "/api/supplier/dispatches", `{"point":{"area":"El Cacao","toLocate":true},"runner":{"name":"Pucha","phone":"6555-0000"},"items":[{"item":"Agua","quantity":1}]}`, jc, jt)
	var res3 struct{ ID, PointID, RunnerID string }
	json.Unmarshal(w.Body.Bytes(), &res3)
	if w.Code != 201 {
		t.Fatalf("generic destination: %d %s", w.Code, w.Body.String())
	}
	if b := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(b, res3.PointID) {
		t.Fatal("a generic destination is not public")
	}
	if b := s.as(s.viewerCookie("JULIE77"), "GET", "/api/full/tolocate", "").Body.String(); !strings.Contains(b, res3.PointID) || !strings.Contains(b, "El Cacao") {
		t.Fatalf("the dispatching center sees it «por ubicar»: %s", b)
	}
	if w := s.agentCall("POST", "/api/supplier/dispatches", `{"runner":{"name":"X"}}`, jc, jt); w.Code != 400 {
		t.Fatalf("destination required: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/dispatches", `{"pointId":"`+s.camp+`","runner":{"phone":"6000-0000"}}`, jc, jt); w.Code != 400 {
		t.Fatalf("runner name required: %d", w.Code)
	}
}
