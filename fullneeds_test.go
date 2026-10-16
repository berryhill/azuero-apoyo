package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCodeUserAssignsRunnerAndPacksFromPublicDetail(t *testing.T) {
	s := newServiceHarness(t)
	fufo := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana","telephone":"6999-2222","roles":["runner"]}`, 201)["id"].(string)
	off := s.expect("POST", "/api/admin/users", `{"name":"Apagado","telephone":"6999-3333","roles":["runner"]}`, 201)["id"].(string)
	s.expect("PATCH", "/api/admin/users/"+off, `{"disabled":true}`, 200)
	builder := s.expect("POST", "/api/admin/users", `{"name":"Bea","telephone":"6999-4444","roles":["builder"],"accessCode":"BEA123"}`, 201)["id"].(string)
	s.team("Norte", fufo, builder)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	done := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Toldas","urgency":"normal","status":"COMPLETED"}`, 201)["id"].(string)
	viewer, beaC := s.viewerCookie("FUFO12"), s.viewerCookie("BEA123")
	call := func(method, path, body string, c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if method == "POST" {
			r.Header.Set("Content-Type", "application/json")
		}
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/api/full/crew", "/api/full/packing-items"} {
		if w := call("GET", path, "", nil); w.Code != 401 {
			t.Fatalf("%s needs a code session: %d", path, w.Code)
		}
	}
	if w := call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+fufo+`"}`, nil); w.Code != 401 {
		t.Fatalf("assigning needs a code session: %d", w.Code)
	}
	// Runner actions need the Runner role: Bea (builder, same team) can't assign, pack or add a pickup photo.
	for _, try := range [][2]string{{"/api/full/needs/" + need + "/runner", `{"runnerId":"` + ana + `"}`}, {"/api/full/needs/" + need + "/packing", `{"items":[{"item":"Linternas","quantity":1}]}`}} {
		if w := call("POST", try[0], try[1], beaC); w.Code != 403 || !strings.Contains(w.Body.String(), "Runner") {
			t.Fatalf("non-runner %s: %d %s", try[0], w.Code, w.Body.String())
		}
	}
	var crew []crewChoice
	json.Unmarshal(call("GET", "/api/full/crew", "", viewer).Body.Bytes(), &crew)
	if len(crew) != 2 || crew[0].ID != fufo || !crew[0].Team || crew[1].ID != ana || crew[1].Team {
		t.Fatalf("active runners, own team first: %+v", crew)
	}
	if body := call("GET", "/api/full/crew", "", viewer).Body.String(); strings.Contains(body, "6999") || strings.Contains(body, builder) {
		t.Fatalf("no phones, no builders: %s", body)
	}
	for _, bad := range []string{off, builder, "nadie"} {
		if w := call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+bad+`"}`, viewer); w.Code != 400 {
			t.Fatalf("only active runners can be assigned (%q): %d", bad, w.Code)
		}
	}
	w := call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+ana+`"}`, viewer)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"IN_PROGRESS"`) || !strings.Contains(w.Body.String(), `"runnerName":"Ana"`) {
		t.Fatalf("assign puts it en camino: %d %s", w.Code, w.Body.String())
	}
	if w := call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+fufo+`"}`, viewer); w.Code != 200 || !strings.Contains(w.Body.String(), `"runnerName":"Fufo"`) {
		t.Fatalf("change runner: %d %s", w.Code, w.Body.String())
	}
	if w := call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":""}`, viewer); w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"OPEN"`) {
		t.Fatalf("unassign reopens it: %d %s", w.Code, w.Body.String())
	}
	call("POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+fufo+`"}`, viewer)
	if w := call("POST", "/api/full/needs/"+done+"/runner", `{"runnerId":"`+fufo+`"}`, viewer); w.Code != 409 {
		t.Fatalf("completed needs can't be reassigned: %d", w.Code)
	}
	if w := call("POST", "/api/full/needs/"+need+"/packing", `{"items":[{"item":"Linternas","quantity":0}]}`, viewer); w.Code != 400 {
		t.Fatalf("bad packing: %d", w.Code)
	}
	if w := call("POST", "/api/full/needs/"+need+"/packing", `{"items":[{"item":"Linternas","quantity":4}]}`, viewer); w.Code != 200 || !strings.Contains(w.Body.String(), `"addedByName":"Fufo"`) {
		t.Fatalf("pack: %d %s", w.Code, w.Body.String())
	}
	if w := call("POST", "/api/full/needs/"+done+"/packing", `{"items":[{"item":"Linternas","quantity":1}]}`, viewer); w.Code != 409 {
		t.Fatalf("closed needs can't be packed: %d", w.Code)
	}
	// The detail shows the runner's name and the packing list to code users only; never the runner ID.
	full := call("GET", "/api/full/camps/"+s.camp+"/needs", "", viewer).Body.String()
	if !strings.Contains(full, `"runnerName":"Fufo"`) || !strings.Contains(full, `"item":"Linternas"`) || strings.Contains(full, `"runnerId"`) {
		t.Fatalf("full detail: %s", full)
	}
	public := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if strings.Contains(public, "Fufo") || strings.Contains(public, "Linternas") || strings.Contains(public, "runnerName") {
		t.Fatalf("public detail stays private: %s", public)
	}
	var stored []CampNeed
	json.Unmarshal(s.call("GET", "/api/admin/needs", "", true).Body.Bytes(), &stored)
	for _, n := range stored {
		if n.ID == need && (n.RunnerID != fufo || n.RunnerName != "") {
			t.Fatalf("stored with the runner ID, never the name: %+v", n)
		}
	}
}

func TestCodeUserAddsPickupPhotoFromPublicDetail(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Bea","telephone":"6999-4444","roles":["agent"],"accessCode":"BEA123"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Rui","telephone":"6999-5555","roles":["runner"],"accessCode":"RUI123"}`, 201)
	need := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	done := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Toldas","urgency":"normal","status":"COMPLETED"}`, 201)["id"].(string)
	viewer := s.viewerCookie("RUI123")
	jpeg := []byte("\xff\xd8\xff fake jpeg")
	if w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=pickup", s.viewerCookie("BEA123"), "", "image/jpeg", jpeg); w.Code != 403 {
		t.Fatalf("an agent without the Runner role can't add a pickup photo: %d", w.Code)
	}
	if w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=pickup", &http.Cookie{Name: "viewer_session", Value: "x"}, "", "image/jpeg", jpeg); w.Code != 401 {
		t.Fatalf("needs a code session: %d", w.Code)
	}
	if w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=delivery", viewer, "", "image/jpeg", jpeg); w.Code != 400 {
		t.Fatalf("only pickup photos from the detail: %d", w.Code)
	}
	if w := rawUpload(s, "/api/full/needs/"+done+"/media?stage=pickup", viewer, "", "image/jpeg", jpeg); w.Code != 404 {
		t.Fatalf("closed needs: %d", w.Code)
	}
	w := rawUpload(s, "/api/full/needs/"+need+"/media?stage=pickup", viewer, "", "image/jpeg", jpeg)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"addedByName":"Rui"`) {
		t.Fatalf("upload: %d %s", w.Code, w.Body.String())
	}
	var item MediaItem
	json.Unmarshal(w.Body.Bytes(), &item)
	get := func(path string, c *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		return w
	}
	full := get("/api/full/camps/"+s.camp+"/needs", viewer).Body.String()
	if !strings.Contains(full, `"stage":"pickup"`) || strings.Contains(full, `"key"`) {
		t.Fatalf("code users see pickup photos without storage keys: %s", full)
	}
	if public := get("/api/camps/"+s.camp+"/needs", nil).Body.String(); strings.Contains(public, "media") {
		t.Fatalf("public sees no photos: %s", public)
	}
	if w := get("/api/full/media/"+need+"/"+item.ID, viewer); w.Code != 200 || w.Body.String() != string(jpeg) {
		t.Fatalf("view photo: %d", w.Code)
	}
	if w := get("/api/full/media/"+need+"/"+item.ID, nil); w.Code != 401 {
		t.Fatalf("photo needs a code session: %d", w.Code)
	}
}
