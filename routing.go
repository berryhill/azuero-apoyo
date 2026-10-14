package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Ruta por calles desde la camioneta del corredor hasta su próxima parada. El servidor la pide a OSRM
// (por defecto el servidor público de demostración) para que los teléfonos no lo contacten directamente.
// ROUTING_URL permite cambiar a otro servidor OSRM (por ejemplo uno propio) sin tocar la página.
const defaultRoutingURL = "https://router.project-osrm.org"

type router struct {
	base   string
	client *http.Client
	mu     sync.Mutex
	last   time.Time
	// El servidor público pide como máximo una consulta por segundo.
	gap   time.Duration
	cache map[string]cachedRoute
}

type streetRoute struct {
	Coords   [][2]float64 `json:"coords"`
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
}

type cachedRoute struct {
	route   streetRoute
	expires time.Time
}

func newRouter() *router {
	base := strings.TrimRight(os.Getenv("ROUTING_URL"), "/")
	if base == "" {
		base = defaultRoutingURL
	}
	return &router{base: base, client: &http.Client{Timeout: 8 * time.Second}, gap: time.Second, cache: map[string]cachedRoute{}}
}

// parseLatLon lee «lat,lon» y lo redondea a 5 decimales (~1 m).
func parseLatLon(s string) ([2]float64, bool) {
	a, b, ok := strings.Cut(s, ",")
	if !ok {
		return [2]float64{}, false
	}
	lat, e1 := strconv.ParseFloat(strings.TrimSpace(a), 64)
	lon, e2 := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if e1 != nil || e2 != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return [2]float64{}, false
	}
	round := func(v float64) float64 { f, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', 5, 64), 64); return f }
	return [2]float64{round(lat), round(lon)}, true
}

func (rt *router) route(ctx context.Context, from, to [2]float64) (streetRoute, error) {
	// Misma consulta (a ~10 m) en los últimos 2 minutos: se reutiliza.
	key := fmt.Sprintf("%.4f,%.4f;%.4f,%.4f", from[0], from[1], to[0], to[1])
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if c, ok := rt.cache[key]; ok && time.Now().Before(c.expires) {
		return c.route, nil
	}
	if wait := rt.gap - time.Since(rt.last); wait > 0 {
		time.Sleep(wait)
	}
	rt.last = time.Now()
	url := fmt.Sprintf("%s/route/v1/driving/%.5f,%.5f;%.5f,%.5f?overview=full&geometries=geojson", rt.base, from[1], from[0], to[1], to[0])
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return streetRoute{}, err
	}
	req.Header.Set("User-Agent", "azuero-apoyo/1.0 (+https://azuero-apoyo.berryhill.dev)")
	res, err := rt.client.Do(req)
	if err != nil {
		return streetRoute{}, err
	}
	defer res.Body.Close()
	var body struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Coordinates [][2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}
	if res.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&body) != nil || body.Code != "Ok" || len(body.Routes) == 0 {
		return streetRoute{}, fmt.Errorf("routing service: status %d code %q", res.StatusCode, body.Code)
	}
	best := body.Routes[0]
	out := streetRoute{Distance: best.Distance, Duration: best.Duration, Coords: make([][2]float64, 0, len(best.Geometry.Coordinates))}
	for _, c := range best.Geometry.Coordinates {
		out.Coords = append(out.Coords, [2]float64{c[1], c[0]}) // GeoJSON es lon,lat; Leaflet usa lat,lon.
	}
	if len(rt.cache) > 500 {
		rt.cache = map[string]cachedRoute{}
	}
	rt.cache[key] = cachedRoute{route: out, expires: time.Now().Add(2 * time.Minute)}
	return out, nil
}

func (a *App) routingRoutes(mux *http.ServeMux, runners *crew) {
	mux.HandleFunc("GET /api/runner/route", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.crewSession(runners, w, r); !ok {
			return
		}
		from, ok1 := parseLatLon(r.URL.Query().Get("from"))
		to, ok2 := parseLatLon(r.URL.Query().Get("to"))
		if !ok1 || !ok2 {
			fail(w, 400, "Ubicación inválida.")
			return
		}
		route, err := a.router.route(r.Context(), from, to)
		if err != nil {
			fail(w, 502, "No se pudo calcular la ruta por calles ahora.")
			return
		}
		response(w, 200, route)
	})
}
