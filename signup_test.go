package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestSignupCreatesPendingUserWithMemorableCode(t *testing.T) {
	s := newServiceHarness(t)
	w := s.call("POST", "/api/session/signup", `{"name":"  José  Pérez ","telephone":"6999-1234"}`, false)
	if w.Code != 201 {
		t.Fatalf("signup: %d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	code, _ := got["accessCode"].(string)
	if !regexp.MustCompile(`^[A-Z]{2,8}[0-9]{2}JOSE[0-9]$`).MatchString(code) || got["pending"] != true || got["name"] != "José Pérez" {
		t.Fatalf("word + 2 digits + first name + digit, pending: %v", got)
	}
	// The admin sees a pending user: no roles, a note, a code; the list never shows the code itself.
	list := s.call("GET", "/api/admin/users", "", true).Body.String()
	if !strings.Contains(list, "José Pérez") || !strings.Contains(list, `"roles":[]`) || !strings.Contains(list, "pendiente de aprobación") || strings.Contains(list, code) {
		t.Fatalf("pending user in the admin list: %s", list)
	}
	// Pending: the code is recognized (no wrong-code strike) but opens nothing.
	login := s.call("POST", "/api/session/code", `{"accessCode":"`+strings.ToLower(code)+`"}`, false)
	if login.Code != 403 || !strings.Contains(login.Body.String(), "pendiente de aprobación") || len(login.Result().Cookies()) != 0 {
		t.Fatalf("pending login: %d %s %v", login.Code, login.Body.String(), login.Result().Cookies())
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
	// Once the admin gives a role, the same code opens that panel.
	id := got["id"].(string)
	s.expect("PATCH", "/api/admin/users/"+id, `{"roles":["runner"]}`, 200)
	if w := s.call("POST", "/api/session/code", `{"accessCode":"`+code+`"}`, false); w.Code != 200 || !strings.Contains(w.Body.String(), "/runner.html") {
		t.Fatalf("approved login: %d %s", w.Code, w.Body.String())
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
