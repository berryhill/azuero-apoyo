package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// «Reparación de casa (Varilla / Clavos)»: con materiales sigue siendo reparación. Al reportarla desde la página pública
// va en una necesidad aparte para constructores (con sus materiales) y nunca sale en público.
func TestRepairWithMaterialsIsSplitForBuilders(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/camps/"+s.camp+"/needs", `{"need":"Agua (Para beber), Reparación de casa (Varilla / Hojas de zinc)","urgency":"normal","requesterName":"Ana","requesterPhone":"6555-1234"}`, false)
	if w.Code != 201 {
		t.Fatalf("report: %d %s", w.Code, w.Body.String())
	}
	var all []CampNeed
	json.Unmarshal(s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true).Body.Bytes(), &all)
	var supply, repair *CampNeed
	for i := range all {
		switch {
		case all[i].Kind == "repair" && strings.Contains(all[i].Need, "Varilla"):
			repair = &all[i]
		case all[i].Need == "Agua (Para beber)":
			supply = &all[i]
		}
	}
	if supply == nil || repair == nil || repair.Need != "Reparación de casa (Varilla / Hojas de zinc)" || repair.RequesterName != "Ana" {
		t.Fatalf("split into supplies and a repair with materials: %+v", all)
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(pub, "Reparación") || strings.Contains(pub, "Varilla") {
		t.Fatalf("repairs are never public: %s", pub)
	}
	if rest, ok := splitRepair("Agua, Reparación de casa (Clavos)"); !ok || rest != "Agua" {
		t.Fatalf("splitRepair with materials: %q %v", rest, ok)
	}
	if got := repairText("Agua, reparacion de casa (Clavos)"); got != "reparacion de casa (Clavos)" {
		t.Fatalf("repairText: %q", got)
	}
}
