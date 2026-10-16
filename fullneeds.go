package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// Página pública con código: en la ficha del punto, quien entró con código y tiene el rol Runner asigna o cambia el conductor (corredor)
// y agrega a la lista de empaque de las necesidades abiertas o en ruta, como en administración.

type crewChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Team: es del equipo de quien asigna (se muestra primero, con la etiqueta «Equipo»).
	Team bool `json:"team,omitempty"`
}

var errNeedClosed = errors.New("Esta necesidad ya no está abierta ni en camino.")
var errOtherTeam = errors.New("Esta carrera la lleva otro equipo.")

const notRunnerMsg = "Solo quien tiene el rol Runner puede asignar conductor, agregar empaque o fotos de partida."

// runnerViewer: las acciones de carrera desde la ficha (conductor, empaque, fotos de partida) piden el rol Runner.
func (a *App) runnerViewer(w http.ResponseWriter, r *http.Request) (User, bool) {
	me, ok := a.viewer(w, r)
	if !ok {
		return me, false
	}
	if !me.has(roleRunner) {
		fail(w, 403, notRunnerMsg)
		return me, false
	}
	return me, true
}

// assignable: a quien entró con código le salen los corredores de su equipo y los que no tienen código;
// los corredores con código de otro equipo (o sin equipo) no.
func assignable(t *teamView, u User) bool {
	return !u.Disabled && u.has(roleRunner) && (u.CodeHash == "" || t.mate(u.ID))
}

// fullNeed carga, para quien entró con código (me), una necesidad de suministros (no reparación) de un punto aprobado;
// abierta o en ruta, y que no lleve un corredor de otro equipo.
func (a *App) fullNeed(tx *bolt.Tx, id, me string) (CampNeed, error) {
	var n CampNeed
	v := tx.Bucket(needsBucket).Get([]byte(id))
	if v == nil {
		return n, errMissing
	}
	if e := json.Unmarshal(v, &n); e != nil {
		return n, e
	}
	if !a.campExists(tx, n.CampID) || isRepair(n) || isMealKind(n) || n.Status == "DELETED" {
		return n, errMissing
	}
	if newTeamView(tx, me).hidden(n) {
		return n, errOtherTeam
	}
	if n.Status != "OPEN" && n.Status != "IN_PROGRESS" {
		return n, errNeedClosed
	}
	return n, nil
}

// fullNeedError: 403 si es de otro equipo, 409 si ya se cerró o el runner ya lleva 10 runs.
func fullNeedError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, errOtherTeam):
		fail(w, 403, err.Error())
	case errors.Is(err, errNeedClosed):
		fail(w, 409, err.Error())
	default:
		return runError(w, err)
	}
	return true
}

func (a *App) fullNeedRoutes(mux *http.ServeMux) {
	// Corredores activos para «Asignar conductor»: primero los del equipo de quien asigna, luego por nombre. Sin teléfonos.
	mux.HandleFunc("GET /api/full/crew", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		list := []crewChoice{}
		a.db.View(func(tx *bolt.Tx) error {
			tv := newTeamView(tx, me.ID)
			return eachUser(tx, func(u User) error {
				if assignable(tv, u) {
					list = append(list, crewChoice{ID: u.ID, Name: u.Name, Team: u.CodeHash != ""})
				}
				return nil
			})
		})
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Team != list[j].Team {
				return list[i].Team
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
	// Asignar o quitar el conductor: como en administración, al asignar una necesidad abierta pasa a «En ruta» y al
	// quitarlo vuelve a «Abierta». Si ya la lleva otro runner, el nuevo se suma con su propio run (runs.go).
	mux.HandleFunc("POST /api/full/needs/{id}/runner", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.runnerViewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"runnerId": 120})
		if !ok {
			return
		}
		runnerID := fieldText(f, "runnerId")
		var n CampNeed
		var name string
		err := a.write(r, func(tx *bolt.Tx) error {
			var e error
			if n, e = a.fullNeed(tx, r.PathValue("id"), me.ID); e != nil {
				return e
			}
			// runnerId vacío quita el conductor: la necesidad vuelve a «Abierta».
			if runnerID != "" {
				u, found := getUser(tx, runnerID)
				if !found || !assignable(newTeamView(tx, me.ID), u) {
					return errInvalid
				}
				name = u.Name
			}
			// Ya la lleva otro runner: se le suma el nuevo con su propio run; el del primero no cambia.
			if otherRunner(n, runnerID) {
				if _, has, e := runOf(tx, n, runnerID); e != nil || has {
					if e == nil {
						e = errAlreadyRunning
					}
					return e
				}
				run, e := newRun(tx, n, runnerID, "IN_PROGRESS")
				if e != nil {
					return e
				}
				n = run
				n.UpdatedBy = me.ID
				return putNeed(tx.Bucket(needsBucket), &n)
			}
			before := n.RunnerID
			n.RunnerID = runnerID
			n.UpdatedBy = me.ID
			enRouteOnAssign(&n, before)
			return putNeed(tx.Bucket(needsBucket), &n)
		})
		if fullNeedError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "status": n.Status, "runnerName": name, "enRouteAt": n.EnRouteAt})
	})
	// Foto de partida desde la ficha: en necesidades abiertas o en ruta; quien entró con código ve las fotos de partida,
	// salvo las de carreras de otro equipo.
	pickupOK := func(me string) func(tx *bolt.Tx, n CampNeed) error {
		return func(tx *bolt.Tx, n CampNeed) error {
			if !a.campExists(tx, n.CampID) || isRepair(n) || n.Status != "OPEN" && n.Status != "IN_PROGRESS" || newTeamView(tx, me).hidden(n) {
				return errMissing
			}
			return nil
		}
	}
	mux.HandleFunc("POST /api/full/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.runnerViewer(w, r)
		if !ok || !sameOrigin(w, r) {
			return
		}
		if r.URL.Query().Get("stage") != stagePickup {
			failField(w, 400, "stage", "Desde la ficha solo se agregan fotos de partida.")
			return
		}
		a.receiveMedia(w, r, u.ID, u.Name, pickupOK(u.ID))
	})
	mux.HandleFunc("GET /api/full/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		a.serveMedia(w, r, func(tx *bolt.Tx, n CampNeed) error {
			if !a.campExists(tx, n.CampID) || isRepair(n) || newTeamView(tx, u.ID).hidden(n) {
				return errMissing
			}
			return nil
		})
	})
	mux.HandleFunc("POST /api/full/needs/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.runnerViewer(w, r)
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
		err = a.write(r, func(tx *bolt.Tx) error {
			var e error
			if n, e = a.fullNeed(tx, r.PathValue("id"), u.ID); e != nil {
				return e
			}
			return addPacking(tx, &n, lines, u.ID, u.Name)
		})
		if fullNeedError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
}
