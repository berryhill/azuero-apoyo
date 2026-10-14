package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestAdminUsersCreateEditDisable(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("GET", "/api/admin/users", "", false); w.Code != 401 {
		t.Fatalf("users list must require admin: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/users", `{"name":"Ana","telephone":"6999-1111","roles":["runner"]}`, false); w.Code != 401 {
		t.Fatalf("user create must require admin: %d", w.Code)
	}
	u := s.expect("POST", "/api/admin/users", `{"name":" Ana  Pérez ","telephone":"6999-1111","roles":["builder","runner"],"accessCode":"ab 12cd"}`, 201)
	id := u["id"].(string)
	if u["name"] != "Ana Pérez" || u["hasAccessCode"] != true || u["disabled"] != false {
		t.Fatalf("created: %v", u)
	}
	if roles := u["roles"].([]any); len(roles) != 2 || roles[0] != "runner" || roles[1] != "builder" {
		t.Fatalf("roles keep a fixed order: %v", roles)
	}
	list := s.call("GET", "/api/admin/users", "", true).Body.String()
	if !strings.Contains(list, id) || strings.Contains(strings.ToUpper(list), "AB12CD") || strings.Contains(list, "codeHash") {
		t.Fatalf("list must include the user and never the code or its hash: %s", list)
	}
	// Runner and builder lists used by assignment show the same person.
	for _, path := range []string{"/api/admin/runners", "/api/admin/builders"} {
		if got := s.call("GET", path, "", true).Body.String(); !strings.Contains(got, id) {
			t.Fatalf("%s: %s", path, got)
		}
	}

	for body, field := range map[string]string{
		`{"name":"","telephone":"1","roles":["runner"]}`:                          "name",
		`{"name":"B","telephone":"","roles":["runner"]}`:                          "telephone",
		`{"name":"B","telephone":"1","roles":[]}`:                                 "roles",
		`{"name":"B","telephone":"1","roles":["admin"]}`:                          "roles",
		`{"name":"B","telephone":"1","roles":"runner"}`:                           "roles",
		`{"name":"B","telephone":"1","roles":["agent"],"accessCode":"x1"}`:        "accessCode",
		`{"name":"B","telephone":"1","roles":["agent"],"password":"x"}`:           "password",
		`{"name":"B","telephone":"1","roles":["agent"],"accessCode":"AB12CD"}`:    "accessCode",
		`{"name":"B","telephone":"1","roles":["agent"],"accessCode":"  ab12cd "}`: "accessCode",
	} {
		w := s.call("POST", "/api/admin/users", body, true)
		if w.Code != 400 && w.Code != 409 || !strings.Contains(w.Body.String(), `"field":"`+field+`"`) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}

	// Editing without an access code keeps the existing one.
	edited := s.expect("PATCH", "/api/admin/users/"+id, `{"telephone":"6999-2222","roles":["agent"]}`, 200)
	if edited["telephone"] != "6999-2222" || edited["hasAccessCode"] != true || len(edited["roles"].([]any)) != 1 {
		t.Fatalf("edited: %v", edited)
	}
	if got := s.call("GET", "/api/admin/runners", "", true).Body.String(); strings.Contains(got, id) {
		t.Fatalf("removed role must leave the runner list: %s", got)
	}
	var before, after string
	s.app.db.View(func(tx *bolt.Tx) error { u, _ := getUser(tx, id); before = u.CodeHash; return nil })
	s.expect("PATCH", "/api/admin/users/"+id, `{"accessCode":"NEW-9999"}`, 200)
	s.app.db.View(func(tx *bolt.Tx) error { u, _ := getUser(tx, id); after = u.CodeHash; return nil })
	if before == "" || before == after {
		t.Fatal("access code must change only when a new one is sent")
	}
	if w := s.call("PATCH", "/api/admin/users/missing", `{"name":"X"}`, true); w.Code != 404 {
		t.Fatalf("missing user: %d", w.Code)
	}
	if d := s.expect("PATCH", "/api/admin/users/"+id, `{"disabled":true}`, 200); d["disabled"] != true {
		t.Fatalf("disable: %v", d)
	}
	if d := s.expect("PATCH", "/api/admin/users/"+id, `{"disabled":false}`, 200); d["disabled"] != false {
		t.Fatalf("enable: %v", d)
	}
}

func TestDisabledRunnerCannotSignInAndLosesSession(t *testing.T) {
	s := newServiceHarness(t)
	login := s.call("POST", "/api/runner/session", `{"name":"Rosa Díaz","telephone":"6123-4567"}`, false)
	if login.Code != 200 {
		t.Fatalf("runner login: %d %s", login.Code, login.Body.String())
	}
	cookie := login.Result().Cookies()[0]
	list := s.call("GET", "/api/admin/users", "", true).Body.String()
	if !strings.Contains(list, "Rosa Díaz") || !strings.Contains(list, `"roles":["runner"]`) {
		t.Fatalf("public runner sign-in must create a user: %s", list)
	}
	var id string
	s.app.db.View(func(tx *bolt.Tx) error {
		return eachUser(tx, func(u User) error { id = u.ID; return nil })
	})
	s.expect("PATCH", "/api/admin/users/"+id, `{"disabled":true}`, 200)
	if w := s.crewCall("GET", "/api/runner/needs", "", cookie, ""); w.Code != 401 {
		t.Fatalf("disabled runner session must end: %d", w.Code)
	}
	if w := s.call("POST", "/api/runner/session", `{"name":"rosa díaz","telephone":"6123-4567"}`, false); w.Code != 403 {
		t.Fatalf("disabled runner must not sign in again: %d %s", w.Code, w.Body.String())
	}
	if n := strings.Count(s.call("GET", "/api/admin/users", "", true).Body.String(), `"id"`); n != 1 {
		t.Fatalf("disabled runner must not be recreated: %d users", n)
	}
}

// legacyDB escribe corredores, constructores y necesidades como los guardaba la versión anterior.
func legacyDB(t *testing.T, write func(tx *bolt.Tx) error) string {
	t.Helper()
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	path := filepath.Join(t.TempDir(), "relief.db")
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{runnersBucket, buildersBucket, needsBucket} {
			if _, e := tx.CreateBucketIfNotExists(name); e != nil {
				return e
			}
		}
		return write(tx)
	}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	return path
}

func TestLegacyRunnersAndBuildersMigrateWithSameIDs(t *testing.T) {
	path := legacyDB(t, func(tx *bolt.Tx) error {
		putJSON(tx.Bucket(runnersBucket), "r1", Runner{ID: "r1", Name: "Rosa", Telephone: "6000-0001"})
		putJSON(tx.Bucket(buildersBucket), "b1", Runner{ID: "b1", Name: "Beto", Telephone: "6000-0002"})
		// Sin ID dentro del registro: se usa la clave.
		putJSON(tx.Bucket(runnersBucket), "r2", Runner{Name: "Sin ID", Telephone: "6000-0003"})
		putJSON(tx.Bucket(needsBucket), "n1", CampNeed{ID: "n1", CampID: "c", Need: "agua", Status: "IN_PROGRESS", RunnerID: "r1"})
		return putJSON(tx.Bucket(needsBucket), "n2", CampNeed{ID: "n2", CampID: "c", Need: "techo", Kind: "repair", Status: "IN_PROGRESS", BuilderID: "b1"})
	})
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	app.db.View(func(tx *bolt.Tx) error {
		if u, ok := getUser(tx, "r1"); !ok || u.Name != "Rosa" || !u.has(roleRunner) || u.has(roleBuilder) {
			t.Fatalf("runner: %+v", u)
		}
		if u, ok := getUser(tx, "b1"); !ok || u.Telephone != "6000-0002" || !u.has(roleBuilder) {
			t.Fatalf("builder: %+v", u)
		}
		if u, ok := getUser(tx, "r2"); !ok || u.Name != "Sin ID" {
			t.Fatalf("runner without ID: %+v", u)
		}
		var n CampNeed
		json.Unmarshal(tx.Bucket(needsBucket).Get([]byte("n1")), &n)
		if n.RunnerID != "r1" {
			t.Fatalf("need keeps its runner: %+v", n)
		}
		return nil
	})
}

func TestLegacySamePersonMergesAndNeedsAreRelinked(t *testing.T) {
	path := legacyDB(t, func(tx *bolt.Tx) error {
		putJSON(tx.Bucket(runnersBucket), "r1", Runner{ID: "r1", Name: "Ana Pérez", Telephone: "6000-0001"})
		putJSON(tx.Bucket(buildersBucket), "b1", Runner{ID: "b1", Name: "ana  pérez", Telephone: "6000 0001"})
		putJSON(tx.Bucket(needsBucket), "n1", CampNeed{ID: "n1", CampID: "c", Need: "techo", Kind: "repair", Status: "IN_PROGRESS", BuilderID: "b1"})
		return putJSON(tx.Bucket(needsBucket), "n2", CampNeed{ID: "n2", CampID: "c", Need: "pared", Kind: "repair", Status: "OPEN", ReleasedBy: "b1"})
	})
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		app.db.View(func(tx *bolt.Tx) error {
			count := 0
			eachUser(tx, func(User) error { count++; return nil })
			if u, _ := getUser(tx, "r1"); count != 1 || !u.has(roleRunner) || !u.has(roleBuilder) {
				t.Fatalf("same name and phone must be one user with both roles: %d %+v", count, u)
			}
			var n1, n2 CampNeed
			json.Unmarshal(tx.Bucket(needsBucket).Get([]byte("n1")), &n1)
			json.Unmarshal(tx.Bucket(needsBucket).Get([]byte("n2")), &n2)
			if n1.BuilderID != "r1" || n2.ReleasedBy != "r1" {
				t.Fatalf("needs must point to the merged user: %+v %+v", n1, n2)
			}
			return nil
		})
	}
	check()
	// Editing the user and restarting must not recreate the legacy builder.
	app.db.Update(func(tx *bolt.Tx) error {
		u, _ := getUser(tx, "r1")
		u.Telephone = "6999-9999"
		return putJSON(tx.Bucket(usersBucket), u.ID, u)
	})
	app.db.Close()
	if app, err = openApp(path, "."); err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	check()
}

func TestNeedLinkedToUnknownRunnerGetsPlaceholderUser(t *testing.T) {
	path := legacyDB(t, func(tx *bolt.Tx) error {
		return putJSON(tx.Bucket(needsBucket), "n1", CampNeed{ID: "n1", CampID: "c", Need: "agua", Status: "IN_PROGRESS", RunnerID: "gone"})
	})
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	app.db.View(func(tx *bolt.Tx) error {
		u, ok := getUser(tx, "gone")
		if !ok || !u.has(roleRunner) || !u.Disabled || u.Note == "" || !strings.Contains(u.Name, "Corredor anterior") {
			t.Fatalf("placeholder: %+v", u)
		}
		return nil
	})
}

func TestAccessCodeSignInOpensEveryPanelOfTheUser(t *testing.T) {
	s := newServiceHarness(t)
	u := s.expect("POST", "/api/admin/users", `{"name":"Ana Pérez","telephone":"6999-1111","roles":["runner","builder"],"accessCode":"ANA-2026"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Solo Agente","telephone":"6999-3333","roles":["agent"],"accessCode":"AGENTE1"}`, 201)

	if w := s.call("POST", "/api/runner/session", `{"accessCode":"nope"}`, false); w.Code != 401 || !strings.Contains(w.Body.String(), `"field":"accessCode"`) {
		t.Fatalf("wrong code: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/builder/session", `{"accessCode":"agente1"}`, false); w.Code != 403 || !strings.Contains(w.Body.String(), "constructores") {
		t.Fatalf("code without builder role: %d %s", w.Code, w.Body.String())
	}
	// Agents enter the runner panel in packing-only mode.
	if w := s.call("POST", "/api/runner/session", `{"accessCode":"agente1"}`, false); w.Code != 200 || !strings.Contains(w.Body.String(), `"agent":true`) {
		t.Fatalf("agent on runner panel: %d %s", w.Code, w.Body.String())
	}
	// Spaces and case don't matter; one sign-in opens both panels the user has.
	login := s.call("POST", "/api/runner/session", `{"accessCode":" ana-2026 "}`, false)
	if login.Code != 200 || !strings.Contains(login.Body.String(), `"name":"Ana Pérez"`) {
		t.Fatalf("code login: %d %s", login.Code, login.Body.String())
	}
	cookies := map[string]*http.Cookie{}
	for _, c := range login.Result().Cookies() {
		cookies[c.Name] = c
	}
	if cookies["runner_session"] == nil || cookies["builder_session"] == nil {
		t.Fatalf("code login must open runner and builder sessions: %v", login.Result().Cookies())
	}
	if w := s.crewCall("GET", "/api/builder/session", "", cookies["builder_session"], ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Ana Pérez") {
		t.Fatalf("builder panel with the same sign-in: %d %s", w.Code, w.Body.String())
	}
	// Someone with a code can't be impersonated by typing their name.
	if w := s.call("POST", "/api/runner/session", `{"name":"ana pérez","telephone":"6000-0000"}`, false); w.Code != 403 || !strings.Contains(w.Body.String(), "código de acceso") {
		t.Fatalf("name sign-in for a user with a code: %d %s", w.Code, w.Body.String())
	}
	s.expect("PATCH", "/api/admin/users/"+u["id"].(string), `{"disabled":true}`, 200)
	if w := s.call("POST", "/api/runner/session", `{"accessCode":"ANA-2026"}`, false); w.Code != 403 {
		t.Fatalf("disabled user code: %d", w.Code)
	}
	if w := s.crewCall("GET", "/api/builder/needs", "", cookies["builder_session"], ""); w.Code != 401 {
		t.Fatalf("disabling ends every panel session: %d", w.Code)
	}
}

func TestNeedsOfAccessCodeRunnersAreReservedFromPublicAndOtherRunners(t *testing.T) {
	s := newServiceHarness(t)
	special := s.expect("POST", "/api/admin/users", `{"name":"Agente Especial","telephone":"6999-1111","roles":["runner"],"accessCode":"ESPECIAL1"}`, 201)["id"].(string)
	plain := s.expect("POST", "/api/admin/users", `{"name":"Corredor Normal","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	secret := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"insulina refrigerada","urgency":"urgente","runnerId":"`+special+`"}`, 201)["id"].(string)
	s.expect("PATCH", "/api/admin/needs/"+secret, `{"status":"IN_PROGRESS"}`, 200)
	normal := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"agua potable","urgency":"normal","runnerId":"`+plain+`"}`, 201)["id"].(string)
	s.expect("PATCH", "/api/admin/needs/"+normal, `{"status":"IN_PROGRESS"}`, 200)

	var public []map[string]any
	pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body
	json.Unmarshal(pub.Bytes(), &public)
	for _, leak := range []string{"insulina", "Agente Especial", special, "restricted", "reserv"} {
		if strings.Contains(pub.String(), leak) {
			t.Fatalf("public needs leak %q: %s", leak, pub.String())
		}
	}
	found := false
	for _, n := range public {
		switch n["id"] {
		case secret:
			found = true
			// Public sees that it's en ruta (and urgent), nothing else.
			want := map[string]any{"id": secret, "campId": s.camp, "need": "", "status": "IN_PROGRESS", "urgency": "urgente"}
			if len(n) != len(want) {
				t.Fatalf("reserved need shows only status and urgency: %v", n)
			}
			for k, v := range want {
				if n[k] != v {
					t.Fatalf("reserved need %s: %v", k, n)
				}
			}
		case normal:
			if n["need"] != "agua potable" || n["enRouteAt"] == nil {
				t.Fatalf("needs of runners without a code stay as before: %v", n)
			}
		}
	}
	if !found {
		t.Fatalf("reserved need must still be listed as en ruta: %s", pub.String())
	}
	// It still counts toward the camp's public status and urgency.
	s.expect("PATCH", "/api/admin/needs/"+normal, `{"status":"COMPLETED"}`, 200)
	if pts := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(pts, `"campStatus":"IN_PROGRESS"`) || !strings.Contains(pts, `"highestUrgency":"urgente"`) {
		t.Fatalf("aggregate: %s", pts)
	}
	// A runner without a code doesn't see it; the special runner sees all of it; admin sees everything.
	cookie, _ := s.crewLogin("/api/runner", "Otra Persona")
	if body := s.crewCall("GET", "/api/runner/needs", "", cookie, "").Body.String(); strings.Contains(body, "insulina") {
		t.Fatalf("runner without code sees a reserved need: %s", body)
	}
	login := s.call("POST", "/api/runner/session", `{"accessCode":"especial1"}`, false)
	if body := s.crewCall("GET", "/api/runner/needs", "", login.Result().Cookies()[0], "").Body.String(); !strings.Contains(body, "insulina") {
		t.Fatalf("special runner must see their own run: %s", body)
	}
	if body := s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true).Body.String(); !strings.Contains(body, "insulina") || strings.Contains(body, "restricted") {
		t.Fatalf("admin sees full details: %s", body)
	}
}

func TestReportedNeedsHideItemsOfReservedRuns(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"needs":"Agua, Alimentos, Kit de bebé"}`, 200)
	special := s.expect("POST", "/api/admin/users", `{"name":"Especial","telephone":"6999-1111","roles":["runner"],"accessCode":"ESPECIAL1"}`, 201)["id"].(string)
	plain := s.expect("POST", "/api/admin/users", `{"name":"Normal","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"agua,  KIT DE BEBE","urgency":"normal","runnerId":"`+special+`"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Alimentos","urgency":"normal","runnerId":"`+plain+`"}`, 201)
	pts := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(pts, `"needs":"Alimentos"`) {
		t.Fatalf("reported needs must drop items of reserved runs only: %s", pts)
	}
	for _, leak := range []string{"Agua", "bebé"} {
		if strings.Contains(pts, leak) {
			t.Fatalf("public points leak %q: %s", leak, pts)
		}
	}
	if adminPts := s.call("GET", "/api/admin/points", "", true).Body.String(); !strings.Contains(adminPts, "Agua, Alimentos, Kit de bebé") {
		t.Fatalf("admin keeps the full reported needs: %s", adminPts)
	}
	s.expect("POST", "/api/admin/boletines", `{"accessCode":"`+boletinAccessCode+`"}`, 201)
	if b := s.call("GET", "/api/boletines/latest", "", false).Body.String(); !strings.Contains(b, `"needs":"Alimentos"`) || strings.Contains(b, "Agua") {
		t.Fatalf("boletín must drop reserved items: %s", b)
	}
}

func TestHiddenCodeSignInFromPublicSearch(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner","builder"],"accessCode":"FUFO12"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Agente","telephone":"6999-2222","roles":["agent"],"accessCode":"AGENTE9"}`, 201)
	if w := s.call("POST", "/api/session/code", `{"accessCode":"nope"}`, false); w.Code != 401 {
		t.Fatalf("wrong code: %d", w.Code)
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"  "}`, false); w.Code != 400 {
		t.Fatalf("blank code: %d", w.Code)
	}
	w := s.call("POST", "/api/session/code", `{"accessCode":"fufo12"}`, false)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"name":"Fufo"`) || !strings.Contains(w.Body.String(), `"telephone":"6999-1111"`) || !strings.Contains(w.Body.String(), `{"label":"Corredores","url":"/runner.html"}`) || !strings.Contains(w.Body.String(), `{"label":"Constructores","url":"/builder.html"}`) {
		t.Fatalf("code sign-in: %d %s", w.Code, w.Body.String())
	}
	cookies := map[string]*http.Cookie{}
	for _, c := range w.Result().Cookies() {
		cookies[c.Name] = c
	}
	if got := s.crewCall("GET", "/api/runner/session", "", cookies["runner_session"], ""); got.Code != 200 {
		t.Fatalf("runner panel opens with the same sign-in: %d", got.Code)
	}
	if got := s.crewCall("GET", "/api/builder/session", "", cookies["builder_session"], ""); got.Code != 200 {
		t.Fatalf("builder panel opens with the same sign-in: %d", got.Code)
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"AGENTE9"}`, false); w.Code != 200 || !strings.Contains(w.Body.String(), `{"label":"Listas de empaque","url":"/runner.html"}`) || strings.Contains(w.Body.String(), "builder") {
		t.Fatalf("agent: %d %s", w.Code, w.Body.String())
	}
}

func TestCodeUserSeesReservedDetailsOnPublicList(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("PATCH", "/api/admin/camps/"+s.camp, `{"needs":"Agua, Alimentos"}`, 200)
	special := s.expect("POST", "/api/admin/users", `{"name":"Especial","telephone":"6999-1111","roles":["runner"],"accessCode":"ESPECIAL1"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"urgente","runnerId":"`+special+`"}`, 201)
	viewerCall := func(path string, c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	if w := viewerCall("/api/full/points", nil); w.Code != 401 {
		t.Fatalf("full view requires a code sign-in: %d", w.Code)
	}
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, "Agua") {
		t.Fatalf("public still hides the reserved run: %s", body)
	}
	login := s.call("POST", "/api/session/code", `{"accessCode":"especial1"}`, false)
	var viewer *http.Cookie
	for _, c := range login.Result().Cookies() {
		if c.Name == "viewer_session" {
			viewer = c
		}
	}
	if viewer == nil || viewer.Path != "/api/full" || !viewer.HttpOnly {
		t.Fatalf("viewer cookie: %+v", viewer)
	}
	if body := viewerCall("/api/full/camps/"+s.camp+"/needs", viewer).Body.String(); !strings.Contains(body, `"need":"Agua"`) || !strings.Contains(body, `"enRouteAt"`) || strings.Contains(body, special) {
		t.Fatalf("code user sees the run details (not the runner ID): %s", body)
	}
	if body := viewerCall("/api/full/points", viewer).Body.String(); !strings.Contains(body, `"needs":"Agua, Alimentos"`) {
		t.Fatalf("code user sees the full reported needs: %s", body)
	}
	s.expect("PATCH", "/api/admin/users/"+special, `{"disabled":true}`, 200)
	if w := viewerCall("/api/full/points", viewer); w.Code != 401 {
		t.Fatalf("disabling ends the full view: %d", w.Code)
	}
}

func TestCodeUserPointIsPublishedWithoutReview(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	post := func(c *http.Cookie, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/full/points", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	point := `{"area":"Cañas","place":"Escuela","needs":"Agua","coordinator":"Fufo","coordinatorPhone":"6999-1111","notes":"interno","code":"MIO-1","addedByName":"Otra persona","addedBy":"x"}`
	if w := post(nil, point); w.Code != 401 {
		t.Fatalf("publishing requires a code sign-in: %d", w.Code)
	}
	var viewer, runner *http.Cookie
	for _, c := range s.call("POST", "/api/session/code", `{"accessCode":"FUFO12"}`, false).Result().Cookies() {
		switch c.Name {
		case "viewer_session":
			viewer = c
		case "runner_session":
			runner = c
		}
	}
	if w := post(viewer, `{"area":"Sin teléfono"}`); w.Code != 400 {
		t.Fatalf("public field rules still apply (phone required): %d %s", w.Code, w.Body.String())
	}
	w := post(viewer, point)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"status":"approved"`) {
		t.Fatalf("code user point: %d %s", w.Code, w.Body.String())
	}
	var created map[string]string
	json.Unmarshal(w.Body.Bytes(), &created)
	pts := s.call("GET", "/api/points", "", false).Body.String()
	if !strings.Contains(pts, created["id"]) || !strings.Contains(pts, `"campStatus":"OPEN"`) {
		t.Fatalf("published at once, with its need: %s", pts)
	}
	var all []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &all)
	for _, p := range all {
		if p.ID == created["id"] && (p.Notes != "" || p.Code == "MIO-1" || p.Source != "user" || p.LastCheckup == "" || p.AddedByName != "Fufo" || p.AddedBy == "" || p.AddedBy == "x") {
			t.Fatalf("internal fields stay admin-only; source and Agregado por mark the user: %+v", p)
		}
	}
	if strings.Contains(pts, "addedBy") || strings.Contains(pts, "Fufo\",\"needs") {
		t.Fatalf("Agregado por is private: %s", pts)
	}
	var needs struct{ Available []runnerNeed }
	json.Unmarshal(s.crewCall("GET", "/api/runner/needs", "", runner, "").Body.Bytes(), &needs)
	team := map[string]bool{}
	for _, n := range needs.Available {
		team[n.Camp.ID] = n.Camp.Team
	}
	if v, ok := team[created["id"]]; !ok || !v || team[s.camp] {
		t.Fatalf("runner needs mark only team points: %v", team)
	}
	// A public submission can't claim to be added by someone.
	pub := s.call("POST", "/api/submissions", `{"area":"X","coordinatorPhone":"6999-0000","addedByName":"Fufo","addedBy":"x"}`, false)
	var pubCreated map[string]string
	json.Unmarshal(pub.Body.Bytes(), &pubCreated)
	all = nil // fresh slice: decoding into the old one would keep earlier points' fields
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &all)
	for _, p := range all {
		if p.ID == pubCreated["id"] && (p.AddedBy != "" || p.AddedByName != "") {
			t.Fatalf("public submissions have no Agregado por: %+v", p)
		}
	}
}

func TestAdminRevealsAccessCodeOnlyOnRequest(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"fu fo12"}`, 201)["id"].(string)
	if list := s.call("GET", "/api/admin/users", "", true).Body.String(); strings.Contains(list, "FUFO12") || strings.Contains(list, "accessCode") {
		t.Fatalf("the list never carries codes: %s", list)
	}
	if w := s.call("GET", "/api/admin/users/"+id+"/access-code", "", false); w.Code != 401 {
		t.Fatalf("reveal requires admin: %d", w.Code)
	}
	if got := s.expect("GET", "/api/admin/users/"+id+"/access-code", "", 200); got["accessCode"] != "FUFO12" {
		t.Fatalf("reveal: %v", got)
	}
	plain := s.expect("POST", "/api/admin/users", `{"name":"Sin","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	if w := s.call("GET", "/api/admin/users/"+plain+"/access-code", "", true); w.Code != 404 {
		t.Fatalf("no code: %d", w.Code)
	}
	// Codes assigned before they were kept readable can't be shown.
	s.app.db.Update(func(tx *bolt.Tx) error {
		u, _ := getUser(tx, id)
		u.AccessCode = ""
		return putJSON(tx.Bucket(usersBucket), id, u)
	})
	if w := s.call("GET", "/api/admin/users/"+id+"/access-code", "", true); w.Code != 404 || !strings.Contains(w.Body.String(), "Asigna uno nuevo") {
		t.Fatalf("legacy code: %d %s", w.Code, w.Body.String())
	}
}

func TestAccessCodeAttemptsAreLimitedPerVisitor(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	try := func(ip, path, code string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", path, strings.NewReader(`{"accessCode":"`+code+`"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Real-IP", ip)
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	// A correct code resets the count.
	for i := 0; i < 9; i++ {
		try("10.0.0.1", "/api/session/code", "NOPE")
	}
	if w := try("10.0.0.1", "/api/session/code", "FUFO12"); w.Code != 200 {
		t.Fatalf("correct code before the limit: %d", w.Code)
	}
	// Ten wrong codes (through Buscar or the panels) block every code attempt from that visitor, even the right one.
	for i := 0; i < 10; i++ {
		path := "/api/session/code"
		if i%2 == 1 {
			path = "/api/runner/session"
		}
		if w := try("10.0.0.1", path, "NOPE"); w.Code != 401 {
			t.Fatalf("wrong code %d: %d", i+1, w.Code)
		}
	}
	w := try("10.0.0.1", "/api/session/code", "FUFO12")
	if w.Code != 429 || !strings.Contains(w.Body.String(), "Espera 10 min") {
		t.Fatalf("blocked even with the right code: %d %s", w.Code, w.Body.String())
	}
	if w := try("10.0.0.1", "/api/builder/session", "FUFO12"); w.Code != 429 {
		t.Fatalf("block covers every code sign-in: %d", w.Code)
	}
	if w := try("10.0.0.2", "/api/session/code", "FUFO12"); w.Code != 200 {
		t.Fatalf("another visitor is not blocked: %d", w.Code)
	}
	// Name and phone sign-in is not affected.
	r := httptest.NewRequest("POST", "/api/runner/session", strings.NewReader(`{"name":"Rosa","telephone":"6000-0000"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Real-IP", "10.0.0.1")
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, r)
	if rec.Code != 200 {
		t.Fatalf("name sign-in during a code block: %d", rec.Code)
	}
	// After 10 minutes the visitor can try again.
	s.app.mu.Lock()
	at := s.app.codeAttempts["10.0.0.1"]
	at.until = time.Now().Add(-time.Second)
	s.app.codeAttempts["10.0.0.1"] = at
	s.app.mu.Unlock()
	if w := try("10.0.0.1", "/api/session/code", "FUFO12"); w.Code != 200 {
		t.Fatalf("after the window: %d", w.Code)
	}
}

func TestSessionCookiesAreSecureBehindHTTPSProxy(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	for _, c := range []struct{ path, body string }{
		{"/api/admin/login", `{"password":"test-password-only-12345"}`},
		{"/api/session/code", `{"accessCode":"FUFO12"}`},
		{"/api/runner/session", `{"name":"Rosa","telephone":"6000-0000"}`},
	} {
		r := httptest.NewRequest("POST", c.path, strings.NewReader(c.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		for _, ck := range w.Result().Cookies() {
			if !ck.Secure || !ck.HttpOnly || ck.SameSite != http.SameSiteStrictMode {
				t.Fatalf("%s cookie %s: %+v", c.path, ck.Name, ck)
			}
		}
		if len(w.Result().Cookies()) == 0 {
			t.Fatalf("%s set no cookies: %d %s", c.path, w.Code, w.Body.String())
		}
	}
}

func TestRunnerSessionReportsUserID(t *testing.T) {
	s := newServiceHarness(t)
	login := s.call("POST", "/api/runner/session", `{"name":"Rosa","telephone":"6000-0000"}`, false)
	var body map[string]any
	json.Unmarshal(login.Body.Bytes(), &body)
	id, _ := body["id"].(string)
	if len(id) != 64 {
		t.Fatalf("name sign-in returns the user ID: %s", login.Body.String())
	}
	if list := s.call("GET", "/api/admin/users", "", true).Body.String(); !strings.Contains(list, id) {
		t.Fatalf("the same ID is in Usuarios: %s", list)
	}
	if got := s.crewCall("GET", "/api/runner/session", "", login.Result().Cookies()[0], "").Body.String(); !strings.Contains(got, `"id":"`+id+`"`) {
		t.Fatalf("session check returns the ID: %s", got)
	}
}

func TestUsuariosShowsOnlineAndLastActive(t *testing.T) {
	s := newServiceHarness(t)
	never := s.expect("POST", "/api/admin/users", `{"name":"Nunca","telephone":"6999-1111","roles":["runner"]}`, 201)["id"].(string)
	login := s.call("POST", "/api/runner/session", `{"name":"Rosa","telephone":"6000-0000"}`, false)
	var body map[string]any
	json.Unmarshal(login.Body.Bytes(), &body)
	rosa := body["id"].(string)
	users := func() map[string]adminUser {
		var list []adminUser
		json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &list)
		out := map[string]adminUser{}
		for _, u := range list {
			out[u.ID] = u
		}
		return out
	}
	got := users()
	if !got[rosa].Online || got[rosa].LastActiveAt == "" {
		t.Fatalf("signed-in runner is online with a last active time: %+v", got[rosa])
	}
	if got[never].Online || got[never].LastActiveAt != "" {
		t.Fatalf("a user who never signed in is offline with no date: %+v", got[never])
	}
	// The time is saved on the user, so it survives a restart.
	s.app.db.View(func(tx *bolt.Tx) error {
		if u, _ := getUser(tx, rosa); u.LastActiveAt == "" {
			t.Fatal("last active time saved on the user")
		}
		return nil
	})
	// Ten minutes without activity: offline, but the last active time stays.
	s.app.mu.Lock()
	s.app.lastSeen[rosa] = time.Now().Add(-10 * time.Minute)
	s.app.mu.Unlock()
	if u := users()[rosa]; u.Online || u.LastActiveAt == "" {
		t.Fatalf("idle runner: %+v", u)
	}
	// Using the panel again brings them back online.
	s.crewCall("GET", "/api/runner/needs", "", login.Result().Cookies()[0], "")
	if !users()[rosa].Online {
		t.Fatal("activity marks the runner online again")
	}
}

// «Solo Nuestro Equipo» on the public page: a point counts when one of its needs has a runner who signs in with an access code.
func TestFullPointsMarkCampsServedByCodeRunners(t *testing.T) {
	s := newServiceHarness(t)
	coded := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	plain := s.expect("POST", "/api/admin/runners", `{"name":"Ana","telephone":"6999-2222"}`, 201)["id"].(string)
	other := s.expect("POST", "/api/admin/points", `{"code":"C-2","area":"Cañas","place":"Escuela","contact":"X","phone":"6555-0000"}`, 201)["id"].(string)
	idle := s.expect("POST", "/api/admin/points", `{"code":"C-3","area":"Tonosí","place":"Iglesia","contact":"X","phone":"6555-0001"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","status":"OPEN","urgency":"urgente","runnerId":"`+coded+`"}`, 201)
	s.expect("POST", "/api/admin/camps/"+other+"/needs", `{"need":"Comida","status":"OPEN","urgency":"urgente","runnerId":"`+plain+`"}`, 201)
	s.expect("POST", "/api/admin/camps/"+idle+"/needs", `{"need":"Ropa","status":"OPEN","urgency":"urgente"}`, 201)
	var viewer *http.Cookie
	for _, c := range s.call("POST", "/api/session/code", `{"accessCode":"FUFO12"}`, false).Result().Cookies() {
		if c.Name == "viewer_session" {
			viewer = c
		}
	}
	r := httptest.NewRequest("GET", "/api/full/points", nil)
	r.AddCookie(viewer)
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	var full []Point
	json.Unmarshal(w.Body.Bytes(), &full)
	team := map[string]bool{}
	for _, p := range full {
		team[p.ID] = p.Team
	}
	if len(team) != 3 || !team[s.camp] || team[other] || team[idle] {
		t.Fatalf("only the camp with a code runner is team: %v", team)
	}
	if pts := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pts, `"team"`) {
		t.Fatalf("the public data has no team mark: %s", pts)
	}
}
