package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// sessionCall: a request with one session cookie and its CSRF token (cook_session, runner_session…).
func (s *serviceHarness) sessionLogin(path, code, cookie string) (*http.Cookie, string) {
	s.t.Helper()
	w := s.crewCall("POST", path, `{"accessCode":"`+code+`"}`, nil, "")
	if w.Code != 200 {
		s.t.Fatalf("login %s %s: %d %s", path, code, w.Code, w.Body.String())
	}
	var out struct{ CSRF string }
	json.Unmarshal(w.Body.Bytes(), &out)
	for _, c := range w.Result().Cookies() {
		if c.Name == cookie {
			return c, out.CSRF
		}
	}
	s.t.Fatalf("no %s cookie", cookie)
	return nil, ""
}

// Prepared meals: a team member asks for N meals at a point; one cook takes it (others see who); the cook asks for
// drivers with a pickup time and/or invites runners, who accept to get their own delivery run; when the last run is
// completed the request closes and the point logs the service «Comidas preparadas · 50 · almuerzo».
func TestMealTicketCookRunnersAndServiceLog(t *testing.T) {
	t.Setenv("MEDIA_LOCAL_DISK", "1")
	s := newServiceHarness(t)
	user := func(body string) string { return s.expect("POST", "/api/admin/users", body, 201)["id"].(string) }
	scout := user(`{"name":"Sara Scout","telephone":"6999-1111","roles":["runner"],"accessCode":"SARA77"}`)
	user(`{"name":"Solo Uno","telephone":"6999-2222","roles":["runner"],"accessCode":"SOLO77"}`)
	user(`{"name":"Cora Cocina","telephone":"6999-3333","roles":["cook"],"accessCode":"CORA77"}`)
	user(`{"name":"Bruno Cocina","telephone":"6999-4444","roles":["cook"],"accessCode":"BRUN77"}`)
	r1 := user(`{"name":"Rita Runner","telephone":"6999-5555","roles":["runner"],"accessCode":"RITA77"}`)
	r2 := user(`{"name":"Raúl Runner","telephone":"6999-6666","roles":["runner"],"accessCode":"RAUL77"}`)
	s.team("Norte", scout)

	// Only someone on a team can ask; people and meal are required.
	if w := s.as(s.viewerCookie("SOLO77"), "POST", "/api/full/camps/"+s.camp+"/meals", `{"people":50,"meal":"almuerzo"}`); w.Code != 400 || !strings.Contains(w.Body.String(), "equipo") {
		t.Fatalf("no team: %d %s", w.Code, w.Body.String())
	}
	sara := s.viewerCookie("SARA77")
	for _, bad := range []string{`{"meal":"almuerzo"}`, `{"people":50}`, `{"people":50,"meal":"merienda"}`} {
		if w := s.as(sara, "POST", "/api/full/camps/"+s.camp+"/meals", bad); w.Code != 400 {
			t.Fatalf("%s: %d", bad, w.Code)
		}
	}
	w := s.as(sara, "POST", "/api/full/camps/"+s.camp+"/meals", `{"people":50,"meal":"almuerzo"}`)
	if w.Code != 201 || !strings.Contains(w.Body.String(), "Comida preparada · 50 personas · almuerzo") {
		t.Fatalf("meal request: %d %s", w.Code, w.Body.String())
	}
	var created struct{ ID string }
	json.Unmarshal(w.Body.Bytes(), &created)
	ticket := created.ID

	// Runners don't see it in «Runs» and can't take it like a normal need.
	rc1, rcsrf1 := s.sessionLogin("/api/runner/session", "RITA77", "runner_session")
	if body := s.crewCall("GET", "/api/runner/needs", "", rc1, "").Body.String(); strings.Contains(body, ticket) {
		t.Fatalf("meal in Runs: %s", body)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+ticket+"/take", `{}`, rc1, rcsrf1); w.Code == 200 {
		t.Fatalf("runner took the meal ticket")
	}

	// One cook at a time.
	cc, ccsrf := s.sessionLogin("/api/cook/session", "CORA77", "cook_session")
	bc, bcsrf := s.sessionLogin("/api/cook/session", "BRUN77", "cook_session")
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/take", `{}`, cc, ccsrf); w.Code != 200 {
		t.Fatalf("take: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/take", `{}`, bc, bcsrf); w.Code != 409 {
		t.Fatalf("second cook: %d", w.Code)
	}
	var lists struct{ Open, Mine, Taken []mealTicket }
	json.Unmarshal(s.crewCall("GET", "/api/cook/tickets", "", bc, "").Body.Bytes(), &lists)
	if len(lists.Taken) != 1 || lists.Taken[0].CookName != "Cora Cocina" || len(lists.Open) != 0 {
		t.Fatalf("Bruno sees who has it: %+v", lists)
	}
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/drivers", `{"pickupAt":"12:30"}`, bc, bcsrf); w.Code != 409 {
		t.Fatalf("other cook asks drivers: %d", w.Code)
	}
	// Packing (meal types), drivers with a pickup time, and an invite for Raúl.
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/packing", `{"items":[{"item":"Arroz con pollo","quantity":50}]}`, cc, ccsrf); w.Code != 200 {
		t.Fatalf("packing: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/drivers", `{"pickupAt":"nunca"}`, cc, ccsrf); w.Code != 400 {
		t.Fatalf("pickup time required: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/drivers", `{"pickupAt":"12:30"}`, cc, ccsrf); w.Code != 200 || !strings.Contains(w.Body.String(), `"driversAskedAt"`) {
		t.Fatalf("ask drivers: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/drivers", `{"pickupAt":"12:30","runnerId":"`+r2+`"}`, cc, ccsrf); w.Code != 200 || !strings.Contains(w.Body.String(), "Raúl Runner") {
		t.Fatalf("invite: %d %s", w.Code, w.Body.String())
	}
	// Rita sees it in «🍲 Comidas» and accepts: her own run, in «Yo», with pickup time, cook and packing.
	var meals []mealTicket
	json.Unmarshal(s.crewCall("GET", "/api/runner/meals", "", rc1, "").Body.Bytes(), &meals)
	if len(meals) != 1 || meals[0].ID != ticket || meals[0].PickupAt == "" || meals[0].CookName != "Cora Cocina" || meals[0].InvitedMe {
		t.Fatalf("Rita's meals: %+v", meals)
	}
	if w := s.crewCall("POST", "/api/runner/meals/"+ticket+"/accept", `{}`, rc1, rcsrf1); w.Code != 200 {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	mine := s.crewCall("GET", "/api/runner/needs", "", rc1, "").Body.String()
	if !strings.Contains(mine, `"kind":"mealrun"`) || !strings.Contains(mine, `"pickupAt"`) || !strings.Contains(mine, "Cora Cocina") || !strings.Contains(mine, "Arroz con pollo") {
		t.Fatalf("Rita's run: %s", mine)
	}
	json.Unmarshal(s.crewCall("GET", "/api/runner/meals", "", rc1, "").Body.Bytes(), &meals)
	if len(meals) != 0 {
		t.Fatalf("accepted meals leave «Comidas»: %+v", meals)
	}
	// Raúl was invited and accepts too (several runners per delivery).
	rc2, rcsrf2 := s.sessionLogin("/api/runner/session", "RAUL77", "runner_session")
	json.Unmarshal(s.crewCall("GET", "/api/runner/meals", "", rc2, "").Body.Bytes(), &meals)
	if len(meals) != 1 || !meals[0].InvitedMe {
		t.Fatalf("Raúl invited: %+v", meals)
	}
	s.crewCall("POST", "/api/runner/meals/"+ticket+"/accept", `{}`, rc2, rcsrf2)
	runOf := func(runner string) string {
		for _, n := range s.needsOf(s.camp) {
			if n.Kind == kindMealRun && n.RunnerID == runner && n.Status == "IN_PROGRESS" {
				return n.ID
			}
		}
		t.Fatal("no run for", runner)
		return ""
	}
	run1, run2 := runOf(r1), runOf(r2)
	ticketNow := func() CampNeed {
		for _, n := range s.needsOf(s.camp) {
			if n.ID == ticket {
				return n
			}
		}
		return CampNeed{}
	}
	// The cook can't close it while runners are on the way; the first delivery doesn't close it either.
	if w := s.crewCall("POST", "/api/cook/tickets/"+ticket+"/done", `{}`, cc, ccsrf); w.Code != 400 {
		t.Fatalf("done with runs on the way: %d", w.Code)
	}
	if w := s.crewCall("POST", "/api/runner/needs/"+run1+"/complete", `{}`, rc1, rcsrf1); w.Code != 200 {
		t.Fatalf("complete 1: %d %s", w.Code, w.Body.String())
	}
	if n := ticketNow(); n.Status != "IN_PROGRESS" {
		t.Fatalf("still one run on the way: %+v", n.Status)
	}
	s.crewCall("POST", "/api/runner/needs/"+run2+"/complete", `{}`, rc2, rcsrf2)
	if n := ticketNow(); n.Status != "COMPLETED" || !n.ServiceLogged {
		t.Fatalf("last delivery closes the request: %+v", n)
	}
	// The point's services show it publicly, without names.
	services := s.call("GET", "/api/camps/"+s.camp+"/services", "", false).Body.String()
	if !strings.Contains(services, `"service":"Comidas preparadas"`) || !strings.Contains(services, `"quantity":50`) || !strings.Contains(services, `"detail":"almuerzo"`) || strings.Contains(services, "Raúl") || strings.Contains(services, "Cora") {
		t.Fatalf("services: %s", services)
	}
	// Public needs: the request shows, the delivery runs and the cook don't.
	public := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String()
	if !strings.Contains(public, "Comida preparada · 50 personas") || strings.Contains(public, "mealrun") || strings.Contains(public, "cookId") {
		t.Fatalf("public needs: %s", public)
	}
	// Admin/Command log another service by hand, valid for 14 days.
	s.expect("POST", "/api/admin/camps/"+s.camp+"/services", `{"service":"Fumigación de mosquitos","validDays":14,"detail":"zona del río"}`, 201)
	if services := s.call("GET", "/api/camps/"+s.camp+"/services", "", false).Body.String(); !strings.Contains(services, "Fumigación de mosquitos") || !strings.Contains(services, `"nextDueAt"`) {
		t.Fatalf("manual service: %s", services)
	}
}

// «Elige tu cocina»: a new name joins the shared list; the same name with other accents/case reuses the saved spelling.
func TestCookKitchenPickAndCreate(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Cora Cocina","telephone":"6999-3333","roles":["cook"],"accessCode":"CORA77"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Bruno Cocina","telephone":"6999-4444","roles":["cook"],"accessCode":"BRUN77"}`, 201)
	cc, ccsrf := s.sessionLogin("/api/cook/session", "CORA77", "cook_session")
	bc, bcsrf := s.sessionLogin("/api/cook/session", "BRUN77", "cook_session")
	if w := s.crewCall("POST", "/api/cook/kitchen", `{"kitchen":"  Fonda   Las Tablas "}`, cc, ccsrf); w.Code != 200 {
		t.Fatalf("new kitchen: %d %s", w.Code, w.Body.String())
	}
	if w := s.crewCall("POST", "/api/cook/kitchen", `{"kitchen":"fonda las TABLAS"}`, bc, bcsrf); w.Code != 200 {
		t.Fatalf("pick kitchen: %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Kitchen  string
		Kitchens []string
	}
	json.Unmarshal(s.crewCall("GET", "/api/cook/tickets", "", bc, "").Body.Bytes(), &out)
	if out.Kitchen != "Fonda Las Tablas" || len(out.Kitchens) != 1 || out.Kitchens[0] != "Fonda Las Tablas" {
		t.Fatalf("kitchen: %+v", out)
	}
}
