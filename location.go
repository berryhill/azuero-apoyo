package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Ubicación en vivo de corredores: cuando un corredor activa «Mi ubicación en vivo» en su panel, el teléfono envía
// su posición y quienes entraron con código la ven en el mapa público. Solo se guarda la última posición de cada
// corredor y solo en memoria (nunca en la base de datos); deja de mostrarse 2 minutos después de la última.
const locationTTL = 2 * time.Minute

type liveLocation struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy float64 `json:"accuracy,omitempty"`
	At       time.Time
}

type runnerOnMap struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy float64 `json:"accuracy,omitempty"`
	At       string  `json:"at"`
	Seconds  int     `json:"seconds"`
}

func (a *App) locationRoutes(mux *http.ServeMux, runners *crew) {
	mux.HandleFunc("POST /api/runner/location", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		if s.agent {
			fail(w, 403, "Solo los corredores comparten su ubicación.")
			return
		}
		a.saveLocation(w, r, s.runnerID)
	})
	// Página pública con código: cualquiera que entró con código puede compartir su ubicación con los demás
	// (botón «Compartir mi ubicación» en el mapa; nunca se activa sola).
	mux.HandleFunc("POST /api/full/location", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			a.saveLocation(w, r, u.ID)
		}
	})
	mux.HandleFunc("POST /api/full/location/stop", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			a.mu.Lock()
			delete(a.locations, u.ID)
			a.mu.Unlock()
			response(w, 200, map[string]string{"status": "stopped"})
		}
	})
	mux.HandleFunc("POST /api/runner/location/stop", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		a.mu.Lock()
		delete(a.locations, s.runnerID)
		a.mu.Unlock()
		response(w, 200, map[string]string{"status": "stopped"})
	})
	// Vista completa (entró con código): los otros corredores con ubicación reciente, sin incluir a quien consulta.
	mux.HandleFunc("GET /api/full/runners", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		response(w, 200, a.liveRunners(me.ID))
	})
	// Panel de corredores: quien entró con código (corredor o agente) también ve a los demás en su mapa.
	// Quien tiene código solo puede entrar con él, así que «tiene código» equivale a «entró con código».
	mux.HandleFunc("GET /api/runner/team", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		var u User
		var found bool
		a.db.View(func(tx *bolt.Tx) error { u, found = getUser(tx, s.runnerID); return nil })
		if !found || u.Disabled || u.CodeHash == "" {
			fail(w, 403, "Solo quienes entran con código ven al equipo en el mapa.")
			return
		}
		response(w, 200, a.liveRunners(u.ID))
	})
}

func (a *App) saveLocation(w http.ResponseWriter, r *http.Request, userID string) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var in liveLocation
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Lat < -90 || in.Lat > 90 || in.Lon < -180 || in.Lon > 180 || in.Accuracy < 0 {
		fail(w, 400, "Ubicación inválida.")
		return
	}
	in.At = time.Now()
	a.mu.Lock()
	a.locations[userID] = in
	a.mu.Unlock()
	response(w, 200, map[string]string{"status": "ok"})
}

// liveRunners: corredores (o usuarios con código) activos con ubicación reciente, sin incluir a exclude; las posiciones vencidas se borran.
func (a *App) liveRunners(exclude string) []runnerOnMap {
	now := time.Now()
	a.mu.Lock()
	live := map[string]liveLocation{}
	for id, loc := range a.locations {
		if now.Sub(loc.At) > locationTTL {
			delete(a.locations, id)
		} else if id != exclude {
			live[id] = loc
		}
	}
	a.mu.Unlock()
	list := []runnerOnMap{}
	a.db.View(func(tx *bolt.Tx) error {
		for id, loc := range live {
			u, found := getUser(tx, id)
			if !found || u.Disabled || !u.has(roleRunner) && u.CodeHash == "" {
				continue
			}
			list = append(list, runnerOnMap{ID: id, Name: u.Name, Lat: loc.Lat, Lon: loc.Lon, Accuracy: loc.Accuracy, At: loc.At.UTC().Format(time.RFC3339), Seconds: int(now.Sub(loc.At).Seconds())})
		}
		return nil
	})
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}
