package main

import (
	"strings"
	"testing"
)

// The command center can add a runner by name and phone (no access code) when the name typed in a need is new.
func TestAgentCreatesRunnerFromNeedWindow(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-1111","roles":["agent"],"accessCode":"ANA123"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Rui","telephone":"6999-2222","roles":["runner"],"accessCode":"RUI123"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Bea","telephone":"6999-3333","roles":["builder"],"accessCode":"BEA123"}`, 201)
	cookie, csrf := s.agentLogin("ANA123")
	if w := s.agentCall("POST", "/api/agent/runners", `{"name":"Leo","telephone":"6000-1234"}`, cookie, ""); w.Code != 403 {
		t.Fatalf("needs the CSRF token: %d", w.Code)
	}
	w := s.agentCall("POST", "/api/agent/runners", `{"name":" Leo  Díaz ","telephone":"6000-1234"}`, cookie, csrf)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"name":"Leo Díaz"`) || !strings.Contains(w.Body.String(), `"roles":["runner"]`) {
		t.Fatalf("new runner: %d %s", w.Code, w.Body.String())
	}
	if list := s.call("GET", "/api/admin/users", "", true).Body.String(); !strings.Contains(list, "Creado desde el centro de comando por Ana Agente") {
		t.Fatalf("admin sees who added it: %s", list)
	}
	// Same phone as an existing runner: that runner is reused; as someone else: refused, untouched.
	if w := s.agentCall("POST", "/api/agent/runners", `{"name":"Rui Otro","telephone":"69992222"}`, cookie, csrf); w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Rui"`) {
		t.Fatalf("existing runner reused: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/agent/runners", `{"name":"Bea Dos","telephone":"6999-3333"}`, cookie, csrf); w.Code != 400 || !strings.Contains(w.Body.String(), "ya es de Bea") {
		t.Fatalf("someone else's phone: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/agent/runners", `{"name":"Sin Tel","telephone":"12"}`, cookie, csrf); w.Code != 400 {
		t.Fatalf("phone required: %d %s", w.Code, w.Body.String())
	}
}
