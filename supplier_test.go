package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
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

type supplierView struct {
	Runs    []supplierRun
	Open    []supplierRun
	Runners []supplierRunner
	Center  string
	Centers []string
}

func (s *serviceHarness) supplierView(cookie *http.Cookie) supplierView {
	s.t.Helper()
	w := s.agentCall("GET", "/api/supplier/runs", "", cookie, "")
	if w.Code != 200 {
		s.t.Fatalf("supplier runs: %d %s", w.Code, w.Body.String())
	}
	var v supplierView
	json.Unmarshal(w.Body.Bytes(), &v)
	return v
}

func findRun(list []supplierRun, id string) *supplierRun {
	for i := range list {
		if list[i].ID == id {
			return &list[i]
		}
	}
	return nil
}

// Supply links runs to its center: it asks runners for one («Pedidos» for runners, with who and where) or assigns a
// runner in person. Only its own linked runs show in «Por despachar»; it packs and marks them «Despachada»
// (still IN_PROGRESS for the runner to complete).
func TestSupplyRequestsRunnersAndDispatchesItsOwnRuns(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Julie Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Otro Acopio","telephone":"6999-1111","roles":["supplier"],"accessCode":"OTRO777"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6999-3030","roles":["agent"],"accessCode":"ANA777"}`, 201)
	if w := s.agentCall("POST", "/api/supplier/session", `{"accessCode":"ANA777"}`, nil, ""); w.Code != 403 {
		t.Fatalf("an agent code is not Supply: %d", w.Code)
	}
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua, Arroz","status":"OPEN"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Carpas","status":"OPEN"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Pañales","status":"OPEN"}`, 201)
	var water, tents, diapers string
	for _, n := range s.needsOf(s.camp) {
		switch n.Need {
		case "Agua, Arroz":
			water = n.ID
		case "Carpas":
			tents = n.ID
		case "Pañales":
			diapers = n.ID
		}
	}
	julie, jcsrf := s.supplierLogin("JULIE77")
	other, ocsrf := s.supplierLogin("OTRO777")
	if w := s.agentCall("POST", "/api/supplier/center", `{"center":"  Villa   Marina "}`, julie, jcsrf); w.Code != 200 || !strings.Contains(w.Body.String(), `"Villa Marina"`) {
		t.Fatalf("center: %d %s", w.Code, w.Body.String())
	}
	// The center joins the shared list; the other Supply picks it with different case/accents and gets the same name.
	if w := s.agentCall("POST", "/api/supplier/center", `{"center":"villa marína"}`, other, ocsrf); w.Code != 200 || !strings.Contains(w.Body.String(), `"Villa Marina"`) {
		t.Fatalf("reuse center: %d %s", w.Code, w.Body.String())
	}
	s.agentCall("POST", "/api/supplier/center", `{"center":"Las Tablas"}`, other, ocsrf)
	if c := s.supplierView(julie).Centers; strings.Join(c, "|") != "Las Tablas|Villa Marina" {
		t.Fatalf("shared centers: %q", c)
	}
	v := s.supplierView(julie)
	if len(v.Runs) != 0 || len(v.Open) != 3 || v.Center != "Villa Marina" {
		t.Fatalf("start: runs %d open %d center %q", len(v.Runs), len(v.Open), v.Center)
	}
	// A run a runner takes on their own is not linked to any Supply.
	rc := s.runnerLogin("Rui Corre")
	if w := rc.call("POST", "/api/runner/needs/"+diapers+"/take", ``); w.Code != 200 {
		t.Fatalf("take: %d", w.Code)
	}
	if v := s.supplierView(julie); len(v.Runs) != 0 || findRun(v.Open, diapers) != nil {
		t.Fatalf("unlinked run stays out of Supply: %+v", v.Runs)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+diapers+"/dispatch", `{}`, julie, jcsrf); w.Code != 409 {
		t.Fatalf("cannot dispatch someone else's run: %d", w.Code)
	}
	// Julie asks for a runner: the other Supply can't take over the request; runners see it with Julie and her center.
	if w := s.agentCall("POST", "/api/supplier/needs/"+water+"/request", `{}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("request: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+water+"/request", `{}`, other, ocsrf); w.Code != 409 {
		t.Fatalf("requested elsewhere: %d", w.Code)
	}
	if r := findRun(s.supplierView(other).Open, water); r == nil || r.RequestMine || r.Supplier == nil || r.Supplier.Name != "Julie Acopio" {
		t.Fatalf("other Supply sees who asked: %+v", r)
	}
	if r := findRun(s.supplierView(julie).Open, water); r == nil || !r.RequestMine {
		t.Fatalf("Julie sees her request: %+v", r)
	}
	body := rc.call("GET", "/api/runner/needs", "").Body.String()
	if !strings.Contains(body, `"requestedAt"`) || !strings.Contains(body, `"center":"Villa Marina"`) || !strings.Contains(body, `"phone":"6999-1010"`) {
		t.Fatalf("runner sees the request with who/where/phone: %s", body)
	}
	// The runner accepts: linked to Julie, gone from the request list, in her «Por despachar».
	if w := rc.call("POST", "/api/runner/needs/"+water+"/take", ``); w.Code != 200 {
		t.Fatalf("accept: %d", w.Code)
	}
	v = s.supplierView(julie)
	if r := findRun(v.Runs, water); r == nil || r.RunnerName != "Rui Corre" || r.DispatchedAt != "" || findRun(v.Open, water) != nil {
		t.Fatalf("accepted request: runs %+v", v.Runs)
	}
	if len(s.supplierView(other).Runs) != 0 {
		t.Fatalf("the other Supply does not see Julie's run")
	}
	// In person: Julie assigns «Carpas» to a runner who is there.
	var runnerID string
	for _, r := range v.Runners {
		if r.Name == "Rui Corre" {
			runnerID = r.ID
			if r.Active != 2 {
				t.Fatalf("active runs: %+v", r)
			}
		}
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+tents+"/assign", `{"runnerId":"nope"}`, julie, jcsrf); w.Code != 400 {
		t.Fatalf("unknown runner: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+tents+"/assign", `{"runnerId":"`+runnerID+`"}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("assign: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(rc.call("GET", "/api/runner/needs", "").Body.String(), "Carpas") {
		t.Fatalf("assigned run is in the runner's Yo")
	}
	if len(s.supplierView(julie).Runs) != 2 {
		t.Fatalf("both linked runs in Por despachar")
	}
	// The runner shares live location: Supply sees where they are on its runs.
	if r := findRun(s.supplierView(julie).Runs, water); r == nil || r.RunnerLive != nil {
		t.Fatalf("no live position yet: %+v", r)
	}
	s.app.mu.Lock()
	s.app.locations[runnerID] = liveLocation{Lat: 7.45, Lon: -80.3, Accuracy: 12, At: time.Now()}
	s.app.mu.Unlock()
	if r := findRun(s.supplierView(julie).Runs, water); r == nil || r.RunnerLive == nil || r.RunnerLive.Lat != 7.45 {
		t.Fatalf("live position: %+v", r)
	}
	s.app.mu.Lock()
	s.app.locations[runnerID] = liveLocation{Lat: 7.45, Lon: -80.3, At: time.Now().Add(-10 * time.Minute)}
	s.app.mu.Unlock()
	if r := findRun(s.supplierView(julie).Runs, water); r.RunnerLive != nil {
		t.Fatalf("stale position is not shown")
	}
	// Pack and dispatch: still IN_PROGRESS; the runner sees it; public hides who.
	if w := s.agentCall("POST", "/api/supplier/needs/"+water+"/packing", `{"items":[{"item":"Agua","quantity":4,"unit":"galones"}]}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("pack: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+water+"/dispatch", `{}`, julie, "bad"); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+water+"/dispatch", `{}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("dispatch: %d %s", w.Code, w.Body.String())
	}
	var n CampNeed
	for _, x := range s.needsOf(s.camp) {
		if x.ID == water {
			n = x
		}
	}
	if n.Status != "IN_PROGRESS" || n.DispatchedAt == "" || n.SupplierID == "" {
		t.Fatalf("dispatched stays en camino: %+v", n)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if !strings.Contains(pub, `"dispatchedAt"`) || strings.Contains(pub, "Julie") || strings.Contains(pub, "supplierId") {
		t.Fatalf("public: %s", pub)
	}
	// Julie cancels nothing on taken runs; a request can be cancelled while open.
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Linternas","status":"OPEN"}`, 201)
	var lamps string
	for _, x := range s.needsOf(s.camp) {
		if x.Need == "Linternas" {
			lamps = x.ID
		}
	}
	s.agentCall("POST", "/api/supplier/needs/"+lamps+"/request", `{}`, julie, jcsrf)
	if w := s.agentCall("POST", "/api/supplier/needs/"+lamps+"/request", `{"cancel":true}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("cancel: %d", w.Code)
	}
	if strings.Contains(rc.call("GET", "/api/runner/needs", "").Body.String(), `"requestedAt"`) {
		t.Fatalf("cancelled request is gone for runners")
	}
	// The runner drops the requested run: back to the request list (still Julie's), dispatch cleared.
	if w := rc.call("POST", "/api/runner/needs/"+water+"/release", ``); w.Code != 200 {
		t.Fatalf("release: %d", w.Code)
	}
	if r := findRun(s.supplierView(julie).Open, water); r == nil || !r.RequestMine {
		t.Fatalf("released request goes back to «Pedir corredor»: %+v", r)
	}
	// The in-person run dropped: free again, not Julie's.
	if w := rc.call("POST", "/api/runner/needs/"+tents+"/release", ``); w.Code != 200 {
		t.Fatalf("release tents: %d", w.Code)
	}
	if r := findRun(s.supplierView(julie).Open, tents); r == nil || r.RequestedAt != "" {
		t.Fatalf("in-person run released is a plain open need: %+v", r)
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"JULIE77"}`, false); !strings.Contains(w.Body.String(), "/supplier.html") {
		t.Fatalf("panels: %s", w.Body.String())
	}
}

// The other way round: a runner with a run asks supply centers; each says whether it can supply it; the runner sees
// the ones that can (to pick the closest) and chooses one, which then dispatches it.
func TestRunnerAsksSupplyCentersAndPicksOne(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Julie Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Pedro Acopio","telephone":"6999-1212","roles":["supplier"],"accessCode":"PEDRO77"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","status":"OPEN"}`, 201)
	need := s.needsOf(s.camp)[0].ID
	julie, jcsrf := s.supplierLogin("JULIE77")
	pedro, pcsrf := s.supplierLogin("PEDRO77")
	s.agentCall("POST", "/api/supplier/center", `{"center":"Villa Marina"}`, julie, jcsrf)
	s.agentCall("POST", "/api/supplier/center", `{"center":"Las Tablas"}`, pedro, pcsrf)
	rc := s.runnerLogin("Rui Corre")
	if w := rc.call("POST", "/api/runner/needs/"+need+"/supply-ask", `{}`); w.Code != 409 {
		t.Fatalf("only on your own run: %d", w.Code)
	}
	rc.call("POST", "/api/runner/needs/"+need+"/take", ``)
	asks := func(c *http.Cookie) []supplierRun {
		var v struct{ Asks []supplierRun }
		json.Unmarshal(s.agentCall("GET", "/api/supplier/runs", "", c, "").Body.Bytes(), &v)
		return v.Asks
	}
	if len(asks(julie)) != 0 {
		t.Fatalf("nothing asked yet")
	}
	if w := rc.call("POST", "/api/runner/needs/"+need+"/supply-ask", `{}`); w.Code != 200 {
		t.Fatalf("ask: %d %s", w.Code, w.Body.String())
	}
	if a := asks(julie); len(a) != 1 || a[0].RunnerName != "Rui Corre" || a[0].MyOffer != "" {
		t.Fatalf("Julie sees the ask: %+v", a)
	}
	// Julie can, Pedro can't (and stops seeing it).
	if w := s.agentCall("POST", "/api/supplier/needs/"+need+"/offer", `{"can":true}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("offer: %d %s", w.Code, w.Body.String())
	}
	s.agentCall("POST", "/api/supplier/needs/"+need+"/offer", `{"can":false}`, pedro, pcsrf)
	if len(asks(pedro)) != 0 {
		t.Fatalf("declined ask hidden for Pedro")
	}
	if a := asks(julie); a[0].MyOffer != "can" || a[0].OfferedBy != 1 {
		t.Fatalf("Julie's offer: %+v", a[0])
	}
	body := rc.call("GET", "/api/runner/needs", "").Body.String()
	if !strings.Contains(body, `"supplyOffers":[{"supplierId"`) || !strings.Contains(body, "Villa Marina") || strings.Contains(body, "Las Tablas") {
		t.Fatalf("runner sees only centers that can: %s", body)
	}
	// Pedro changes his mind and can after all; the runner picks Julie.
	s.agentCall("POST", "/api/supplier/needs/"+need+"/offer", `{"can":true}`, pedro, pcsrf)
	if w := rc.call("POST", "/api/runner/needs/"+need+"/supply-choose", `{"supplierId":"nope"}`); w.Code != 409 {
		t.Fatalf("only an offering center: %d", w.Code)
	}
	var julieID string
	var users []map[string]any
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	for _, u := range users {
		if u["name"] == "Julie Acopio" {
			julieID = u["id"].(string)
		}
	}
	if w := rc.call("POST", "/api/runner/needs/"+need+"/supply-choose", `{"supplierId":"`+julieID+`"}`); w.Code != 200 {
		t.Fatalf("choose: %d %s", w.Code, w.Body.String())
	}
	if len(asks(julie)) != 0 || len(asks(pedro)) != 0 {
		t.Fatalf("chosen: no longer an ask")
	}
	if v := s.supplierView(julie); findRun(v.Runs, need) == nil {
		t.Fatalf("in Julie's Por despachar: %+v", v.Runs)
	}
	if v := s.supplierView(pedro); findRun(v.Runs, need) != nil {
		t.Fatalf("not Pedro's")
	}
	if w := rc.call("POST", "/api/runner/needs/"+need+"/supply-ask", `{}`); w.Code != 409 {
		t.Fatalf("already has a center: %d", w.Code)
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+need+"/packing", `{"items":[{"item":"Agua","quantity":2}]}`, julie, jcsrf); w.Code != 200 {
		t.Fatalf("Julie packs it: %d", w.Code)
	}
	// Withdraw and cancel on another run.
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Carpas","status":"OPEN"}`, 201)
	var tents string
	for _, n := range s.needsOf(s.camp) {
		if n.Need == "Carpas" {
			tents = n.ID
		}
	}
	rc.call("POST", "/api/runner/needs/"+tents+"/take", ``)
	rc.call("POST", "/api/runner/needs/"+tents+"/supply-ask", `{}`)
	s.agentCall("POST", "/api/supplier/needs/"+tents+"/offer", `{"can":true}`, julie, jcsrf)
	s.agentCall("POST", "/api/supplier/needs/"+tents+"/offer", `{"withdraw":true}`, julie, jcsrf)
	if a := asks(julie); len(a) != 1 || a[0].MyOffer != "" {
		t.Fatalf("withdrawn: %+v", a)
	}
	rc.call("POST", "/api/runner/needs/"+tents+"/supply-ask", `{"cancel":true}`)
	if len(asks(julie)) != 0 {
		t.Fatalf("cancelled ask is gone")
	}
	if w := s.agentCall("POST", "/api/supplier/needs/"+tents+"/offer", `{"can":true}`, julie, jcsrf); w.Code != 409 {
		t.Fatalf("no ask to answer: %d", w.Code)
	}
	// Releasing the run clears the ask.
	rc.call("POST", "/api/runner/needs/"+tents+"/supply-ask", `{}`)
	rc.call("POST", "/api/runner/needs/"+tents+"/release", ``)
	for _, n := range s.needsOf(s.camp) {
		if n.ID == tents && (n.SupplyAskedAt != "" || len(n.SupplyOffers) != 0) {
			t.Fatalf("released run keeps no ask: %+v", n)
		}
	}
}
