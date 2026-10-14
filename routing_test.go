package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunnerStreetRouteProxiesOSRM(t *testing.T) {
	s := newServiceHarness(t)
	var calls atomic.Int32
	var gotPath, gotAgent string
	fail := false
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		gotPath, gotAgent = r.URL.Path+"?"+r.URL.RawQuery, r.Header.Get("User-Agent")
		if fail {
			http.Error(w, "down", 503)
			return
		}
		w.Write([]byte(`{"code":"Ok","routes":[{"distance":4210.5,"duration":540,"geometry":{"coordinates":[[-80.30,7.40],[-80.31,7.41],[-80.32,7.42]]}}]}`))
	}))
	defer osrm.Close()
	s.app.router.base, s.app.router.gap = osrm.URL, 0

	if w := s.call("GET", "/api/runner/route?from=7.4,-80.3&to=7.42,-80.32", "", false); w.Code != 401 {
		t.Fatalf("route requires a runner session: %d", w.Code)
	}
	ana := s.runnerLogin("Ana")
	for _, q := range []string{"from=7.4&to=7.42,-80.32", "from=abc,1&to=7.42,-80.32", "from=95,1&to=7.42,-80.32", "to=7.42,-80.32"} {
		if w := ana.call("GET", "/api/runner/route?"+q, ""); w.Code != 400 {
			t.Fatalf("%s: %d", q, w.Code)
		}
	}
	w := ana.call("GET", "/api/runner/route?from=7.400001,-80.3&to=7.42,-80.32", "")
	if w.Code != 200 {
		t.Fatalf("route: %d %s", w.Code, w.Body.String())
	}
	var route streetRoute
	json.Unmarshal(w.Body.Bytes(), &route)
	if route.Distance != 4210.5 || route.Duration != 540 || len(route.Coords) != 3 || route.Coords[0] != [2]float64{7.40, -80.30} {
		t.Fatalf("route converted to lat,lon: %+v", route)
	}
	if gotPath != "/route/v1/driving/-80.30000,7.40000;-80.32000,7.42000?overview=full&geometries=geojson" || !strings.HasPrefix(gotAgent, "azuero-apoyo/") {
		t.Fatalf("OSRM request: %s %q", gotPath, gotAgent)
	}
	// The same trip (within ~10 m) is served from the cache.
	ana.call("GET", "/api/runner/route?from=7.40002,-80.30001&to=7.42,-80.32", "")
	if calls.Load() != 1 {
		t.Fatalf("cached route must not call OSRM again: %d calls", calls.Load())
	}
	fail = true
	if w := ana.call("GET", "/api/runner/route?from=7.5,-80.3&to=7.42,-80.32", ""); w.Code != 502 {
		t.Fatalf("routing down: %d", w.Code)
	}
}
