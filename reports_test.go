package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Reports: team managers, the «📰 Boletín» role or Supply, and only with a team. Needs, en route and deliveries in the
// period, scoped to the team or everyone, without phones; «📰 Reportes» shows up in the pages of who can use it.
func TestReportsAccessAndData(t *testing.T) {
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	mia := user(`{"name":"Mia Manager","telephone":"6999-1111","roles":["runner"],"accessCode":"MIA123"}`)
	bea := user(`{"name":"Bea Boletin","telephone":"6999-2222","roles":["reports"],"accessCode":"BEA123"}`)
	leo := user(`{"name":"Leo Runner","telephone":"6999-3333","roles":["runner"],"accessCode":"LEO123"}`)
	user(`{"name":"Sola Boletin","telephone":"6999-4444","roles":["reports"],"accessCode":"SOL123"}`)
	s.team("Norte", mia, bea, leo)
	s.expect("PATCH", "/api/admin/users/"+mia, `{"roles":["runner","teammanager"]}`, 200)
	from := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	to := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	url := "/api/full/reports/supply?from=" + from + "&to=" + to + "&scope=all"
	for code, want := range map[string]int{"MIA123": 200, "BEA123": 200, "LEO123": 403} {
		if w := s.as(s.viewerCookie(code), "GET", url, ""); w.Code != want {
			t.Fatalf("%s: %d %s", code, w.Code, w.Body.String())
		}
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"BEA123"}`, false); !strings.Contains(w.Body.String(), "/reports.html") {
		t.Fatalf("the Boletín role gets «Reportes»: %s", w.Body.String())
	}
	if w := s.call("POST", "/api/session/code", `{"accessCode":"SOL123"}`, false); strings.Contains(w.Body.String(), "/reports.html") {
		t.Fatalf("no team, no reports page")
	}
	if w := s.as(s.viewerCookie("MIA123"), "GET", "/api/full/reports/supply?from="+to+"&to="+from, ""); w.Code != 400 {
		t.Fatalf("to before from: %d", w.Code)
	}
	var d map[string]any
	w := s.as(s.viewerCookie("MIA123"), "GET", url, "")
	json.Unmarshal(w.Body.Bytes(), &d)
	if d["team"] != "Norte" || d["template"] != "unidos" || d["scope"] != "all" {
		t.Fatalf("report header: %v", d)
	}
	for _, k := range []string{"needs", "enroute", "delivered"} {
		if _, ok := d[k].([]any); !ok {
			t.Fatalf("%s list: %v", k, d[k])
		}
	}
	if strings.Contains(w.Body.String(), "6999") {
		t.Fatalf("no phones in reports: %s", w.Body.String())
	}
	if w := s.as(s.viewerCookie("MIA123"), "GET", "/api/full/reports", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"team":"Norte"`) {
		t.Fatalf("reports page info: %d %s", w.Code, w.Body.String())
	}
	// A team manager can give the Boletín role.
	if w := s.as(s.viewerCookie("MIA123"), "PATCH", "/api/full/team/members/"+leo, `{"roles":["runner","reports"]}`); w.Code != 200 {
		t.Fatalf("team manager grants Boletín: %d %s", w.Code, w.Body.String())
	}
}
