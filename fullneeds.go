package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// Página pública con código: en la ficha del punto, quien entró con código asigna o cambia el conductor (corredor)
// y agrega a la lista de empaque de las necesidades abiertas o en ruta, como en administración.

type crewChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code bool   `json:"code,omitempty"`
}

var errNeedClosed = errors.New("Esta necesidad ya no está abierta ni en ruta.")

// fullNeed carga una necesidad de suministros (no reparación) de un punto aprobado; abierta o en ruta.
func (a *App) fullNeed(tx *bolt.Tx, id string) (CampNeed, error) {
	var n CampNeed
	v := tx.Bucket(needsBucket).Get([]byte(id))
	if v == nil {
		return n, errMissing
	}
	if e := json.Unmarshal(v, &n); e != nil {
		return n, e
	}
	if !a.campExists(tx, n.CampID) || isRepair(n) || n.Status == "DELETED" {
		return n, errMissing
	}
	if n.Status != "OPEN" && n.Status != "IN_PROGRESS" {
		return n, errNeedClosed
	}
	return n, nil
}

func (a *App) fullNeedRoutes(mux *http.ServeMux) {
	// Corredores activos para «Asignar conductor»: primero los que entran con código, luego por nombre. Sin teléfonos.
	mux.HandleFunc("GET /api/full/crew", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); !ok {
			return
		}
		list := []crewChoice{}
		a.db.View(func(tx *bolt.Tx) error {
			return eachUser(tx, func(u User) error {
				if !u.Disabled && u.has(roleRunner) {
					list = append(list, crewChoice{ID: u.ID, Name: u.Name, Code: u.CodeHash != ""})
				}
				return nil
			})
		})
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Code != list[j].Code {
				return list[i].Code
			}
			return foldText(list[i].Name) < foldText(list[j].Name)
		})
		response(w, 200, list)
	})
	mux.HandleFunc("GET /api/full/packing-items", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); !ok {
			return
		}
		var list []PackingItem
		err := a.db.View(func(tx *bolt.Tx) error {
			var e error
			list, e = packingItems(tx)
			return e
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	// Asignar, cambiar o quitar el conductor: como en administración, al asignar una necesidad abierta pasa a «En ruta»
	// y al quitarlo vuelve a «Abierta».
	mux.HandleFunc("POST /api/full/needs/{id}/runner", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"runnerId": 120})
		if !ok {
			return
		}
		runnerID := fieldText(f, "runnerId")
		var n CampNeed
		var name string
		err := a.db.Update(func(tx *bolt.Tx) error {
			var e error
			if n, e = a.fullNeed(tx, r.PathValue("id")); e != nil {
				return e
			}
			// runnerId vacío quita el conductor: la necesidad vuelve a «Abierta».
			if runnerID != "" {
				u, found := getUser(tx, runnerID)
				if !found || u.Disabled || !u.has(roleRunner) {
					return errInvalid
				}
				name = u.Name
			}
			before := n.RunnerID
			n.RunnerID = runnerID
			enRouteOnAssign(&n, before)
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if errors.Is(err, errNeedClosed) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "status": n.Status, "runnerName": name, "enRouteAt": n.EnRouteAt})
	})
	// Foto de partida desde la ficha: en necesidades abiertas o en ruta; quien entró con código ve las fotos de partida.
	pickupOK := func(tx *bolt.Tx, n CampNeed) error {
		if !a.campExists(tx, n.CampID) || isRepair(n) || n.Status != "OPEN" && n.Status != "IN_PROGRESS" {
			return errMissing
		}
		return nil
	}
	mux.HandleFunc("POST /api/full/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok || !sameOrigin(w, r) {
			return
		}
		if r.URL.Query().Get("stage") != stagePickup {
			failField(w, 400, "stage", "Desde la ficha solo se agregan fotos de partida.")
			return
		}
		a.receiveMedia(w, r, u.ID, u.Name, pickupOK)
	})
	mux.HandleFunc("GET /api/full/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.viewer(w, r); !ok {
			return
		}
		a.serveMedia(w, r, func(tx *bolt.Tx, n CampNeed) error {
			if !a.campExists(tx, n.CampID) || isRepair(n) {
				return errMissing
			}
			return nil
		})
	})
	mux.HandleFunc("POST /api/full/needs/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"items": 0})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		var n CampNeed
		err = a.db.Update(func(tx *bolt.Tx) error {
			var e error
			if n, e = a.fullNeed(tx, r.PathValue("id")); e != nil {
				return e
			}
			return addPacking(tx, &n, lines, u.ID, u.Name)
		})
		if errors.Is(err, errNeedClosed) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
}
