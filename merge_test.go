package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMergeDuplicateIntoApprovedCamp(t *testing.T) {
	s := newServiceHarness(t)
	sub := s.call("POST", "/api/submissions", `{"area":"Venao","needs":"Agua\nArroz","people":"30","members":[{"name":"Ana","phone":"6111-2222"}]}`, false)
	if sub.Code != 201 {
		t.Fatalf("submission: %s", sub.Body.String())
	}
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	src := created["id"]
	path := "/api/admin/points/" + src + "/merge"

	if w := s.call("POST", path, `{"targetId":"`+s.camp+`","fields":{"note":"x"}}`, true); w.Code != 400 {
		t.Fatalf("note must be server-managed: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", path, `{"targetId":"`+src+`"}`, true); w.Code != 400 {
		t.Fatalf("self merge: %d", w.Code)
	}
	if w := s.call("POST", path, `{"targetId":"`+s.camp+`","needs":["  "]}`, true); w.Code != 400 {
		t.Fatalf("blank need: %d", w.Code)
	}

	body := `{"targetId":"` + s.camp + `","fields":{"people":"30","members":[{"name":"Private camp","phone":"6555-1234"},{"name":"Ana","phone":"6111-2222"}]},"needs":["Arroz"],"log":"Necesidad no creada: Agua"}`
	w := s.call("POST", path, body, true)
	if w.Code != 200 {
		t.Fatalf("merge: %d %s", w.Code, w.Body.String())
	}
	var target Point
	json.Unmarshal(w.Body.Bytes(), &target)
	if target.People != "30" || len(target.Members) != 2 || target.SecondPhone != "6111-2222" || target.Phone != "6555-1234" {
		t.Fatalf("merged fields: %+v", target)
	}
	if !strings.Contains(target.Note, "— Fusión del reporte duplicado") || !strings.Contains(target.Note, "Necesidad no creada: Agua") {
		t.Fatalf("merge log missing from private note: %q", target.Note)
	}

	needs := s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true)
	if !strings.Contains(needs.Body.String(), `"need":"Arroz"`) || strings.Contains(needs.Body.String(), `"need":"Agua"`) {
		t.Fatalf("needs after merge: %s", needs.Body.String())
	}

	all := s.call("GET", "/api/admin/points", "", true)
	var points []Point
	json.Unmarshal(all.Body.Bytes(), &points)
	for _, p := range points {
		if p.ID == src && (p.Status != "rejected" || p.MergedInto != s.camp) {
			t.Fatalf("source not marked merged: %+v", p)
		}
	}

	public := s.call("GET", "/api/points", "", false).Body.String()
	if strings.Contains(public, "Fusión") || strings.Contains(public, "6111-2222") || strings.Contains(public, "Ana") {
		t.Fatalf("merge data leaked publicly: %s", public)
	}

	if w := s.call("POST", "/api/admin/decisions", `{"id":"`+src+`","decision":"approved"}`, true); w.Code != 409 {
		t.Fatalf("merged report must not be re-approved: %d", w.Code)
	}
	if w := s.call("POST", path, body, true); w.Code != 404 {
		t.Fatalf("second merge of same report: %d", w.Code)
	}
}

func TestMergeRequiresApprovedTarget(t *testing.T) {
	s := newServiceHarness(t)
	var ids []string
	for _, phone := range []string{"6111-0001", "6111-0002"} {
		w := s.call("POST", "/api/submissions", `{"needs":"Agua","contact":"P","phone":"`+phone+`"}`, false)
		var c map[string]string
		json.Unmarshal(w.Body.Bytes(), &c)
		ids = append(ids, c["id"])
	}
	if w := s.call("POST", "/api/admin/points/"+ids[0]+"/merge", `{"targetId":"`+ids[1]+`"}`, true); w.Code != 404 {
		t.Fatalf("pending target accepted: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/admin/points/"+ids[0]+"/merge", `{"targetId":"`+s.camp+`"}`, false); w.Code == 200 {
		t.Fatal("merge without session succeeded")
	}
}

func TestMergeConsolidatesIntoOpenNeedAndMarksCheckup(t *testing.T) {
	s := newServiceHarness(t)
	open := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"urgente","status":"OPEN"}`, true)
	route := s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Carpas","urgency":"urgente","status":"IN_PROGRESS"}`, true)
	if open.Code != 201 && open.Code != 200 || route.Code != 201 && route.Code != 200 {
		t.Fatalf("seed needs: %d %d", open.Code, route.Code)
	}
	var openNeed, routeNeed CampNeed
	json.Unmarshal(open.Body.Bytes(), &openNeed)
	json.Unmarshal(route.Body.Bytes(), &routeNeed)

	sub := s.call("POST", "/api/submissions", `{"needs":"Kit de bebé, Pañales","contact":"P","phone":"6111-0003","lastCheckup":"2020-01-01T00:00:00Z","mergedInto":"x"}`, false)
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	w := s.call("POST", "/api/admin/points/"+created["id"]+"/merge", `{"targetId":"`+s.camp+`","needs":["Kit de bebé","Pañales"]}`, true)
	if w.Code != 200 {
		t.Fatalf("merge: %d %s", w.Code, w.Body.String())
	}
	var target Point
	json.Unmarshal(w.Body.Bytes(), &target)
	if target.LastCheckup == "" {
		t.Fatal("merge must set lastCheckup")
	}

	var needs []CampNeed
	json.Unmarshal(s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true).Body.Bytes(), &needs)
	if len(needs) != 2 {
		t.Fatalf("merge must not add separate needs: %+v", needs)
	}
	for _, n := range needs {
		switch n.ID {
		case openNeed.ID:
			if n.Need != "Agua, Kit de bebé, Pañales" || n.RevisedAt == "" || n.CreatedAt != openNeed.CreatedAt {
				t.Fatalf("open need not consolidated: %+v", n)
			}
		case routeNeed.ID:
			if n.Need != "Carpas" || n.RevisedAt != "" {
				t.Fatalf("en-route need changed: %+v", n)
			}
		}
	}
	if pub := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); !strings.Contains(pub, `"revisedAt"`) {
		t.Fatalf("public needs should show revisedAt: %s", pub)
	}
}

func TestLastCheckupOnApproveEditAndButton(t *testing.T) {
	s := newServiceHarness(t)
	sub := s.call("POST", "/api/submissions", `{"needs":"Agua","contact":"P","phone":"6111-0004","lastCheckup":"2020-01-01T00:00:00Z"}`, false)
	var created map[string]string
	json.Unmarshal(sub.Body.Bytes(), &created)
	point := func(id string) Point {
		var points []Point
		json.Unmarshal(s.call("GET", "/api/admin/points", "", true).Body.Bytes(), &points)
		for _, p := range points {
			if p.ID == id {
				return p
			}
		}
		return Point{}
	}
	if p := point(created["id"]); p.LastCheckup != "" {
		t.Fatalf("public submission set lastCheckup: %q", p.LastCheckup)
	}
	s.call("POST", "/api/admin/decisions", `{"id":"`+created["id"]+`","decision":"approved"}`, true)
	if point(created["id"]).LastCheckup == "" {
		t.Fatal("approval must set lastCheckup")
	}
	if point(s.camp).LastCheckup == "" {
		t.Fatal("admin-created camp must start with lastCheckup")
	}
	if w := s.call("PATCH", "/api/admin/camps/"+s.camp, `{"lastCheckup":"2020-01-01T00:00:00Z"}`, true); w.Code != 400 {
		t.Fatalf("lastCheckup must be read-only: %d", w.Code)
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/checkup", `{}`, true); w.Code != 200 || !strings.Contains(w.Body.String(), `"lastCheckup"`) {
		t.Fatalf("checkup: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/admin/camps/"+s.camp+"/checkup", `{}`, false); w.Code == 200 {
		t.Fatal("checkup without session succeeded")
	}
	if !strings.Contains(s.call("GET", "/api/points", "", false).Body.String(), `"lastCheckup"`) {
		t.Fatal("lastCheckup should be public")
	}
}
