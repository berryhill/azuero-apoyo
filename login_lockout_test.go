package main

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidAdminPasswordWorksAfterSharedAddressLockout(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "test-password-only-12345")
	app, err := openApp(filepath.Join(t.TempDir(), "relief.db"), ".")
	if err != nil {
		t.Fatal(err)
	}
	defer app.db.Close()
	login := func(password string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{"password":"`+password+`"}`))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		app.routes().ServeHTTP(w, r)
		return w
	}
	for i := 0; i < 5; i++ {
		if got := login("wrong-password"); got.Code != 401 {
			t.Fatalf("attempt %d: %d", i, got.Code)
		}
	}
	if got := login("wrong-password"); got.Code != 429 {
		t.Fatalf("invalid attempts not limited: %d", got.Code)
	}
	if got := login("test-password-only-12345"); got.Code != 200 {
		t.Fatalf("valid password blocked: %d %s", got.Code, got.Body.String())
	}
}
