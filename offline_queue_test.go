package main

import (
	"strings"
	"testing"
)

// Offline queue: the browser resends a point with the same clientRef until it gets an answer; it is saved only once.
func TestPointClientRefIsIdempotent(t *testing.T) {
	s := newServiceHarness(t)
	body := `{"area":"Cola Prueba","coordinatorPhone":"6000-5656","clientRef":"q-abc123"}`
	first := s.expect("POST", "/api/submissions", body, 201)
	again := s.call("POST", "/api/submissions", body, false)
	if again.Code != 200 || !strings.Contains(again.Body.String(), first["id"].(string)) || !strings.Contains(again.Body.String(), `"duplicate":true`) {
		t.Fatalf("resend returns the saved point: %d %s", again.Code, again.Body.String())
	}
	if n := strings.Count(s.call("GET", "/api/admin/points", "", true).Body.String(), "Cola Prueba"); n != 1 {
		t.Fatalf("one point, not %d", n)
	}
	if strings.Contains(s.call("GET", "/api/points", "", false).Body.String(), "q-abc123") {
		t.Fatal("clientRef is not public")
	}
	if w := s.call("POST", "/api/submissions", `{"area":"X","coordinatorPhone":"6000-5656","clientRef":"`+strings.Repeat("a", 65)+`"}`, false); w.Code != 400 {
		t.Fatalf("too long: %d", w.Code)
	}
	// Without clientRef nothing changes: two sends, two points.
	s.expect("POST", "/api/submissions", `{"area":"Sin Ref","coordinatorPhone":"6000-5656"}`, 201)
	s.expect("POST", "/api/submissions", `{"area":"Sin Ref","coordinatorPhone":"6000-5656"}`, 201)
}
