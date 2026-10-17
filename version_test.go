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
