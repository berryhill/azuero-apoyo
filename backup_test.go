package main

import (
	"bytes"
	"compress/gzip"
	"net/http/httptest"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

func TestBackupDownloadAndImportIntoLocal(t *testing.T) {
	live := newServiceHarness(t)
	live.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)
	if w := live.call("GET", "/api/admin/backup", "", false); w.Code != 401 {
		t.Fatalf("backup needs an admin session: %d", w.Code)
	}
	w := live.call("GET", "/api/admin/backup", "", true)
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/gzip" || !strings.Contains(w.Header().Get("Content-Disposition"), `filename="azuero-respaldo-`) {
		t.Fatalf("backup download: %d %v", w.Code, w.Header())
	}
	backup := w.Body.Bytes()

	local := newServiceHarness(t)
	localHash, _ := bcrypt.GenerateFromPassword([]byte("local-password-123"), bcrypt.MinCost)
	local.app.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(meta).Put([]byte("password_hash"), localHash) })
	restore := func(body []byte, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/admin/restore", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/gzip")
		if auth {
			r.AddCookie(local.cookie)
			r.Header.Set("X-CSRF-Token", local.csrf)
		}
		w := httptest.NewRecorder()
		local.h.ServeHTTP(w, r)
		return w
	}
	// The live site never accepts an import: it needs ALLOW_IMPORT=1.
	if got := local.expect("GET", "/api/admin/settings", "", 200); got["importEnabled"] != false {
		t.Fatalf("import is off by default: %v", got)
	}
	if w := restore(backup, true); w.Code != 403 {
		t.Fatalf("import refused without ALLOW_IMPORT: %d", w.Code)
	}
	t.Setenv("ALLOW_IMPORT", "1")
	if got := local.expect("GET", "/api/admin/settings", "", 200); got["importEnabled"] != true {
		t.Fatalf("import on: %v", got)
	}
	if w := restore(backup, false); w.Code != 401 {
		t.Fatalf("import needs an admin session: %d", w.Code)
	}
	var notBolt bytes.Buffer
	gz := gzip.NewWriter(&notBolt)
	gz.Write([]byte("hola"))
	gz.Close()
	for _, bad := range [][]byte{[]byte("not gzip"), notBolt.Bytes()} {
		if w := restore(bad, true); w.Code != 400 || !strings.Contains(w.Body.String(), "no es un respaldo válido") {
			t.Fatalf("bad file: %d %s", w.Code, w.Body.String())
		}
	}
	// A runner signed in before the import loses that session: it belonged to the old data.
	runnerCookie, _ := local.crewLogin("/api/runner", "Rosa")
	w = restore(backup, true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"users":1`) {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	if !local.app.passwordValid("local-password-123") || local.app.passwordValid("test-password-only-12345") {
		t.Fatal("import keeps the local admin password")
	}
	if got := local.call("GET", "/api/admin/points", "", true); got.Code != 200 || !strings.Contains(got.Body.String(), live.camp) || strings.Contains(got.Body.String(), local.camp) {
		t.Fatalf("the admin stays signed in and sees the imported points only: %d %s", got.Code, got.Body.String())
	}
	if w := local.call("POST", "/api/session/code", `{"accessCode":"FUFO12"}`, false); w.Code != 200 {
		t.Fatalf("imported users sign in with their codes: %d %s", w.Code, w.Body.String())
	}
	if w := local.crewCall("GET", "/api/runner/session", "", runnerCookie, ""); w.Code == 200 {
		t.Fatalf("old runner session is dropped: %s", w.Body.String())
	}
}
