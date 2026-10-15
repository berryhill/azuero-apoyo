package main

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Recorrido GPS: además de la última posición (live_locations), cada posición en vivo se agrega al historial del usuario
// (bucket location_history, un sub-bucket por usuario, claves por hora UTC de ancho fijo para que queden en orden).
// Para no llenar la base con alguien detenido, se guarda un punto nuevo solo si se movió trackMinMeters o pasó trackMinGap.
// Se conserva trackKeep; solo administración lo ve (Usuarios → abrir usuario → Recorrido GPS).
var trackBucket = []byte("location_history")

const (
	trackMinMeters = 10
	trackMinGap    = 2 * time.Minute
	trackKeep      = 30 * 24 * time.Hour
	trackMaxPoints = 20000
	trackKeyLayout = "2006-01-02T15:04:05.000000000Z"
)

type trackPoint struct {
	Lat      float64   `json:"lat"`
	Lon      float64   `json:"lon"`
	Accuracy float64   `json:"accuracy,omitempty"`
	At       time.Time `json:"at"`
}

// meters: distancia aproximada (haversine) entre dos posiciones.
func meters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0
	rad := math.Pi / 180
	dLat, dLon := (lat2-lat1)*rad, (lon2-lon1)*rad
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * r * math.Asin(math.Min(1, math.Sqrt(h)))
}

// recordTrack agrega la posición al historial del usuario si se movió o pasó tiempo suficiente desde el último punto.
func recordTrack(tx *bolt.Tx, userID string, loc liveLocation) error {
	root, err := tx.CreateBucketIfNotExists(trackBucket)
	if err != nil {
		return err
	}
	b, err := root.CreateBucketIfNotExists([]byte(userID))
	if err != nil {
		return err
	}
	if _, v := b.Cursor().Last(); v != nil {
		var last trackPoint
		if json.Unmarshal(v, &last) == nil && loc.At.Sub(last.At) < trackMinGap && meters(last.Lat, last.Lon, loc.Lat, loc.Lon) < trackMinMeters {
			return nil
		}
	}
	return putJSON(b, loc.At.UTC().Format(trackKeyLayout), trackPoint(loc))
}

// pruneTracks borra los puntos anteriores a before (y los sub-buckets que quedan vacíos).
func pruneTracks(tx *bolt.Tx, before time.Time) error {
	root := tx.Bucket(trackBucket)
	if root == nil {
		return nil
	}
	cut := before.UTC().Format(trackKeyLayout)
	var users [][]byte
	root.ForEach(func(k, v []byte) error {
		if v == nil {
			users = append(users, append([]byte(nil), k...))
		}
		return nil
	})
	for _, u := range users {
		b := root.Bucket(u)
		var old [][]byte
		c := b.Cursor()
		for k, _ := c.First(); k != nil && string(k) < cut; k, _ = c.Next() {
			old = append(old, append([]byte(nil), k...))
		}
		for _, k := range old {
			if e := b.Delete(k); e != nil {
				return e
			}
		}
		if k, _ := b.Cursor().First(); k == nil {
			if e := root.DeleteBucket(u); e != nil {
				return e
			}
		}
	}
	return nil
}

// userTrack lee los puntos de [from, to); truncated si había más de trackMaxPoints (se devuelven los primeros).
func userTrack(tx *bolt.Tx, userID string, from, to time.Time) ([]trackPoint, bool) {
	list := []trackPoint{}
	root := tx.Bucket(trackBucket)
	if root == nil {
		return list, false
	}
	b := root.Bucket([]byte(userID))
	if b == nil {
		return list, false
	}
	end := to.UTC().Format(trackKeyLayout)
	c := b.Cursor()
	for k, v := c.Seek([]byte(from.UTC().Format(trackKeyLayout))); k != nil && string(k) < end; k, v = c.Next() {
		if len(list) == trackMaxPoints {
			return list, true
		}
		var p trackPoint
		if json.Unmarshal(v, &p) == nil {
			list = append(list, p)
		}
	}
	return list, false
}

func (a *App) trackRoutes(mux *http.ServeMux) {
	// ?from=&to= en RFC 3339; sin ellos, las últimas 24 horas. Como máximo trackKeep hacia atrás.
	mux.HandleFunc("GET /api/admin/users/{id}/track", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		to, from := time.Now(), time.Now().Add(-24*time.Hour)
		for _, q := range []struct {
			name string
			t    *time.Time
		}{{"from", &from}, {"to", &to}} {
			if v := r.URL.Query().Get(q.name); v != "" {
				t, e := time.Parse(time.RFC3339, v)
				if e != nil {
					failField(w, 400, q.name, "Fecha inválida.")
					return
				}
				*q.t = t
			}
		}
		if !from.Before(to) {
			failField(w, 400, "from", "El inicio debe ser antes del final.")
			return
		}
		var points []trackPoint
		var truncated, found bool
		a.db.View(func(tx *bolt.Tx) error {
			if _, found = getUser(tx, r.PathValue("id")); found {
				points, truncated = userTrack(tx, r.PathValue("id"), from, to)
			}
			return nil
		})
		if !found {
			fail(w, 404, "not found")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		response(w, 200, map[string]any{"points": points, "truncated": truncated, "from": from.UTC().Format(time.RFC3339), "to": to.UTC().Format(time.RFC3339)})
	})
}
