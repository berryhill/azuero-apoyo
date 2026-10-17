package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// Public sign-up: approved at once as a scout (no manual review; runner is given by their team manager or admin); the generated code signs in right away.
func TestSignupCreatesActiveScoutWithMemorableCode(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/session/signup", `{"name":"  José  Pérez ","telephone":"6999-1234","consent":true}`, false)
	if w.Code != 201 {
		t.Fatalf("signup: %d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	code, _ := got["accessCode"].(string)
	if !regexp.MustCompile(`^[A-Z]{2,8}[0-9]{2}JOSE[0-9]$`).MatchString(code) || got["name"] != "José Pérez" || got["pending"] != nil {
		t.Fatalf("word + 2 digits + first name + digit, not pending: %v", got)
	}
	// The admin sees an active scout with a note; the list never shows the code itself.
	list := s.call("GET", "/api/admin/users", "", true).Body.String()
	if !strings.Contains(list, "José Pérez") || !strings.Contains(list, `"roles":["scout"]`) || !strings.Contains(list, "sin aprobación manual") || strings.Contains(list, code) {
		t.Fatalf("active scout in the admin list: %s", list)
	}
	// The code opens the Scout panel immediately (any case).
	if login := s.call("POST", "/api/session/code", `{"accessCode":"`+strings.ToLower(code)+`"}`, false); login.Code != 200 || !strings.Contains(login.Body.String(), "/scout.html") {
		t.Fatalf("signed in right away: %d %s", login.Code, login.Body.String())
	}
	// Same phone again (maybe the same person who lost their code): a new account with its own code, never the old one;
	// the same full name gets a number. Teams tell them apart by user ID.
	again := s.call("POST", "/api/session/signup", `{"name":"José Pérez","telephone":"69991234","consent":true}`, false)
	if again.Code != 201 || strings.Contains(again.Body.String(), code) || !strings.Contains(again.Body.String(), "José Pérez 2") {
		t.Fatalf("duplicate phone is allowed: %d %s", again.Code, again.Body.String())
	}
	for body, field := range map[string]string{`{"name":"Ana Ruiz","telephone":"6999-0000"}`: "consent", `{"name":"Ana Ruiz","telephone":"6999-0000","consent":false}`: "consent", `{"name":"Solo","telephone":"6999-0000"}`: "name", `{"name":"Ana Ruiz","telephone":"12"}`: "telephone"} {
		if w := s.call("POST", "/api/session/signup", body, false); w.Code != 400 || !strings.Contains(w.Body.String(), field) {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body.String())
		}
	}
	// Admin can still take the access away.
	s.expect("PATCH", "/api/admin/users/"+got["id"].(string), `{"disabled":true}`, 200)
	if w := s.call("POST", "/api/session/code", `{"accessCode":"`+code+`"}`, false); w.Code == 200 {
		t.Fatalf("disabled: %d", w.Code)
	}
}

func TestSignupIsLimitedPerConnection(t *testing.T) {
	s := newServiceHarness(t)
	for i := 0; i < codeMaxWrong; i++ {
		body := `{"name":"Persona Prueba","telephone":"6000-00` + string(rune('0'+i/10)) + string(rune('0'+i%10)) + `","consent":true}`
		if w := s.call("POST", "/api/session/signup", body, false); w.Code != 201 {
			t.Fatalf("signup %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	if w := s.call("POST", "/api/session/signup", `{"name":"Una Más","telephone":"6111-1111","consent":true}`, false); w.Code != 429 {
		t.Fatalf("11th signup from one connection: %d %s", w.Code, w.Body.String())
	}
}

func TestSignupNamePart(t *testing.T) {
	for in, want := range map[string]string{"José Pérez": "JOSE", "Maximiliano": "MAXIM", "Ñoño": "NONO", "  ": "AMIGO", "1 2": "AMIGO"} {
		if got := signupNamePart(in); got != want {
			t.Fatalf("%q → %q, want %q", in, got, want)
		}
	}
}

// Sign-up asks for an optional email so admin can recognize the person if they lose their code; admin can edit it.
func TestSignupStoresEmailForCodeRecovery(t *testing.T) {
	s := newServiceHarness(t)
	if w := s.call("POST", "/api/session/signup", `{"name":"Eva Correo","telephone":"6999-7777","email":"nope","consent":true}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"email"`) {
		t.Fatalf("bad email: %d %s", w.Code, w.Body.String())
	}
	w := s.call("POST", "/api/session/signup", `{"name":"Eva Correo","telephone":"6999-7777","email":"  Eva@Gmail.COM ","consent":true}`, false)
	if w.Code != 201 || strings.Contains(w.Body.String(), "gmail") {
		t.Fatalf("signup: %d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	id := got["id"].(string)
	if list := s.call("GET", "/api/admin/users", "", true).Body.String(); !strings.Contains(list, `"email":"eva@gmail.com"`) {
		t.Fatalf("admin sees the email: %s", list)
	}
	s.expect("PATCH", "/api/admin/users/"+id, `{"email":"eva.nueva@correo.pa"}`, 200)
	if w := s.call("PATCH", "/api/admin/users/"+id, `{"email":"sin arroba"}`, true); w.Code != 400 || !strings.Contains(w.Body.String(), "email") {
		t.Fatalf("admin bad email: %d", w.Code)
	}
	if w := s.call("POST", "/api/session/signup", `{"name":"Sin Correo","telephone":"6999-8888","consent":true}`, false); w.Code != 201 {
		t.Fatalf("email stays optional: %d %s", w.Code, w.Body.String())
	}
}

// Codes assigned before (any format, any case) keep working: the page sends them without the «!».
func TestExistingAccessCodesStillSignIn(t *testing.T) {
	s := newServiceHarness(t)
	s.expect("POST", "/api/admin/users", `{"name":"Jader Surf","telephone":"6999-3333","roles":["runner"],"accessCode":"jadersurf3"}`, 201)
	s.expect("POST", "/api/admin/users", `{"name":"Prueba Corre","telephone":"6999-4444","roles":["runner"],"accessCode":"PRUEBACORRE"}`, 201)
	for _, code := range []string{"jadersurf3", "JADERSURF3", " jader surf3 ", "pruebacorre"} {
		if w := s.call("POST", "/api/session/code", `{"accessCode":"`+code+`"}`, false); w.Code != 200 {
			t.Fatalf("%q: %d %s", code, w.Code, w.Body.String())
		}
	}
}

// From the Constructores sign-in, a new account is a builder.
func TestSignupFromBuilderPanelIsBuilder(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/session/signup", `{"name":"Beto Obra","telephone":"6999-5555","panel":"builder","consent":true}`, false)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"roles":["builder"]`) {
		t.Fatalf("builder signup: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/session/signup", `{"name":"Ana Corre","telephone":"6999-6666","panel":"admin","consent":true}`, false); w.Code != 201 || !strings.Contains(w.Body.String(), `"roles":["scout"]`) {
		t.Fatalf("anything else is a scout: %d %s", w.Code, w.Body.String())
	}
}

// Same full name as someone else (any case or accents): a separate account, told apart by a number.
func TestSignupSameNameGetsANumber(t *testing.T) {
	s := newServiceHarness(t)
	names := []string{}
	for i, body := range []string{
		`{"name":"Juan Pérez","telephone":"6100-0001","consent":true}`,
		`{"name":"juan  perez","telephone":"6100-0002","consent":true}`,
		`{"name":"Juan Pérez","telephone":"6100-0003","consent":true}`,
	} {
		w := s.call("POST", "/api/session/signup", body, false)
		if w.Code != 201 {
			t.Fatalf("%d: %d %s", i, w.Code, w.Body.String())
		}
		var got map[string]any
		json.Unmarshal(w.Body.Bytes(), &got)
		names = append(names, got["name"].(string))
	}
	if names[0] != "Juan Pérez" || names[1] != "juan perez 2" || names[2] != "Juan Pérez 3" {
		t.Fatalf("names: %q", names)
	}
	var users []map[string]any
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	if len(users) != 3 {
		t.Fatalf("three separate users: %d", len(users))
	}
}

// A scout who opens Runners can become a runner in one tap: runner role, no team, signed into Runners. /api/full/me
// then counts the runs they are carrying, for the menu to show «en curso».
func TestScoutBecomesRunner(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/session/signup", `{"name":"Lola Scout","telephone":"6222-1111","consent":true}`, false)
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	code := got["accessCode"].(string)
	vc := s.viewerCookie(code)
	if w := s.as(vc, "POST", "/api/full/become-runner", `{}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"roles":["scout","runner"]`) && !strings.Contains(w.Body.String(), `"roles":["runner","scout"]`) || !strings.Contains(w.Body.String(), "/runner.html") {
		t.Fatalf("become runner: %d %s", w.Code, w.Body.String())
	}
	var users []adminUser
	json.Unmarshal(s.call("GET", "/api/admin/users", "", true).Body.Bytes(), &users)
	id := ""
	for _, u := range users {
		if u.Name == "Lola Scout" {
			id = u.ID
			if u.TeamID != "" || !slicesContains(u.Roles, "runner") {
				t.Fatalf("runner without a team: %+v", u)
			}
		}
	}
	if w := s.as(vc, "GET", "/api/full/me", ""); !strings.Contains(w.Body.String(), `"activeRuns":0`) {
		t.Fatalf("no runs yet: %s", w.Body.String())
	}
	s.call("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal","status":"IN_PROGRESS","runnerId":"`+id+`"}`, true)
	if w := s.as(vc, "GET", "/api/full/me", ""); !strings.Contains(w.Body.String(), `"activeRuns":1`) {
		t.Fatalf("one run in progress: %s", w.Body.String())
	}
	if w := s.call("POST", "/api/full/become-runner", `{}`, false); w.Code != 401 {
		t.Fatalf("needs a session: %d", w.Code)
	}
}
