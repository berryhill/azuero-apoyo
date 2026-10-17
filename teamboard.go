package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// «Mi equipo» → 📊 Actividad: lo que hizo el equipo (sus runners y quienes despachan), para todos sus miembros.
type teamEvent struct {
	Kind  string `json:"kind"` // delivered | enroute | dispatched
	Who   string `json:"who"`
	At    string `json:"at"`
	Need  string `json:"need,omitempty"`
	Code  string `json:"code,omitempty"`
	Place string `json:"place,omitempty"`
	Area  string `json:"area,omitempty"`
}

type teamBoard struct {
	Members      int         `json:"members"`
	Online       int         `json:"online"`
	Active24     int         `json:"active24"`
	EnRoute      int         `json:"enRoute"`
	Delivered7   int         `json:"delivered7"`
	Dispatched7  int         `json:"dispatched7"`
	DeliveredAll int         `json:"deliveredAll"`
	Recent       []teamEvent `json:"recent"`
}

const teamRecentMax = 25

func teamActivity(tx *bolt.Tx, ids map[string]string, now time.Time, board *teamBoard) {
	board.Recent = []teamEvent{}
	within := func(at string, d time.Duration) bool {
		t, e := time.Parse(time.RFC3339, at)
		return e == nil && now.Sub(t) < d
	}
	points := tx.Bucket(submissions)
	tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if json.Unmarshal(v, &n) != nil || isRepair(n) {
			return nil
		}
		runnerName, runner := ids[n.RunnerID]
		_, dispatcher := ids[n.DispatchedBy]
		if !runner && !dispatcher {
			return nil
		}
		var p Point
		json.Unmarshal(points.Get([]byte(n.CampID)), &p)
		ev := func(kind, who, at string) {
			if within(at, 30*24*time.Hour) {
				board.Recent = append(board.Recent, teamEvent{Kind: kind, Who: who, At: at, Need: n.Need, Code: p.Code, Place: p.Place, Area: p.Area})
			}
		}
		if runner {
			switch n.Status {
			case "IN_PROGRESS":
				board.EnRoute++
				ev("enroute", runnerName, n.EnRouteAt)
			case "COMPLETED":
				board.DeliveredAll++
				if within(n.CompletedAt, 7*24*time.Hour) {
					board.Delivered7++
				}
				ev("delivered", runnerName, n.CompletedAt)
			}
		}
		if dispatcher && n.DispatchedAt != "" {
			if within(n.DispatchedAt, 7*24*time.Hour) {
				board.Dispatched7++
			}
			who := ids[n.DispatchedBy]
			if who == "" {
				who = n.DispatchedByName
			}
			ev("dispatched", who, n.DispatchedAt)
		}
		return nil
	})
	sort.Slice(board.Recent, func(i, j int) bool { return board.Recent[i].At > board.Recent[j].At })
	if len(board.Recent) > teamRecentMax {
		board.Recent = board.Recent[:teamRecentMax]
	}
}

// Mapa del equipo: dónde está cada compañero ahora (ubicación en vivo, menos de 2 minutos) o su última posición de las
// últimas 24 horas. Solo los del mismo equipo; quien apagó su ubicación ya no aparece en vivo (dropLocation), pero su
// última posición del recorrido sí, con la hora.
const teamLastSeen = 24 * time.Hour

// Time-lapse del mapa del equipo: hasta 7 días atrás (el recorrido completo, 30 días, solo lo ve administración).
const teamTrackKeep = 7 * 24 * time.Hour

type teamMapPin struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy float64 `json:"accuracy,omitempty"`
	At       string  `json:"at"`
	Live     bool    `json:"live,omitempty"`
}

func (a *App) teamBoardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/full/team/map", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		now := time.Now()
		a.mu.Lock()
		live := map[string]liveLocation{}
		for id, loc := range a.locations {
			if now.Sub(loc.At) <= locationTTL {
				live[id] = loc
			}
		}
		a.mu.Unlock()
		pins := []teamMapPin{}
		inTeam := false
		a.db.View(func(tx *bolt.Tx) error {
			team, found := getTeam(tx, me.TeamID)
			if !found {
				return nil
			}
			inTeam = true
			return eachUser(tx, func(u User) error {
				if u.TeamID != team.ID || u.Disabled {
					return nil
				}
				if loc, ok := live[u.ID]; ok {
					pins = append(pins, teamMapPin{ID: u.ID, Name: u.Name, Lat: loc.Lat, Lon: loc.Lon, Accuracy: loc.Accuracy, At: loc.At.UTC().Format(time.RFC3339), Live: true})
					return nil
				}
				if root := tx.Bucket(trackBucket); root != nil {
					if b := root.Bucket([]byte(u.ID)); b != nil {
						if _, v := b.Cursor().Last(); v != nil {
							var p trackPoint
							if json.Unmarshal(v, &p) == nil && now.Sub(p.At) < teamLastSeen {
								pins = append(pins, teamMapPin{ID: u.ID, Name: u.Name, Lat: p.Lat, Lon: p.Lon, Accuracy: p.Accuracy, At: p.At.UTC().Format(time.RFC3339)})
							}
						}
					}
				}
				return nil
			})
		})
		if !inTeam {
			fail(w, 403, "No eres parte de un equipo.")
			return
		}
		sort.Slice(pins, func(i, j int) bool { return pins[i].Name < pins[j].Name })
		response(w, 200, pins)
	})
	// Recorrido de un compañero (time-lapse del mapa del equipo): sus posiciones entre from y to (RFC3339), solo de los
	// últimos teamTrackKeep días y solo si es de tu mismo equipo.
	mux.HandleFunc("GET /api/full/team/members/{id}/track", func(w http.ResponseWriter, r *http.Request) {
		me, ok := a.viewer(w, r)
		if !ok {
			return
		}
		now := time.Now()
		from, e1 := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
		to, e2 := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
		if e1 != nil || e2 != nil || !to.After(from) || to.Sub(from) > 26*time.Hour {
			fail(w, 400, "Elige un día.")
			return
		}
		if oldest := now.Add(-teamTrackKeep); from.Before(oldest) {
			from = oldest
		}
		var list []trackPoint
		allowed := false
		a.db.View(func(tx *bolt.Tx) error {
			mine, okMe := getUser(tx, me.ID)
			them, okThem := getUser(tx, r.PathValue("id"))
			if !okMe || !okThem || them.Disabled || mine.TeamID == "" || mine.TeamID != them.TeamID {
				return nil
			}
			allowed = true
			if to.After(from) {
				list, _ = userTrack(tx, them.ID, from, to)
			}
			return nil
		})
		if !allowed {
			fail(w, 404, "Esa persona no está en tu equipo.")
			return
		}
		if list == nil {
			list = []trackPoint{}
		}
		response(w, 200, list)
	})
	// El team manager cambia el nombre de su equipo.
	mux.HandleFunc("PATCH /api/full/team", func(w http.ResponseWriter, r *http.Request) {
		_, team, ok := a.teamManager(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"name": 80})
		if !ok {
			return
		}
		name := strings.Join(strings.Fields(fieldText(f, "name")), " ")
		if name == "" {
			failField(w, 400, "name", "Escribe el nombre del equipo.")
			return
		}
		err := a.write(r, func(tx *bolt.Tx) error {
			t, found := getTeam(tx, team.ID)
			if !found {
				return errMissing
			}
			t.Name, t.UpdatedAt = name, nowStamp()
			team = t
			return putJSON(tx.Bucket(teamsBucket), t.ID, t)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"team": team})
	})
}
