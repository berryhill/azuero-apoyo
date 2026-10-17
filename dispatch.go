package main

import (
	"encoding/json"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// «＋ Nuevo despacho» (Supply): el acopio despacha en un solo paso, aunque el destino o el runner todavía no estén en el
// sistema. Destino: un punto aprobado existente, o uno nuevo (zona obligatoria; lugar, indicaciones, enlace de Maps o
// GPS opcionales; sin teléfono obligatorio) que queda publicado a nombre del acopio, o solo una zona («Pedasí», «El
// Cacao», sin saber exactamente dónde) que queda «📍 Por ubicar». Runner: uno existente, o una
// persona nueva (nombre obligatorio; teléfono y nota opcionales) que solo se enlaza con alguien de SU MISMO equipo
// (mismo teléfono o nombre único) y si no, se crea en ese equipo. Queda como un run en camino, despachado ya por este
// acopio, con su lista de empaque; las fotos de partida se suben después con /api/supplier/needs/{id}/media.

func (a *App) dispatchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/supplier/dispatches", func(w http.ResponseWriter, r *http.Request) {
		me, _, ok := a.supplierSession(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"pointId": 128, "point": 0, "runnerId": 128, "runner": 0, "items": 0, "need": 400})
		if !ok {
			return
		}
		var newPoint newPointInput
		var newRunner newRunnerInput
		pointID, runnerID := fieldText(f, "pointId"), fieldText(f, "runnerId")
		if pointID == "" && (f["point"] == nil || json.Unmarshal(f["point"], &newPoint) != nil || strings.TrimSpace(newPoint.Area) == "") {
			failField(w, 400, "point", "Elige el destino o escribe al menos la zona del lugar nuevo.")
			return
		}
		if runnerID == "" && (f["runner"] == nil || json.Unmarshal(f["runner"], &newRunner) != nil || strings.TrimSpace(newRunner.Name) == "") {
			failField(w, 400, "runner", "Elige el runner o escribe al menos su nombre.")
			return
		}
		if len(newRunner.Note) > userNotesMax || len(newPoint.Directions) > 8192 || len(newPoint.Place) > 200 || len(newPoint.Area) > 80 {
			failField(w, 400, "point", "Texto demasiado largo.")
			return
		}
		var lines []packingInput
		if f["items"] != nil {
			var err error
			if lines, err = parsePacking(f["items"]); userError(w, err) {
				return
			}
		}
		if len(lines) == 0 {
			failField(w, 400, "items", "Agrega al menos un artículo a la lista de empaque.")
			return
		}
		var coords []float64
		if gps := strings.TrimSpace(newPoint.GPS); pointID == "" && gps != "" {
			c, e := parseMapLink(gps) // fuera de la transacción: puede seguir un enlace corto
			if e != nil {
				failField(w, 400, "gps", "Usa un enlace de Google Maps de la zona.")
				return
			}
			coords = c
		}
		need := strings.TrimSpace(fieldText(f, "need"))
		if need == "" {
			need = "Despacho del acopio"
		}
		now := nowStamp()
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			// Destino.
			if pointID != "" {
				if !a.campExists(tx, pointID) {
					return userFieldError{"pointId", "Ese punto ya no está publicado. Elige otro o crea el destino."}
				}
			} else {
				p, e := newRunPoint(tx, me, newPoint, coords, "Supply al despachar")
				if e != nil {
					return e
				}
				pointID = p.ID
			}
			// Runner (regla de seguridad: solo se enlaza con alguien del equipo de este acopio).
			runner, e := dispatchRunner(tx, me, runnerID, newRunner)
			if e != nil {
				return e
			}
			n = CampNeed{ID: randomToken(), CampID: pointID, Need: need, Status: "IN_PROGRESS", Urgency: "normal", RunnerID: runner.ID, SupplierID: me.ID,
				CreatedAt: now, EnRouteAt: now, DispatchedAt: now, DispatchedBy: me.ID, DispatchedByName: me.Name, SupplyCenter: strings.TrimSpace(me.SupplyCenter), CreatedBy: me.ID, CreatedByName: me.Name}
			verifyStamp(tx, &me, &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			if e := putNeed(tx.Bucket(needsBucket), &n); e != nil {
				return e
			}
			if len(lines) > 0 {
				if e := addPacking(tx, &n, lines, me.ID, me.Name); e != nil {
					return e
				}
			}
			runner.RunOrder = append(runner.RunOrder, n.ID)
			return putJSON(tx.Bucket(usersBucket), runner.ID, runner)
		})
		if userError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": n.ID, "pointId": pointID, "runnerId": n.RunnerID})
	})
}

type newRunnerInput struct{ Name, Phone, Note string }

// dispatchRunner: el runner de un despacho desde Supply. Con runnerID, alguien activo con rol runner; sin él, la persona
// nueva (nombre obligatorio) se enlaza solo con alguien del equipo de este acopio (importUser) o se crea en ese equipo.
func dispatchRunner(tx *bolt.Tx, me User, runnerID string, nr newRunnerInput) (User, error) {
	if runnerID != "" {
		runner, found := getUser(tx, runnerID)
		if !found || runner.Disabled || !runner.has(roleRunner) {
			return User{}, userFieldError{"runnerId", "Elige un runner activo."}
		}
		return runner, nil
	}
	if strings.TrimSpace(nr.Name) == "" {
		return User{}, userFieldError{"runner", "Elige el runner o escribe al menos su nombre."}
	}
	team := ""
	if t, found := getTeam(tx, teamOf(tx, me.ID)); found {
		team = t.Name
	}
	runner, e := importUser(tx, strings.Join(strings.Fields(nr.Name), " "), strings.TrimSpace(nr.Phone), roleRunner, team)
	if e != nil {
		return User{}, e
	}
	// Solo a quien se acaba de crear: su nota y el aviso para administración (a quien ya existía no se le toca).
	if strings.HasPrefix(runner.Note, "Creado al importar") {
		runner.Notes = strings.TrimSpace(nr.Note)
		runner.Note = "Agregado desde Supply por " + me.Name + " al despacharle una entrega: revisa su teléfono y accesos."
	}
	return runner, nil
}

// readNewRunner: el campo opcional «runner» ({name, phone, note}) de una petición; ok=false si ya respondió con error.
func readNewRunner(w http.ResponseWriter, f map[string]json.RawMessage) (nr newRunnerInput, ok bool) {
	if f["runner"] == nil {
		return nr, true
	}
	if json.Unmarshal(f["runner"], &nr) != nil || strings.TrimSpace(nr.Name) == "" {
		failField(w, 400, "runner", "Escribe al menos el nombre del runner nuevo.")
		return nr, false
	}
	if len(nr.Note) > userNotesMax || len(nr.Name) > 120 || len(nr.Phone) > 40 {
		failField(w, 400, "runner", "Texto demasiado largo.")
		return nr, false
	}
	return nr, true
}

// newPointInput: un destino nuevo escrito al despachar o al empezar un run (zona obligatoria).
type newPointInput struct {
	Area, Place, Directions, GPS string
	ToLocate                     bool // solo la zona, sin saber exactamente dónde: «📍 Por ubicar»
}

// newRunPoint guarda el destino nuevo de un despacho o run: publicado a nombre de quien lo crea (verificado por su
// equipo), o solo la zona («Pedasí», «El Cacao») que queda «📍 Por ubicar», no público hasta que se marque el lugar
// exacto o se una a un punto existente (locate.go).
func newRunPoint(tx *bolt.Tx, me User, np newPointInput, coords []float64, from string) (Point, error) {
	now := nowStamp()
	p := Point{ID: randomToken(), Area: strings.Join(strings.Fields(np.Area), " "), Place: strings.TrimSpace(np.Place), Directions: strings.TrimSpace(np.Directions), GPS: strings.TrimSpace(np.GPS), Coords: coords,
		Source: "user", Status: "approved", Created: now, Decided: now, AddedBy: me.ID, AddedByName: me.Name, Note: "Creado desde " + from + "."}
	if np.ToLocate {
		p.ToLocate, p.Place, p.Directions, p.GPS, p.Coords = true, "", strings.TrimSpace(np.Directions), "", nil
		p.Note = "Creado desde " + from + ", solo con la zona: por ubicar."
	}
	codeFor(tx, &p)
	verifyStamp(tx, &me, &p.VerifiedTeam, &p.VerifiedBy, &p.VerifiedAt)
	return p, putPointJSON(tx.Bucket(submissions), p)
}
