package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// An agent from no team or another team can't overwrite a need someone else last updated: the change becomes a
// request for that person, who accepts (it applies) or rejects (nothing changes). Same-team agents edit directly.
func TestAgentChangeNeedsOwnersOK(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	fufo := user(`{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`)
	user(`{"name":"Gus","telephone":"6999-2222","roles":["agent"],"accessCode":"GUS123"}`)
	mate := user(`{"name":"Mar","telephone":"6999-3333","roles":["agent"],"accessCode":"MAR123"}`)
	s.team("Norte", fufo, mate)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	// Fufo takes it on the runner page: he is now the one who last updated it.
	var runner *http.Cookie
	login := s.crewCall("POST", "/api/runner/session", `{"accessCode":"FUFO12"}`, nil, "")
	for _, c := range login.Result().Cookies() {
		if c.Name == "runner_session" {
			runner = c
		}
	}
	var sess struct{ CSRF string }
	json.Unmarshal(login.Body.Bytes(), &sess)
	rc := sess.CSRF
	if w := s.crewCall("POST", "/api/runner/needs/"+need+"/take", `{}`, runner, rc); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	stored := func() CampNeed {
		for _, n := range s.needsOf(s.camp) {
			if n.ID == need {
				return n
			}
		}
		t.Fatal("need missing")
		return CampNeed{}
	}
	if n := stored(); n.UpdatedBy != fufo {
		t.Fatalf("taking marks the runner: %+v", n)
	}
	gusC, gusCSRF := s.agentLogin("GUS123")
	w := s.agentCall("PATCH", "/api/agent/needs/"+need, `{"status":"COMPLETED","urgency":"urgente"}`, gusC, gusCSRF)
	var out struct{ Request ChangeRequest }
	json.Unmarshal(w.Body.Bytes(), &out)
	if w.Code != 202 || out.Request.ToName != "Fufo" || out.Request.FromName != "Gus" || len(out.Request.Changes) != 2 {
		t.Fatalf("request instead of change: %d %s", w.Code, w.Body.String())
	}
	if n := stored(); n.Status != "IN_PROGRESS" || n.Urgency != "normal" {
		t.Fatalf("need untouched until Fufo answers: %+v", n)
	}
	// Asking again replaces the pending request (one per agent and need).
	s.agentCall("PATCH", "/api/agent/needs/"+need, `{"status":"COMPLETED"}`, gusC, gusCSRF)
	var list struct{ ForMe, Sent []ChangeRequest }
	json.Unmarshal(s.crewCall("GET", "/api/runner/change-requests", "", runner, "").Body.Bytes(), &list)
	if len(list.ForMe) != 1 || list.ForMe[0].ID != out.Request.ID || len(list.ForMe[0].Changes) != 1 {
		t.Fatalf("Fufo's requests: %+v", list)
	}
	json.Unmarshal(s.agentCall("GET", "/api/agent/change-requests", "", gusC, "").Body.Bytes(), &list)
	if len(list.Sent) != 1 || list.Sent[0].Status != "pending" || len(list.ForMe) != 0 {
		t.Fatalf("Gus's sent: %+v", list)
	}
	// Gus can't answer his own request; Fufo rejects: nothing changes.
	if w := s.agentCall("POST", "/api/agent/change-requests/"+out.Request.ID+"/accept", `{}`, gusC, gusCSRF); w.Code != 409 {
		t.Fatalf("only the owner answers: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/change-requests/"+out.Request.ID+"/reject", `{}`, runner, rc); w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"rejected"`) {
		t.Fatalf("reject: %d %s", w.Code, w.Body.String())
	}
	if n := stored(); n.Status != "IN_PROGRESS" {
		t.Fatalf("rejected: %+v", n)
	}
	// A new request accepted by Fufo applies, and the need stays his.
	w = s.agentCall("PATCH", "/api/agent/needs/"+need, `{"urgency":"urgente"}`, gusC, gusCSRF)
	json.Unmarshal(w.Body.Bytes(), &out)
	if w := s.crewCall("POST", "/api/runner/change-requests/"+out.Request.ID+"/accept", `{}`, runner, rc); w.Code != 200 {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	if n := stored(); n.Urgency != "urgente" || n.UpdatedBy != fufo {
		t.Fatalf("accepted: %+v", n)
	}
	if w := s.crewCall("POST", "/api/runner/change-requests/"+out.Request.ID+"/accept", `{}`, runner, rc); w.Code != 409 {
		t.Fatalf("answered once: %d", w.Code)
	}
	// Mar (agent on Fufo's team) edits directly; then the need is Mar's, so Gus asks Mar.
	marC, marCSRF := s.agentLogin("MAR123")
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need, `{"urgency":"normal"}`, marC, marCSRF); w.Code != 200 {
		t.Fatalf("same team edits: %d %s", w.Code, w.Body.String())
	}
	w = s.agentCall("PATCH", "/api/agent/needs/"+need, `{"urgency":"urgente"}`, gusC, gusCSRF)
	json.Unmarshal(w.Body.Bytes(), &out)
	if w.Code != 202 || out.Request.ToName != "Mar" {
		t.Fatalf("ask the last one who updated it: %d %s", w.Code, w.Body.String())
	}
	// Admin with the master password is not an agent: it still edits directly. A no-op edit is not a request.
	s.expect("PATCH", "/api/admin/needs/"+need, `{"urgency":"urgente"}`, 200)
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need, `{"urgency":"urgente"}`, gusC, gusCSRF); w.Code != 200 {
		t.Fatalf("no change, no request: %d %s", w.Code, w.Body.String())
	}
	// The public never sees who updated it.
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, "updatedBy") {
		t.Fatalf("public: %s", body)
	}
}
