package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Team manager: alguien de un equipo (requisito: tener equipo) que invita gente a su equipo con enlaces de un solo uso
// que vencen. El enlace (/signup.html?invite=…) se acepta solo con sesión de código: así el usuario que entra queda
// ligado al equipo por su ID. Se guarda el hash del token, nunca el token.
const roleTeamManager = "teammanager"

var teamInvitesBucket = []byte("team_invites")

type TeamInvite struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	TeamID    string `json:"teamId"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
	UsedBy    string `json:"usedBy,omitempty"`
	UsedAt    string `json:"usedAt,omitempty"`
	Revoked   bool   `json:"revoked,omitempty"`
}

// teamInviteView: la invitación como la ve el team manager (sin token ni hash).
type teamInviteView struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"createdAt"`
	ExpiresAt  string `json:"expiresAt"`
	CreatedBy  string `json:"createdByName,omitempty"`
	UsedByName string `json:"usedByName,omitempty"`
	UsedAt     string `json:"usedAt,omitempty"`
	Status     string `json:"status"` // pending | used | expired | revoked
}

// Roles que un team manager puede dar o quitar en su equipo; agente y donaciones (datos privados) solo administración.
var teamManagerRoles = []string{roleRunner, roleBuilder, roleSupplier, roleTeamManager}

func teamManagerCan(role string) bool {
	for _, r := range teamManagerRoles {
		if r == role {
			return true
		}
	}
	return false
}

var errInviteInvalid = errors.New("Este enlace de invitación no sirve: ya se usó, venció o lo anularon. Pide uno nuevo.")
var errNotTeamManager = errors.New("Solo un team manager con equipo puede invitar.")

func ensureTeamInvites(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists(teamInvitesBucket)
	return err
}

func inviteHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func (inv TeamInvite) status(now time.Time) string {
	switch {
	case inv.Revoked:
		return "revoked"
	case inv.UsedBy != "":
		return "used"
	}
	if t, e := time.Parse(time.RFC3339, inv.ExpiresAt); e != nil || now.After(t) {
		return "expired"
	}
	return "pending"
}

// manager: el usuario de la sesión, si es team manager y su equipo existe.
func (a *App) teamManager(w http.ResponseWriter, r *http.Request) (User, Team, bool) {
	u, ok := a.viewer(w, r)
	if !ok {
		return u, Team{}, false
	}
	var t Team
	var found bool
	a.db.View(func(tx *bolt.Tx) error { t, found = getTeam(tx, u.TeamID); return nil })
	if !u.has(roleTeamManager) || !found {
		fail(w, 403, errNotTeamManager.Error())
		return u, t, false
	}
	return u, t, true
}

func (a *App) teamInviteRoutes(mux *http.ServeMux) {
	// «Mi equipo»: miembros e invitaciones del equipo del team manager.
	mux.HandleFunc("GET /api/full/team", func(w http.ResponseWriter, r *http.Request) {
		_, team, ok := a.teamManager(w, r)
		if !ok {
			return
		}
		type member struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Telephone    string   `json:"telephone,omitempty"`
			Roles        []string `json:"roles"`
			LastActiveAt string   `json:"lastActiveAt,omitempty"`
			CreatedAt    string   `json:"createdAt,omitempty"`
			Online       bool     `json:"online,omitempty"`
		}
		members, invites := []member{}, []teamInviteView{}
		now := time.Now()
		err := a.db.View(func(tx *bolt.Tx) error {
			if e := eachUser(tx, func(u User) error {
				if u.TeamID == team.ID && !u.Disabled {
					members = append(members, member{ID: u.ID, Name: u.Name, Telephone: u.Telephone, Roles: u.Roles, LastActiveAt: u.LastActiveAt, CreatedAt: u.CreatedAt, Online: a.withPresence(u.admin()).Online})
				}
				return nil
			}); e != nil {
				return e
			}
			name := func(id string) string {
				if u, ok := getUser(tx, id); ok {
					return u.Name
				}
				return ""
			}
			return tx.Bucket(teamInvitesBucket).ForEach(func(_, v []byte) error {
				var inv TeamInvite
				if e := json.Unmarshal(v, &inv); e != nil {
					return e
				}
				if inv.TeamID != team.ID {
					return nil
				}
				invites = append(invites, teamInviteView{ID: inv.ID, CreatedAt: inv.CreatedAt, ExpiresAt: inv.ExpiresAt, CreatedBy: name(inv.CreatedBy), UsedByName: name(inv.UsedBy), UsedAt: inv.UsedAt, Status: inv.status(now)})
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.Slice(members, func(i, j int) bool { return runnerNameKey(members[i].Name) < runnerNameKey(members[j].Name) })
		sort.Slice(invites, func(i, j int) bool { return invites[i].CreatedAt > invites[j].CreatedAt })
		response(w, 200, map[string]any{"team": team, "members": members, "invites": invites, "grantable": teamManagerRoles})
	})
	// Nueva invitación de un solo uso; vence en hours horas (24, 72 o 168). El token se devuelve una vez.
	mux.HandleFunc("POST /api/full/team/invites", func(w http.ResponseWriter, r *http.Request) {
		u, team, ok := a.teamManager(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"hours": 0})
		if !ok {
			return
		}
		hours := 72
		if raw, has := f["hours"]; has {
			if json.Unmarshal(raw, &hours) != nil || (hours != 24 && hours != 72 && hours != 168) {
				failField(w, 400, "hours", "Elige 24 horas, 3 días o 7 días.")
				return
			}
		}
		token := randomToken()
		now := time.Now().UTC()
		inv := TeamInvite{ID: randomToken()[:16], Hash: inviteHash(token), TeamID: team.ID, CreatedBy: u.ID, CreatedAt: now.Format(time.RFC3339), ExpiresAt: now.Add(time.Duration(hours) * time.Hour).Format(time.RFC3339)}
		err := a.db.Update(func(tx *bolt.Tx) error { return putJSON(tx.Bucket(teamInvitesBucket), inv.Hash, inv) })
		if serviceError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": inv.ID, "token": token, "expiresAt": inv.ExpiresAt, "team": team.Name})
	})
	mux.HandleFunc("POST /api/full/team/invites/{id}/revoke", func(w http.ResponseWriter, r *http.Request) {
		_, team, ok := a.teamManager(w, r)
		if !ok {
			return
		}
		if _, ok := decodeFields(w, r, map[string]int{}); !ok {
			return
		}
		found := false
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(teamInvitesBucket)
			return b.ForEach(func(k, v []byte) error {
				var inv TeamInvite
				if e := json.Unmarshal(v, &inv); e != nil {
					return e
				}
				if inv.ID != r.PathValue("id") || inv.TeamID != team.ID {
					return nil
				}
				found = true
				if inv.UsedBy == "" {
					inv.Revoked = true
					return putJSON(b, string(k), inv)
				}
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		if !found {
			fail(w, 404, "Invitación no encontrada.")
			return
		}
		response(w, 200, map[string]bool{"ok": true})
	})
	// Cambiar roles de un miembro del equipo (solo los de teamManagerRoles; los demás quedan como estaban) o sacarlo del
	// equipo. No sobre sí mismo. Si pierde un rol, se cierran sus sesiones, como en administración.
	memberChange := func(w http.ResponseWriter, r *http.Request, apply func(m *User) error) {
		me, team, ok := a.teamManager(w, r)
		if !ok {
			return
		}
		var m User
		var before []string
		err := a.db.Update(func(tx *bolt.Tx) error {
			var found bool
			if m, found = getUser(tx, r.PathValue("id")); !found || m.TeamID != team.ID {
				return errMissing
			}
			if m.ID == me.ID {
				return userFieldError{"id", "No puedes cambiarte a ti mismo; pídeselo a administración."}
			}
			before = append([]string(nil), m.Roles...)
			if e := apply(&m); e != nil {
				return e
			}
			m.UpdatedAt = nowStamp()
			return putJSON(tx.Bucket(usersBucket), m.ID, m)
		})
		if errors.Is(err, errMissing) {
			fail(w, 404, "Esa persona no está en tu equipo.")
			return
		}
		if userError(w, err) {
			return
		}
		for _, role := range before {
			if !m.has(role) {
				a.dropCrewSessions(m.ID)
				break
			}
		}
		response(w, 200, map[string]any{"id": m.ID, "roles": m.Roles, "teamId": m.TeamID})
	}
	mux.HandleFunc("PATCH /api/full/team/members/{id}", func(w http.ResponseWriter, r *http.Request) {
		f, ok := decodeFields(w, r, map[string]int{"roles": 0})
		if !ok {
			return
		}
		var want []string
		if json.Unmarshal(f["roles"], &want) != nil {
			failField(w, 400, "roles", "Roles inválidos.")
			return
		}
		memberChange(w, r, func(m *User) error {
			roles := []string{}
			for _, role := range m.Roles {
				if !teamManagerCan(role) {
					roles = append(roles, role) // agente / donaciones: solo administración
				}
			}
			for _, role := range want {
				if !teamManagerCan(role) {
					return userFieldError{"roles", "Ese rol solo lo asigna administración."}
				}
				if !slicesContains(roles, role) {
					roles = append(roles, role)
				}
			}
			if len(roles) == 0 {
				return userFieldError{"roles", "Deja al menos un rol."}
			}
			m.Roles = roles
			return nil
		})
	})
	mux.HandleFunc("POST /api/full/team/members/{id}/remove", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := decodeFields(w, r, map[string]int{}); !ok {
			return
		}
		memberChange(w, r, func(m *User) error {
			m.TeamID = ""
			// Team manager necesita equipo: al salir pierde ese rol.
			roles := []string{}
			for _, role := range m.Roles {
				if role != roleTeamManager {
					roles = append(roles, role)
				}
			}
			m.Roles = roles
			return nil
		})
	})
	// Aceptar: solo con sesión de código; el usuario queda en el equipo de la invitación (cambia de equipo si tenía otro).
	mux.HandleFunc("POST /api/full/invites/accept", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"token": 128})
		if !ok {
			return
		}
		var team Team
		already := false
		err := a.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(teamInvitesBucket)
			key := inviteHash(fieldText(f, "token"))
			v := b.Get([]byte(key))
			if v == nil {
				return errInviteInvalid
			}
			var inv TeamInvite
			if e := json.Unmarshal(v, &inv); e != nil {
				return e
			}
			var found bool
			if team, found = getTeam(tx, inv.TeamID); !found {
				return errInviteInvalid
			}
			if inv.UsedBy == u.ID && u.TeamID == team.ID {
				already = true
				return nil
			}
			if inv.status(time.Now()) != "pending" {
				return errInviteInvalid
			}
			me, found := getUser(tx, u.ID)
			if !found {
				return errMissing
			}
			already = me.TeamID == team.ID
			me.TeamID, me.UpdatedAt = team.ID, nowStamp()
			inv.UsedBy, inv.UsedAt = me.ID, nowStamp()
			if e := putJSON(b, key, inv); e != nil {
				return e
			}
			return putJSON(tx.Bucket(usersBucket), me.ID, me)
		})
		if errors.Is(err, errInviteInvalid) {
			fail(w, 410, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"team": team.Name, "already": already})
	})
}

func slicesContains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
