package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

type runnerClient struct {
	s      *serviceHarness
	cookie string
	csrf   string
}

func (s *serviceHarness) runnerLogin(name string) *runnerClient {
	s.t.Helper()
	w := s.call("POST", "/api/runner/session", `{"name":"`+name+`","telephone":"6000-0000"}`, false)
	if w.Code != 200 {
		s.t.Fatalf("runner login: %d %s", w.Code, w.Body.String())
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	return &runnerClient{s: s, cookie: w.Result().Cookies()[0].Value, csrf: body["csrf"]}
}
func (c *runnerClient) call(method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", "runner_session="+c.cookie)
	r.Header.Set("X-CSRF-Token", c.csrf)
	w := httptest.NewRecorder()
	c.s.h.ServeHTTP(w, r)
	return w
}
func (c *runnerClient) needs() (current *runnerNeed, available []runnerNeed, raw string) {
	w := c.call("GET", "/api/runner/needs", "")
	var body struct {
		Current   *runnerNeed  `json:"current"`
		Available []runnerNeed `json:"available"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	return body.Current, body.Available, w.Body.String()
}

func TestRunnerSessionReusesRunnerByName(t *testing.T) {
	s := newServiceHarness(t)
	existing := s.expect("POST", "/api/admin/runners", `{"name":"Ana Pérez","telephone":"6999-1111"}`, 201)
	c := s.runnerLogin("  ana   pérez ")
	if w := c.call("GET", "/api/runner/session", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Ana Pérez"`) || strings.Contains(w.Body.String(), "6999") {
		t.Fatalf("session: %d %s", w.Code, w.Body.String())
	}
	s.runnerLogin("Luis")
	list := s.call("GET", "/api/admin/runners", "", true).Body.String()
	if strings.Count(list, `"id"`) != 2 || !strings.Contains(list, existing["id"].(string)) || !strings.Contains(list, `"name":"Luis","telephone":"6000-0000"`) || !strings.Contains(list, `"telephone":"6999-1111"`) {
		t.Fatalf("runners (new runner keeps phone, curated phone not overwritten): %s", list)
	}
	if w := s.call("POST", "/api/runner/session", `{"name":"  ","telephone":"6000-0000"}`, false); w.Code != 400 {
		t.Fatalf("blank name: %d", w.Code)
	}
	for _, body := range []string{`{"name":"Marta"}`, `{"name":"Marta","telephone":"  "}`, `{"name":"Marta","telephone":"12-34"}`} {
		if w := s.call("POST", "/api/runner/session", body, false); w.Code != 400 || !strings.Contains(w.Body.String(), "teléfono") {
			t.Fatalf("missing phone %s: %d %s", body, w.Code, w.Body.String())
		}
	}
	if w := s.call("GET", "/api/runner/needs", "", false); w.Code != 401 {
		t.Fatalf("no session: %d", w.Code)
	}
	if w := s.call("GET", "/api/admin/needs", "", false); w.Code != 401 {
		t.Fatalf("runner cookie must not open admin: %d", w.Code)
	}
}

func TestRunnerTakeCompleteReleaseAndFollowup(t *testing.T) {
	s := newServiceHarness(t)
	n1 := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Tomorrow"}`, 201)["id"].(string)
	n2 := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Food","urgency":"Today"}`, 201)["id"].(string)
	ana, luis := s.runnerLogin("Ana"), s.runnerLogin("Luis")

	_, available, raw := ana.needs()
	if len(available) != 2 || available[0].ID != n2 || available[0].Camp.Code != "C-1" || !strings.Contains(raw, `"contacts":[{"name":"Private camp","phone":"6555-1234"}]`) || strings.Contains(raw, "runnerId") {
		t.Fatalf("available: %s", raw)
	}
	if w := ana.call("POST", "/api/runner/needs/"+n1+"/take", "{}"); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	if w := ana.call("POST", "/api/runner/needs/"+n2+"/take", "{}"); w.Code != 409 {
		t.Fatalf("second run must be refused: %d", w.Code)
	}
	if w := luis.call("POST", "/api/runner/needs/"+n1+"/take", "{}"); w.Code != 409 {
		t.Fatalf("taken need: %d", w.Code)
	}
	if w := luis.call("POST", "/api/runner/needs/"+n1+"/complete", "{}"); w.Code != 409 {
		t.Fatalf("other runner completes: %d", w.Code)
	}
	current, available, _ := ana.needs()
	if current == nil || current.ID != n1 || len(available) != 1 {
		t.Fatalf("current: %+v %+v", current, available)
	}
	if _, available, _ := luis.needs(); len(available) != 1 {
		t.Fatalf("luis sees taken need: %+v", available)
	}
	if w := ana.call("POST", "/api/runner/needs/"+n1+"/release", "{}"); w.Code != 200 {
		t.Fatalf("release: %d", w.Code)
	}
	if _, available, _ := luis.needs(); len(available) != 2 {
		t.Fatalf("released need back: %+v", available)
	}
	if w := luis.call("POST", "/api/runner/needs/"+n2+"/take", "{}"); w.Code != 200 {
		t.Fatalf("luis take: %d", w.Code)
	}
	if w := luis.call("POST", "/api/runner/needs/"+n2+"/complete", "{}"); w.Code != 200 {
		t.Fatalf("complete: %d", w.Code)
	}
	if w := luis.call("POST", "/api/runner/camps/"+s.camp+"/needs", `{"need":"More water","urgency":"Urgent"}`); w.Code != 201 {
		t.Fatalf("followup: %d %s", w.Code, w.Body.String())
	}
	all := s.call("GET", "/api/admin/needs", "", true).Body.String()
	if !strings.Contains(all, `"status":"COMPLETED"`) || !strings.Contains(all, `"need":"More water","status":"OPEN","urgency":"Urgent"`) {
		t.Fatalf("admin view: %s", all)
	}
	r := httptest.NewRequest("POST", "/api/runner/needs/"+n1+"/take", strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", "runner_session="+ana.cookie)
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("missing CSRF: %d", w.Code)
	}
}

func TestNeedTimestamps(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Today"}`, 201)["id"].(string)
	ana := s.runnerLogin("Ana")
	ana.call("POST", "/api/runner/needs/"+id+"/take", "{}")
	ana.call("POST", "/api/runner/needs/"+id+"/release", "{}")
	var list []CampNeed
	list = nil
	json.Unmarshal(s.call("GET", "/api/admin/needs", "", true).Body.Bytes(), &list)
	if len(list) != 1 || list[0].CreatedAt == "" || list[0].ReleasedAt == "" || list[0].ReleasedBy == "" || list[0].CompletedAt != "" || list[0].EnRouteAt != "" {
		t.Fatalf("after release: %+v", list)
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(pub, list[0].ReleasedBy) || strings.Contains(pub, "releasedBy") {
		t.Fatalf("public leaks runner: %s", pub)
	}
	ana.call("POST", "/api/runner/needs/"+id+"/take", "{}")
	ana.call("POST", "/api/runner/needs/"+id+"/complete", "{}")
	list = nil
	json.Unmarshal(s.call("GET", "/api/admin/needs", "", true).Body.Bytes(), &list)
	if list[0].Status != "COMPLETED" || list[0].CompletedAt == "" || list[0].EnRouteAt == "" || list[0].EnRouteAt > list[0].CompletedAt {
		t.Fatalf("after complete: %+v", list[0])
	}
	s.expect("PATCH", "/api/admin/needs/"+id, `{"status":"OPEN"}`, 200)
	list = nil
	json.Unmarshal(s.call("GET", "/api/admin/needs", "", true).Body.Bytes(), &list)
	if list[0].CompletedAt != "" || list[0].EnRouteAt != "" {
		t.Fatalf("reopened keeps completedAt: %+v", list[0])
	}
	s.expect("PATCH", "/api/admin/needs/"+id, `{"status":"IN_PROGRESS"}`, 200)
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); !strings.Contains(pub, `"enRouteAt":"20`) {
		t.Fatalf("public needs should show when it went en ruta: %s", pub)
	}
	if w := s.call("PATCH", "/api/admin/needs/"+id, `{"createdAt":"2000-01-01T00:00:00Z"}`, true); w.Code != 400 {
		t.Fatalf("timestamps must be server-set: %d", w.Code)
	}
}

func TestRunnerSessionFillsMissingPhone(t *testing.T) {
	s := newServiceHarness(t)
	// Corredores registrados antes de pedir teléfono quedaron sin número.
	if err := s.app.db.Update(func(tx *bolt.Tx) error {
		return putJSON(tx.Bucket(runnersBucket), "legacy", Runner{ID: "legacy", Name: "Rosa"})
	}); err != nil {
		t.Fatal(err)
	}
	if w := s.call("POST", "/api/runner/session", `{"name":"rosa","telephone":" 6123  4567 "}`, false); w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	if list := s.call("GET", "/api/admin/runners", "", true).Body.String(); strings.Count(list, `"id"`) != 1 || !strings.Contains(list, `"telephone":"6123 4567"`) {
		t.Fatalf("runners: %s", list)
	}
}
