package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// crewCall hace una petición con la sesión de un corredor o constructor.
func (s *serviceHarness) crewCall(method, path, body string, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	s.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if method == "POST" {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
		r.Header.Set("X-CSRF-Token", csrf)
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}
func (s *serviceHarness) crewLogin(path, name string) (*http.Cookie, string) {
	s.t.Helper()
	w := s.crewCall("POST", path+"/session", `{"name":"`+name+`","telephone":"6777-8888"}`, nil, "")
	if w.Code != 200 {
		s.t.Fatalf("login %s: %d %s", path, w.Code, w.Body.String())
	}
	var out map[string]string
	json.Unmarshal(w.Body.Bytes(), &out)
	return w.Result().Cookies()[0], out["csrf"]
}
func (s *serviceHarness) needsOf(camp string) []CampNeed {
	var needs []CampNeed
	json.Unmarshal(s.call("GET", "/api/admin/camps/"+camp+"/needs", "", true).Body.Bytes(), &needs)
	return needs
}

func TestRepairNeedIsSeparateAndPrivate(t *testing.T) {
	s := newServiceHarness(t)
	sub := s.call("POST", "/api/submissions", `{"needs":"Agua, Reparación de casa","contact":"P","phone":"6111-0005"}`, false)
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	id := created["id"]
	s.call("POST", "/api/admin/decisions", `{"id":"`+id+`","decision":"approved"}`, true)

	needs := s.needsOf(id)
	if len(needs) != 2 {
		t.Fatalf("approval should create a supply and a repair need: %+v", needs)
	}
	var supply, repair CampNeed
	for _, n := range needs {
		if n.Kind == "repair" {
			repair = n
		} else {
			supply = n
		}
	}
	if supply.Need != "Agua" || repair.Need != repairLabel {
		t.Fatalf("split needs: supply=%q repair=%q", supply.Need, repair.Need)
	}

	public := s.call("GET", "/api/points", "", false).Body.String()
	if strings.Contains(public, "Reparaci") {
		t.Fatalf("repair leaked in public points: %s", public)
	}
	if pub := s.call("GET", "/api/camps/"+id+"/needs", "", false).Body.String(); strings.Contains(pub, repair.ID) || strings.Contains(pub, "Reparaci") {
		t.Fatalf("repair leaked in public needs: %s", pub)
	}

	// La reparación abierta no cambia el estado público: con el agua completada, el punto queda completado.
	s.call("PATCH", "/api/admin/needs/"+supply.ID, `{"status":"COMPLETED"}`, true)
	var points []Point
	json.Unmarshal(s.call("GET", "/api/points", "", false).Body.Bytes(), &points)
	for _, p := range points {
		if p.ID == id && p.CampStatus != "COMPLETED" {
			t.Fatalf("open repair must not affect public status: %q", p.CampStatus)
		}
	}

	// Corredores no ven ni toman reparaciones.
	rc, rcsrf := s.crewLogin("/api/runner", "Corredor Uno")
	if list := s.crewCall("GET", "/api/runner/needs", "", rc, rcsrf).Body.String(); strings.Contains(list, repair.ID) {
		t.Fatalf("runner sees repair: %s", list)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+repair.ID+"/take", `{}`, rc, rcsrf); w.Code != 409 {
		t.Fatalf("runner took repair: %d", w.Code)
	}
}

func TestBuilderTakesUpToTwentyRepairs(t *testing.T) {
	s := newServiceHarness(t)
	supply := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"urgente"}`, true)
	var supplyNeed CampNeed
	json.Unmarshal(supply.Body.Bytes(), &supplyNeed)
	ids := []string{}
	for i := 0; i < 21; i++ {
		w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", fmt.Sprintf(`{"need":"Techo %d","urgency":"normal","kind":"repair"}`, i), true)
		if w.Code != 201 {
			t.Fatalf("create repair: %d %s", w.Code, w.Body.String())
		}
		var n CampNeed
		json.Unmarshal(w.Body.Bytes(), &n)
		ids = append(ids, n.ID)
	}
	bc, bcsrf := s.crewLogin("/api/builder", "Constructora Ana Pérez")
	list := s.crewCall("GET", "/api/builder/needs", "", bc, bcsrf).Body.String()
	if strings.Contains(list, supplyNeed.ID) || !strings.Contains(list, ids[0]) || !strings.Contains(list, `"max":20`) {
		t.Fatalf("builder list: %s", list)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+supplyNeed.ID+"/take", `{}`, bc, bcsrf); w.Code != 409 {
		t.Fatalf("builder took supply need: %d", w.Code)
	}
	for i := 0; i < 20; i++ {
		if w := s.crewCall("POST", "/api/builder/needs/"+ids[i]+"/take", `{}`, bc, bcsrf); w.Code != 200 {
			t.Fatalf("take %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+ids[20]+"/take", `{}`, bc, bcsrf); w.Code != 409 || !strings.Contains(w.Body.String(), "20 reparaciones") {
		t.Fatalf("21st repair: %d %s", w.Code, w.Body.String())
	}
	var mine struct {
		Mine []runnerNeed `json:"mine"`
	}
	json.Unmarshal(s.crewCall("GET", "/api/builder/needs", "", bc, bcsrf).Body.Bytes(), &mine)
	if len(mine.Mine) != 20 {
		t.Fatalf("mine: %d", len(mine.Mine))
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+ids[0]+"/complete", `{}`, bc, bcsrf); w.Code != 200 {
		t.Fatalf("complete: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+ids[1]+"/release", `{}`, bc, bcsrf); w.Code != 200 {
		t.Fatalf("release: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+ids[20]+"/take", `{}`, bc, bcsrf); w.Code != 200 {
		t.Fatalf("take after freeing slots: %d", w.Code)
	}
	// La sesión de constructor no sirve en rutas de corredor ni al revés.
	if w := s.crewCall("GET", "/api/runner/needs", "", bc, bcsrf); w.Code != 401 {
		t.Fatalf("builder cookie on runner route: %d", w.Code)
	}

	var builders []Runner
	json.Unmarshal(s.call("GET", "/api/admin/builders", "", true).Body.Bytes(), &builders)
	if len(builders) != 1 || builders[0].Name != "Constructora Ana Pérez" {
		t.Fatalf("admin builders: %+v", builders)
	}
	var runners []Runner
	json.Unmarshal(s.call("GET", "/api/admin/runners", "", true).Body.Bytes(), &runners)
	if len(runners) != 0 {
		t.Fatalf("builder registered as runner: %+v", runners)
	}
}

func TestAdminAssignmentMatchesNeedKind(t *testing.T) {
	s := newServiceHarness(t)
	runner := s.call("POST", "/api/admin/runners", `{"name":"R","telephone":"6000-0001"}`, true)
	builder := s.call("POST", "/api/admin/builders", `{"name":"B","telephone":"6000-0002"}`, true)
	var r, b Runner
	json.Unmarshal(runner.Body.Bytes(), &r)
	json.Unmarshal(builder.Body.Bytes(), &b)
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Techo","kind":"repair","runnerId":"`+r.ID+`"}`, true); w.Code != 400 {
		t.Fatalf("runner on repair: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","builderId":"`+b.ID+`"}`, true); w.Code != 400 {
		t.Fatalf("builder on supply: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Techo","kind":"house"}`, true); w.Code != 400 {
		t.Fatalf("unknown kind: %d", w.Code)
	}
	w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Techo","kind":"repair","builderId":"`+b.ID+`","status":"IN_PROGRESS"}`, true)
	if w.Code != 201 {
		t.Fatalf("repair with builder: %d %s", w.Code, w.Body.String())
	}
	var n CampNeed
	json.Unmarshal(w.Body.Bytes(), &n)
	// Cambiar a suministros suelta al constructor.
	w = s.call("PATCH", "/api/admin/needs/"+n.ID, `{"kind":"supply"}`, true)
	n = CampNeed{}
	json.Unmarshal(w.Body.Bytes(), &n)
	if w.Code != 200 || n.Kind != "" || n.BuilderID != "" {
		t.Fatalf("kind change: %d %+v", w.Code, n)
	}
}

func TestMergeKeepsRepairSeparate(t *testing.T) {
	s := newServiceHarness(t)
	s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"urgente"}`, true)
	sub := s.call("POST", "/api/submissions", `{"needs":"Alimentos","contact":"P","phone":"6111-0006"}`, false)
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	w := s.call("POST", "/api/admin/points/"+created["id"]+"/merge", `{"targetId":"`+s.camp+`","needs":["Alimentos","Reparación de casa"]}`, true)
	if w.Code != 200 {
		t.Fatalf("merge: %d %s", w.Code, w.Body.String())
	}
	needs := s.needsOf(s.camp)
	got := map[string]string{}
	for _, n := range needs {
		got[n.Kind] = n.Need
	}
	if len(needs) != 2 || got[""] != "Agua, Alimentos" || got["repair"] != repairLabel {
		t.Fatalf("merge needs: %+v", needs)
	}
}

func TestLegacySupplyNeedWithRepairIsSplitOnOpen(t *testing.T) {
	s := newServiceHarness(t)
	put := func(n CampNeed) {
		if e := s.app.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(needsBucket), n.ID, n) }); e != nil {
			t.Fatal(e)
		}
	}
	put(CampNeed{ID: "legacy-mixed", CampID: s.camp, Need: "Agua, Alimentos, Reparación de casa", Status: "OPEN", Urgency: "Urgent", CreatedAt: "2026-10-01T00:00:00Z"})
	if e := s.app.db.Update(splitLegacyRepairs); e != nil {
		t.Fatal(e)
	}
	if e := s.app.db.Update(splitLegacyRepairs); e != nil { // idempotente
		t.Fatal(e)
	}
	needs := s.needsOf(s.camp)
	if len(needs) != 2 {
		t.Fatalf("expected supply + repair: %+v", needs)
	}
	for _, n := range needs {
		if n.Kind == "repair" && (n.Need != repairLabel || n.Status != "OPEN") || n.Kind == "" && n.Need != "Agua, Alimentos" {
			t.Fatalf("split: %+v", n)
		}
	}
	rc, rcsrf := s.crewLogin("/api/runner", "Corredor Dos")
	if list := s.crewCall("GET", "/api/runner/needs", "", rc, rcsrf).Body.String(); strings.Contains(list, "Reparaci") {
		t.Fatalf("runner sees repair text: %s", list)
	}
	// Una necesidad de suministros escrita a mano con la reparación tampoco la muestra a corredores ni al público.
	put(CampNeed{ID: "typed", CampID: s.camp, Need: "Reparación de casa", Status: "OPEN", Urgency: "Today"})
	if list := s.crewCall("GET", "/api/runner/needs", "", rc, rcsrf).Body.String(); strings.Contains(list, "Reparaci") || strings.Contains(list, `"typed"`) {
		t.Fatalf("runner sees typed repair: %s", list)
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(pub, "Reparaci") {
		t.Fatalf("public sees typed repair: %s", pub)
	}
}
