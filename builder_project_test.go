package main

import (
	"strings"
	"testing"
)

// Builder project: before/progress/after photos, what is still needed, materials with quantities and % progress;
// only the assigned builder can change it, and completing sets 100 %.
func TestBuilderProjectNotesMaterialsProgressAndPhotos(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","urgency":"normal","kind":"repair"}`, 201)["id"].(string)
	luis, lcsrf := s.crewLogin("/api/builder", "Luis Obra")
	otro, ocsrf := s.crewLogin("/api/builder", "Otro Constructor")
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/take", "{}", luis, lcsrf); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	jpeg := []byte("\xff\xd8\xff fake jpeg")
	for _, stage := range []string{"before", "progress"} {
		if w := rawUpload(s, "/api/builder/needs/"+id+"/media?stage="+stage, luis, lcsrf, "image/jpeg", jpeg); w.Code != 201 && w.Code != 200 {
			t.Fatalf("upload %s: %d %s", stage, w.Code, w.Body.String())
		}
	}
	if w := rawUpload(s, "/api/builder/needs/"+id+"/media?stage=before", otro, ocsrf, "image/jpeg", jpeg); w.Code == 200 || w.Code == 201 {
		t.Fatalf("another builder can't document it: %d", w.Code)
	}
	body := `{"scope":"Techo caído en la cocina; falta repellar","materials":[{"item":"Bolsas de cemento","quantity":12},{"item":"  Láminas  de zinc ","quantity":8}],"progress":40}`
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/project", body, luis, lcsrf); w.Code != 200 {
		t.Fatalf("project: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/project", `{"progress":40}`, otro, ocsrf); w.Code != 409 {
		t.Fatalf("not his job: %d %s", w.Code, w.Body.String())
	}
	for _, bad := range []string{`{"progress":140}`, `{"materials":[{"item":"Cemento","quantity":0}]}`} {
		if w := s.crewCall("POST", "/api/builder/needs/"+id+"/project", bad, luis, lcsrf); w.Code != 400 {
			t.Fatalf("%s: %d", bad, w.Code)
		}
	}
	var n CampNeed
	for _, x := range s.needsOf(s.camp) {
		if x.ID == id {
			n = x
		}
	}
	if n.Scope != "Techo caído en la cocina; falta repellar" || n.Progress != 40 || len(n.Materials) != 2 || n.Materials[1].Item != "Láminas de zinc" || n.Materials[0].Quantity != 12 || len(n.Media) != 2 || n.Media[0].Stage != "before" {
		t.Fatalf("saved project: %+v", n)
	}
	list := s.crewCall("GET", "/api/builder/needs", "", luis, "").Body.String()
	if !strings.Contains(list, `"scope":"Techo caído`) || !strings.Contains(list, `"progress":40`) {
		t.Fatalf("builder sees the project: %s", list)
	}
	if cat := s.crewCall("GET", "/api/builder/materials", "", luis, "").Body.String(); !strings.Contains(cat, "Carriolas") || !strings.Contains(cat, "Electrodos de soldar") {
		t.Fatalf("materials catalog: %s", cat)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/complete", "{}", luis, lcsrf); w.Code != 200 {
		t.Fatalf("complete: %d", w.Code)
	}
	if w := rawUpload(s, "/api/builder/needs/"+id+"/media?stage=after", luis, lcsrf, "image/jpeg", jpeg); w.Code != 201 && w.Code != 200 {
		t.Fatalf("final photo right after completing: %d %s", w.Code, w.Body.String())
	}
	for _, x := range s.needsOf(s.camp) {
		if x.ID == id && (x.Progress != 100 || x.Status != "COMPLETED") {
			t.Fatalf("completed = 100 %%: %+v", x)
		}
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(pub, "cemento") {
		t.Fatalf("repairs stay private: %s", pub)
	}
}
