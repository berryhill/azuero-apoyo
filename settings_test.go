package main

import (
	"encoding/json"
	"testing"
)

func TestRequireReviewsToggle(t *testing.T) {
	s := newServiceHarness(t)
	if got := s.expect("GET", "/api/admin/settings", "", 200); got["requireReviews"] != true {
		t.Fatalf("reviews are required by default: %v", got)
	}
	if w := s.call("GET", "/api/admin/settings", "", false); w.Code != 401 {
		t.Fatalf("settings need admin: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/settings", `{"requireReviews":false}`, false); w.Code != 401 {
		t.Fatalf("changing needs admin: %d", w.Code)
	}
	s.expect("POST", "/api/admin/settings", `{}`, 400)
	submit := func() map[string]string {
		w := s.call("POST", "/api/submissions", `{"needs":"Agua","contact":"Privada","phone":"6555-1234","code":"MIO","notes":"interna"}`, false)
		if w.Code != 201 {
			t.Fatalf("submit: %d %s", w.Code, w.Body.String())
		}
		var v map[string]string
		json.Unmarshal(w.Body.Bytes(), &v)
		return v
	}
	if got := submit(); got["status"] != "pending" {
		t.Fatalf("with reviews on, public points wait: %v", got)
	}
	if got := s.expect("POST", "/api/admin/settings", `{"requireReviews":false}`, 200); got["requireReviews"] != false {
		t.Fatalf("turn off: %v", got)
	}
	if got := s.expect("GET", "/api/admin/settings", "", 200); got["requireReviews"] != false {
		t.Fatalf("saved: %v", got)
	}
	auto := submit()
	if auto["status"] != "approved" {
		t.Fatalf("with reviews off, public points publish: %v", auto)
	}
	var public []Point
	json.Unmarshal(s.call("GET", "/api/points", "", false).Body.Bytes(), &public)
	found := false
	for _, p := range public {
		if p.ID == auto["id"] {
			found = true
			if p.Contact != "" || p.Phone != "" || p.Notes != "" || p.Code == "MIO" {
				t.Fatalf("still no private or internal fields: %+v", p)
			}
		}
	}
	if !found {
		t.Fatal("auto-approved point is public")
	}
	var needs []map[string]any
	json.Unmarshal(s.call("GET", "/api/admin/needs", "", true).Body.Bytes(), &needs)
	hasNeed := false
	for _, n := range needs {
		hasNeed = hasNeed || n["campId"] == auto["id"]
	}
	if !hasNeed {
		t.Fatal("its need opens for runners like an approved point")
	}
	if w := s.call("POST", "/api/submissions", `{"needs":"Agua"}`, false); w.Code != 400 {
		t.Fatalf("still validated: %d", w.Code)
	}
	s.expect("POST", "/api/admin/settings", `{"requireReviews":true}`, 200)
	if got := submit(); got["status"] != "pending" {
		t.Fatalf("back on: %v", got)
	}
}
