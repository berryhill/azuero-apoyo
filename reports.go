package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Reportes (📰 Reportes y «🖨️ Informe PDF» de Supply): solo para quien es de un equipo y además es team manager, tiene
// el rol «📰 Boletín» (reports) o es de Supply. Un mismo informe (necesidades abiertas, lo que va en camino y las
// entregas de un período, de un día a otro) en dos formatos que dibuja reports.js:
//   - «🧾 Sencillo»: tablas para imprimir.
//   - «📰 Boletín»: el estilo del boletín del equipo (encabezado, tarjetas, barras, prioridades). Por ahora todos usan
//     el de Unidos por Azuero; cada equipo podrá tener el suyo (Team.Template, nombre de un estilo en reports.js).
//
// Nada lleva teléfonos ni contactos.
const roleReports = "reports"

// reportTemplate: el estilo de boletín del equipo; vacío = «unidos».
func reportTemplate(t Team) string {
	if t.Template != "" {
		return t.Template
	}
	return "unidos"
}

var errNoReports = errors.New("Los reportes son para team managers, Supply o quien tiene el rol «📰 Boletín», y hay que ser de un equipo.")

// reporter: quien entró con código, con equipo y con permiso de reportes.
func (a *App) reporter(w http.ResponseWriter, r *http.Request) (User, Team, bool) {
	u, ok := a.viewer(w, r)
	if !ok {
		return u, Team{}, false
	}
	var t Team
	var found bool
	a.db.View(func(tx *bolt.Tx) error {
		if fresh, ok := getUser(tx, u.ID); ok {
			u = fresh
		}
		t, found = getTeam(tx, u.TeamID)
		return nil
	})
	if !found || !(u.has(roleTeamManager) || u.has(roleReports) || u.has(roleSupplier)) {
		fail(w, 403, errNoReports.Error())
		return u, t, false
	}
	return u, t, true
}

// canReport: para las páginas (userPanels): «📰 Reportes» la ven el team manager y el rol Boletín, con equipo.
func canReportPage(u User) bool {
	return u.TeamID != "" && (u.has(roleTeamManager) || u.has(roleReports))
}

type reportCamp struct {
	ID     string `json:"id"`
	Code   string `json:"code,omitempty"`
	Area   string `json:"area,omitempty"`
	Place  string `json:"place,omitempty"`
	People string `json:"people,omitempty"`
	Team   bool   `json:"team,omitempty"`
}
type reportMedia struct {
	Stage string `json:"stage"`
}
type reportNeed struct {
	ID               string        `json:"id"`
	Need             string        `json:"need"`
	Urgency          string        `json:"urgency,omitempty"`
	CreatedAt        string        `json:"createdAt,omitempty"`
	RevisedAt        string        `json:"revisedAt,omitempty"`
	EnRouteAt        string        `json:"enRouteAt,omitempty"`
	DispatchedAt     string        `json:"dispatchedAt,omitempty"`
	CompletedAt      string        `json:"completedAt,omitempty"`
	RunnerName       string        `json:"runnerName,omitempty"`
	DispatchedByName string        `json:"dispatchedByName,omitempty"`
	SupplyCenter     string        `json:"supplyCenter,omitempty"`
	Packing          []PackingLine `json:"packing,omitempty"`
	Media            []reportMedia `json:"media,omitempty"`
	Camp             reportCamp    `json:"camp"`
}

func parseDay(v string) (time.Time, bool) {
	t, e := time.Parse(time.RFC3339, v)
	return t, e == nil
}

func (a *App) reportRoutes(mux *http.ServeMux) {
	// Datos del «Informe de Supply»: scope=team (puntos y gente del equipo) o all; from/to (RFC3339) para las entregas.
	mux.HandleFunc("GET /api/full/reports/supply", func(w http.ResponseWriter, r *http.Request) {
		me, team, ok := a.reporter(w, r)
		if !ok {
			return
		}
		q := r.URL.Query()
		from, okF := parseDay(q.Get("from"))
		to, okT := parseDay(q.Get("to"))
		if !okF || !okT || !to.After(from) {
			fail(w, 400, "Elige el período: desde y hasta.")
			return
		}
		onlyTeam := q.Get("scope") == "team"
		out := map[string]any{"team": team.Name, "template": reportTemplate(team), "center": me.SupplyCenter, "by": me.Name, "scope": map[bool]string{true: "team", false: "all"}[onlyTeam], "from": from.UTC().Format(time.RFC3339), "to": to.UTC().Format(time.RFC3339)}
		needs, way, done := []reportNeed{}, []reportNeed{}, []reportNeed{}
		err := a.db.View(func(tx *bolt.Tx) error {
			tv := newTeamView(tx, me.ID)
			mine, e := teamCamps(tv)
			if e != nil {
				return e
			}
			points := map[string]Point{}
			tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if json.Unmarshal(v, &p) == nil && p.Status == "approved" && !hiddenFromPublic(p, tv.team) {
					points[p.ID] = p
				}
				return nil
			})
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil || isRepair(n) {
					return nil
				}
				p, ok := points[n.CampID]
				if !ok {
					return nil
				}
				ours := mine[p.ID] || tv.mate(n.RunnerID) || tv.mate(n.DispatchedBy)
				if onlyTeam && !ours {
					return nil
				}
				item := reportNeed{ID: n.ID, Need: n.Need, Urgency: n.Urgency, CreatedAt: n.CreatedAt, RevisedAt: n.RevisedAt, EnRouteAt: n.EnRouteAt, DispatchedAt: n.DispatchedAt, CompletedAt: n.CompletedAt, SupplyCenter: n.SupplyCenter, Packing: n.Packing, Camp: reportCamp{ID: p.ID, Code: p.Code, Area: p.Area, Place: p.Place, People: p.People, Team: mine[p.ID]}}
				// El nombre de quien lleva un run de otro equipo no se muestra (como en la ficha pública).
				if n.RunnerID != "" && !tv.hidden(n) {
					item.RunnerName = nameOf(tx, n.RunnerID)
				}
				item.DispatchedByName = n.DispatchedByName
				for _, m := range n.Media {
					item.Media = append(item.Media, reportMedia{Stage: m.Stage})
				}
				switch n.Status {
				case "OPEN":
					item.Packing, item.Media = nil, nil
					needs = append(needs, item)
				case "IN_PROGRESS":
					way = append(way, item)
				case "COMPLETED":
					if t, ok := parseDay(n.CompletedAt); ok && !t.Before(from) && t.Before(to) {
						done = append(done, item)
					}
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.Slice(done, func(i, j int) bool { return done[i].CompletedAt > done[j].CompletedAt })
		out["needs"], out["enroute"], out["delivered"] = needs, way, done
		response(w, 200, out)
	})
	// Para la página: equipo, formato de boletín y si puede.
	mux.HandleFunc("GET /api/full/reports", func(w http.ResponseWriter, r *http.Request) {
		me, team, ok := a.reporter(w, r)
		if !ok {
			return
		}
		response(w, 200, map[string]any{"team": team.Name, "template": reportTemplate(team), "center": me.SupplyCenter, "name": me.Name})
	})
}
