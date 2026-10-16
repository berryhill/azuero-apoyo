package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// The camp's public log shows all activity in time order with public fields only and no names or phones; teammates
// see each other's full name. Scout lists the points you or your team added.
func TestCampLogAndScout(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	fufo := user(`{"name":"Fufo Pérez","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`)
	ana := user(`{"name":"Ana Ruiz","telephone":"6999-2222","roles":["runner"],"accessCode":"ANA123"}`)
	user(`{"name":"Zed Otro","telephone":"6999-3333","roles":["runner"],"accessCode":"ZED123"}`)
	s.team("Norte", fufo, ana)
	fufoC, anaC, zedC := s.viewerCookie("FUFO12"), s.viewerCookie("ANA123"), s.viewerCookie("ZED123")
	w := s.as(fufoC, "POST", "/api/full/points", `{"area":"Sigua","place":"Escuela","coordinator":"Doña Rosa","coordinatorPhone":"6555-1234","needs":"Agua","people":"30"}`)
	if w.Code != 201 {
		t.Fatalf("Fufo adds a point: %d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	camp := created["id"].(string)
	var need string
	for _, n := range s.needsOf(camp) {
		need = n.ID
	}
	s.expect("PATCH", "/api/admin/camps/"+camp, `{"people":"45"}`, 200)
	if w := s.as(anaC, "POST", "/api/full/needs/"+need+"/runner", `{"runnerId":"`+ana+`"}`); w.Code != 200 {
		t.Fatalf("Ana assigns herself: %d %s", w.Code, w.Body.String())
	}
	s.as(zedC, "POST", "/api/full/points", `{"area":"Otro","coordinatorPhone":"6555-9999","needs":"Arroz"}`)

	logOf := func(path string) []LogEntry {
		var list []LogEntry
		json.Unmarshal(s.call("GET", path, "", false).Body.Bytes(), &list)
		return list
	}
	public := s.call("GET", "/api/camps/"+camp+"/log", "", false).Body.String()
	for _, secret := range []string{"Fufo", "Ana", "6999", "6555", "Doña Rosa", "Norte"} {
		if strings.Contains(public, secret) {
			t.Fatalf("public log leaks %q: %s", secret, public)
		}
	}
	list := logOf("/api/camps/" + camp + "/log")
	if len(list) < 3 || list[0].At < list[len(list)-1].At {
		t.Fatalf("newest first, all activity: %+v", list)
	}
	text := public
	for _, want := range []string{"Un voluntario de un equipo", "Administración", "Un voluntario de un equipo", `"after":"45"`, `"after":"asignado"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("public log misses %q: %s", want, text)
		}
	}
	// Ana (Norte) sees full names of her team; Zed (other team) sees roles only.
	mate := s.as(anaC, "GET", "/api/full/camps/"+camp+"/log", "").Body.String()
	if !strings.Contains(mate, "Fufo Pérez") || !strings.Contains(mate, "Ana Ruiz") || strings.Contains(mate, "6999") {
		t.Fatalf("teammate log: %s", mate)
	}
	if other := s.as(zedC, "GET", "/api/full/camps/"+camp+"/log", "").Body.String(); strings.Contains(other, "Fufo") || strings.Contains(other, "Ana") {
		t.Fatalf("other team log: %s", other)
	}
	// Scout: everyone sees every published point; the Norte point is «mine» for Fufo and «my team's» for Ana (with
	// who added it); Zed sees it without either flag.
	var mine []scoutPoint
	find := func(c *http.Cookie) scoutPoint {
		mine = nil
		json.Unmarshal(s.as(c, "GET", "/api/full/scout", "").Body.Bytes(), &mine)
		if len(mine) != 3 {
			t.Fatalf("all published points: %+v", mine)
		}
		for _, p := range mine {
			if p.ID == camp {
				return p
			}
		}
		t.Fatal("Norte point missing")
		return scoutPoint{}
	}
	if p := find(anaC); p.AddedByMate != "Fufo Pérez" || !p.AddedByTeam || p.AddedByMe || p.PointStatus != "approved" || len(p.ActiveNeeds) != 1 {
		t.Fatalf("Ana's scout: %+v", p)
	}
	if p := find(fufoC); !p.AddedByMe || p.AddedByTeam {
		t.Fatalf("Fufo's scout: %+v", p)
	}
	if p := find(zedC); p.AddedByMe || p.AddedByTeam || p.AddedByMate != "" {
		t.Fatalf("Zed's scout: %+v", p)
	}
	if w := s.call("GET", "/api/full/scout", "", false); w.Code != 401 {
		t.Fatalf("scout needs a code session: %d", w.Code)
	}
	if strings.Contains(s.as(zedC, "GET", "/api/full/scout", "").Body.String(), "6555") {
		t.Fatal("scout never returns contact phones")
	}
}
