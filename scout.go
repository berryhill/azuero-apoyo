package main

import (
	"encoding/json"
	"net/http"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// Scout (scout.html): todos los puntos publicados, más los por revisar o retenidos que agregaste tú o tu equipo, cada uno
// con sus necesidades activas como en Runners. Cualquiera que entró con código puede actualizar las necesidades.
type scoutPoint struct {
	Point
	AddedByMe   bool       `json:"addedByMe,omitempty"`
	AddedByTeam bool       `json:"addedByTeam,omitempty"`
	AddedByMate string     `json:"addedByMate,omitempty"`
	PointStatus string     `json:"pointStatus"`
	ActiveNeeds []CampNeed `json:"activeNeeds"`
	// Contacts: nombres y teléfonos del punto (quien reporta y las personas del grupo), solo para quien es de un equipo.
	// Sin cédula ni edad. Sin equipo no llegan: cualquiera puede crear una cuenta y entrar a Scout.
	Contacts []scoutContact `json:"contacts,omitempty"`
}

type scoutContact struct {
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Reporter bool   `json:"reporter,omitempty"`
}

// scoutContacts: quien reporta (si tiene teléfono) y las personas con nombre o teléfono, como peopleOf en Command.
func scoutContacts(p Point) []scoutContact {
	var out []scoutContact
	if p.CoordinatorPhone != "" {
		out = append(out, scoutContact{Name: p.Coordinator, Phone: p.CoordinatorPhone, Reporter: true})
	}
	people := p.Members
	if len(people) == 0 {
		for _, c := range [][2]string{{p.Contact, p.Phone}, {p.SecondContact, p.SecondPhone}, {p.ThirdContact, p.ThirdPhone}} {
			if c[0] != "" || c[1] != "" {
				people = append(people, Member{Name: c[0], Phone: c[1]})
			}
		}
	}
	for _, m := range people {
		if m.Name != "" || m.Phone != "" {
			out = append(out, scoutContact{Name: m.Name, Phone: m.Phone})
		}
	}
	return out
}

// scoutEditable: necesidades que Scout puede cambiar: abiertas o en camino, normales (no reparaciones, comidas ni
// servicios). Quitar solo las abiertas: una en camino la lleva un runner.
func scoutEditable(n CampNeed) bool {
	return (n.Status == "OPEN" || n.Status == "IN_PROGRESS") && n.Kind == ""
}

// errScoutNoTeam: en Scout actualizan necesidades solo las personas de un equipo (cualquier punto, sin solicitud de cambio).
var errScoutNoTeam = userFieldError{"team", "Solo las personas de un equipo pueden actualizar necesidades. Pide a un team manager que te invite."}

func (a *App) scoutRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/full/scout", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		list := []scoutPoint{}
		err := a.db.View(func(tx *bolt.Tx) error {
			tv := newTeamView(tx, u.ID)
			guard := newRunnerPhoneGuard(tx, tv)
			needs := map[string][]CampNeed{}
			if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) != nil || n.Status != "OPEN" && n.Status != "IN_PROGRESS" {
					return nil
				}
				if n, keep := viewerNeed(tx, tv, u.ID, n); keep {
					needs[n.CampID] = append(needs[n.CampID], n)
				}
				return nil
			}); e != nil {
				return e
			}
			return tx.Bucket(submissions).ForEach(func(_, v []byte) error {
				var p Point
				if json.Unmarshal(v, &p) != nil || p.MergedInto != "" {
					return nil
				}
				mine := p.AddedBy != "" && p.AddedBy == u.ID
				team := !mine && tv.mate(p.AddedBy)
				status := p.Status
				switch {
				case status == "approved" && hiddenFromPublic(p, tv.team):
					return nil
				case status == "approved" && p.HeldBy != "" && p.HeldUntil > nowStamp():
					status = "held"
				case status == "pending" && (mine || team):
				case status != "approved":
					return nil
				}
				if p.Status == "approved" {
					p.CampStatus, p.HighestUrgency, _ = aggregateCamp(tx, p.ID)
				}
				sp := scoutPoint{Point: publicPoint(p), AddedByMe: mine, AddedByTeam: team, PointStatus: status, ActiveNeeds: needs[p.ID]}
				sp.HeldUntil = p.HeldUntil
				if tv.team != "" {
					sp.Contacts = guard.scoutContacts(scoutContacts(p))
				}
				if team {
					sp.AddedByMate = p.AddedByName
				}
				if sp.ActiveNeeds == nil {
					sp.ActiveNeeds = []CampNeed{}
				}
				sort.SliceStable(sp.ActiveNeeds, func(i, j int) bool { return sp.ActiveNeeds[i].CreatedAt > sp.ActiveNeeds[j].CreatedAt })
				list = append(list, sp)
				return nil
			})
		})
		if serviceError(w, err) {
			return
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].Created > list[j].Created })
		response(w, 200, list)
	})
	// Nueva necesidad para un punto publicado (como «Actualizar necesidad» al terminar un run).
	mux.HandleFunc("POST /api/full/scout/camps/{id}/needs", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 32})
		if !ok {
			return
		}
		n := CampNeed{ID: randomToken(), CampID: r.PathValue("id"), Need: fieldText(f, "need"), Status: "OPEN", Urgency: normalizeUrgency(fieldText(f, "urgency")), CreatedAt: nowStamp()}
		if n.Need == "" || !validUrgency(n.Urgency) {
			fail(w, 400, "Marca qué necesita y elige una urgencia válida.")
			return
		}
		err := a.write(r, func(tx *bolt.Tx) error {
			if teamOf(tx, u.ID) == "" {
				return errScoutNoTeam
			}
			if !a.campExists(tx, n.CampID) || campHeldFrom(tx, n.CampID, teamOf(tx, u.ID)) {
				return errMissing
			}
			verifyStamp(tx, &u, &n.VerifiedTeam, &n.VerifiedBy, &n.VerifiedAt)
			n.UpdatedBy = u.ID
			var e error
			n, e = putNeedSplit(tx, n)
			return e
		})
		if runError(w, err) {
			return
		}
		response(w, 201, map[string]string{"id": n.ID, "status": n.Status})
	})
	// Cambiar lo que pide una necesidad o su urgencia, o quitarla si está abierta («Ya no lo necesitan»: status DELETED).
	// Quien es de un equipo actualiza cualquier punto directamente; todo queda en el historial con su nombre y «Scout».
	mux.HandleFunc("PATCH /api/full/scout/needs/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.viewer(w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"need": 8192, "urgency": 32, "status": 32})
		if !ok {
			return
		}
		if _, yes := f["status"]; yes && fieldText(f, "status") != "DELETED" {
			failField(w, 400, "status", "Solo se puede quitar la necesidad.")
			return
		}
		if _, yes := f["need"]; yes && fieldText(f, "need") == "" {
			failField(w, 400, "need", "Marca qué necesita, o quítala.")
			return
		}
		id := r.PathValue("id")
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			if teamOf(tx, u.ID) == "" {
				return errScoutNoTeam
			}
			v := tx.Bucket(needsBucket).Get([]byte(id))
			if v == nil || json.Unmarshal(v, &n) != nil || campHeldFrom(tx, n.CampID, teamOf(tx, u.ID)) || !a.campExists(tx, n.CampID) {
				return errMissing
			}
			if !scoutEditable(n) {
				return userFieldError{"status", "Esta necesidad no se cambia desde aquí."}
			}
			if fieldText(f, "status") == "DELETED" && n.Status != "OPEN" {
				return userFieldError{"status", "Un runner ya la lleva: no se puede quitar mientras va en camino."}
			}
			var e error
			n, e = a.patchNeed(tx, id, f, &u, "")
			return e
		})
		if runError(w, err) {
			return
		}
		response(w, 200, map[string]string{"id": n.ID, "status": n.Status})
	})
}
