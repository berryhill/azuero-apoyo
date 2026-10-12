package main

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicCampDetailRoute(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	w := httptest.NewRecorder()
	app.routes().ServeHTTP(w, httptest.NewRequest("GET", "/camp.html?id=seed%3Aab-1", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Ficha del punto") {
		t.Fatalf("camp detail route: %d", w.Code)
	}
}
