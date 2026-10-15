package main

import (
	"net/url"
	"strings"
	"testing"
)

// Snapshot points have IDs like «seed:s/c-1»: the command center must reach them like admin does (one escaped segment).
func TestAgentProxyKeepsEscapedSlashInIDs(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("POST", "/api/admin/seed", `[{"code":"S/C-1","area":"Cañas","contact":"Ana","phone":"6000-1111"}]`, true); w.Code != 200 {
		t.Fatalf("seed: %d %s", w.Code, w.Body.String())
	}
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-1111","roles":["agent"],"accessCode":"ANA123"}`, 201)
	cookie, csrf := s.agentLogin("ANA123")
	id := url.PathEscape("seed:s/c-1")
	if w := s.agentCall("POST", "/api/agent/camps/"+id+"/needs", `{"need":"Agua","urgency":"normal"}`, cookie, csrf); w.Code != 201 || !strings.Contains(w.Body.String(), `"campId":"seed:s/c-1"`) {
		t.Fatalf("need on a seed point through /api/agent: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("PATCH", "/api/agent/camps/"+id, `{"code":"S/C-1","area":"Cañas","place":"Escuela"}`, cookie, csrf); w.Code != 200 || !strings.Contains(w.Body.String(), "Escuela") {
		t.Fatalf("edit a seed point through /api/agent: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("GET", "/api/agent/camps/"+id+"/needs", "", cookie, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Agua") {
		t.Fatalf("list its needs: %d %s", w.Code, w.Body.String())
	}
}
