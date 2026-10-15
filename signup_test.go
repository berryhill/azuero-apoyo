package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// Public sign-up: approved at once as a runner (no manual review); the generated code signs in right away.
func TestSignupCreatesActiveRunnerWithMemorableCode(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/session/signup", `{"name":"  José  Pérez ","telephone":"6999-1234"}`, false)
	if w.Code != 201 {
		t.Fatalf("signup: %d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	code, _ := got["accessCode"].(string)
	if !regexp.MustCompile(`^[A-Z]{2,8}[0-9]{2}JOSE[0-9]$`).MatchString(code) || got["name"] != "José Pérez" || got["pending"] != nil {
		t.Fatalf("word + 2 digits + first name + digit, not pending: %v", got)
	}
	// The admin sees an active runner with a note; the list never shows the code itself.
	list := s.call("GET", "/api/admin/users", "", true).Body.String()
	if !strings.Contains(list, "José Pérez") || !strings.Contains(list, `"roles":["runner"]`) || !strings.Contains(list, "sin aprobación manual") || strings.Contains(list, code) {
		t.Fatalf("active runner in the admin list: %s", list)
	}
	// The code opens the runner panel immediately (any case).
	if login := s.call("POST", "/api/session/code", `{"accessCode":"`+strings.ToLower(code)+`"}`, false); login.Code != 200 || !strings.Contains(login.Body.String(), "/runner.html") {
		t.Fatalf("signed in right away: %d %s", login.Code, login.Body.String())
	}
	// Same phone again (any format): refused without revealing the code.
	again := s.call("POST", "/api/session/signup", `{"name":"Otra Persona","telephone":"69991234"}`, false)
	if again.Code != 400 || !strings.Contains(again.Body.String(), "Ya hay una cuenta con ese teléfono") || strings.Contains(again.Body.String(), code) {
		t.Fatalf("duplicate phone: %d %s", again.Code, again.Body.String())
	}
	for body, field := range map[string]string{`{"name":"Solo","telephone":"6999-0000"}`: "name", `{"name":"Ana Ruiz","telephone":"12"}`: "telephone"} {
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
		body := `{"name":"Persona Prueba","telephone":"6000-00` + string(rune('0'+i/10)) + string(rune('0'+i%10)) + `"}`
		if w := s.call("POST", "/api/session/signup", body, false); w.Code != 201 {
			t.Fatalf("signup %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	if w := s.call("POST", "/api/session/signup", `{"name":"Una Más","telephone":"6111-1111"}`, false); w.Code != 429 {
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
	if w := s.call("POST", "/api/session/signup", `{"name":"Eva Correo","telephone":"6999-7777","email":"nope"}`, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"field":"email"`) {
		t.Fatalf("bad email: %d %s", w.Code, w.Body.String())
	}
	w := s.call("POST", "/api/session/signup", `{"name":"Eva Correo","telephone":"6999-7777","email":"  Eva@Gmail.COM "}`, false)
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
	if w := s.call("POST", "/api/session/signup", `{"name":"Sin Correo","telephone":"6999-8888"}`, false); w.Code != 201 {
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
	w := s.call("POST", "/api/session/signup", `{"name":"Beto Obra","telephone":"6999-5555","panel":"builder"}`, false)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"roles":["builder"]`) {
		t.Fatalf("builder signup: %d %s", w.Code, w.Body.String())
	}
	if w := s.call("POST", "/api/session/signup", `{"name":"Ana Corre","telephone":"6999-6666","panel":"admin"}`, false); w.Code != 201 || !strings.Contains(w.Body.String(), `"roles":["runner"]`) {
		t.Fatalf("anything else is a runner: %d %s", w.Code, w.Body.String())
	}
}

// Same full name as someone else (any case or accents): a separate account, told apart by a number.
func TestSignupSameNameGetsANumber(t *testing.T) {
	s := newServiceHarness(t)
	names := []string{}
	for i, body := range []string{
		`{"name":"Juan Pérez","telephone":"6100-0001"}`,
		`{"name":"juan  perez","telephone":"6100-0002"}`,
		`{"name":"Juan Pérez","telephone":"6100-0003"}`,
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
