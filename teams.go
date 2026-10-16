package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Equipos: administración agrupa a los usuarios con código de acceso. Lo que antes veía «cualquiera con código»
// (carreras reservadas, «Solo Nuestro Equipo», ubicación en vivo, asignar o empacar una carrera) ahora lo ve
// solo quien está en el mismo equipo que quien agregó el punto o lleva la carrera.
var teamsBucket = []byte("teams")

type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// adminTeam es lo que ve el panel: el equipo con la cantidad de miembros.
type adminTeam struct {
	Team
	Members int `json:"members"`
}

const defaultTeamName = "Equipo principal"

func getTeam(tx *bolt.Tx, id string) (Team, bool) {
	var t Team
	if id == "" {
		return t, false
	}
	v := tx.Bucket(teamsBucket).Get([]byte(id))
	return t, v != nil && json.Unmarshal(v, &t) == nil
}

// migrateTeams crea el registro de equipos. La primera vez, si ya hay usuarios con código, los pone a todos en
// «Equipo principal» para que sigan viéndose entre sí como antes de que existieran los equipos.
func migrateTeams(tx *bolt.Tx) error {
	if tx.Bucket(teamsBucket) != nil {
		return nil
	}
	b, err := tx.CreateBucket(teamsBucket)
	if err != nil {
		return err
	}
	coded := []User{}
	if e := eachUser(tx, func(u User) error {
		if u.CodeHash != "" && u.TeamID == "" {
			coded = append(coded, u)
		}
		return nil
	}); e != nil || len(coded) == 0 {
		return e
	}
	t := Team{ID: randomToken(), Name: defaultTeamName, CreatedAt: nowStamp()}
	t.UpdatedAt = t.CreatedAt
	if e := putJSON(b, t.ID, t); e != nil {
		return e
	}
	for _, u := range coded {
		u.TeamID = t.ID
		if e := putJSON(tx.Bucket(usersBucket), u.ID, u); e != nil {
			return e
		}
	}
	return nil
}

// teamOf: el equipo de un usuario activo con código; vacío si no tiene (o el equipo ya no existe).
func teamOf(tx *bolt.Tx, id string) string {
	u, ok := getUser(tx, id)
	if !ok || u.Disabled || u.CodeHash == "" {
		return ""
	}
	if _, ok := getTeam(tx, u.TeamID); !ok {
		return ""
	}
	return u.TeamID
}

// verifyStamp anota «verificado por equipo» cuando quien crea o actualiza (u) es de un equipo: nombre del equipo,
// persona y hora. Si u no es de ningún equipo, lo quita: el dato cambió sin una revisión de equipo. u nil
// (administración con la contraseña maestra) no lo cambia.
func verifyStamp(tx *bolt.Tx, u *User, team, by, at *string) {
	if u == nil {
		return
	}
	if t, ok := getTeam(tx, teamOf(tx, u.ID)); ok {
		*team, *by, *at = t.Name, u.Name, nowStamp()
		return
	}
	*team, *by, *at = "", "", ""
}

// teamView responde, para quien mira (me), si otra persona es de su equipo. me vacío es el público: nadie lo es.
type teamView struct {
	tx       *bolt.Tx
	me, team string
	seen     map[string]bool
}

func newTeamView(tx *bolt.Tx, me string) *teamView {
	t := &teamView{tx: tx, me: me, seen: map[string]bool{}}
	if me != "" {
		t.team = teamOf(tx, me)
	}
	return t
}

// mate: es quien mira, o es de su mismo equipo.
func (t *teamView) mate(id string) bool {
	if id == "" || t.me == "" {
		return false
	}
	if id == t.me {
		return true
	}
	if t.team == "" {
		return false
	}
	ok, seen := t.seen[id]
	if !seen {
		ok = teamOf(t.tx, id) == t.team
		t.seen[id] = ok
	}
	return ok
}

// hidden: la carrera la lleva un corredor con código de otro equipo; quien mira ve solo el estado y la urgencia.
func (t *teamView) hidden(n CampNeed) bool {
	return reservedNeed(t.tx, n) && !t.mate(n.RunnerID)
}

// teamCamps: puntos que agregó alguien del equipo de quien mira o con una necesidad de suministros asignada
// a alguien de su equipo («Solo Nuestro Equipo»).
func teamCamps(t *teamView) (map[string]bool, error) {
	team := map[string]bool{}
	if t.me == "" {
		return team, nil
	}
	if e := t.tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if e := json.Unmarshal(v, &p); e != nil {
			return e
		}
		if t.mate(p.AddedBy) {
			team[p.ID] = true
		}
		return nil
	}); e != nil {
		return nil, e
	}
	err := t.tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if !team[n.CampID] && n.Status != "DELETED" && !isRepair(n) && t.mate(n.RunnerID) {
			team[n.CampID] = true
		}
		return nil
	})
	return team, err
}

func (a *App) teamRoutes(mux *http.ServeMux) {
	fields := map[string]int{"name": 80}
	name := func(w http.ResponseWriter, f map[string]json.RawMessage) (string, bool) {
		n := strings.Join(strings.Fields(fieldText(f, "name")), " ")
		if n == "" {
			failField(w, 400, "name", "Escribe el nombre del equipo.")
		}
		return n, n != ""
	}
	mux.HandleFunc("GET /api/admin/teams", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		list := []adminTeam{}
		err := a.db.View(func(tx *bolt.Tx) error {
			count := map[string]int{}
			if e := eachUser(tx, func(u User) error { count[u.TeamID]++; return nil }); e != nil {
				return e
			}
			return tx.Bucket(teamsBucket).ForEach(func(_, v []byte) error {
				var t Team
				if e := json.Unmarshal(v, &t); e != nil {
					return e
				}
				list = append(list, adminTeam{Team: t, Members: count[t.ID]})
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(list, func(i, j int) bool { return foldText(list[i].Name) < foldText(list[j].Name) })
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/admin/teams", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, fields)
		if !ok {
			return
		}
		t := Team{ID: randomToken(), CreatedAt: nowStamp()}
		if t.Name, ok = name(w, f); !ok {
			return
		}
		t.UpdatedAt = t.CreatedAt
		if serviceError(w, a.write(r, func(tx *bolt.Tx) error { return putJSON(tx.Bucket(teamsBucket), t.ID, t) })) {
			return
		}
		response(w, 201, adminTeam{Team: t})
	})
	mux.HandleFunc("PATCH /api/admin/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, fields)
		if !ok {
			return
		}
		n, ok := name(w, f)
		if !ok {
			return
		}
		var t Team
		err := a.write(r, func(tx *bolt.Tx) error {
			var found bool
			if t, found = getTeam(tx, r.PathValue("id")); !found {
				return errMissing
			}
			t.Name, t.UpdatedAt = n, nowStamp()
			return putJSON(tx.Bucket(teamsBucket), t.ID, t)
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, adminTeam{Team: t})
	})
}
