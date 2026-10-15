package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «Resumen» on the public map: need counts of published points only, items without private data, repairs left out.
func TestPublicStatsCountsNeeds(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua, Alimentos","urgency":"urgente"}`, 201)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"agua","urgency":"normal"}`, 201)
	done := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Kit de bebé","urgency":"normal"}`, 201)
	s.expect("PATCH", "/api/admin/needs/"+done["id"].(string), `{"status":"COMPLETED"}`, 200)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Techo","urgency":"normal","kind":"repair"}`, 201)
	w := s.call("GET", "/api/stats", "", false)
	if w.Code != 200 {
		t.Fatalf("stats: %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "Techo") {
		t.Fatalf("no private data or repairs in the public stats: %s", body)
	}
	var st publicStats
	json.Unmarshal(w.Body.Bytes(), &st)
	if st.Open != 2 || st.InProgress != 0 || st.Completed != 1 || st.Urgent != 1 || st.Completed24h != 1 || st.Completed7d != 1 || st.Reported24h != 3 {
		t.Fatalf("counts: %+v", st)
	}
	items := map[string]statItem{}
	for _, it := range st.Items {
		items[strings.ToLower(it.Name)] = it
	}
	if items["agua"].Active != 2 || items["alimentos"].Active != 1 || items["kit de bebé"].Completed != 1 || len(items) != 3 {
		t.Fatalf("items (Agua and agua are one): %+v", st.Items)
	}
}

// Repair materials in the public summary: one list per active job (builder's list, else the latest assessment), summed.
func TestPublicStatsRepairMaterials(t *testing.T) {
	s := newServiceHarness(t)
	a := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","urgency":"normal","kind":"repair"}`, 201)["id"].(string)
	b := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","urgency":"normal","kind":"repair"}`, 201)["id"].(string)
	s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","urgency":"normal","kind":"repair"}`, 201)
	scout, scsrf := s.crewLogin("/api/builder", "Explorador")
	s.crewCall("POST", "/api/builder/needs/"+a+"/assessments", `{"materials":[{"item":"Bolsas de cemento","quantity":5}]}`, scout, scsrf)
	s.crewCall("POST", "/api/builder/needs/"+a+"/assessments", `{"materials":[{"item":"Bolsas de cemento","quantity":8},{"item":"Carriolas","quantity":4}]}`, scout, scsrf)
	builder, bcsrf := s.crewLogin("/api/builder", "Constructor")
	s.crewCall("POST", "/api/builder/needs/"+b+"/assessments", `{"materials":[{"item":"Bolsas de cemento","quantity":99}]}`, scout, scsrf)
	s.crewCall("POST", "/api/builder/needs/"+b+"/take", "{}", builder, bcsrf)
	s.crewCall("POST", "/api/builder/needs/"+b+"/project", `{"materials":[{"item":"bolsas de cemento","quantity":12}]}`, builder, bcsrf)
	var st publicStats
	json.Unmarshal(s.call("GET", "/api/stats", "", false).Body.Bytes(), &st)
	if st.RepairJobs != 3 || st.RepairCosted != 2 || len(st.RepairMaterials) != 2 {
		t.Fatalf("jobs: %+v", st)
	}
	if m := st.RepairMaterials[0]; strings.ToLower(m.Name) != "bolsas de cemento" || m.Quantity != 20 || m.Jobs != 2 {
		t.Fatalf("latest assessment (8) + builder list (12), not 5 or 99: %+v", st.RepairMaterials)
	}
}

// «¿Qué donar?»: people per need category (once per point) and mouths to feed per zone.
func TestPublicStatsCoverage(t *testing.T) {
	s := newServiceHarness(t)
	mk := func(code, area, people, babies string) string {
		p := s.expect("POST", "/api/admin/points", `{"code":"`+code+`","area":"`+area+`","people":"`+people+`","babies":"`+babies+`"}`, 201)
		return p["id"].(string)
	}
	a := mk("A-1", "Cañas", "10", "2")
	b := mk("B-1", "Cañas", "6", "")
	c := mk("C-2", "Tonosí", "4", "1")
	s.expect("POST", "/api/admin/camps/"+a+"/needs", `{"need":"Agua, Alimentos","urgency":"normal"}`, 201)
	s.expect("POST", "/api/admin/camps/"+a+"/needs", `{"need":"agua y arroz","urgency":"normal"}`, 201)
	s.expect("POST", "/api/admin/camps/"+b+"/needs", `{"need":"Refugio (carpa / lona / toldo), Saneamiento / baño","urgency":"normal"}`, 201)
	s.expect("POST", "/api/admin/camps/"+c+"/needs", `{"need":"Kit de bebé, comida","urgency":"normal"}`, 201)
	var st publicStats
	json.Unmarshal(s.call("GET", "/api/stats", "", false).Body.Bytes(), &st)
	if w := st.Cover["water"]; w.Points != 1 || w.People != 10 {
		t.Fatalf("water counted once per point: %+v", w)
	}
	if f := st.Cover["food"]; f.Points != 2 || f.People != 14 {
		t.Fatalf("food (Alimentos, arroz, comida): %+v", f)
	}
	if sh := st.Cover["shelter"]; sh.People != 6 || st.Cover["sanitation"].People != 6 || st.Cover["baby"].Babies != 1 {
		t.Fatalf("shelter / sanitation / baby: %+v %+v %+v", sh, st.Cover["sanitation"], st.Cover["baby"])
	}
	if len(st.Zones) < 2 || st.Zones[0].Area != "Cañas" || st.Zones[0].FoodPeople != 10 || st.Zones[0].People != 16 || st.Zones[0].Needing != 2 {
		t.Fatalf("zones (most mouths to feed first): %+v", st.Zones)
	}
}
