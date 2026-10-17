package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// After the center dispatches a run, the runner asks to change the load (the center accepts or not) instead of editing
// it; and «No puedo continuar» with the load on board either returns it to the center or takes it somewhere else.
func TestRunStepsPackingChangeAndStop(t *testing.T) {
	s := newServiceHarness(t)
	julie := s.expect("POST", "/api/admin/users", `{"name":"Julie","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)["id"].(string)
	pucha := s.expect("POST", "/api/admin/users", `{"name":"Pucha","telephone":"6555-0000","roles":["runner"],"accessCode":"PUCHA77"}`, 201)["id"].(string)
	s.team("Unidos", julie, pucha)
	jc, jt := s.supplierLogin("JULIE77")
	dispatch := func() string {
		w := s.agentCall("POST", "/api/supplier/dispatches", `{"pointId":"`+s.camp+`","runnerId":"`+pucha+`","items":[{"item":"Agua","quantity":3}]}`, jc, jt)
		var res struct{ ID string }
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &res) != nil {
			t.Fatalf("dispatch: %d %s", w.Code, w.Body.String())
		}
		return res.ID
	}
	login := s.crewCall("POST", "/api/runner/session", `{"accessCode":"PUCHA77"}`, nil, "")
	var body map[string]any
	json.Unmarshal(login.Body.Bytes(), &body)
	rc, rt := login.Result().Cookies()[0], body["csrf"].(string)
	run := dispatch()
	if w := s.crewCall("POST", "/api/runner/needs/"+run+"/packing", `{"items":[{"item":"Toallas","quantity":1}]}`, rc, rt); w.Code != 409 || !strings.Contains(w.Body.String(), "Pedir cambio") {
		t.Fatalf("dispatched: no direct edit: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+run+"/packing/change", `{"items":[{"item":"Agua","quantity":2},{"item":"Toallas","quantity":1}],"note":"Se rompió un galón"}`, rc, rt); w.Code != 200 {
		t.Fatalf("ask for a change: %d %s", w.Code, w.Body.String())
	}
	// Adding to the pending request: the new list replaces it and the reason is appended.
	if w := s.crewCall("POST", "/api/runner/needs/"+run+"/packing/change", `{"items":[{"item":"Agua","quantity":2},{"item":"Toallas","quantity":1}],"note":"y faltan toallas"}`, rc, rt); w.Code != 200 || !strings.Contains(w.Body.String(), `Se rompió un galón\ny faltan toallas`) {
		t.Fatalf("add to the pending change: %d %s", w.Code, w.Body.String())
	}
	if b := s.agentCall("GET", "/api/supplier/runs", "", jc, jt).Body.String(); !strings.Contains(b, `"packingChange":{"items":[{"item":"Agua","quantity":2`) || !strings.Contains(b, "Se rompió un galón") {
		t.Fatalf("the center sees the request: %s", b)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+run+"/packing/change", `{"accept":true}`, jc, jt); w.Code != 200 || !strings.Contains(w.Body.String(), `"item":"Toallas"`) {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	// A second request, rejected: a reason is required, and the runner sees both answers.
	s.crewCall("POST", "/api/runner/needs/"+run+"/packing/change", `{"items":[{"item":"Agua","quantity":9}],"note":"más agua"}`, rc, rt)
	if w := s.agentCall("POST", "/api/supplier/needs/"+run+"/packing/change", `{"accept":false}`, jc, jt); w.Code != 400 {
		t.Fatalf("reject needs a reason: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+run+"/packing/change", `{"accept":false,"reason":"No tenemos más agua"}`, jc, jt); w.Code != 200 {
		t.Fatalf("reject: %d %s", w.Code, w.Body.String())
	}
	if b := s.crewCall("GET", "/api/runner/needs", "", rc, "").Body.String(); !strings.Contains(b, `"decision":"accepted"`) || !strings.Contains(b, `"decision":"rejected","decidedByName":"Julie"`) || !strings.Contains(b, "No tenemos más agua") {
		t.Fatalf("the runner sees the log: %s", b)
	}
	// «No puedo continuar» → back to the center: the need reopens as that center's «Necesito runner» offer, not dispatched.
	if w := s.crewCall("POST", "/api/runner/needs/"+run+"/stop", `{"action":"return","note":"Se me dañó el carro"}`, rc, rt); w.Code != 200 {
		t.Fatalf("return: %d %s", w.Code, w.Body.String())
	}
	for _, n := range s.needsOf(s.camp) {
		if n.ID == run && (n.Status != "OPEN" || n.RunnerID != "" || n.DispatchedAt != "" || n.ReturnedAt == "" || len(n.Packing) != 2 || len(n.OpenOffers) != 1 || n.OpenOffers[0].SupplierID != julie) {
			t.Fatalf("returned: %+v", n)
		}
	}
	// Another dispatch, taken somewhere else (an area only): a new run there with the load; the original reopens empty.
	run2 := dispatch()
	w := s.crewCall("POST", "/api/runner/needs/"+run2+"/stop", `{"action":"redirect","point":{"area":"El Cacao","toLocate":true},"note":"El camino al punto está cerrado"}`, rc, rt)
	var moved struct{ ID, CampID, Status string }
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &moved) != nil || moved.Status != "IN_PROGRESS" || moved.CampID == s.camp {
		t.Fatalf("redirect: %d %s", w.Code, w.Body.String())
	}
	for _, n := range s.needsOf(moved.CampID) {
		if n.ID == moved.ID && (n.RunnerID != pucha || n.SupplierID != julie || n.DispatchedAt == "" || len(n.Packing) != 1 || n.RedirectedFrom != run2) {
			t.Fatalf("the load goes on as a new run: %+v", n)
		}
	}
	for _, n := range s.needsOf(s.camp) {
		if n.ID == run2 && (n.Status != "OPEN" || len(n.Packing) != 0 || n.RedirectedTo != moved.ID) {
			t.Fatalf("the original need reopens without the load: %+v", n)
		}
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+moved.ID+"/stop", `{"action":"redirect","note":"x"}`, rc, rt); w.Code != 400 {
		t.Fatalf("redirect needs a place: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+moved.ID+"/stop", `{"action":"return"}`, rc, rt); w.Code != 400 || !strings.Contains(w.Body.String(), "Cuenta qué pasó") {
		t.Fatalf("«¿Qué pasó?» is required: %d %s", w.Code, w.Body.String())
	}
}

// Everyone working at the same supply center sees its runs in «Despachando» and can work them (load, «Despachada»);
// another center only sees them as «en camino».
func TestCenterColleaguesShareDespachando(t *testing.T) {
	s := newServiceHarness(t)
	julie := s.expect("POST", "/api/admin/users", `{"name":"Julie","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/users", `{"name":"Yas","telephone":"6999-2020","roles":["supplier"],"accessCode":"YASYAS7"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Otro","telephone":"6999-3030","roles":["supplier"],"accessCode":"OTRO777"}`, 201)
	pucha := s.expect("POST", "/api/admin/users", `{"name":"Pucha","telephone":"6555-0000","roles":["runner"],"accessCode":"PUCHA77"}`, 201)["id"].(string)
	s.team("Unidos", julie, pucha)
	jc, jt := s.supplierLogin("JULIE77")
	yc, yt := s.supplierLogin("YASYAS7")
	oc, ot := s.supplierLogin("OTRO777")
	s.agentCall("POST", "/api/supplier/center", `{"center":"Donde Rubén"}`, jc, jt)
	s.agentCall("POST", "/api/supplier/center", `{"center":"donde ruben"}`, yc, yt)
	s.agentCall("POST", "/api/supplier/center", `{"center":"Las Tablas"}`, oc, ot)
	var n CampNeed
	json.Unmarshal(s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, true).Body.Bytes(), &n)
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/assign", `{"runnerId":"`+pucha+`"}`, jc, jt); w.Code != 200 {
		t.Fatalf("Julie assigns: %d %s", w.Code, w.Body.String())
	}
	runs := func(c *http.Cookie, tok string) (mine, enroute []string) {
		var data struct{ Runs, Enroute []supplierRun }
		json.Unmarshal(s.agentCall("GET", "/api/supplier/runs", "", c, tok).Body.Bytes(), &data)
		for _, r := range data.Runs {
			mine = append(mine, r.ID)
		}
		for _, r := range data.Enroute {
			enroute = append(enroute, r.ID)
		}
		return
	}
	if mine, _ := runs(yc, yt); len(mine) != 1 || mine[0] != n.ID {
		t.Fatalf("Yas (same center) sees it in «Despachando»: %v", mine)
	}
	if mine, enroute := runs(oc, ot); len(mine) != 0 || len(enroute) != 1 {
		t.Fatalf("another center: only «en camino»: %v %v", mine, enroute)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/packing", `{"items":[{"item":"Agua","quantity":2}]}`, yc, yt); w.Code != 200 {
		t.Fatalf("Yas adds to the load: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/dispatch", `{}`, yc, yt); w.Code != 200 || !strings.Contains(w.Body.String(), "Yas") {
		t.Fatalf("Yas marks it «Despachada»: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/dispatch", `{"undo":true}`, oc, ot); w.Code == 200 {
		t.Fatalf("another center cannot touch it: %d", w.Code)
	}
	// From Historial, the center completes the run for the runner and records what they still need.
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/complete", `{}`, oc, ot); w.Code == 200 {
		t.Fatalf("another center cannot complete it: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+n.ID+"/complete", `{"need":"Lonas","urgency":"urgente"}`, yc, yt); w.Code != 200 {
		t.Fatalf("complete for the runner: %d %s", w.Code, w.Body.String())
	}
	done, follow := false, false
	for _, x := range s.needsOf(s.camp) {
		if x.ID == n.ID && x.Status == "COMPLETED" && x.CompletedAt != "" {
			done = true
		}
		if x.Need == "Lonas" && x.Status == "OPEN" && x.Urgency == "urgente" {
			follow = true
		}
	}
	if !done || !follow {
		t.Fatalf("completed %v, follow-up need %v", done, follow)
	}
}

// Runners → 🕓 Historial: only my delivered runs, newest first, with place, center and load.
func TestRunnerHistory(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Pucha","telephone":"6555-0000","roles":["runner"],"accessCode":"PUCHA77"}`, 201)
	login := s.crewCall("POST", "/api/runner/session", `{"accessCode":"PUCHA77"}`, nil, "")
	var body map[string]any
	json.Unmarshal(login.Body.Bytes(), &body)
	rc, rt := login.Result().Cookies()[0], body["csrf"].(string)
	var n CampNeed
	json.Unmarshal(s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, true).Body.Bytes(), &n)
	if w := s.crewCall("POST", "/api/runner/needs/"+n.ID+"/take", "{}", rc, rt); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("GET", "/api/runner/history", "", rc, ""); w.Body.String() != "[]\n" && w.Body.String() != "[]" {
		t.Fatalf("nothing delivered yet: %s", w.Body.String())
	}
	s.crewCall("POST", "/api/runner/needs/"+n.ID+"/complete", "{}", rc, rt)
	if w := s.crewCall("GET", "/api/runner/history", "", rc, ""); !strings.Contains(w.Body.String(), n.ID) || !strings.Contains(w.Body.String(), `"completedAt"`) {
		t.Fatalf("delivered run in the history: %s", w.Body.String())
	}
}
