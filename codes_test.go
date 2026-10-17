package main

import (
	"encoding/json"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// Point codes by area: abbreviation + running number, continuing the codes typed by hand; Maintenance fixes the
// points that still show the automatic «P-…» code (preview first, then apply).
func TestAreaCodes(t *testing.T) {
	s := newServiceHarness(t)
	add := func(body string) Point {
		var p Point
		if w := s.call("POST", "/api/admin/points", body, true); w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatalf("point %s: %d %s", body, w.Code, w.Body.String())
		}
		return p
	}
	// Codes typed by hand (as in the real data).
	add(`{"code":"AB-1","area":"Agua Buena"}`)
	add(`{"code":"AB-6","area":"Agua Buena"}`)
	add(`{"code":"C-2","area":"Cañas"}`)
	add(`{"code":"C-4","area":"Cañas"}`)
	add(`{"code":"CA-1","area":"Rio Viejo"}`)
	add(`{"code":"RV-2","area":"Rio Viejo"}`)
	add(`{"code":"RV-3","area":"Rio Viejo"}`)
	// New points get their area's code right away.
	codeOf := func(body string) string { return add(body).Code }
	if c := codeOf(`{"area":"Agua Buena","place":"Escuela"}`); c != "AB-7" {
		t.Fatalf("continues the area's numbering: %q", c)
	}
	if c := codeOf(`{"area":"agua  buena"}`); c != "AB-8" {
		t.Fatalf("same area in other case/spacing: %q", c)
	}
	if c := codeOf(`{"area":"Río Viejo"}`); c != "RV-4" {
		t.Fatalf("accents and the most used prefix: %q", c)
	}
	if c := codeOf(`{"area":"Canas"}`); c != "C-5" {
		t.Fatalf("Cañas = Canas: %q", c)
	}
	if c := codeOf(`{"area":"Cacao"}`); c != "CAC-1" {
		t.Fatalf("a new area avoids prefixes of other areas (C = Cañas, CA = Río Viejo): %q", c)
	}
	if c := codeOf(`{"area":"El cacao"}`); c != "CAC-2" {
		t.Fatalf("«El Cacao» = «Cacao»: %q", c)
	}
	if c := codeOf(`{"area":"Pedasí"}`); c != "PE-1" {
		t.Fatalf("«P» is reserved for the automatic code: %q", c)
	}
	if c := codeOf(`{"area":"Ojo de Agua"}`); c != "OA-1" {
		t.Fatalf("initials without «de»: %q", c)
	}
	// Old points with the automatic code (as before this change), one with its code typed in the area.
	var ids []string
	if err := s.app.db.Update(func(tx *bolt.Tx) error {
		for _, x := range []Point{{ID: "a1aaaaaaaaaaaaaa", Area: "Tonosí", Status: "approved", Created: "2026-10-01T00:00:00Z"}, {ID: "b2bbbbbbbbbbbbbb", Area: "Tonosi", Status: "pending", Created: "2026-10-02T00:00:00Z"}, {ID: "c3cccccccccccccc", Area: "FL-3 Flores", Status: "approved", Created: "2026-10-03T00:00:00Z"}, {ID: "d4dddddddddddddd", Area: "Por ubicar", Status: "approved", ToLocate: true}, {ID: "e5eeeeeeeeeeeeee", Area: "Tonosí", Status: "archived"}} {
			x.Code = "P-" + x.ID[:12]
			ids = append(ids, x.ID)
			if e := putPointJSON(tx.Bucket(submissions), x); e != nil {
				return e
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var preview []codeChange
	pw := s.call("GET", "/api/admin/maintenance/codes", "", true)
	json.Unmarshal(pw.Body.Bytes(), &preview)
	got := map[string]codeChange{}
	for _, c := range preview {
		got[c.ID] = c
	}
	if len(preview) != 3 || got[ids[0]].New != "T-1" || got[ids[1]].New != "T-2" || got[ids[2]].New != "FL-3" || got[ids[2]].NewArea != "Flores" {
		t.Fatalf("preview (oldest first; «Por ubicar» and archived skipped; code typed in the area): %+v", preview)
	}
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, p := range pts {
		if p.ID == ids[0] && p.Code != "P-a1aaaaaaaaaa" {
			t.Fatalf("the preview changes nothing: %q", p.Code)
		}
	}
	w := s.call("POST", "/api/admin/maintenance/codes", `{}`, true)
	var res struct{ Changed int }
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &res) != nil || res.Changed != 3 {
		t.Fatalf("apply: %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	for _, p := range pts {
		switch p.ID {
		case ids[0]:
			if p.Code != "T-1" {
				t.Fatalf("applied: %+v", p)
			}
		case ids[2]:
			if p.Code != "FL-3" || p.Area != "Flores" {
				t.Fatalf("code taken out of the area: %+v", p)
			}
		case ids[3]:
			if p.Code != "P-d4dddddddddd" {
				t.Fatalf("«Por ubicar» keeps its automatic code until located: %+v", p)
			}
		}
	}
	if w := s.call("GET", "/api/admin/maintenance/codes", "", true); w.Body.String() != "[]\n" && w.Body.String() != "[]" {
		t.Fatalf("nothing left to fix: %s", w.Body.String())
	}
}

// Codes stay unique: many points in many spellings of the same areas, codes typed in the area, «Por ubicar», old
// automatic codes and rejected points; then «Generar códigos». An edit cannot take another point's code either.
func TestAreaCodesAreUnique(t *testing.T) {
	s := newServiceHarness(t)
	areas := []string{"Tonosí", "Tonosi", "tonosi", "Cañas", "Canas", "El Cacao", "cacao", "Cacao", "Agua Buena", "agua buena", "Ojo de Agua", "Pedasí", "P", "Río Viejo", "Rio Viejo Solar", "FL-3 Flores", "FL-3 Flores", "RV-2 Río Viejo", "Sin ubicación", "Macaracas", "Mariabé", "Mensabé", "Muñoz"}
	var ids []string
	for i := 0; i < 4; i++ {
		for _, a := range areas {
			var p Point
			w := s.call("POST", "/api/admin/points", `{"area":"`+a+`","place":"Casa `+a+`"}`, true)
			if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
				t.Fatalf("%s: %d %s", a, w.Code, w.Body.String())
			}
			ids = append(ids, p.ID)
		}
	}
	// Old points still on the automatic code (some in areas with a code typed in), a «Por ubicar» and a rejected one.
	if err := s.app.db.Update(func(tx *bolt.Tx) error {
		for i, a := range []string{"Tonosí", "FL-3 Flores", "FL-9 Flores", "Cañas", "Mensabé", "RV-2 Río Viejo"} {
			id := strings.Repeat(string("abcdef"[i]), 16)
			if e := putPointJSON(tx.Bucket(submissions), Point{ID: id, Code: "P-" + id[:12], Area: a, Status: "approved", Created: "2026-10-0" + string(rune('1'+i)) + "T00:00:00Z"}); e != nil {
				return e
			}
		}
		return putPointJSON(tx.Bucket(submissions), Point{ID: "0123456789abcdef", Code: "T-1", Area: "Tonosí", Status: "rejected"})
	}); err != nil {
		t.Fatal(err)
	}
	if w := s.call("POST", "/api/admin/maintenance/codes", `{}`, true); w.Code != 200 {
		t.Fatalf("apply: %d %s", w.Code, w.Body.String())
	}
	var pts []Point
	json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &pts)
	seen := map[string]string{}
	for _, p := range pts {
		k := strings.ToUpper(strings.TrimSpace(p.Code))
		if p.Status == "rejected" {
			continue
		}
		if other, dup := seen[k]; dup {
			t.Fatalf("duplicate code %q: %s and %s (%s)", p.Code, other, p.ID, p.Area)
		}
		seen[k] = p.ID
		if strings.HasPrefix(k, "P-") && len(k) == 14 {
			t.Fatalf("still an automatic code: %+v", p)
		}
	}
	// Editing a point to another point's code (any case) is refused; keeping its own is fine.
	var a, b Point
	for _, p := range pts {
		if p.ID == ids[0] {
			a = p
		}
		if p.ID == ids[1] {
			b = p
		}
	}
	if w := s.call("PATCH", "/api/admin/camps/"+a.ID, `{"code":"`+strings.ToLower(b.Code)+`"}`, true); w.Code/100 == 2 || !strings.Contains(w.Body.String(), "Ya hay otro punto con el código") {
		t.Fatalf("edit onto another point's code: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("PATCH", "/api/admin/camps/"+a.ID, `{"code":"`+a.Code+`","place":"Otra casa"}`, true); w.Code/100 != 2 {
		t.Fatalf("keeping its own code: %d %s", w.Code, w.Body.String())
	}
}

// Checkup: an agent changes a point's code only if someone of their team created it; the «Código» type-ahead offers
// only free codes, each once.
func TestCheckupCodeEditAndSuggestions(t *testing.T) {
	s := newServiceHarness(t)
	ana := s.expect("POST", "/api/admin/users", `{"name":"Ana Agente","telephone":"6111-0001","roles":["agent"],"accessCode":"ANAAG1"}`, 201)["id"].(string)
	beto := s.expect("POST", "/api/admin/users", `{"name":"Beto Runner","telephone":"6111-0002","roles":["runner"],"accessCode":"BETORU1"}`, 201)["id"].(string)
	caro := s.expect("POST", "/api/admin/users", `{"name":"Caro Agente","telephone":"6111-0003","roles":["agent"],"accessCode":"CAROAG1"}`, 201)["id"].(string)
	s.team("Equipo A", ana, beto)
	s.team("Equipo B", caro)
	for _, c := range []string{"S-1", "S-2", "S-4"} {
		s.call("POST", "/api/admin/points", `{"code":"`+c+`","area":"Sigua"}`, true)
	}
	var p Point
	json.Unmarshal(s.call("POST", "/api/admin/points", `{"area":"Sigua","place":"Casa de Beto","addedBy":"`+beto+`"}`, true).Body.Bytes(), &p)
	if p.Code != "S-5" {
		t.Fatalf("new Sigua point: %q", p.Code)
	}
	ac, at := s.agentLogin("ANAAG1")
	cc, ct := s.agentLogin("CAROAG1")
	var list []Point
	json.Unmarshal(s.agentCall("GET", "/api/agent/points", "", ac, at).Body.Bytes(), &list)
	for _, x := range list {
		if x.ID == p.ID && !x.CodeEditable {
			t.Fatal("Ana's teammate created it: editable for her")
		}
	}
	list = nil
	json.Unmarshal(s.agentCall("GET", "/api/agent/points", "", cc, ct).Body.Bytes(), &list)
	for _, x := range list {
		if x.ID == p.ID && x.CodeEditable {
			t.Fatal("another team: not editable")
		}
	}
	if w := s.agentCall("PATCH", "/api/agent/camps/"+p.ID, `{"code":"S-9"}`, cc, ct); w.Code != 400 || !strings.Contains(w.Body.String(), "Solo el equipo que creó el punto") {
		t.Fatalf("another team changes the code: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("PATCH", "/api/agent/camps/"+p.ID, `{"code":"s-2"}`, ac, at); w.Code != 400 || !strings.Contains(w.Body.String(), "Ya hay otro punto") {
		t.Fatalf("teammate, but the code is taken: %d %s", w.Code, w.Body.String())
	}
	if w := s.agentCall("PATCH", "/api/agent/camps/"+p.ID, `{"code":"S-3"}`, ac, at); w.Code != 200 {
		t.Fatalf("teammate, free code: %d %s", w.Code, w.Body.String())
	}
	suggest := func(q string) []string {
		var out []string
		json.Unmarshal(s.agentCall("GET", "/api/agent/maintenance/codes/suggest?q="+q+"&area=Sigua&id="+p.ID, "", ac, at).Body.Bytes(), &out)
		return out
	}
	for _, q := range []string{"", "S", "s-", "S-6", "s-2"} {
		got := suggest(q)
		seen := map[string]bool{}
		for _, c := range got {
			if seen[c] {
				t.Fatalf("%q: repeated suggestion %q in %v", q, c, got)
			}
			seen[c] = true
			if c != "S-3" && (c == "S-1" || c == "S-2" || c == "S-4") {
				t.Fatalf("%q: taken code suggested %q in %v", q, c, got)
			}
		}
		if len(got) == 0 {
			t.Fatalf("%q: no suggestions", q)
		}
	}
	if got := suggest(""); got[0] != "S-3" || got[1] != "S-5" {
		t.Fatalf("its own code first, then the next free ones (S-5 was freed when it moved to S-3): %v", got)
	}
	if got := suggest("S-6"); got[0] != "S-3" && got[0] != "S-6" {
		t.Fatalf("a typed free code is offered: %v", got)
	}
}
