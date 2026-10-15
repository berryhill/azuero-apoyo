package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func (s *serviceHarness) supplierLogin(code string) (*http.Cookie, string) {
	s.t.Helper()
	w := s.agentCall("POST", "/api/supplier/session", `{"accessCode":"`+code+`"}`, nil, "")
	if w.Code != 200 {
		s.t.Fatalf("supplier login %s: %d %s", code, w.Code, w.Body.String())
	}
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	for _, c := range w.Result().Cookies() {
		if c.Name == "supplier_session" {
			return c, out["csrf"].(string)
		}
	}
	s.t.Fatal("no supplier_session cookie")
	return nil, ""
}

// Acopio sees runs in progress (with the runner), packs them, adds departure photos and marks them «Despachada»
// — a step inside «En camino»: the run stays IN_PROGRESS for the runner to complete.
func TestSupplierPacksAndDispatchesRuns(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Sara Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"SARA77"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua, Arroz","status":"OPEN"}`, 201)
	open := s.needsOf(s.camp)[0]
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-3030","roles":["agent"],"accessCode":"ANA777"}`, 201)
	if w := s.agentCall("POST", "/api/supplier/session", `{"accessCode":"ANA777"}`, nil, ""); w.Code != 403 {
		t.Fatalf("an agent code is not Acopio: %d", w.Code)
	}
	cookie, csrf := s.supplierLogin("SARA77")
	list := func() []supplierRun {
		w := s.agentCall("GET", "/api/supplier/runs", "", cookie, "")
		if w.Code != 200 {
			t.Fatalf("runs: %d %s", w.Code, w.Body.String())
		}
		var out struct{ Runs []supplierRun }
		json.Unmarshal(w.Body.Bytes(), &out)
		return out.Runs
	}
	if runs := list(); len(runs) != 0 {
		t.Fatalf("open needs are not runs yet: %+v", runs)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/dispatch", `{}`, cookie, csrf); w.Code != 409 {
		t.Fatalf("cannot dispatch an open need: %d", w.Code)
	}
	// The runner takes it: Acopio sees it with the runner's name and phone.
	rc := s.runnerLogin("Rui Corre")
	if w := rc.call("POST", "/api/runner/needs/"+open.ID+"/take", ``); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	runs := list()
	if len(runs) != 1 || runs[0].RunnerName != "Rui Corre" || runs[0].RunnerPhone != "6000-0000" || runs[0].DispatchedAt != "" {
		t.Fatalf("run for Acopio: %+v", runs)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/packing", `{"items":[{"item":"Agua","quantity":4,"unit":"galones"}]}`, cookie, csrf); w.Code != 200 || !strings.Contains(w.Body.String(), "Sara Acopio") {
		t.Fatalf("pack: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/dispatch", `{}`, cookie, "bad"); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/dispatch", `{}`, cookie, csrf); w.Code != 200 || !strings.Contains(w.Body.String(), "Sara Acopio") {
		t.Fatalf("dispatch: %d %s", w.Code, w.Body.String())
	}
	n := s.needsOf(s.camp)[0]
	if n.Status != "IN_PROGRESS" || n.DispatchedAt == "" || n.DispatchedByName != "Sara Acopio" {
		t.Fatalf("dispatched stays en camino: %+v", n)
	}
	// The runner sees it; the public list shows the step but not who dispatched.
	if w := rc.call("GET", "/api/runner/needs", ""); !strings.Contains(w.Body.String(), `"dispatchedAt"`) {
		t.Fatalf("runner sees dispatched: %s", w.Body.String())
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if !strings.Contains(pub, `"dispatchedAt"`) || strings.Contains(pub, "Sara Acopio") || strings.Contains(pub, "dispatchedBy\"") {
		t.Fatalf("public: %s", pub)
	}
	// Undo, then the runner releases it: back to OPEN, dispatch cleared, gone from Acopio.
	s.expect("POST", "/api/supplier/needs/"+open.ID+"/dispatch", ``, 401)
	if w := s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/dispatch", `{"undo":true}`, cookie, csrf); w.Code != 200 || strings.Contains(w.Body.String(), "Sara") {
		t.Fatalf("undo: %d %s", w.Code, w.Body.String())
	}
	s.agentCall("POST", "/api/supplier/needs/"+open.ID+"/dispatch", `{}`, cookie, csrf)
	if w := rc.call("POST", "/api/runner/needs/"+open.ID+"/release", ``); w.Code != 200 {
		t.Fatalf("release: %d %s", w.Code, w.Body.String())
	}
	if n := s.needsOf(s.camp)[0]; n.Status != "OPEN" || n.DispatchedAt != "" {
		t.Fatalf("released clears dispatch: %+v", n)
	}
	if runs := list(); len(runs) != 0 {
		t.Fatalf("released run leaves Acopio: %+v", runs)
	}
	// Code login lists the Acopio panel.
	if w := s.call("POST", "/api/session/code", `{"accessCode":"SARA77"}`, false); !strings.Contains(w.Body.String(), "/supplier.html") {
		t.Fatalf("panels: %s", w.Body.String())
	}
}
