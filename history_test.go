package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCampEditRecordsPeopleAndMembersHistory(t *testing.T) {
	s := newServiceHarness(t)
	path := "/api/admin/camps/" + s.camp
	history := func(p map[string]any) []any { h, _ := p["history"].([]any); return h }
	// Pasar los contactos sueltos a la lista de personas sin cambiarlos no es un cambio.
	p := s.expect("PATCH", path, `{"members":[{"name":"Private camp","phone":"6555-1234"}]}`, 200)
	if h := history(p); len(h) != 0 {
		t.Fatalf("legacy contacts logged: %v", h)
	}
	p = s.expect("PATCH", path, `{"people":"10"}`, 200)
	if h := history(p); len(h) != 1 || h[0].(map[string]any)["field"] != "people" || h[0].(map[string]any)["until"] == "" {
		t.Fatalf("first people change: %v", h)
	}
	p = s.expect("PATCH", path, `{"people":"10","members":[{"name":"Private camp","phone":"6555-1234"}],"place":"Otro"}`, 200)
	if h := history(p); len(h) != 1 {
		t.Fatalf("unchanged values logged: %v", h)
	}
	p = s.expect("PATCH", path, `{"people":"12","members":[{"name":"Ana","phone":"6000-1111"}]}`, 200)
	h := history(p)
	if len(h) != 3 {
		t.Fatalf("history: %v", h)
	}
	people, members := h[1].(map[string]any), h[2].(map[string]any)
	if people["field"] != "people" || people["people"] != "10" {
		t.Fatalf("people entry: %v", people)
	}
	list := members["members"].([]any)
	if members["field"] != "members" || len(list) != 1 || list[0].(map[string]any)["name"] != "Private camp" {
		t.Fatalf("members entry: %v", members)
	}
	if pub := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pub, "history") || strings.Contains(pub, "Private camp") {
		t.Fatalf("history exposed: %s", pub)
	}
	if w := s.call("PATCH", path, `{"history":[]}`, true); w.Code != 400 {
		t.Fatalf("history writable: %d %s", w.Code, w.Body.String())
	}
	// Un reporte público no puede traer su propio historial.
	sub := s.call("POST", "/api/submissions", `{"needs":"Agua","phone":"6555-9999","history":[{"field":"people","people":"99","until":"2020-01-01T00:00:00Z"}]}`, false)
	if sub.Code != 201 {
		t.Fatalf("submit: %d %s", sub.Code, sub.Body.String())
	}
	var created map[string]any
	json.Unmarshal(sub.Body.Bytes(), &created)
	if created["history"] != nil {
		t.Fatalf("submitted history kept: %v", created)
	}
}

func TestCheckupClicksAreLoggedPrivately(t *testing.T) {
	s := newServiceHarness(t)
	path := "/api/admin/camps/" + s.camp
	s.expect("PATCH", path, `{"place":"Escuela"}`, 200) // editar actualiza el último checkup pero no es un clic
	s.expect("POST", path+"/checkup", `{}`, 200)
	p := s.expect("POST", path+"/checkup", `{}`, 200)
	log, _ := p["checkups"].([]any)
	if len(log) != 2 || log[1] != p["lastCheckup"] {
		t.Fatalf("checkups: %v last %v", log, p["lastCheckup"])
	}
	if pub := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pub, "checkups") {
		t.Fatalf("checkup log exposed: %s", pub)
	}
	if w := s.call("PATCH", path, `{"checkups":[]}`, true); w.Code != 400 {
		t.Fatalf("checkups writable: %d %s", w.Code, w.Body.String())
	}
	if got := s.expect("PATCH", path, `{"people":"5"}`, 200); len(got["checkups"].([]any)) != 2 {
		t.Fatalf("edit changed checkup log: %v", got["checkups"])
	}
}
