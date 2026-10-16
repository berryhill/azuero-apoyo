package main

import (
	"strings"
	"testing"
)

// Several runners on the same need: each one gets a run of their own (linked by SameAs), so nobody's runner, packing
// or photos are overwritten; edits by someone else to a runner's run become a new open need.
func TestSeveralRunnersKeepTheirOwnRuns(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	fufo := user(`{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`)
	ana := user(`{"name":"Ana","telephone":"6999-2222","roles":["runner"],"accessCode":"ANA123"}`)
	zoe := user(`{"name":"Zoe","telephone":"6999-3333","roles":["runner"]}`)
	user(`{"name":"Gus","telephone":"6999-4444","roles":["agent"],"accessCode":"GUS123"}`)
	s.team("Norte", fufo, ana)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua, Arroz","urgency":"normal","runnerId":"`+fufo+`"}`, 201)["id"].(string)
	fufoC := s.viewerCookie("FUFO12")
	if w := s.as(fufoC, "POST", "/api/full/needs/"+need+"/packing", `{"items":[{"item":"Agua 5 gal","quantity":3}]}`); w.Code != 200 {
		t.Fatalf("Fufo packs: %d %s", w.Code, w.Body.String())
	}
	if w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=pickup", fufoC, "", "image/jpeg", []byte("\xff\xd8\xff x")); w.Code != 201 {
		t.Fatalf("Fufo's photo: %d %s", w.Code, w.Body.String())
	}
	byID := func() map[string]CampNeed {
		out := map[string]CampNeed{}
		for _, n := range s.needsOf(s.camp) {
			out[n.ID] = n
		}
		return out
	}
	fufoIntact := func(when string) {
		t.Helper()
		n := byID()[need]
		if n.RunnerID != fufo || n.Status != "IN_PROGRESS" || n.Need != "Agua, Arroz" || len(n.Packing) != 1 || len(n.Media) != 1 {
			t.Fatalf("%s: Fufo's run changed: %+v", when, n)
		}
	}

	// Home «Asignar conductor» with Ana: Ana gets her own run; asking again for Ana is refused.
	w := s.as(fufoC, "POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+ana+`"}`)
	if w.Code != 200 || strings.Contains(w.Body.String(), need) || !strings.Contains(w.Body.String(), `"runnerName":"Ana"`) {
		t.Fatalf("second runner from home: %d %s", w.Code, w.Body.String())
	}
	fufoIntact("home assign")
	var anaRun CampNeed
	for _, n := range byID() {
		if n.RunnerID == ana {
			anaRun = n
		}
	}
	if anaRun.SameAs != need || anaRun.Status != "IN_PROGRESS" || anaRun.Need != "Agua, Arroz" || len(anaRun.Packing) != 0 || len(anaRun.Media) != 0 {
		t.Fatalf("Ana's run: %+v", anaRun)
	}
	if w := s.as(fufoC, "POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+ana+`"}`); w.Code != 400 {
		t.Fatalf("Ana already runs it: %d", w.Code)
	}
	// Ana's packing goes on her run only.
	if w := s.as(s.viewerCookie("ANA123"), "POST", "/api/full/needs/"+anaRun.ID+"/packing", `{"items":[{"item":"Arroz 25 lb","quantity":2}]}`); w.Code != 200 {
		t.Fatalf("Ana packs: %d", w.Code)
	}
	fufoIntact("Ana packs")

	// Admin/Command puts Zoe on Fufo's need and later says Zoe completed it: one run for Zoe, completed; Fufo's untouched.
	if n := s.expect("PATCH", "/api/admin/needs/"+need, `{"runnerId":"`+zoe+`","status":"IN_PROGRESS"}`, 200); n["id"] == need || n["runnerId"] != zoe {
		t.Fatalf("admin assigns Zoe: %+v", n)
	}
	if n := s.expect("PATCH", "/api/admin/needs/"+need, `{"runnerId":"`+zoe+`","status":"COMPLETED"}`, 200); n["id"] == need || n["status"] != "COMPLETED" {
		t.Fatalf("Zoe fulfilled it: %+v", n)
	}
	fufoIntact("admin Zoe")
	zoeRuns := 0
	for _, n := range byID() {
		if n.RunnerID == zoe {
			zoeRuns++
			if n.Status != "COMPLETED" || n.SameAs != need || n.CompletedAt == "" {
				t.Fatalf("Zoe's run: %+v", n)
			}
		}
	}
	if zoeRuns != 1 {
		t.Fatalf("one run for Zoe, got %d", zoeRuns)
	}

	// Someone else edits what Fufo's run asks for: the run keeps «Agua, Arroz»; the added item is a new open need.
	agentC, csrf := s.agentLogin("GUS123")
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need, `{"need":"Agua, Arroz, Pañales","urgency":"urgente"}`, agentC, csrf); w.Code != 200 {
		t.Fatalf("agent edit: %d %s", w.Code, w.Body.String())
	}
	fufoIntact("agent edit")
	if byID()[need].Urgency != "urgente" {
		t.Fatalf("urgency still changes on the run: %+v", byID()[need])
	}
	var added CampNeed
	for _, n := range byID() {
		if n.Need == "Pañales" {
			added = n
		}
	}
	if added.Status != "OPEN" || added.SameAs != need || added.CreatedByName != "Gus" || added.RunnerID != "" {
		t.Fatalf("added need: %+v", added)
	}
	// Only removing things from a runner's run is refused; the run is not changed.
	if w := s.agentCall("PATCH", "/api/agent/needs/"+need, `{"need":"Agua"}`, agentC, csrf); w.Code != 409 {
		t.Fatalf("removing from a run: %d %s", w.Code, w.Body.String())
	}
	fufoIntact("removal refused")
	// The public never sees who added the new need.
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, "Gus") || strings.Contains(body, "createdBy") {
		t.Fatalf("public: %s", body)
	}
}
