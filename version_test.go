package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The app version (VERSION, major.minor.patch) is served for the menu and matches the Helm chart's appVersion.
func TestAppVersion(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(appVersion()) {
		t.Fatalf("VERSION is not major.minor.patch: %q", appVersion())
	}
	s := newServiceHarness(t)
	if w := s.call("GET", "/api/version", "", false); w.Code != 200 || !strings.Contains(w.Body.String(), `"version":"`+appVersion()+`"`) {
		t.Fatalf("/api/version: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("GET", "/api/points", "", false); w.Header().Get("X-App-Version") != appVersion() {
		t.Fatal("API responses carry X-App-Version (pages notice a new version with the checks they already make)")
	}
	chart, _ := os.ReadFile("helm/Chart.yaml")
	if !strings.Contains(string(chart), `appVersion: "`+appVersion()+`"`) {
		t.Fatal("helm/Chart.yaml appVersion must match VERSION (scripts/bump-version.sh keeps them together)")
	}
}

// A stale CSRF token is refused with a message the pages recognize (/csrf/i); the page's own GET session returns the
// current token, so the page retries the same save once instead of losing what was typed.
func TestStaleCSRFRetryToken(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Julie Acopio","telephone":"6999-1010","roles":["supplier"],"accessCode":"JULIE77"}`, 201)
	jc, jt := s.supplierLogin("JULIE77")
	w := s.agentCall("POST", "/api/supplier/center", `{"center":"Villa Marina"}`, jc, "stale")
	if w.Code != 403 || !strings.Contains(strings.ToLower(w.Body.String()), "csrf") {
		t.Fatalf("stale token: %d %s", w.Code, w.Body.String())
	}
	if b := s.agentCall("GET", "/api/supplier/session", "", jc, "").Body.String(); !strings.Contains(b, `"csrf":"`+jt+`"`) {
		t.Fatalf("GET session gives the current token: %s", b)
	}
	if w := s.agentCall("POST", "/api/supplier/center", `{"center":"Villa Marina"}`, jc, jt); w.Code != 200 {
		t.Fatalf("retry with the fresh token: %d %s", w.Code, w.Body.String())
	}
}
