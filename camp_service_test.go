package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type serviceHarness struct {
	t      *testing.T
	app    *App
	h      http.Handler
	cookie *http.Cookie
	csrf   string
	camp   string
}

func newServiceHarness(t *testing.T) *serviceHarness {
	t.Helper()
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.db.Close() })
	s := &serviceHarness{t: t, app: app, h: app.routes()}
	login := s.call("POST", "/api/admin/login", `{"password":"test-password-only-12345"}`, false)
	if login.Code != 200 {
		t.Fatalf("login: %s", login.Body.String())
	}
	s.cookie = login.Result().Cookies()[0]
	var auth map[string]string
	json.Unmarshal(login.Body.Bytes(), &auth)
	s.csrf = auth["csrf"]
	created := s.call("POST", "/api/admin/points", `{"code":"C-1","area":"Venao","place":"Camp","contact":"Private camp","phone":"6555-1234"}`, true)
	if created.Code != 201 {
		t.Fatalf("camp create: %s", created.Body.String())
	}
	var p map[string]string
	json.Unmarshal(created.Body.Bytes(), &p)
	s.camp = p["id"]
	return s
}
func (s *serviceHarness) call(method, path, body string, auth bool) *httptest.ResponseRecorder {
	s.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if method == "POST" || method == "PATCH" {
		r.Header.Set("Content-Type", "application/json")
	}
	if auth && s.cookie != nil {
		r.AddCookie(s.cookie)
		r.Header.Set("X-CSRF-Token", s.csrf)
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}
func (s *serviceHarness) expect(method, path, body string, status int) map[string]any {
	s.t.Helper()
	w := s.call(method, path, body, true)
	if w.Code != status {
		s.t.Fatalf("%s %s: got %d %s want %d", method, path, w.Code, w.Body.String(), status)
	}
	var v map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		s.t.Fatal(err)
	}
	return v
}
func TestCampServiceRunnerCreateAndList(t *testing.T) {
	s := newServiceHarness(t)
	r := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	if r["name"] != "Ana" || r["telephone"] != "6999-1111" || r["id"] == "" {
		t.Fatalf("runner: %v", r)
	}
	got := s.call("GET", "/api/admin/runners", "", true)
	if got.Code != 200 || !bytes.Contains(got.Body.Bytes(), []byte("6999-1111")) {
		t.Fatalf("list: %d %s", got.Code, got.Body.String())
	}
}
func TestCampServiceNeedsAndPublicPrivacy(t *testing.T) {
	s := newServiceHarness(t)
	runner := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	id := runner["id"].(string)
	n := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Bring potable water","status":"OPEN","urgency":"urgente","runnerId":"`+id+`"}`, 201)
	if n["need"] != "Bring potable water" || n["runnerId"] != id {
		t.Fatalf("need: %v", n)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false)
	if pub.Code != 200 || !bytes.Contains(pub.Body.Bytes(), []byte("Bring potable water")) || bytes.Contains(pub.Body.Bytes(), []byte("6999-1111")) || bytes.Contains(pub.Body.Bytes(), []byte("runnerId")) {
		t.Fatalf("public: %d %s", pub.Code, pub.Body.String())
	}
	points := s.call("GET", "/api/points", "", false)
	// Assigning a runner puts the need en ruta.
	if n["status"] != "IN_PROGRESS" || n["enRouteAt"] == nil {
		t.Fatalf("assigned need must be en ruta: %v", n)
	}
	if !bytes.Contains(points.Body.Bytes(), []byte(`"campStatus":"IN_PROGRESS"`)) || !bytes.Contains(points.Body.Bytes(), []byte(`"highestUrgency":"urgente"`)) || bytes.Contains(points.Body.Bytes(), []byte("6999-1111")) {
		t.Fatalf("aggregate: %s", points.Body.String())
	}
}
func TestCampServiceWriteProtectionAndValidation(t *testing.T) {
	s := newServiceHarness(t)
	for _, path := range []string{"/api/admin/runners", "/api/admin/camps/" + s.camp + "/needs"} {
		if w := s.call("POST", path, `{}`, false); w.Code != 401 {
			t.Fatalf("no login %s: %d", path, w.Code)
		}
		r := httptest.NewRequest("POST", path, strings.NewReader(`{}`))
		r.AddCookie(s.cookie)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("no csrf %s: %d", path, w.Code)
		}
	}
	if w := s.call("POST", "/api/admin/runners", `{"name":"","telephone":"123"}`, true); w.Code != 400 {
		t.Fatalf("invalid runner: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"","urgency":"Now"}`, true); w.Code != 400 {
		t.Fatalf("invalid need: %d %s", w.Code, w.Body.String())
	}
}
func TestCampServicePatchAggregationAndCampEdit(t *testing.T) {
	s := newServiceHarness(t)
	first := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"normal"}`, 201)
	second := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Medicine","urgency":"urgente"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+first["id"].(string), `{"status":"IN_PROGRESS"}`, 200)
	s.expect("PATCH", "/api/admin/needs/"+second["id"].(string), `{"status":"COMPLETED"}`, 200)
	points := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"IN_PROGRESS"`) || !strings.Contains(points, `"highestUrgency":"normal"`) {
		t.Fatalf("aggregate: %s", points)
	}
	s.expect("PATCH", "/api/admin/needs/"+first["id"].(string), `{"status":"DELETED"}`, 200)
	points = s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"COMPLETED"`) || strings.Contains(points, "highestUrgency") {
		t.Fatalf("completed: %s", points)
	}
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if strings.Contains(pub, "Water") {
		t.Fatalf("deleted exposed: %s", pub)
	}
	edited := s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"note":"Private note","coordinator":"Lead","phone":"6000-2222","place":"New place"}`, 200)
	if edited["note"] != "Private note" || edited["area"] != "Venao" || edited["place"] != "New place" {
		t.Fatalf("partial edit: %v", edited)
	}
	points = s.call("GET", "/api/points", "", false).Body.String()
	if strings.Contains(points, "Private note") || strings.Contains(points, "6000-2222") {
		t.Fatalf("private exposed: %s", points)
	}
}
func TestCampServicePatchAndNoNeeds(t *testing.T) {
	s := newServiceHarness(t)
	points := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"NO_NEEDS"`) {
		t.Fatalf("empty: %s", points)
	}
	r := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)
	runnerID := r["id"].(string)
	req := httptest.NewRequest("PATCH", "/api/admin/runners/"+runnerID, strings.NewReader(`{"telephone":"6999-2222"}`))
	req.AddCookie(s.cookie)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("PATCH without csrf: %d %s", w.Code, w.Body.String())
	}
	edited := s.expect("PATCH", "/api/admin/runners/"+runnerID, `{"telephone":"6999-2222"}`, 200)
	if edited["name"] != "Ana" || edited["telephone"] != "6999-2222" {
		t.Fatalf("runner edit: %v", edited)
	}
	n := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"urgente"}`, 201)
	if n["updatedAt"] == nil || n["updatedAt"] == "" {
		t.Fatalf("created need lacks updatedAt: %v", n)
	}
	if patched := s.expect("PATCH", "/api/admin/needs/"+n["id"].(string), `{"status":"DELETED"}`, 200); patched["updatedAt"] == nil || patched["updatedAt"] == "" {
		t.Fatalf("patched need lacks updatedAt: %v", patched)
	}
	points = s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(points, `"campStatus":"NO_NEEDS"`) || strings.Contains(points, "highestUrgency") {
		t.Fatalf("deleted aggregate: %s", points)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","urgency":"Bad"}`, true); w.Code != 400 {
		t.Fatalf("invalid urgency: %d %s", w.Code, w.Body.String())
	}
}
func TestCampServiceCampEditFieldsAndReadOnly(t *testing.T) {
	s := newServiceHarness(t)
	edited := s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"secondContact":"Backup","secondPhone":"6111-1111","thirdContact":"Other","thirdPhone":"6222-2222","gps":"https://maps.google.com/?q=7.45,-80.35","coords":[7.45,-80.35],"note":"operations only"}`, 200)
	if edited["secondPhone"] != "6111-1111" || edited["thirdPhone"] != "6222-2222" || edited["note"] != "operations only" {
		t.Fatalf("edit: %v", edited)
	}
	if w := s.call("PATCH", "/api/admin/camps/"+s.camp, `{"status":"archived"}`, true); w.Code != 400 {
		t.Fatalf("status override: %d %s", w.Code, w.Body.String())
	}
	public := s.call("GET", "/api/points", "", false).Body.String()
	for _, secret := range []string{"operations only", "6111-1111", "6222-2222"} {
		if strings.Contains(public, secret) {
			t.Fatalf("leaked %s: %s", secret, public)
		}
	}
}
func TestCampServiceSeedRefreshPreservesCampNote(t *testing.T) {
	s := newServiceHarness(t)
	seeded := s.call("POST", "/api/admin/seed", `[{"code":"S-2","area":"Original"}]`, true)
	if seeded.Code != 200 {
		t.Fatalf("seed: %s", seeded.Body.String())
	}
	s.expect("PATCH", "/api/admin/camps/seed:s-2", `{"note":"Keep this note"}`, 200)
	refreshed := s.call("POST", "/api/admin/seed", `[{"code":"S-2","area":"Updated"}]`, true)
	if refreshed.Code != 200 {
		t.Fatalf("refresh: %s", refreshed.Body.String())
	}
	all := s.call("GET", "/api/admin/points", "", true).Body.String()
	if !strings.Contains(all, "Keep this note") {
		t.Fatalf("note lost: %s", all)
	}
}
func TestCampServicePersists(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Fuel","urgency":"normal"}`, 201)
	path := s.app.db.Path()
	s.app.db.Close()
	reopened, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.db.Close()
	w := httptest.NewRecorder()
	reopened.routes().ServeHTTP(w, httptest.NewRequest("GET", "/api/camps/"+s.camp+"/needs", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Fuel") {
		t.Fatalf("persisted: %d %s", w.Code, w.Body.String())
	}
}
func TestAdminNeedsListAllCamps(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("GET", "/api/admin/needs", "", false); w.Code != 401 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","status":"OPEN","urgency":"normal"}`, 201)
	gone := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Old","status":"OPEN","urgency":"urgente"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+gone["id"].(string), `{"status":"DELETED"}`, 200)
	w := s.call("GET", "/api/admin/needs", "", true)
	var list []CampNeed
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list) != 1 || list[0].CampID != s.camp || list[0].Urgency != "normal" {
		t.Fatalf("needs: %d %s", w.Code, w.Body.String())
	}
}
func TestPublicPointsListNeedStatuses(t *testing.T) {
	s := newServiceHarness(t)
	if points := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(points, "needStatuses") {
		t.Fatalf("no needs: %s", points)
	}
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Water","status":"OPEN","urgency":"urgente"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Food","status":"COMPLETED","urgency":"normal"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Old","status":"DELETED","urgency":"urgente"}`, 201)
	if points := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(points, `"needStatuses":["OPEN","COMPLETED"]`) {
		t.Fatalf("statuses: %s", points)
	}
	if w := s.call("PATCH", "/api/admin/camps/"+s.camp, `{"needStatuses":["OPEN"]}`, true); w.Code != 400 {
		t.Fatalf("read-only: %d %s", w.Code, w.Body.String())
	}
}

func TestPublicNeedReportRequiresRequesterAndOpensImmediately(t *testing.T) {
	s := newServiceHarness(t)
	path := "/api/camps/" + s.camp + "/needs"
	for body, field := range map[string]string{
		`{"requesterName":"Eva","requesterPhone":"6555-1234"}`:                              "need",
		`{"need":"Agua","requesterPhone":"6555-1234"}`:                                      "requesterName",
		`{"need":"Agua","requesterName":"Eva"}`:                                             "requesterPhone",
		`{"need":"Agua","requesterName":"Eva","requesterPhone":"12-34"}`:                    "requesterPhone",
		`{"need":"Agua","requesterName":"Eva","requesterPhone":"6555-1234","urgency":"Ya"}`: "urgency",
	} {
		if w := s.call("POST", path, body, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"`+field+`"`) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	if w := s.call("POST", "/api/camps/missing/needs", `{"need":"Agua","requesterName":"Eva","requesterPhone":"6555-1234"}`, false); w.Code != 404 {
		t.Fatalf("unknown camp: %d", w.Code)
	}
	if w := s.call("POST", path, `{"need":"Agua potable","urgency":"urgente","requesterName":" Eva  Ruiz ","requesterPhone":"6555-1234"}`, false); w.Code != 201 || !strings.Contains(w.Body.String(), `"status":"OPEN"`) {
		t.Fatalf("report: %d %s", w.Code, w.Body.String())
	}
	public := s.call("GET", path, "", false).Body.String()
	if !strings.Contains(public, `"need":"Agua potable"`) || !strings.Contains(public, `"requesterName":"Eva Ruiz"`) || !strings.Contains(public, `"requesterPhone":"6555-1234"`) {
		t.Fatalf("public list: %s", public)
	}
	runner := s.runnerLogin("Luis")
	if _, available, raw := runner.needs(); len(available) != 1 || !strings.Contains(raw, `"requesterPhone":"6555-1234"`) || available[0].CreatedAt == "" {
		t.Fatalf("runner list: %s", raw)
	}
}

func TestNewPointNeedsBecomeOpenCampNeedOnceLive(t *testing.T) {
	s := newServiceHarness(t)
	needsOf := func(id string) []CampNeed {
		var list []CampNeed
		json.Unmarshal(s.call("GET", "/api/admin/camps/"+id+"/needs", "", true).Body.Bytes(), &list)
		return list
	}
	sub := s.call("POST", "/api/submissions", `{"needs":"  Agua y pañales ","contact":"Privada","phone":"6555-1234"}`, false)
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	if all := s.call("GET", "/api/admin/needs", "", true).Body.String(); strings.Contains(all, "pañales") {
		t.Fatalf("pending point must not create a need yet: %s", all)
	}
	decide := func(decision string) {
		s.expect("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"`+decision+`"}`, 200)
	}
	decide("approved")
	list := needsOf(created["id"])
	if len(list) != 1 || list[0].Need != "Agua y pañales" || list[0].Status != "OPEN" || list[0].Urgency != "normal" || list[0].RequesterPhone != "" {
		t.Fatalf("approved point need: %+v", list)
	}
	decide("archived")
	decide("approved")
	if list := needsOf(created["id"]); len(list) != 1 {
		t.Fatalf("re-approval duplicated need: %+v", list)
	}
	rej := s.call("POST", "/api/submissions", `{"needs":"Linternas","contact":"Privada","phone":"6555-4321"}`, false)
	var rejected map[string]string
	json.Unmarshal(rej.Body.Bytes(), &rejected)
	s.expect("POST", "/api/admin/decisions", `{"id":"`+rejected["id"]+`","decision":"rejected"}`, 200)
	if pub := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pub, "Linternas") {
		t.Fatalf("rejected point public: %s", pub)
	}
	s.expect("POST", "/api/admin/decisions", `{"id":"`+rejected["id"]+`","decision":"archived"}`, 409)
	s.expect("POST", "/api/admin/decisions", `{"id":"`+rejected["id"]+`","decision":"approved"}`, 200)
	if list := needsOf(rejected["id"]); len(list) != 1 || list[0].Need != "Linternas" {
		t.Fatalf("approved rejected point need: %+v", list)
	}
	admin := s.expect("POST", "/api/admin/points", `{"code":"ADM-1","needs":"Colchones"}`, 201)
	if list := needsOf(admin["id"].(string)); len(list) != 1 || list[0].Need != "Colchones" {
		t.Fatalf("admin point need: %+v", list)
	}
	empty := s.expect("POST", "/api/admin/points", `{"code":"ADM-2"}`, 201)
	if list := needsOf(empty["id"].(string)); len(list) != 0 {
		t.Fatalf("point without needs created one: %+v", list)
	}
}

func TestArchivedPointHidesItAndItsNeedsUntilRecovered(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua archivada","urgency":"urgente"}`, 201)
	runner := s.runnerLogin("Luis")
	visible := func() (public, runnerRaw string, needs int) {
		_, _, runnerRaw = runner.needs()
		return s.call("GET", "/api/points", "", false).Body.String(), runnerRaw, s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Code
	}
	decide := func(decision string) {
		s.expect("POST", "/api/admin/decisions", `{"id":"`+s.camp+`","decision":"`+decision+`"}`, 200)
	}
	decide("archived")
	if public, runnerRaw, needs := visible(); strings.Contains(public, "C-1") || strings.Contains(runnerRaw, "Agua archivada") || needs != 404 {
		t.Fatalf("archived still visible: points=%s runner=%s needs=%d", public, runnerRaw, needs)
	}
	if w := s.call("POST", "/api/camps/"+s.camp+"/needs", `{"need":"Más","requesterName":"Eva","requesterPhone":"6555-1234"}`, false); w.Code != 404 {
		t.Fatalf("public need on archived point: %d", w.Code)
	}
	if admin := s.call("GET", "/api/admin/points", "", true).Body.String(); !strings.Contains(admin, `"status":"archived"`) {
		t.Fatalf("admin lost archived point: %s", admin)
	}
	decide("approved")
	if public, runnerRaw, needs := visible(); !strings.Contains(public, "C-1") || !strings.Contains(runnerRaw, "Agua archivada") || needs != 200 {
		t.Fatalf("recovered point not visible: points=%s runner=%s needs=%d", public, runnerRaw, needs)
	}
}

func TestPointMembersStayPrivateFillContactsAndIgnorePublicAdminFields(t *testing.T) {
	s := newServiceHarness(t)
	members := `[{"name":" Ana  Ruiz ","cedula":"8-123-456","age":"34"},{},{"name":"Luis"},{"name":"Rosa","cedula":"8-9-9","phone":"6555-1234"}]`
	body := `{"area":"Cañas","needs":"Agua","code":"MINE","service":"YES","serviceNotes":"x","notes":"y","ordinal":"3","row":"9","members":` + members + `}`
	w := s.call("POST", "/api/submissions", body, false)
	if w.Code != 201 {
		t.Fatalf("submit (phone on 3rd non-empty person): %d %s", w.Code, w.Body.String())
	}
	var created map[string]string
	json.Unmarshal(w.Body.Bytes(), &created)
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	var p Point
	for _, x := range pts {
		if x.ID == created["id"] {
			p = x
		}
	}
	if len(p.Members) != 3 || p.Members[0].Name != "Ana Ruiz" || p.Members[0].Cedula != "8-123-456" || p.Contact != "Ana Ruiz" || p.SecondContact != "Luis" || p.ThirdContact != "Rosa" || p.ThirdPhone != "6555-1234" {
		t.Fatalf("members/contacts: %+v", p)
	}
	if p.Code == "MINE" || p.Service != "" || p.ServiceNotes != "" || p.Notes != "" || p.Ordinal != "" || p.Row != "" {
		t.Fatalf("public submission set admin-only fields: %+v", p)
	}
	if w := s.call("POST", "/api/submissions", `{"needs":"Agua","members":[{"name":"Sin teléfono"}]}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"phone"`) {
		t.Fatalf("members without phone: %d %s", w.Code, w.Body.String())
	}
	many := "[" + strings.TrimSuffix(strings.Repeat(`{"phone":"6555-1234"},`, 51), ",") + "]"
	if w := s.call("POST", "/api/submissions", `{"members":`+many+`}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"members"`) {
		t.Fatalf("too many members: %d %s", w.Code, w.Body.String())
	}
	s.expect("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, 200)
	if public := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(public, "8-123-456") || strings.Contains(public, "members") || strings.Contains(public, "6555-1234") {
		t.Fatalf("public points leak members: %s", public)
	}
	edited := s.expect("PATCH", "/api/admin/camps/"+created["id"], `{"members":[{"name":"Eva","age":"50","phone":"6000-1111"},{"name":"Pedro","phone":"6000-2222"},{"name":"Juan","phone":"6000-3333"},{"name":"Cuarta","phone":"6000-4444"}]}`, 200)
	if edited["contact"] != "Eva" || edited["thirdPhone"] != "6000-3333" || len(edited["members"].([]any)) != 4 {
		t.Fatalf("patch members: %v", edited)
	}
	_, _, raw := s.runnerLogin("Luis").needs()
	if !strings.Contains(raw, `{"name":"Cuarta","phone":"6000-4444"}`) || strings.Contains(raw, "cedula") || strings.Contains(raw, `"age"`) {
		t.Fatalf("runner contacts should list every person's phone without cédula/edad: %s", raw)
	}
}

func TestPublicFormFormatsAndPrivateCoordinatorPhone(t *testing.T) {
	s := newServiceHarness(t)
	for body, field := range map[string]string{
		`{"coordinatorPhone":"6555-1234","adults":"diez"}`:                            "adults",
		`{"coordinatorPhone":"6555-1234","people":"1000"}`:                            "people",
		`{"coordinatorPhone":"llamar a Ana"}`:                                         "coordinatorPhone",
		`{"coordinatorPhone":"6555-1234","members":[{"name":"Eva","age":"treinta"}]}`: "members",
		`{"members":[{"name":"Eva","phone":"6555-1234 Ana"}]}`:                        "members",
	} {
		if w := s.call("POST", "/api/submissions", body, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"`+field+`"`) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	w := s.call("POST", "/api/submissions", `{"coordinator":"Marta","coordinatorPhone":"6555-9999","people":"12","adults":"8","needs":"Agua, Kit de bebé"}`, false)
	if w.Code != 201 {
		t.Fatalf("coordinator phone alone should satisfy the phone rule: %d %s", w.Code, w.Body.String())
	}
	var created map[string]string
	json.Unmarshal(w.Body.Bytes(), &created)
	s.expect("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, 200)
	if public := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(public, `"coordinator":"Marta"`) || strings.Contains(public, "6555-9999") {
		t.Fatalf("coordinator name public, phone private: %s", public)
	}
	_, _, raw := s.runnerLogin("Luis").needs()
	if !strings.Contains(raw, `{"name":"Marta (coordinador)","phone":"6555-9999"}`) {
		t.Fatalf("runner should see coordinator phone: %s", raw)
	}
}

func TestAssigningRunnerPutsNeedEnRuta(t *testing.T) {
	s := newServiceHarness(t)
	runner := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-1111"}`, 201)["id"].(string)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	n := s.expect("PATCH", "/api/admin/needs/"+id, `{"runnerId":"`+runner+`"}`, 200)
	if n["status"] != "IN_PROGRESS" || n["enRouteAt"] == nil {
		t.Fatalf("assigning a runner must put the need en ruta: %v", n)
	}
	// Moving it back to Abierta on purpose keeps the same runner without forcing en ruta again.
	if n := s.expect("PATCH", "/api/admin/needs/"+id, `{"status":"OPEN","runnerId":"`+runner+`"}`, 200); n["status"] != "OPEN" {
		t.Fatalf("explicit Abierta with the same runner: %v", n)
	}
	// Changing to another runner keeps it en ruta with a fresh time; removing the runner reopens it.
	s.expect("PATCH", "/api/admin/needs/"+id, `{"runnerId":"`+runner+`","status":"IN_PROGRESS"}`, 200)
	other := s.expect("POST", "/api/admin/runners", `{"name":"Luis","telephone":"6999-2222"}`, 201)["id"].(string)
	if n := s.expect("PATCH", "/api/admin/needs/"+id, `{"runnerId":"`+other+`","status":"IN_PROGRESS"}`, 200); n["status"] != "IN_PROGRESS" || n["runnerId"] != other || n["enRouteAt"] == nil {
		t.Fatalf("reassigned: %v", n)
	}
	// The admin form always sends the status it shows; removing the runner still wins.
	if n := s.expect("PATCH", "/api/admin/needs/"+id, `{"runnerId":"","status":"IN_PROGRESS"}`, 200); n["status"] != "OPEN" || n["enRouteAt"] != nil {
		t.Fatalf("unassigned need must go back to Abierta: %v", n)
	}
	// A completed need stays completed when someone is assigned.
	s.expect("PATCH", "/api/admin/needs/"+id, `{"status":"COMPLETED","runnerId":""}`, 200)
	if n := s.expect("PATCH", "/api/admin/needs/"+id, `{"runnerId":"`+runner+`"}`, 200); n["status"] != "COMPLETED" {
		t.Fatalf("completed need: %v", n)
	}
}
