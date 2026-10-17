package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	bolt "go.etcd.io/bbolt"
)

// Vista previa al compartir (WhatsApp, Facebook, Telegram…): cada página lleva etiquetas Open Graph con la necesidad
// abierta más reciente y su imagen /og/latest.png. Solo datos públicos: los mismos filtros del mapa (punto aprobado y
// visible, sin reparaciones ni runs de comida ni servicios programados, sin la carrera reservada de un equipo) y nada
// de contactos.

const ogW, ogH = 1200, 630

// ogNeed: lo que muestra la vista previa (la necesidad más reciente, o un punto con todo lo que necesita).
type ogNeed struct {
	ID, Code, Area, Place, Coordinator, People, Urgency, CreatedAt string
	Breakdown, Items                                               []string
	// Counts: el desglose para dibujar, como en la tarjeta de la app: {ícono, cantidad, etiqueta en minúscula}.
	Counts [][3]string
	Point  bool
	// GPS: el punto tiene coordenadas (pin del mapa y «Abrir en Maps» activos, como en la tarjeta).
	GPS bool
}

// ogCounts: el desglose de personas con las etiquetas de la ficha.
var ogCounts = [][4]string{{"Adultos", "adults", "adult", "adultos"}, {"Niños", "children", "child", "niños"}, {"Bebés", "babies", "baby", "bebés"}, {"Mayores", "olderAdults", "elder", "mayores"}, {"Discapacitados", "disabled", "disabled", "discapacidad"}, {"Familias", "families", "family", "familias"}}

func ogFromPoint(p Point) ogNeed {
	n := ogNeed{Code: p.Code, Area: p.Area, Place: p.Place, Coordinator: strings.TrimSpace(p.Coordinator), People: strings.TrimSpace(p.People), GPS: len(p.Coords) == 2}
	for _, c := range ogCounts {
		v := map[string]string{"adults": p.Adults, "children": p.Children, "babies": p.Babies, "olderAdults": p.OlderAdults, "disabled": p.Disabled, "families": p.Families}[c[1]]
		if v = strings.TrimSpace(v); v != "" && v != "0" {
			n.Breakdown = append(n.Breakdown, v+" "+c[0])
			n.Counts = append(n.Counts, [3]string{c[2], v, c[3]})
		}
	}
	return n
}

// publicOpenNeeds: las necesidades Abiertas de un punto tal como las ve el público (sin reparaciones, comida ni
// servicios programados; la carrera reservada de un equipo sin texto, así que no cuenta).
func publicOpenNeeds(tx *bolt.Tx, campID string, each func(CampNeed)) {
	tv := newTeamView(tx, "")
	tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var n CampNeed
		if json.Unmarshal(v, &n) != nil || n.Status != "OPEN" || isMealKind(n) || campID != "" && n.CampID != campID {
			return nil
		}
		if n, keep := viewerNeed(tx, tv, "", n); keep && n.Need != "" {
			each(n)
		}
		return nil
	})
}

// latestPublicNeed: la necesidad Abierta más reciente (por creación) que el público puede ver; ok=false si no hay.
func latestPublicNeed(tx *bolt.Tx) (ogNeed, bool) {
	points := map[string]Point{}
	tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if json.Unmarshal(v, &p) == nil && p.Status == "approved" && !hiddenFromPublic(p, "") {
			points[p.ID] = p
		}
		return nil
	})
	var best ogNeed
	found := false
	publicOpenNeeds(tx, "", func(n CampNeed) {
		p, ok := points[n.CampID]
		if !ok || found && n.CreatedAt <= best.CreatedAt {
			return
		}
		best, found = ogFromPoint(p), true
		best.ID, best.Urgency, best.CreatedAt, best.Items = n.ID, n.Urgency, n.CreatedAt, needItems(n.Need)
	})
	return best, found
}

// findPublicPoint: el punto por su código (sin distinguir mayúsculas) o su ID, si el público puede verlo.
func findPublicPoint(tx *bolt.Tx, ref string) (Point, bool) {
	var hit Point
	found := false
	visible := func(p Point) bool { return p.Status == "approved" && !hiddenFromPublic(p, "") }
	if v := tx.Bucket(submissions).Get([]byte(ref)); v != nil && json.Unmarshal(v, &hit) == nil && visible(hit) {
		return hit, true
	}
	tx.Bucket(submissions).ForEach(func(_, v []byte) error {
		var p Point
		if !found && json.Unmarshal(v, &p) == nil && p.Code != "" && strings.EqualFold(p.Code, ref) && visible(p) {
			hit, found = p, true
		}
		return nil
	})
	return hit, found
}

// pointCard: la vista previa de un punto: dónde, cuántas personas (con desglose) y todo lo que necesita ahora.
func pointCard(tx *bolt.Tx, p Point) ogNeed {
	n := ogFromPoint(p)
	n.ID, n.Point, n.CreatedAt = p.ID, true, ""
	seen := map[string]bool{}
	publicOpenNeeds(tx, p.ID, func(x CampNeed) {
		if normalizeUrgency(x.Urgency) == urgencyUrgente {
			n.Urgency = urgencyUrgente
		}
		if x.CreatedAt > n.CreatedAt {
			n.CreatedAt = x.CreatedAt
		}
		for _, it := range needItems(x.Need) {
			if k := foldText(it); !seen[k] {
				seen[k] = true
				n.Items = append(n.Items, it)
			}
		}
	})
	return n
}

// version cambia cuando cambia lo que muestra la tarjeta: va en la URL de la imagen para que las apps no usen una vieja.
func (n ogNeed) version() string {
	h := fnv.New64a()
	fmt.Fprintf(h, "%#v", n)
	return fmt.Sprintf("%x", h.Sum64())[:12]
}

// title: título de la vista previa de un punto («Punto AB-7 · Agua buena»).
func (n ogNeed) title() string {
	where := n.Area
	if where == "" {
		where = n.Place
	}
	// El código automático (P-3f9a…) no dice nada a quien recibe el enlace: solo el lugar.
	if n.Code == "" || autoCodeRe.MatchString(n.Code) {
		return strings.Join(nonEmpty(where, "Punto de apoyo"), " · ")
	}
	return strings.Join(nonEmpty("Punto "+n.Code, where), " · ")
}

// peopleLine: «12 personas · 3 Adultos · 2 Niños».
func (n ogNeed) peopleLine() string {
	parts := []string{}
	if n.People != "" && n.People != "0" {
		word := "personas"
		if n.People == "1" {
			word = "persona"
		}
		parts = append(parts, n.People+" "+word)
	}
	return strings.Join(append(parts, n.Breakdown...), " · ")
}

// description: texto de la vista previa.
func (n ogNeed) description() string {
	if n.ID == "" {
		return "Mapa de puntos de apoyo y sus necesidades en Azuero. Reporta un lugar que necesita ayuda."
	}
	what := strings.Join(n.Items, ", ")
	if len([]rune(what)) > 160 {
		what = string([]rune(what)[:157]) + "…"
	}
	parts := []string{}
	if normalizeUrgency(n.Urgency) == urgencyUrgente {
		parts = append(parts, "URGENTE")
	}
	if !n.Point {
		parts = append(parts, strings.Join(nonEmpty(n.Code, n.Area, n.Place), " · "))
	} else if n.Place != "" && n.Place != n.Area {
		parts = append(parts, n.Place)
	}
	if pl := n.peopleLine(); pl != "" {
		parts = append(parts, pl)
	}
	if what == "" {
		what = "sin necesidades abiertas ahora"
	}
	parts = append(parts, "Necesitan: "+what)
	return strings.Join(parts, " — ")
}

func nonEmpty(v ...string) []string {
	out := []string{}
	for _, s := range v {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

var (
	ogCacheMu sync.Mutex
	ogCache   = map[string][]byte{}
	ogTitleRe = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
)

func (a *App) ogLatest() ogNeed {
	var n ogNeed
	a.db.View(func(tx *bolt.Tx) error {
		n, _ = latestPublicNeed(tx)
		return nil
	})
	return n
}

// ogPoint: la tarjeta de un punto por código o ID; ok=false si no existe o el público no puede verlo.
func (a *App) ogPoint(ref string) (ogNeed, bool) {
	var n ogNeed
	found := false
	a.db.View(func(tx *bolt.Tx) error {
		if p, ok := findPublicPoint(tx, ref); ok {
			n, found = pointCard(tx, p), true
		}
		return nil
	})
	return n, found
}

// ogPNG: la imagen de la tarjeta, guardada en memoria mientras no cambie.
func (a *App) ogPNG(n ogNeed) ([]byte, error) {
	key := n.ID + "/" + n.version()
	ogCacheMu.Lock()
	b, ok := ogCache[key]
	ogCacheMu.Unlock()
	if ok {
		return b, nil
	}
	b, err := ogImage(n)
	if err != nil {
		return nil, err
	}
	ogCacheMu.Lock()
	if len(ogCache) >= 200 {
		ogCache = map[string][]byte{}
	}
	ogCache[key] = b
	ogCacheMu.Unlock()
	return b, nil
}

// ogTags: etiquetas Open Graph y Twitter para la página pedida (URLs absolutas, como piden los rastreadores).
func ogTags(r *http.Request, title string, n ogNeed) string {
	scheme := "http"
	if secureCookie(r) {
		scheme = "https"
	}
	base := scheme + "://" + r.Host
	img := base + "/og/latest.png?v=" + n.version()
	alt := "Necesidad más reciente en Azuero Apoyo"
	switch {
	case n.Point:
		img, alt = base+"/og/p/"+url.PathEscape(n.ID)+".png?v="+n.version(), n.title()
	case n.ID == "":
		alt = "Mapa de apoyo de Azuero"
	}
	e := html.EscapeString
	desc := e(n.description())
	tags := `<meta property="og:type" content="website"><meta property="og:site_name" content="Azuero Apoyo"><meta property="og:locale" content="es_PA">` +
		`<meta property="og:title" content="` + e(title) + `"><meta property="og:description" content="` + desc + `"><meta name="description" content="` + desc + `">` +
		`<meta property="og:url" content="` + e(base+r.URL.RequestURI()) + `">` +
		`<meta property="og:image" content="` + e(img) + `"><meta property="og:image:secure_url" content="` + e(img) + `"><meta property="og:image:type" content="image/png">` +
		fmt.Sprintf(`<meta property="og:image:width" content="%d"><meta property="og:image:height" content="%d">`, ogW, ogH) +
		`<meta property="og:image:alt" content="` + e(alt) + `">` +
		`<meta name="twitter:card" content="summary_large_image"><meta name="twitter:title" content="` + e(title) + `"><meta name="twitter:description" content="` + desc + `"><meta name="twitter:image" content="` + e(img) + `">`
	return tags
}

// serveHTML envía la página con las etiquetas Open Graph de la necesidad más reciente justo antes de </head>.
func (a *App) serveHTML(w http.ResponseWriter, r *http.Request, path string) {
	a.servePage(w, r, path, nil)
}

// servePage envía la página con las etiquetas Open Graph justo antes de </head>: las del punto si card no es nil (el
// título de la pestaña también dice qué punto es); si no, las de la necesidad más reciente.
func (a *App) servePage(w http.ResponseWriter, r *http.Request, path string, card *ogNeed) {
	page, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	title := "Azuero Apoyo"
	if m := ogTitleRe.FindSubmatch(page); m != nil {
		title = html.UnescapeString(strings.TrimSpace(string(m[1])))
	}
	n := ogNeed{}
	if card != nil {
		n, title = *card, card.title()
		page = ogTitleRe.ReplaceAll(page, []byte("<title>"+html.EscapeString(title+" · Azuero")+"</title>"))
	} else {
		n = a.ogLatest()
	}
	if i := bytes.Index(page, []byte("</head>")); i >= 0 {
		page = append(page[:i:i], append([]byte(ogTags(r, title, n)), page[i:]...)...)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(page)
}

func (a *App) ogRoutes(mux *http.ServeMux) {
	png := func(w http.ResponseWriter, n ogNeed) {
		b, err := a.ogPNG(n)
		if err != nil {
			fail(w, 500, "image error")
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Write(b)
	}
	mux.HandleFunc("GET /og/latest.png", func(w http.ResponseWriter, r *http.Request) { png(w, a.ogLatest()) })
	mux.HandleFunc("GET /og/p/{file}", func(w http.ResponseWriter, r *http.Request) {
		n, ok := a.ogPoint(strings.TrimSuffix(r.PathValue("file"), ".png"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		png(w, n)
	})
	// Enlace para compartir una donación: /runner/<código> (o su ID) es el inicio, igual que para el público, con la
	// pestaña «📦 Esta donación» de ese punto (index.html la arma desde la dirección; quien ya es runner pasa a Runners) y
	// la vista previa del punto (WhatsApp, Facebook…). Un punto desconocido u oculto abre el inicio con la vista previa
	// general. /p/<código> (enlaces ya compartidos) lleva aquí.
	// /supply/<código>: el enlace corto que comparte Supply (como /runner/<código>): abre Supply con la búsqueda en ese
	// punto, con su tarjeta de vista previa; sin sesión, primero se entra con el código y se llega ahí mismo.
	mux.HandleFunc("GET /supply/{ref}", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(a.static, "supplier.html")
		if n, ok := a.ogPoint(r.PathValue("ref")); ok {
			a.servePage(w, r, path, &n)
			return
		}
		a.servePage(w, r, path, nil)
	})
	mux.HandleFunc("GET /runner/{ref}", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(a.static, "index.html")
		if n, ok := a.ogPoint(r.PathValue("ref")); ok {
			a.servePage(w, r, path, &n)
			return
		}
		a.servePage(w, r, path, nil)
	})
	mux.HandleFunc("GET /p/{ref}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/runner/"+url.PathEscape(r.PathValue("ref")), http.StatusMovedPermanently)
	})
}
