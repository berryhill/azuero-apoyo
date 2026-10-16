package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// team crea un equipo y le asigna a los usuarios (con código) indicados.
func (s *serviceHarness) team(name string, members ...string) string {
	s.t.Helper()
	id := s.expect("POST", "/api/admin/teams", `{"name":"`+name+`"}`, 201)["id"].(string)
	for _, u := range members {
		s.expect("PATCH", "/api/admin/users/"+u, `{"teamId":"`+id+`"}`, 200)
	}
	return id
}

// viewerCookie entra con código en la página pública y devuelve la cookie de la vista completa.
func (s *serviceHarness) viewerCookie(code string) *http.Cookie {
	s.t.Helper()
	for _, c := range s.call("POST", "/api/session/code", `{"accessCode":"`+code+`"}`, false).Result().Cookies() {
		if c.Name == "viewer_session" {
			return c
		}
	}
	s.t.Fatalf("no viewer session for %s", code)
	return nil
}

func (s *serviceHarness) as(c *http.Cookie, method, path, body string) *httptest.ResponseRecorder {
	s.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if method == "POST" || method == "PATCH" {
		r.Header.Set("Content-Type", "application/json")
	}
	if c != nil {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}

func TestAdminListsCreatesRenamesTeamsAndAssignsCodeUsers(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("GET", "/api/admin/teams", "", false); w.Code != 401 {
		t.Fatalf("teams are admin-only: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/teams", `{"name":"   "}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "nombre del equipo") {
		t.Fatalf("a team needs a name: %d %s", w.Code, w.Body.String())
	}
	norte := s.expect("POST", "/api/admin/teams", `{"name":"  Equipo   Norte "}`, 201)
	if norte["name"] != "Equipo Norte" {
		t.Fatalf("name is tidied: %v", norte)
	}
	id := norte["id"].(string)
	if got := s.expect("PATCH", "/api/admin/teams/"+id, `{"name":"Brigada Norte"}`, 200); got["name"] != "Brigada Norte" {
		t.Fatalf("rename: %v", got)
	}
	if w := s.call("PATCH", "/api/admin/teams/nada", `{"name":"X"}`, true); w.Code != 404 {
		t.Fatalf("unknown team: %d", w.Code)
	}
	fufo := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12","teamId":"`+id+`"}`, 201)
	if fufo["teamId"] != id {
		t.Fatalf("created straight into the team: %v", fufo)
	}
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	if w := s.call("PATCH", "/api/admin/users/"+ana, `{"teamId":"`+id+`"}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "código de acceso") {
		t.Fatalf("only code users join a team: %d %s", w.Code, w.Body.String())
	}
	// Giving her a code in the same save lets her join.
	if got := s.expect("PATCH", "/api/admin/users/"+ana, `{"teamId":"`+id+`","accessCode":"ANA123"}`, 200); got["teamId"] != id {
		t.Fatalf("code + team in one save: %v", got)
	}
	if w := s.call("PATCH", "/api/admin/users/"+ana, `{"teamId":"nada"}`, true); w.Code != 400 {
		t.Fatalf("unknown team id: %d", w.Code)
	}
	var list []adminTeam
	json.Unmarshal(s.call("GET", "/api/admin/teams", "", true).Body.Bytes(), &list)
	if len(list) != 1 || list[0].Name != "Brigada Norte" || list[0].Members != 2 {
		t.Fatalf("list with member count: %+v", list)
	}
	// Removing from the team.
	if got := s.expect("PATCH", "/api/admin/users/"+ana, `{"teamId":""}`, 200); got["teamId"] != nil {
		t.Fatalf("left the team: %v", got)
	}
}

func TestExistingCodeUsersStartInOneTeam(t *testing.T) {
	s := newServiceHarness(t)
	fufo := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	// A database from before teams existed: no teams bucket.
	s.app.db.Update(func(tx *bolt.Tx) error { return tx.DeleteBucket(teamsBucket) })
	if err := s.app.migrate(); err != nil {
		t.Fatal(err)
	}
	var list []adminTeam
	json.Unmarshal(s.call("GET", "/api/admin/teams", "", true).Body.Bytes(), &list)
	if len(list) != 1 || list[0].Name != defaultTeamName || list[0].Members != 1 {
		t.Fatalf("one default team with the code user: %+v", list)
	}
	s.app.db.View(func(tx *bolt.Tx) error {
		if u, _ := getUser(tx, fufo); u.TeamID != list[0].ID {
			t.Fatalf("code user joined: %+v", u)
		}
		if u, _ := getUser(tx, ana); u.TeamID != "" {
			t.Fatalf("users without a code stay out: %+v", u)
		}
		return nil
	})
	// Running it again changes nothing.
	s.app.migrate()
	json.Unmarshal(s.call("GET", "/api/admin/teams", "", true).Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("idempotent: %+v", list)
	}
}

// Everything that was «anyone with a code» is now «someone on the same team»: reserved run details, the team mark,
// live locations, the runner picker, and assigning/packing a run.
func TestCodeUsersOnlySeeAndActOnTheirOwnTeamsRuns(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	fufo := user(`{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`)
	bea := user(`{"name":"Bea","telephone":"6999-2222","roles":["agent"],"accessCode":"BEA123"}`)
	zed := user(`{"name":"Zed","telephone":"6999-3333","roles":["runner"],"accessCode":"ZED123"}`)
	yan := user(`{"name":"Yan","telephone":"6999-4444","roles":["agent"],"accessCode":"YAN123"}`)
	ana := user(`{"name":"Ana","telephone":"6999-5555","roles":["runner"]}`)
	solo := user(`{"name":"Solo","telephone":"6999-6666","roles":["runner"],"accessCode":"SOLO12"}`)
	s.team("Norte", fufo, bea)
	s.team("Sur", zed, yan)
	s.camp = s.expect("POST", "/api/admin/points", `{"code":"C-9","area":"Pedasí","place":"Escuela","needs":"Agua, Arroz","contact":"X","phone":"6555-0000"}`, 201)["id"].(string)
	run := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"urgente","runnerId":"`+fufo+`"}`, 201)["id"].(string)
	open := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Toldas","urgency":"normal"}`, 201)["id"].(string)
	beaC, yanC := s.viewerCookie("BEA123"), s.viewerCookie("YAN123")
	fufoC, zedC := s.viewerCookie("FUFO12"), s.viewerCookie("ZED123")

	// Camp detail: Bea (Norte) sees Fufo's run; Yan (Sur) sees only its status and urgency.
	if body := s.as(beaC, "GET", "/api/full/camps/"+s.camp+"/needs", "").Body.String(); !strings.Contains(body, `"need":"Agua"`) || !strings.Contains(body, `"runnerName":"Fufo"`) {
		t.Fatalf("teammate sees the run: %s", body)
	}
	if body := s.as(yanC, "GET", "/api/full/camps/"+s.camp+"/needs", "").Body.String(); strings.Contains(body, `"need":"Agua"`) || strings.Contains(body, "Fufo") || !strings.Contains(body, `"need":"Toldas"`) {
		t.Fatalf("other team sees the run like the public: %s", body)
	}
	// Map/list: the reported needs keep the run's items and the team mark only for Norte.
	points := func(c *http.Cookie) Point {
		var list []Point
		json.Unmarshal(s.as(c, "GET", "/api/full/points", "").Body.Bytes(), &list)
		for _, p := range list {
			if p.ID == s.camp {
				return p
			}
		}
		t.Fatal("camp missing")
		return Point{}
	}
	if p := points(beaC); !p.Team || p.Needs != "Agua, Arroz" {
		t.Fatalf("Norte: %+v", p)
	}
	if p := points(yanC); p.Team || p.Needs != "Arroz" {
		t.Fatalf("Sur: %+v", p)
	}
	// Runner picker: own team first, runners without a code; never another team's code runners or one with no team.
	var crew []crewChoice
	json.Unmarshal(s.as(yanC, "GET", "/api/full/crew", "").Body.Bytes(), &crew)
	if len(crew) != 2 || crew[0].ID != zed || !crew[0].Team || crew[1].ID != ana || crew[1].Team {
		t.Fatalf("Sur's picker: %+v", crew)
	}
	for _, other := range []string{fufo, solo} {
		if w := s.as(zedC, "POST", "/api/full/needs/"+open+"/runner", `{"runnerId":"`+other+`"}`); w.Code != 400 {
			t.Fatalf("can't hand a run to another team's runner: %d", w.Code)
		}
	}
	// Another team's run can't be reassigned, unassigned, packed or photographed.
	for _, body := range []string{`{"runnerId":"` + zed + `"}`, `{"runnerId":""}`} {
		if w := s.as(zedC, "POST", "/api/full/needs/"+run+"/runner", body); w.Code != 403 {
			t.Fatalf("other team's run: %d %s", w.Code, w.Body.String())
		}
	}
	if w := s.as(zedC, "POST", "/api/full/needs/"+run+"/packing", `{"items":[{"item":"Linternas","quantity":1}]}`); w.Code != 403 {
		t.Fatalf("other team's packing: %d", w.Code)
	}
	if w := rawUpload(s, "/api/full/needs/"+run+"/media?stage=pickup", zedC, "", "image/jpeg", []byte("\xff\xd8\xff x")); w.Code != 404 {
		t.Fatalf("other team's photo: %d", w.Code)
	}
	if w := s.as(fufoC, "POST", "/api/full/needs/"+run+"/packing", `{"items":[{"item":"Linternas","quantity":1}]}`); w.Code != 200 {
		t.Fatalf("teammate packs: %d %s", w.Code, w.Body.String())
	}
	// A teammate without the Runner role (Bea, agent) sees the run but can't pack it or hand it out.
	if w := s.as(beaC, "POST", "/api/full/needs/"+run+"/packing", `{"items":[{"item":"Linternas","quantity":1}]}`); w.Code != 403 {
		t.Fatalf("non-runner teammate packs: %d", w.Code)
	}
	if w := s.as(yanC, "POST", "/api/full/needs/"+open+"/runner", `{"runnerId":"`+zed+`"}`); w.Code != 403 {
		t.Fatalf("non-runner assigns: %d", w.Code)
	}
	if w := s.as(zedC, "POST", "/api/full/needs/"+open+"/runner", `{"runnerId":"`+zed+`"}`); w.Code != 200 {
		t.Fatalf("own team: %d %s", w.Code, w.Body.String())
	}
	// Live map: Fufo shares; Bea sees him, Yan doesn't.
	s.as(fufoC, "POST", "/api/full/location/consent", `{"accept":true}`)
	s.as(fufoC, "POST", "/api/full/location", `{"lat":7.5,"lon":-80.3}`)
	if body := s.as(beaC, "GET", "/api/full/runners", "").Body.String(); !strings.Contains(body, "Fufo") {
		t.Fatalf("teammate on the map: %s", body)
	}
	if body := s.as(yanC, "GET", "/api/full/runners", "").Body.String(); strings.Contains(body, "Fufo") {
		t.Fatalf("other team off the map: %s", body)
	}
	// The public still sees no run detail at all.
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, `"need":"Agua"`) || strings.Contains(body, "Fufo") {
		t.Fatalf("public: %s", body)
	}
}
