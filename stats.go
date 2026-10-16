package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Resumen público (pestaña «Resumen» del mapa): solo conteos de las necesidades de suministros de puntos publicados.
// Sin nombres, teléfonos ni conductores; de las necesidades reservadas a corredores con código se cuenta el estado
// pero no sus artículos, igual que en la lista pública. Las reparaciones nunca se publican.
type statItem struct {
	Name      string `json:"name"`
	Active    int    `json:"active"`
	Completed int    `json:"completed"`
}

// statCover: para cada categoría (agua, comida, refugio, saneamiento, bebé) cuántos puntos la piden y cuántas personas hay en ellos.
type statCover struct {
	Points      int `json:"points"`
	People      int `json:"people"`
	Babies      int `json:"babies"`
	Children    int `json:"children"`
	OlderAdults int `json:"olderAdults"`
}
type statZone struct {
	Area       string `json:"area"`
	Points     int    `json:"points"`
	Needing    int    `json:"needing"`
	People     int    `json:"people"`
	FoodPeople int    `json:"foodPeople"`
}
type statMaterial struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Jobs     int    `json:"jobs"`
}
type publicStats struct {
	Open         int        `json:"open"`
	InProgress   int        `json:"inProgress"`
	Completed    int        `json:"completed"`
	Urgent       int        `json:"urgent"`
	Completed24h int        `json:"completed24h"`
	Completed7d  int        `json:"completed7d"`
	Reported24h  int        `json:"reported24h"`
	Items        []statItem `json:"items"`
	// Reparaciones de casa en curso o por tomar: cuántas, cuántas tienen materiales calculados y el total de materiales.
	RepairJobs      int            `json:"repairJobs"`
	RepairCosted    int            `json:"repairCosted"`
	RepairMaterials []statMaterial `json:"repairMaterials"`
	// Para «¿Qué donar?»: personas por categoría de necesidad activa (cada punto cuenta una vez por categoría) y por zona.
	Cover map[string]*statCover `json:"cover"`
	Zones []statZone            `json:"zones"`
	// Donaciones ofrecidas (por confirmar), comprometidas y entregadas por artículo; solo totales.
	Pledges     []statPledge `json:"pledges"`
	GeneratedAt string       `json:"generatedAt"`
}

func computePublicStats(tx *bolt.Tx, now time.Time) (publicStats, error) {
	st := publicStats{Items: []statItem{}, RepairMaterials: []statMaterial{}, GeneratedAt: now.UTC().Format(time.RFC3339)}
	mats, matOrder := map[string]*statMaterial{}, []string{}
	published, pts := map[string]bool{}, map[string]Point{}
	cats := map[string]map[string]bool{}
	if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if e := json.Unmarshal(v, &p); e != nil {
			return e
		}
		if p.Status == "approved" && !hiddenFromPublic(p, "") {
			published[p.ID] = true
			pts[p.ID] = p
		}
		return nil
	}); e != nil {
		return st, e
	}
	within := func(stamp string, d time.Duration) bool {
		t, e := time.Parse(time.RFC3339, stamp)
		return e == nil && now.Sub(t) <= d
	}
	byName, order := map[string]*statItem{}, []string{}
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if published[n.CampID] && isRepair(n) && (n.Status == "OPEN" || n.Status == "IN_PROGRESS") {
			// Una obra cuenta una vez: los materiales del constructor que la tomó o, si no, la evaluación más reciente con materiales.
			st.RepairJobs++
			list := n.Materials
			if len(list) == 0 {
				for i := len(n.Assessments) - 1; i >= 0; i-- {
					if len(n.Assessments[i].Materials) > 0 {
						list = n.Assessments[i].Materials
						break
					}
				}
			}
			if len(list) > 0 {
				st.RepairCosted++
			}
			seenHere := map[string]bool{}
			for _, m := range list {
				key := strings.ToLower(foldText(m.Item))
				it := mats[key]
				if it == nil {
					it = &statMaterial{Name: m.Item}
					mats[key] = it
					matOrder = append(matOrder, key)
				}
				it.Quantity += m.Quantity
				if !seenHere[key] {
					seenHere[key] = true
					it.Jobs++
				}
			}
			return nil
		}
		if !published[n.CampID] || isRepair(n) || serviceScheduled(n, now) {
			return nil
		}
		active := n.Status == "OPEN" || n.Status == "IN_PROGRESS"
		switch n.Status {
		case "OPEN":
			st.Open++
		case "IN_PROGRESS":
			st.InProgress++
		case "COMPLETED":
			st.Completed++
			if within(n.CompletedAt, 24*time.Hour) {
				st.Completed24h++
			}
			if within(n.CompletedAt, 7*24*time.Hour) {
				st.Completed7d++
			}
		default:
			return nil
		}
		if active && normalizeUrgency(n.Urgency) == "urgente" {
			st.Urgent++
		}
		if within(n.CreatedAt, 24*time.Hour) {
			st.Reported24h++
		}
		if active {
			for _, item := range needItems(n.Need) {
				for _, c := range needCategories(item) {
					if cats[n.CampID] == nil {
						cats[n.CampID] = map[string]bool{}
					}
					cats[n.CampID][c] = true
				}
			}
			if cats[n.CampID] == nil {
				cats[n.CampID] = map[string]bool{}
			}
		}
		if reservedNeed(tx, n) {
			return nil
		}
		for _, item := range needItems(n.Need) {
			key := strings.ToLower(foldText(item))
			it := byName[key]
			if it == nil {
				it = &statItem{Name: item}
				byName[key] = it
				order = append(order, key)
			}
			if active {
				it.Active++
			} else {
				it.Completed++
			}
		}
		return nil
	})
	for _, k := range order {
		st.Items = append(st.Items, *byName[k])
	}
	st.Cover = map[string]*statCover{}
	for _, c := range []string{"water", "food", "shelter", "sanitation", "baby"} {
		st.Cover[c] = &statCover{}
	}
	zones, zoneOrder := map[string]*statZone{}, []string{}
	for id, p := range pts {
		people, babies, children, older := pointPeople(p)
		key := strings.ToLower(foldText(p.Area))
		z := zones[key]
		if z == nil {
			z = &statZone{Area: strings.TrimSpace(p.Area)}
			zones[key] = z
			zoneOrder = append(zoneOrder, key)
		}
		z.Points++
		z.People += people
		set, needing := cats[id]
		if needing {
			z.Needing++
		}
		for c := range set {
			cv := st.Cover[c]
			cv.Points++
			cv.People += people
			cv.Babies += babies
			cv.Children += children
			cv.OlderAdults += older
		}
		if set["food"] {
			z.FoodPeople += people
		}
	}
	for _, k := range zoneOrder {
		st.Zones = append(st.Zones, *zones[k])
	}
	sort.SliceStable(st.Zones, func(i, j int) bool {
		return st.Zones[i].FoodPeople > st.Zones[j].FoodPeople || st.Zones[i].FoodPeople == st.Zones[j].FoodPeople && st.Zones[i].People > st.Zones[j].People
	})
	for _, k := range matOrder {
		st.RepairMaterials = append(st.RepairMaterials, *mats[k])
	}
	sort.SliceStable(st.RepairMaterials, func(i, j int) bool { return st.RepairMaterials[i].Quantity > st.RepairMaterials[j].Quantity })
	if err == nil {
		st.Pledges, err = donationPledges(tx)
	}
	return st, err
}

func (a *App) statsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		var st publicStats
		err := a.db.View(func(tx *bolt.Tx) (e error) {
			st, e = computePublicStats(tx, time.Now())
			return e
		})
		if err != nil {
			fail(w, 500, "database error")
			return
		}
		response(w, 200, st)
	})
}

// needCategories reconoce a qué categoría de «¿Qué donar?» pertenece un artículo escrito (opciones del formulario o texto libre).
func needCategories(item string) []string {
	t := foldText(item)
	out := []string{}
	for _, c := range []struct {
		name  string
		words []string
	}{
		{"water", []string{"agua"}},
		{"food", []string{"aliment", "comida", "arroz", "viver", "comestible", "proteina", "frijol", "lenteja", "enlatad"}},
		{"shelter", []string{"refugio", "carpa", "toldo", "lona", "tienda de campana"}},
		{"sanitation", []string{"saneamiento", "bano", "letrina", "sanitario", "inodoro"}},
		{"baby", []string{"bebe", "panal", "pampers", "formula", "leche"}},
	} {
		for _, w := range c.words {
			if strings.Contains(t, w) {
				out = append(out, c.name)
				break
			}
		}
	}
	return out
}

// pointPeople: total de personas del punto (o la suma de adultos, niños, bebés y mayores si no hay total) y sus grupos.
func pointPeople(p Point) (people, babies, children, older int) {
	babies, children, older = countOf(p.Babies), countOf(p.Children), countOf(p.OlderAdults)
	people = countOf(p.People)
	if people == 0 {
		people = countOf(p.Adults) + children + babies + older
	}
	return
}
