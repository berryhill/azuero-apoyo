package main

import (
	"strings"
	"testing"
)

// Assessments: several builders inspect a repair before anyone takes it (with photos all builders can see);
// each edits only their own; another builder can then take the job.
func TestRepairAssessmentsBeforeTaking(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Reparación de casa","urgency":"normal","kind":"repair"}`, 201)["id"].(string)
	scout, scsrf := s.crewLogin("/api/builder", "Explorador Uno")
	other, ocsrf := s.crewLogin("/api/builder", "Constructor Dos")
	body := `{"work":["Techo / zinc","Paredes / bloque"," "],"materials":[{"item":"Láminas de zinc","quantity":10}],"days":4,"workers":2,"notes":"Techo caído en la sala"}`
	w := s.crewCall("POST", "/api/builder/needs/"+id+"/assessments", body, scout, scsrf)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"byName":"Explorador Uno"`) {
		t.Fatalf("assess: %d %s", w.Code, w.Body.String())
	}
	aid := strings.Split(strings.Split(w.Body.String(), `"id":"`)[1], `"`)[0]
	jpeg := []byte("\xff\xd8\xff fake jpeg")
	if w := rawUpload(s, "/api/builder/needs/"+id+"/media?stage=inspect&ref="+aid, scout, scsrf, "image/jpeg", jpeg); w.Code != 201 {
		t.Fatalf("inspection photo: %d %s", w.Code, w.Body.String())
	}
	if w := rawUpload(s, "/api/builder/needs/"+id+"/media?stage=inspect&ref="+aid, other, ocsrf, "image/jpeg", jpeg); w.Code == 201 {
		t.Fatal("photos only on your own assessment")
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/assessments/"+aid, `{"work":["Piso"]}`, other, ocsrf); w.Code != 404 {
		t.Fatalf("can't edit someone else's: %d", w.Code)
	}
	s.crewCall("POST", "/api/builder/needs/"+id+"/assessments", `{"work":["Electricidad"],"days":1}`, other, ocsrf)
	list := s.crewCall("GET", "/api/builder/needs", "", other, "").Body.String()
	if strings.Count(list, `"byName"`) != 2 || !strings.Contains(list, `"stage":"inspect"`) || !strings.Contains(list, "Techo caído") {
		t.Fatalf("every builder sees all assessments and their photos before taking: %s", list)
	}
	media := strings.Split(strings.Split(list, `"stage":"inspect"`)[0], `"id":"`)
	mid := strings.Split(media[len(media)-1], `"`)[0]
	if w := s.crewCall("GET", "/api/builder/media/"+id+"/"+mid, "", other, ""); w.Code != 200 {
		t.Fatalf("other builder sees the inspection photo: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/take", "{}", other, ocsrf); w.Code != 200 {
		t.Fatalf("someone else takes the job: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/builder/needs/"+id+"/assessments", `{"work":["x"],"days":400}`, scout, scsrf); w.Code != 400 {
		t.Fatalf("bad days: %d", w.Code)
	}
}
