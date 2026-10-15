package main

import (
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Recurring camp service: due → a runner does it → scheduled for repeatDays later (hidden as an open need until then).
func TestRecurringServiceLifeCycle(t *testing.T) {
	s := newServiceHarness(t)
	svc := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Recolección de basura","urgency":"normal","kind":"service","repeatDays":7}`, 201)
	id := svc["id"].(string)
	if svc["kind"] != "service" || svc["repeatDays"].(float64) != 7 {
		t.Fatalf("created: %v", svc)
	}
	ana := s.runnerLogin("Ana")
	if _, available, raw := ana.needs(); len(available) != 1 || !strings.Contains(raw, "Recolección de basura") {
		t.Fatalf("due service is a run: %s", raw)
	}
	if w := ana.call("POST", "/api/runner/needs/"+id+"/take", "{}"); w.Code != 200 {
		t.Fatalf("take: %d", w.Code)
	}
	if w := ana.call("POST", "/api/runner/needs/"+id+"/complete", "{}"); w.Code != 200 {
		t.Fatalf("complete: %d %s", w.Code, w.Body.String())
	}
	var n CampNeed
	for _, x := range s.needsOf(s.camp) {
		if x.ID == id {
			n = x
		}
	}
	due, _ := time.Parse(time.RFC3339, n.NextDueAt)
	if n.Status != "OPEN" || n.RunnerID != "" || len(n.DoneLog) != 1 || due.Sub(time.Now()) < 6*24*time.Hour || due.Sub(time.Now()) > 8*24*time.Hour {
		t.Fatalf("rescheduled in 7 days, unassigned, logged: %+v", n)
	}
	if _, available, raw := ana.needs(); len(available) != 0 {
		t.Fatalf("scheduled service is not a run yet: %s", raw)
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(pub, "Recolección") {
		t.Fatalf("not public while scheduled: %s", pub)
	}
	if pts := s.call("GET", "/api/points", "", false).Body.String(); strings.Contains(pts, `"needStatuses":["OPEN"]`) {
		t.Fatalf("scheduled service does not count as open: %s", pts)
	}
	// When the date comes, it is a run again.
	s.app.db.Update(func(tx *bolt.Tx) error {
		n.NextDueAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		return putJSON(tx.Bucket(needsBucket), n.ID, n)
	})
	if _, available, _ := ana.needs(); len(available) != 1 {
		t.Fatalf("due again: %+v", available)
	}
	// Admin edits the interval and the date; a bad date is refused.
	s.expect("PATCH", "/api/admin/needs/"+id, `{"repeatDays":14,"nextDueAt":"2030-01-05"}`, 200)
	if w := s.call("PATCH", "/api/admin/needs/"+id, `{"nextDueAt":"mañana"}`, true); w.Code == 200 {
		t.Fatalf("bad date accepted")
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"X","urgency":"normal","kind":"service","repeatDays":400}`, true); w.Code != 400 {
		t.Fatalf("400 days: %d", w.Code)
	}
}
