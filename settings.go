package main

import (
	"encoding/json"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

// Ajustes de administración guardados en meta. «require_reviews»: sin valor (o "on"), los puntos que agrega el público
// quedan en «Puntos por revisar»; si administración lo apaga ("off"), se publican al momento, como los de un usuario con código.
var requireReviewsKey = []byte("require_reviews")

func (a *App) requireReviews() bool {
	on := true
	a.db.View(func(tx *bolt.Tx) error {
		on = string(tx.Bucket(meta).Get(requireReviewsKey)) != "off"
		return nil
	})
	return on
}

func (a *App) settingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/settings", func(w http.ResponseWriter, r *http.Request) {
		if a.admin(w, r) {
			response(w, 200, map[string]bool{"requireReviews": a.requireReviews(), "importEnabled": importEnabled()})
		}
	})
	mux.HandleFunc("POST /api/admin/settings", func(w http.ResponseWriter, r *http.Request) {
		if !a.guard(w, r) || !a.admin(w, r) {
			return
		}
		var in struct {
			RequireReviews *bool `json:"requireReviews"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.RequireReviews == nil {
			fail(w, 400, "invalid settings")
			return
		}
		value := []byte("off")
		if *in.RequireReviews {
			value = []byte("on")
		}
		if err := a.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(meta).Put(requireReviewsKey, value) }); err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, map[string]bool{"requireReviews": *in.RequireReviews})
	})
}
