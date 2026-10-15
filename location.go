package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Ubicación en vivo: cuando un corredor activa «Mi ubicación en vivo» en su panel, o alguien con código toca
// «Compartir mi ubicación» en el mapa público, el teléfono envía su posición y quienes entraron con código la ven en
// ambos mapas. La última posición de cada usuario se guarda en la base (bucket live_locations) con copia en
// memoria para las consultas; así sobrevive a un reinicio o despliegue. Deja de mostrarse 2 minutos después de la última.
// Aparte, cada posición se agrega a su recorrido GPS (track.go), que solo ve administración.
const locationTTL = 2 * time.Minute

var locationsBucket = []byte("live_locations")

type liveLocation struct {
	Lat      float64   `json:"lat"`
	Lon      float64   `json:"lon"`
	Accuracy float64   `json:"accuracy,omitempty"`
	At       time.Time `json:"at"`
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
	mux.HandleFunc("POST /api/full/location/consent", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok && sameOrigin(w, r) {
			a.acceptLocationNotice(w, r, u.ID)
		}
	})
	mux.HandleFunc("POST /api/runner/location/consent", func(w http.ResponseWriter, r *http.Request) {
		if s, ok := a.crewSession(runners, w, r); ok {
			a.acceptLocationNotice(w, r, s.runnerID)
		}
	})
	mux.HandleFunc("POST /api/full/location/stop", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.viewer(w, r); ok {
			a.dropLocation(u.ID)
			response(w, 200, map[string]string{"status": "stopped"})
		}
	})
	mux.HandleFunc("POST /api/runner/location/stop", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		a.dropLocation(s.runnerID)
		response(w, 200, map[string]string{"status": "stopped"})
	})
	// Vista completa (entró con código): los otros corredores de su equipo con ubicación reciente (solo su equipo).
	mux.HandleFunc("GET /api/full/runners", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		response(w, 200, a.liveRunners(me.ID))
	})
	// Panel de corredores: quien entró con código (corredor o agente) también ve a su equipo en su mapa.
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

// Aviso de ubicación: antes de compartirla por primera vez, la persona lee el aviso (PointForm.locationNotice en
// pointform.js) y marca «Acepto». Se guarda en su usuario con la versión del texto; si el texto cambia en algo importante,
// se sube la versión y todos lo vuelven a aceptar. Sin aceptación vigente no se guarda ninguna posición (428).
const locationNoticeVersion = "2026-10-15"

// locationConsent: cuándo aceptó el aviso vigente, o vacío.
func (u User) locationConsent() string {
	if u.LocationConsentVersion != locationNoticeVersion {
		return ""
	}
	return u.LocationConsentAt
}

func (a *App) acceptLocationNotice(w http.ResponseWriter, r *http.Request, userID string) {
	f, ok := decodeFields(w, r, map[string]int{"accept": 0})
	if !ok {
		return
	}
	var accept bool
	if json.Unmarshal(f["accept"], &accept) != nil || !accept {
		failField(w, 400, "accept", "Marca «Acepto» para compartir tu ubicación.")
		return
	}
	var u User
	err := a.db.Update(func(tx *bolt.Tx) error {
		var found bool
		if u, found = getUser(tx, userID); !found {
			return errMissing
		}
		u.LocationConsentAt, u.LocationConsentVersion = nowStamp(), locationNoticeVersion
		return putJSON(tx.Bucket(usersBucket), u.ID, u)
	})
	if serviceError(w, err) {
		return
	}
	response(w, 200, map[string]string{"locationConsentAt": u.LocationConsentAt, "version": locationNoticeVersion})
}

func (a *App) saveLocation(w http.ResponseWriter, r *http.Request, userID string) {
	var consent string
	a.db.View(func(tx *bolt.Tx) error {
		if u, ok := getUser(tx, userID); ok {
			consent = u.locationConsent()
		}
		return nil
	})
	if consent == "" {
		failField(w, 428, "locationConsent", "Acepta el aviso de ubicación para compartirla.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var in liveLocation
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Lat < -90 || in.Lat > 90 || in.Lon < -180 || in.Lon > 180 || in.Accuracy < 0 {
		fail(w, 400, "Ubicación inválida.")
		return
	}
	in.At = time.Now().UTC()
	if e := a.db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists(locationsBucket)
		if e != nil {
			return e
		}
		if e := putJSON(b, userID, in); e != nil {
			return e
		}
		// Recorrido GPS (track.go); una vez por hora se borra lo que pasó de trackKeep.
		if e := recordTrack(tx, userID, in); e != nil {
			return e
		}
		a.mu.Lock()
		prune := time.Since(a.trackPruned) > time.Hour
		if prune {
			a.trackPruned = time.Now()
		}
		a.mu.Unlock()
		if prune {
			return pruneTracks(tx, in.At.Add(-trackKeep))
		}
		return nil
	}); e != nil {
		fail(w, 500, "No se pudo guardar la ubicación.")
		return
	}
	a.mu.Lock()
	a.locations[userID] = in
	a.mu.Unlock()
	response(w, 200, map[string]string{"status": "ok"})
}

// dropLocation deja de compartir la ubicación de un usuario (detener, salir, desactivado), en memoria y en la base.
// No llamar con a.mu tomado.
func (a *App) dropLocation(ids ...string) {
	a.mu.Lock()
	for _, id := range ids {
		delete(a.locations, id)
	}
	a.mu.Unlock()
	a.db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(locationsBucket); b != nil {
			for _, id := range ids {
				b.Delete([]byte(id))
			}
		}
		return nil
	})
}

// loadLocations, al arrancar, recupera de la base las posiciones aún vigentes (las vencidas se borran).
func (a *App) loadLocations() error {
	if e := a.pruneStoredLocations(); e != nil {
		return e
	}
	return a.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(locationsBucket).ForEach(func(k, v []byte) error {
			var loc liveLocation
			if json.Unmarshal(v, &loc) == nil {
				a.mu.Lock()
				a.locations[string(k)] = loc
				a.mu.Unlock()
			}
			return nil
		})
	})
}

// pruneStoredLocations borra de la base las posiciones con más de locationTTL (o ilegibles).
func (a *App) pruneStoredLocations() error {
	now := time.Now()
	return a.db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists(locationsBucket)
		if e != nil {
			return e
		}
		var stale [][]byte
		b.ForEach(func(k, v []byte) error {
			var loc liveLocation
			if json.Unmarshal(v, &loc) != nil || now.Sub(loc.At) > locationTTL {
				stale = append(stale, append([]byte(nil), k...))
			}
			return nil
		})
		for _, k := range stale {
			if e := b.Delete(k); e != nil {
				return e
			}
		}
		return nil
	})
}

// liveRunners: compañeros de equipo de exclude (con código, activos) con ubicación reciente, sin incluir a exclude.
// Regla estricta: la ubicación en vivo solo la ve el propio equipo; nadie sin equipo ni de otro equipo sale (tampoco
// quienes entran sin código, que no tienen equipo). Las posiciones vencidas se borran.
func (a *App) liveRunners(exclude string) []runnerOnMap {
	now := time.Now()
	a.mu.Lock()
	live := map[string]liveLocation{}
	pruned := false
	for id, loc := range a.locations {
		if now.Sub(loc.At) > locationTTL {
			delete(a.locations, id)
			pruned = true
		} else if id != exclude {
			live[id] = loc
		}
	}
	a.mu.Unlock()
	if pruned {
		a.pruneStoredLocations()
	}
	list := []runnerOnMap{}
	a.db.View(func(tx *bolt.Tx) error {
		tv := newTeamView(tx, exclude)
		for id, loc := range live {
			u, found := getUser(tx, id)
			if !found || u.Disabled || u.CodeHash == "" || !tv.mate(id) {
				continue
			}
			list = append(list, runnerOnMap{ID: id, Name: u.Name, Lat: loc.Lat, Lon: loc.Lon, Accuracy: loc.Accuracy, At: loc.At.UTC().Format(time.RFC3339), Seconds: int(now.Sub(loc.At).Seconds())})
		}
		return nil
	})
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}
