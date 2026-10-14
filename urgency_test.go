package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestNormalizeUrgencyMapsLegacyValues(t *testing.T) {
	for in, want := range map[string]string{
		"": "normal", "normal": "normal", "Normal": "normal", "urgente": "urgente", " URGENTE ": "urgente",
		"Today": "urgente", "Urgent": "urgente", "Hoy": "urgente",
		"Tomorrow": "normal", "Eventually": "normal", "Mañana": "normal", "Eventualmente": "normal",
		"Now": "", "Ya": "",
	} {
		if got := normalizeUrgency(in); got != want {
			t.Errorf("normalizeUrgency(%q) = %q, want %q", in, got, want)
		}
	}
}

// Una base viva con prioridades antiguas se migra al abrir, sin reordenar la actividad y de forma idempotente.
func TestLegacyUrgencyMigratesOnOpen(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	path := filepath.Join(t.TempDir(), "relief.db")
	app, err := openApp(path, ".")
	if err != nil {
		t.Fatal(err)
	}
	legacy := map[string]string{"t": "Today", "u": "Urgent", "m": "Tomorrow", "e": "Eventually", "x": "", "ok": "urgente"}
	if e := app.db.Update(func(tx *bolt.Tx) error {
		for id, u := range legacy {
			n := CampNeed{ID: id, CampID: "c", Need: "Agua", Status: "OPEN", Urgency: u, UpdatedAt: "2026-10-01T00:00:00Z"}
			if e := putJSON(tx.Bucket(needsBucket), id, n); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	app.db.Close()
	for range 2 {
		if app, err = openApp(path, "."); err != nil {
			t.Fatal(err)
		}
		want := map[string]string{"t": "urgente", "u": "urgente", "m": "normal", "e": "normal", "x": "normal", "ok": "urgente"}
		app.db.View(func(tx *bolt.Tx) error {
			for id, w := range want {
				var n CampNeed
				if e := json.Unmarshal(tx.Bucket(needsBucket).Get([]byte(id)), &n); e != nil {
					t.Fatal(e)
				}
				if n.Urgency != w || n.UpdatedAt != "2026-10-01T00:00:00Z" {
					t.Errorf("%s: %+v", id, n)
				}
			}
			return nil
		})
		app.db.Close()
	}
}

// Páginas viejas en caché siguen mandando Today/Eventually: se guardan como urgente/normal y la urgente va primero.
func TestLegacyUrgencyInputIsStoredCanonically(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"Eventually"}`, 201)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Medicina","urgency":"Today"}`, 201)["id"].(string)
	if pts := s.call("GET", "/api/points", "", false).Body.String(); !strings.Contains(pts, `"highestUrgency":"urgente"`) {
		t.Fatalf("aggregate: %s", pts)
	}
	s.expect("PATCH", "/api/admin/needs/"+id, `{"urgency":"Tomorrow"}`, 200)
	for _, n := range s.needsOf(s.camp) {
		if n.Urgency != "normal" {
			t.Fatalf("stored: %+v", n)
		}
	}
	if w := s.call("PATCH", "/api/admin/needs/"+id, `{"urgency":"Bad"}`, true); w.Code != 400 {
		t.Fatalf("invalid urgency: %d %s", w.Code, w.Body.String())
	}
}
