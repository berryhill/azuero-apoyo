package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	bolt "go.etcd.io/bbolt"
)

// Códigos de punto por zona: abreviatura del área + número correlativo («AB-7» = Agua Buena n.º 7), como los que se
// pusieron a mano (AB-1…AB-6, OA-1, S-1…). Las variantes de una misma zona («Tonosi» / «Tonosí», «El Cacao» / «cacao»)
// comparten abreviatura; una zona que ya tiene códigos sigue con la suya y su numeración. Un punto nuevo recibe su
// código al crearse; los que todavía muestran el código automático («P-a591a94b6343») se arreglan desde Administración →
// Mantenimiento → «Generar códigos de puntos» (vista previa y luego aplicar). «📍 Por ubicar» no recibe código de zona
// hasta que se ubica.

var (
	autoCodeRe     = regexp.MustCompile(`^P-[0-9a-f]{12}$`)
	areaCodeRe     = regexp.MustCompile(`^(.+)-(\d+)$`)
	codeInAreaRe   = regexp.MustCompile(`^([A-Za-z]{1,4})-(\d+)\s+(.+)$`) // «FL-3 Flores»: el código se escribió en la zona
	areaStopWords  = map[string]bool{"el": true, "la": true, "los": true, "las": true, "de": true, "del": true, "y": true}
	reservedPrefix = map[string]bool{"P": true} // «P-…» es el código automático
)

// needsAreaCode: sin código o con el automático.
func needsAreaCode(code string) bool { code = strings.TrimSpace(code); return code == "" || autoCodeRe.MatchString(code) }

// areaWords: las palabras de la zona sin tildes, en minúsculas y sin artículos («El Cacao» → [cacao]).
func areaWords(area string) []string {
	var b strings.Builder
	for _, r := range foldText(area) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteRune(' ')
		}
	}
	var out []string
	for _, w := range strings.Fields(b.String()) {
		if !areaStopWords[w] {
			out = append(out, w)
		}
	}
	return out
}

func areaKey(area string) string { return strings.Join(areaWords(area), " ") }

// codeBook: qué abreviatura usa cada zona y el último número de cada abreviatura, a partir de los códigos existentes.
type codeBook struct {
	owner  map[string]string         // abreviatura → zona (areaKey)
	prefix map[string]string         // zona → abreviatura
	last   map[string]int            // abreviatura → último número usado
	used   map[string]bool           // códigos usados (en mayúsculas)
	votes  map[string]map[string]int // zona → abreviatura → cuántos puntos la usan
}

func newCodeBook(tx *bolt.Tx) *codeBook {
	cb := &codeBook{owner: map[string]string{}, prefix: map[string]string{}, last: map[string]int{}, used: map[string]bool{}, votes: map[string]map[string]int{}}
	tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if json.Unmarshal(v, &p) != nil {
			return nil
		}
		// Un código escrito en la zona («RV-4 Río Viejo») cuenta como usado aunque el punto todavía tenga el automático.
		if m := codeInAreaRe.FindStringSubmatch(strings.TrimSpace(p.Area)); m != nil {
			cb.note(areaKey(m[3]), strings.ToUpper(m[1]), m[2])
		}
		code := strings.TrimSpace(p.Code)
		if code == "" {
			return nil
		}
		cb.used[strings.ToUpper(code)] = true
		if m := areaCodeRe.FindStringSubmatch(code); m != nil && !autoCodeRe.MatchString(code) {
			cb.note(areaKey(p.Area), strings.ToUpper(m[1]), m[2])
		}
		return nil
	})
	// Cada zona se queda con la abreviatura que más usan sus puntos (a igualdad, la que sale de sus iniciales).
	zones := make([]string, 0, len(cb.votes))
	for z := range cb.votes {
		zones = append(zones, z)
	}
	sort.Strings(zones)
	for _, z := range zones {
		best, n := "", -1
		for pre, c := range cb.votes[z] {
			if c > n || c == n && (pre == initials(z) || best != initials(z) && pre < best) {
				best, n = pre, c
			}
		}
		cb.prefix[z] = best
		for pre := range cb.votes[z] {
			if _, taken := cb.owner[pre]; !taken {
				cb.owner[pre] = z
			}
		}
	}
	return cb
}

func (cb *codeBook) note(zone, pre, num string) {
	n, _ := strconv.Atoi(num)
	if n > cb.last[pre] {
		cb.last[pre] = n
	}
	if zone == "" {
		return
	}
	if cb.votes[zone] == nil {
		cb.votes[zone] = map[string]int{}
	}
	cb.votes[zone][pre]++
}

func initials(zone string) string {
	var b strings.Builder
	for i, w := range strings.Fields(zone) {
		if i < 3 {
			b.WriteString(strings.ToUpper(w[:1]))
		}
	}
	return b.String()
}

// prefixFor: la abreviatura de la zona; si es nueva, la primera libre de: iniciales, primeras letras de la primera
// palabra (2, 3, 4) + iniciales de las demás.
func (cb *codeBook) prefixFor(zone string) string {
	if pre, ok := cb.prefix[zone]; ok {
		return pre
	}
	words := strings.Fields(zone)
	if len(words) == 0 {
		return ""
	}
	var cands []string
	rest := ""
	for i, w := range words[1:] {
		if i < 2 {
			rest += strings.ToUpper(w[:1])
		}
	}
	first := strings.ToUpper(asciiLetters(words[0]))
	for n := 1; n <= len(first) && n <= 4; n++ {
		cands = append(cands, first[:n]+rest)
	}
	for i := 2; i < 10; i++ {
		cands = append(cands, cands[0]+strconv.Itoa(i))
	}
	for _, c := range cands {
		if c == "" || reservedPrefix[c] {
			continue
		}
		if owner, taken := cb.owner[c]; !taken || owner == zone {
			cb.owner[c], cb.prefix[zone] = zone, c
			return c
		}
	}
	return ""
}

func asciiLetters(w string) string {
	var b strings.Builder
	for _, r := range w {
		if r < 128 && unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "X"
	}
	return b.String()
}

// next: el siguiente código libre de la zona («AB-7»), o "" si la zona no sirve («Por ubicar», «Sin ubicación»…).
func (cb *codeBook) next(area string) string {
	zone := areaKey(area)
	if zone == "" || zone == "por ubicar" {
		return ""
	}
	pre := cb.prefixFor(zone)
	if pre == "" {
		return ""
	}
	for {
		cb.last[pre]++
		code := pre + "-" + strconv.Itoa(cb.last[pre])
		if !cb.used[strings.ToUpper(code)] {
			cb.used[strings.ToUpper(code)] = true
			return code
		}
	}
}

// codeFor: el código de zona para un punto nuevo o recién ubicado; el automático si todavía no hay zona («Por ubicar»).
func codeFor(tx *bolt.Tx, p *Point) {
	if !needsAreaCode(p.Code) {
		return
	}
	if !p.ToLocate {
		if code := newCodeBook(tx).next(p.Area); code != "" {
			p.Code = code
			return
		}
	}
	if p.Code == "" {
		p.Code = "P-" + p.ID[:12]
	}
}

// uniqueCode: ningún otro punto (salvo los rechazados) puede tener el mismo código, sin importar mayúsculas. Se
// revisa al editar un punto, al pedir un cambio y al fusionar (al crear lo revisa createPoint; los códigos de zona
// generados nunca repiten uno usado).
func uniqueCode(tx *bolt.Tx, p *Point) error {
	code := strings.TrimSpace(p.Code)
	if code == "" {
		return nil
	}
	taken := false
	tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var o Point
		if !taken && json.Unmarshal(v, &o) == nil && o.ID != p.ID && o.Status != "rejected" && strings.EqualFold(strings.TrimSpace(o.Code), code) {
			taken = true
		}
		return nil
	})
	if taken {
		return userFieldError{"code", "Ya hay otro punto con el código «" + code + "». Usa otro."}
	}
	return nil
}

type codeChange struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Area    string `json:"area"`
	Place   string `json:"place,omitempty"`
	Old     string `json:"old"`
	New     string `json:"new"`
	NewArea string `json:"newArea,omitempty"` // el código estaba escrito en la zona: se quita de ahí
}

// planCodes: los puntos (aprobados o por revisar, ya ubicados) que siguen sin código o con el automático, del más
// antiguo al más nuevo, con el código que les toca. apply los guarda (cada cambio queda en el historial del punto).
func planCodes(tx *bolt.Tx, apply bool) ([]codeChange, error) {
	cb := newCodeBook(tx)
	var pts []Point
	if e := tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if json.Unmarshal(v, &p) == nil && (p.Status == "approved" || p.Status == "pending") && !p.ToLocate && p.MergedInto == "" && needsAreaCode(p.Code) {
			pts = append(pts, p)
		}
		return nil
	}); e != nil {
		return nil, e
	}
	sort.SliceStable(pts, func(i, j int) bool {
		if pts[i].Created != pts[j].Created {
			return pts[i].Created < pts[j].Created
		}
		return pts[i].ID < pts[j].ID
	})
	changes := []codeChange{}
	for _, p := range pts {
		c := codeChange{ID: p.ID, Status: p.Status, Area: p.Area, Place: p.Place, Old: p.Code}
		if m := codeInAreaRe.FindStringSubmatch(strings.TrimSpace(p.Area)); m != nil {
			// «FL-3 Flores»: ese es su código (si nadie más lo tiene) y la zona queda «Flores».
			if code := strings.ToUpper(m[1]) + "-" + m[2]; !cb.used[code] {
				cb.used[code] = true
				c.New, c.NewArea = code, strings.TrimSpace(m[3])
			}
		}
		if c.New == "" {
			if c.New = cb.next(p.Area); c.New == "" {
				continue
			}
		}
		changes = append(changes, c)
		if apply {
			p.Code = c.New
			if c.NewArea != "" {
				p.Area = c.NewArea
			}
			if e := putPointJSON(tx.Bucket(submissions), p); e != nil {
				return nil, e
			}
		}
	}
	return changes, nil
}

// suggestCodes: códigos LIBRES para escribir en «Código» (cada uno una sola vez). Con letras escritas («S», «s-», «S-3»)
// sigue esa abreviatura; sin nada, la de la zona del punto. El código actual del punto va primero si coincide.
func suggestCodes(tx *bolt.Tx, q, area, current string) []string {
	cb := newCodeBook(tx)
	q = strings.ToUpper(strings.TrimSpace(q))
	pre := strings.TrimRight(strings.TrimRight(q, "0123456789"), "-")
	if pre == "" {
		pre = cb.prefixFor(areaKey(area))
	}
	out, seen := []string{}, map[string]bool{}
	add := func(c string) {
		if c != "" && !seen[c] && len(out) < 8 {
			seen[c] = true
			out = append(out, c)
		}
	}
	if cur := strings.ToUpper(strings.TrimSpace(current)); cur != "" && (q == "" || strings.HasPrefix(cur, q)) && !autoCodeRe.MatchString(current) {
		add(strings.TrimSpace(current))
	}
	if pre == "" || reservedPrefix[pre] {
		return out
	}
	// Lo escrito con número («S-9»), si está libre.
	if m := areaCodeRe.FindStringSubmatch(q); m != nil && !cb.used[q] {
		add(q)
	}
	for n := cb.last[pre] + 1; len(out) < 6; n++ {
		if c := pre + "-" + strconv.Itoa(n); !cb.used[c] {
			add(c)
		}
	}
	return out
}

func (a *App) codeRoutes(mux *http.ServeMux) {
	// Sugerencias para el campo «Código» (Administración y Checkup): solo códigos libres, sin repetir.
	mux.HandleFunc("GET /api/admin/maintenance/codes/suggest", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		q := r.URL.Query()
		var list []string
		a.db.View(func(tx *bolt.Tx) error {
			current := ""
			if id := q.Get("id"); id != "" {
				var p Point
				if v := tx.Bucket(submissions).Get([]byte(id)); v != nil && json.Unmarshal(v, &p) == nil {
					current = p.Code
				}
			}
			list = suggestCodes(tx, q.Get("q"), q.Get("area"), current)
			return nil
		})
		response(w, 200, list)
	})
	// Vista previa: qué código recibiría cada punto, sin cambiar nada.
	mux.HandleFunc("GET /api/admin/maintenance/codes", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		var list []codeChange
		err := a.db.View(func(tx *bolt.Tx) (e error) { list, e = planCodes(tx, false); return e })
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/admin/maintenance/codes", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		var list []codeChange
		err := a.write(r, func(tx *bolt.Tx) (e error) { list, e = planCodes(tx, true); return e })
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"changed": len(list), "changes": list})
	})
}
