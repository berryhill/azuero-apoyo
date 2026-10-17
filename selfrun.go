package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// «＋ Nuevo run» (Runners): el runner arma su propio run en un solo paso. Destino: un punto publicado (y, si quiere,
// una de sus necesidades abiertas sin runner, que pasa a ser suya) o un lugar nuevo (zona obligatoria; lugar,
// indicaciones, enlace de Maps o GPS opcionales) que queda publicado a su nombre, o solo la zona («📍 Por ubicar»).
// Lleva su lista de empaque (obligatoria); las fotos de lo que lleva se suben después con /api/runner/needs/{id}/media.
// Queda en camino, en su «Yo», con el mismo recorrido de siempre (Maps, «📍 Llegué», completar con fotos y necesidades).

// runDestination: un punto para elegir destino, con sus necesidades abiertas o en camino (sin contactos ni teléfonos).
type runDestination struct {
	ID             string               `json:"id"`
	Code           string               `json:"code"`
	Area           string               `json:"area"`
	Place          string               `json:"place,omitempty"`
	Directions     string               `json:"directions,omitempty"`
	GPS            string               `json:"gps,omitempty"`
	Coords         []float64            `json:"coords,omitempty"`
	HighestUrgency string               `json:"highestUrgency,omitempty"`
	Needs          []runDestinationNeed `json:"needs"`
}

type runDestinationNeed struct {
	ID       string `json:"id"`
	Need     string `json:"need"`
	Urgency  string `json:"urgency,omitempty"`
	Status   string `json:"status"`
	Takeable bool   `json:"takeable,omitempty"`
	Mine     bool   `json:"mine,omitempty"`
}

func (a *App) selfRunRoutes(mux *http.ServeMux, c *crew) {
	mux.HandleFunc("GET /api/runner/destinations", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(c, w, r)
		if !ok {
			return
		}
		list := []runDestination{}
		err := a.db.View(func(tx *bolt.Tx) error {
			team := newTeamView(tx, s.runnerID).team
			needs := map[string][]runDestinationNeed{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil || n.Status != "OPEN" && n.Status != "IN_PROGRESS" || isRepair(n) || isMealKind(n) || n.Kind != "" {
					return nil
				}
				needs[n.CampID] = append(needs[n.CampID], runDestinationNeed{ID: n.ID, Need: n.Need, Urgency: n.Urgency, Status: n.Status,
					Takeable: n.Status == "OPEN" && n.RunnerID == "", Mine: n.RunnerID == s.runnerID})
				return nil
			}); e != nil {
				return e
			}
			return tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if json.Unmarshal(v, &p) != nil || p.MergedInto != "" || p.Status != "approved" || hiddenFromPublic(p, team) {
					return nil
				}
				_, urgency, _ := aggregateCamp(tx, p.ID)
				d := runDestination{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, Directions: p.Directions, GPS: p.GPS, Coords: p.Coords, HighestUrgency: urgency, Needs: needs[p.ID]}
				if d.Needs == nil {
					d.Needs = []runDestinationNeed{}
				}
				list = append(list, d)
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(list, func(i, j int) bool {
			return strings.ToLower(list[i].Area+list[i].Place) < strings.ToLower(list[j].Area+list[j].Place)
		})
		response(w, 200, list)
	})
	// Vista previa de un enlace de Maps (también los cortos): sus coordenadas, para verlo junto a los puntos que ya hay.
	mux.HandleFunc("GET /api/runner/maplink", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.crewSession(c, w, r); !ok {
			return
		}
		link := strings.TrimSpace(r.URL.Query().Get("url"))
		if link == "" || len(link) > 1000 {
			failField(w, 400, "gps", "Pega un enlace de Google Maps.")
			return
		}
		coords, err := parseMapLink(link)
		if err != nil || len(coords) != 2 {
			failField(w, 400, "gps", "No se pudo leer la ubicación de ese enlace.")
			return
		}
		response(w, 200, map[string]any{"coords": coords})
	})
	mux.HandleFunc("POST /api/runner/runs", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(c, w, r)
		if !ok {
			return
		}
		if s.agent {
			fail(w, 403, errAgentNoRuns.Error())
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"pointId": 128, "point": 0, "needId": 128, "items": 0, "need": 400})
		if !ok {
			return
		}
		var newPoint newPointInput
		pointID, needID := fieldText(f, "pointId"), fieldText(f, "needId")
		if pointID == "" && (f["point"] == nil || json.Unmarshal(f["point"], &newPoint) != nil || strings.TrimSpace(newPoint.Area) == "") {
			failField(w, 400, "point", "Elige el destino o escribe al menos la zona del lugar nuevo.")
			return
		}
		if len(newPoint.Directions) > 8192 || len(newPoint.Place) > 200 || len(newPoint.Area) > 80 || len(newPoint.GPS) > 1000 {
			failField(w, 400, "point", "Texto demasiado largo.")
			return
		}
		if needID != "" && pointID == "" {
			failField(w, 400, "needId", "Esa necesidad es de un punto que ya existe: elige ese punto.")
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
			failField(w, 400, "items", "Agrega al menos un artículo a lo que llevas.")
			return
		}
		var coords []float64
		if gps := strings.TrimSpace(newPoint.GPS); pointID == "" && gps != "" && !newPoint.ToLocate {
			cs, e := parseMapLink(gps) // fuera de la transacción: puede seguir un enlace corto
			if e != nil {
				failField(w, 400, "gps", "Usa un enlace de Google Maps del lugar.")
				return
			}
			coords = cs
		}
		need := strings.TrimSpace(fieldText(f, "need"))
		if need == "" {
			need = "Run propio"
		}
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			me, found := getUser(tx, s.runnerID)
			if !found || me.Disabled {
				return errMissing
			}
			b := tx.Bucket(needsBucket)
			active := map[string]bool{}
			if e := b.ForEach(func(_, v []byte) error {
				var other CampNeed
				if e := json.Unmarshal(v, &other); e != nil {
					return e
				}
				if other.RunnerID == me.ID && other.Status == "IN_PROGRESS" && !isRepair(other) {
					active[other.ID] = true
				}
				return nil
			}); e != nil {
				return e
			}
			if len(active) >= c.max {
				return c.busy
			}
			if pointID != "" {
				if !a.campExists(tx, pointID) {
					return userFieldError{"pointId", "Ese punto ya no está publicado. Elige otro o crea el destino."}
				}
			} else {
				p, e := newRunPoint(tx, me, newPoint, coords, "Runners al empezar un run")
				if e != nil {
					return e
				}
				pointID = p.ID
			}
			now := nowStamp()
			if needID != "" {
				v := b.Get([]byte(needID))
				if v == nil {
					return userFieldError{"needId", "Esa necesidad ya no está."}
				}
				if e := json.Unmarshal(v, &n); e != nil {
					return e
				}
				if n.CampID != pointID || n.Status != "OPEN" || n.RunnerID != "" || isRepair(n) || isMealKind(n) || n.Kind != "" {
					return userFieldError{"needId", "Esa necesidad ya la lleva alguien o ya no está abierta. Elige otra o haz tu propio run."}
				}
				n.RunnerID, n.Status, n.EnRouteAt, n.UpdatedBy = me.ID, "IN_PROGRESS", now, me.ID
			} else {
				n = CampNeed{ID: randomToken(), CampID: pointID, Need: need, Status: "IN_PROGRESS", Urgency: "normal", RunnerID: me.ID,
					CreatedAt: now, EnRouteAt: now, CreatedBy: me.ID, CreatedByName: me.Name, UpdatedBy: me.ID}
				verifyStamp(tx, &me, &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			}
			if e := putNeed(b, &n); e != nil {
				return e
			}
			if e := addPacking(tx, &n, lines, me.ID, me.Name); e != nil {
				return e
			}
			// Al final de su orden, como si la hubiera tomado (se descartan las que ya no están en curso).
			order := []string{}
			for _, id := range me.RunOrder {
				if active[id] {
					order = append(order, id)
				}
			}
			me.RunOrder = append(order, n.ID)
			return putJSON(tx.Bucket(usersBucket), me.ID, me)
		})
		if errors.Is(err, c.busy) {
			fail(w, 409, err.Error())
			return
		}
		if userError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": n.ID, "pointId": pointID})
	})
}
