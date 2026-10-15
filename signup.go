package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Alta desde la página pública («Entrar» → «Crear mi código»): cualquiera crea su usuario con nombre y teléfono y el
// servidor le genera un código fácil de recordar pero difícil de adivinar: palabra + 2 dígitos + parte del nombre + 1 dígito,
// p. ej. LORO47JUAN3 (unas 80 × 100 × 10 combinaciones por nombre, y entrar sigue limitado a 10 códigos incorrectos cada
// 10 minutos por conexión). Queda aprobado de inmediato como corredor (Runners); administración puede cambiarle el acceso
// o desactivarlo en Usuarios.
var signupWords = []string{
	"AGUILA", "ALCE", "ARDILLA", "BALLENA", "BUHO", "BURRO", "CABRA", "CAIMAN", "CAMELLO", "CANGREJO",
	"CASTOR", "CEBRA", "CERDO", "CIERVO", "CISNE", "COBRA", "CONDOR", "CONEJO", "COYOTE", "DELFIN",
	"DRAGON", "ERIZO", "FOCA", "GACELA", "GALLO", "GANSO", "GATO", "GAVIOTA", "GORILA", "GRILLO",
	"HALCON", "HIENA", "HORMIGA", "IGUANA", "JAGUAR", "JIRAFA", "KOALA", "LEON", "LINCE", "LOBO",
	"LORO", "MANATI", "MANGO", "MAPACHE", "MONO", "MORSA", "NUTRIA", "OCELOTE", "OSO", "OVEJA",
	"PALOMA", "PANDA", "PANTERA", "PATO", "PELICANO", "PERICO", "PEZ", "PIÑA", "PINGUINO", "PUMA",
	"RANA", "RATON", "RENO", "SAPO", "SALMON", "TAPIR", "TIBURON", "TIGRE", "TOPO", "TORO",
	"TORTUGA", "TUCAN", "VENADO", "YAK", "ZORRO", "COCO", "LIMON", "MAREA", "OLA", "SOL",
}

func randomBelow(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(v.Int64())
}

// signupNamePart: hasta 5 letras del primer nombre, sin tildes ni signos («José» → JOSE); AMIGO si no queda ninguna.
func signupNamePart(name string) string {
	first := strings.Fields(name)
	if len(first) == 0 {
		return "AMIGO"
	}
	var b strings.Builder
	for _, c := range strings.ToUpper(foldText(first[0])) {
		if c >= 'A' && c <= 'Z' && b.Len() < 5 {
			b.WriteRune(c)
		}
	}
	if b.Len() < 2 {
		return "AMIGO"
	}
	return b.String()
}

func signupCode(name string) string {
	word := strings.ReplaceAll(signupWords[randomBelow(len(signupWords))], "Ñ", "N")
	return word + string(rune('0'+randomBelow(10))) + string(rune('0'+randomBelow(10))) + signupNamePart(name) + string(rune('0'+randomBelow(10)))
}


const signupNote = "Se registró solo (página pública o entrada de un panel), sin aprobación manual. Cambia su acceso o desactívalo si hace falta."

func (a *App) signupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/session/signup", func(w http.ResponseWriter, r *http.Request) {
		// Altas desde una misma conexión: como los códigos, 10 cada 10 minutos, para que nadie llene la lista de usuarios.
		key, now := "signup|"+clientIP(r), time.Now()
		a.mu.Lock()
		at := a.codeAttempts[key]
		if now.After(at.until) {
			at = attempt{}
		}
		if at.count >= codeMaxWrong {
			a.mu.Unlock()
			fail(w, 429, "Demasiadas cuentas creadas desde esta conexión. Espera unos minutos e intenta de nuevo.")
			return
		}
		a.mu.Unlock()
		f, ok := decodeFields(w, r, map[string]int{"name": 120, "telephone": 40, "email": 160, "panel": 16})
		if !ok {
			return
		}
		name := strings.Join(strings.Fields(fieldText(f, "name")), " ")
		phone := strings.Join(strings.Fields(fieldText(f, "telephone")), " ")
		if len(strings.Fields(name)) < 2 {
			failField(w, 400, "name", "Escribe tu nombre y apellido.")
			return
		}
		if len(phoneDigits(phone)) < 7 {
			failField(w, 400, "telephone", "Escribe un teléfono válido (por ejemplo 6000-0000).")
			return
		}
		// Correo opcional: sirve para que administración reconozca a la persona si pierde su código.
		email := normalizeEmail(fieldText(f, "email"))
		if email != "" && !validEmail(email) {
			failField(w, 400, "email", "Correo inválido (por ejemplo nombre@gmail.com).")
			return
		}
		// Desde la entrada de Constructores queda como constructor; desde cualquier otra, como corredor.
		role := roleRunner
		if fieldText(f, "panel") == "builder" {
			role = roleBuilder
		}
		u := User{ID: randomToken(), Name: name, Telephone: phone, Email: email, Roles: []string{role}, CreatedAt: nowStamp(), Note: signupNote}
		u.UpdatedAt = u.CreatedAt
		var code string
		err := a.db.Update(func(tx *bolt.Tx) error {
			// Un teléfono repetido se acepta: puede ser la misma persona que perdió su código. Al ponerla en un equipo, se
			// le pide su ID de usuario («Copiar mi ID») para saber cuál cuenta es cuál.
			names := map[string]bool{}
			if e := eachUser(tx, func(other User) error {
				names[foldText(other.Name)] = true
				return nil
			}); e != nil {
				return e
			}
			// Mismo nombre completo que otra persona: es otra cuenta; se le agrega un número para distinguirlas
			// («Juan Pérez», «Juan Pérez 2», «Juan Pérez 3»…).
			if names[foldText(u.Name)] {
				for n := 2; ; n++ {
					if try := fmt.Sprintf("%s %d", name, n); !names[foldText(try)] {
						u.Name = try
						break
					}
				}
			}
			for tries := 0; ; tries++ {
				code = signupCode(name)
				if _, e := userByAccessCode(tx, code); e == errMissing {
					break
				} else if e != nil {
					return e
				}
				if tries > 50 {
					return errBusyCode
				}
			}
			u.CodeHash, u.AccessCode = accessCodeHash(tx, code), code
			return putJSON(tx.Bucket(usersBucket), u.ID, u)
		})
		if err == nil {
			a.mu.Lock()
			at = a.codeAttempts[key]
			if now.After(at.until) {
				at = attempt{}
			}
			at.count++
			at.until = now.Add(codeWindow)
			a.codeAttempts[key] = at
			a.mu.Unlock()
		}
		if userError(w, err) {
			return
		}
		response(w, 201, map[string]any{"id": u.ID, "name": u.Name, "accessCode": code, "roles": u.Roles})
	})
}

var errBusyCode = userFieldError{"accessCode", "No se pudo generar un código libre. Intenta de nuevo."}
