package main

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

// Imagen de la vista previa al compartir (1200×630): tres líneas grandes y juntas, legibles en miniatura (WhatsApp,
// Facebook…): el lugar; qué necesita (íconos y nombres cortos, «+N» lo que no cabe); y las personas, solo íconos y
// números. Abajo, la barra «¡Yo lo llevo!». Íconos y colores copiados de pointform.js.

// ogIcons: los íconos de pointform.js (trazo 2 en una cuadrícula de 24), los que usa la tarjeta.
var ogIcons = map[string]string{
	"pin":       `<path d="M12 21s-7-6.2-7-11a7 7 0 0 1 14 0c0 4.8-7 11-7 11z"/><circle cx="12" cy="10" r="2.5"/>`,
	"megaphone": `<path d="M3 11v2a1 1 0 0 0 1 1h2l5 4V6L6 10H4a1 1 0 0 0-1 1z"/><path d="M15 8.5a5 5 0 0 1 0 7"/><path d="M18 6a8.5 8.5 0 0 1 0 12"/>`,
	"people":    `<circle cx="9" cy="8" r="3"/><path d="M3 20a6 6 0 0 1 12 0"/><circle cx="17" cy="9" r="2.5"/><path d="M15.5 14.2A5 5 0 0 1 21 19"/>`,
	"adult":     `<circle cx="12" cy="4.5" r="2.5"/><path d="M12 8v7M8 11h8M12 15l-3 6M12 15l3 6"/>`,
	"child":     `<circle cx="12" cy="8" r="2.5"/><path d="M12 11v5M9 13h6M12 16l-2 4M12 16l2 4"/>`,
	"baby":      `<circle cx="12" cy="12" r="8"/><path d="M9.5 11h.01M14.5 11h.01M9.5 14.5a3.5 3.5 0 0 0 5 0"/>`,
	"family":    `<circle cx="7" cy="6" r="2.5"/><circle cx="17" cy="6" r="2.5"/><circle cx="12" cy="12" r="2"/><path d="M3 21v-5a4 4 0 0 1 8 0"/><path d="M13 21v-5a4 4 0 0 1 8 0"/><path d="M9.5 21v-2.5a2.5 2.5 0 0 1 5 0V21"/>`,
	"elder":     `<circle cx="10" cy="4.5" r="2.2"/><path d="M10 7.5l-1 6 3 3v5M9 13.5l-3 7.5M10 9l4 2M16 11v10"/>`,
	"disabled":  `<circle cx="11" cy="4" r="2"/><path d="M11 7v6h5l3 6M11 10h5"/><path d="M8 11.5a5 5 0 1 0 6.5 6.5"/>`,
	"drop":      `<path d="M12 3s6 7 6 11a6 6 0 0 1-12 0c0-4 6-11 6-11z"/>`,
	"food":      `<path d="M7 3v8M5 3v5a2 2 0 0 0 4 0V3M7 11v10M17 3c-2 2-2 6 0 8v10"/>`,
	"tent":      `<path d="M3 20L12 4l9 16H3zM9 20l3-6 3 6"/>`,
	"pot":       `<path d="M4 10h16v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-6zM2 10h2M20 10h2M9 6c0-1 1-1 1-2M14 6c0-1 1-1 1-2"/>`,
	"bottle":    `<path d="M10 2h4M10 5h4l1 3v12a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2V8l1-3zM9 12h6"/>`,
	"fuel":      `<path d="M4 21V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v16M3 21h12M14 9h2a2 2 0 0 1 2 2v6a1.5 1.5 0 0 0 3 0V8l-3-3M7 7h4"/>`,
	"sim":       `<path d="M7 3h7l5 5v12a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/><path d="M9 11h6v6H9z"/><path d="M12 11v6M9 14h6"/>`,
	"gas":       `<path d="M9 3h6v3H9z"/><path d="M7 6h10a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2z"/><path d="M12 10c1.5 1.6 2 2.7 2 3.7a2 2 0 0 1-4 0c0-1 .5-2.1 2-3.7z"/>`,
	"other":     `<circle cx="12" cy="12" r="9"/><path d="M12 8v8M8 12h8"/>`,
	"calendar":  `<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M3 10h18M8 3v4M16 3v4"/>`,
	"share":     `<circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4"/>`,
	"nav":       `<path d="M3 11l18-8-8 18-2-8z"/>`,
	"nogps":     `<path d="M12 22s7-6.1 7-12a7 7 0 0 0-12.3-4.6"/><path d="M5.2 7.6A7 7 0 0 0 5 10c0 5.9 7 12 7 12"/><circle cx="12" cy="10" r="2.5"/><path d="M3 3l18 18"/>`,
}

// ogNeedIcons: NEEDS de pointform.js (nombre → ícono). La reparación nunca sale en público.
var ogNeedIcons = [][2]string{{"Agua", "drop"}, {"Alimentos", "food"}, {"Refugio (carpa / lona / toldo)", "tent"}, {"Kit de cocina", "pot"}, {"Kit de bebé", "bottle"}, {"Combustible para generador", "fuel"}, {"Tarjeta de datos prepago", "sim"}, {"Gas de cocina", "gas"}}

// ogBrand: una opción de una necesidad (BRANDED de pointform.js): nombre, texto de la insignia, alias y colores
// (.pf-net-*: fondo, texto y la raya de abajo si la marca la tiene).
type ogBrand struct {
	name, badge string
	alias       *regexp.Regexp
	bg, fg, bar string
}

var ogBranded = map[string][]ogBrand{
	"Agua":                       {{"Para beber", "para beber", nil, "#d8ecfb", "#134d7a", ""}, {"Para cocinar", "para cocinar", nil, "#d9f2ea", "#14594a", ""}},
	"Alimentos":                  {{"Comida preparada", "comida preparada", nil, "#fde3cf", "#8a3a09", ""}, {"Ingredientes para cocinar", "ingredientes", nil, "#e6f3d5", "#3c5a14", ""}},
	"Combustible para generador": {{"Gasolina", "Gasolina", nil, "#f2a900", "#2b1d00", ""}, {"Diésel", "Diésel", nil, "#2f3a40", "#ffffff", ""}},
	"Tarjeta de datos prepago":   {{"Tigo", "tigo", nil, "#00377b", "#ffffff", "#00c8ff"}, {"+Móvil", "+móvil", regexp.MustCompile(`^(mas|\+) ?movil$|^claro$`), "#0153ba", "#ffffff", ""}},
	"Gas de cocina":              {{"Tropigas", "Tropigas", nil, "#e35205", "#ffffff", ""}, {"Panagas", "Panagas", nil, "#272974", "#ffffff", "#e11e26"}},
}

var (
	ogBrandedRe = func() *regexp.Regexp {
		names := []string{}
		for l := range ogBranded {
			names = append(names, regexp.QuoteMeta(l))
		}
		return regexp.MustCompile(`(?i)^\s*(` + strings.Join(names, "|") + `)\s*(?:\(([^)]*)\))?\s*$`)
	}()
	ogLegacy = [][3]string{{`(?i)^\s*gasolina para generador\s*$`, "Combustible para generador", "Gasolina"}, {`(?i)^\s*di[eé]sel para generador\s*$`, "Combustible para generador", "Diésel"}}
)

// ogChip: una ficha de «Necesita»: ícono (vacío si es texto libre), nombre e insignias.
type ogChip struct {
	icon, label string
	badges      []ogBrand
	text        bool
}

// ogChipOf lee una parte del texto de necesidades como en pointform.js (brandOf / needChipsHtml).
func ogChipOf(part string) ogChip {
	label, brand := "", ""
	for _, l := range ogLegacy {
		if regexp.MustCompile(l[0]).MatchString(part) {
			label, brand = l[1], l[2]
		}
	}
	if label == "" {
		if m := ogBrandedRe.FindStringSubmatch(part); m != nil {
			for l := range ogBranded {
				if strings.EqualFold(l, m[1]) {
					label = l
				}
			}
			brand = strings.TrimSpace(m[2])
		}
	}
	if label == "" {
		for _, n := range ogNeedIcons {
			if foldText(n[0]) == foldText(part) {
				return ogChip{icon: n[1], label: n[0]}
			}
		}
		return ogChip{label: part, text: true}
	}
	c := ogChip{icon: "other", label: label}
	for _, n := range ogNeedIcons {
		if n[0] == label {
			c.icon = n[1]
		}
	}
	for _, b := range strings.Split(brand, " / ") {
		if b = strings.TrimSpace(b); b == "" {
			continue
		}
		hit := ogBrand{name: b, badge: b, bg: "#e7eef0", fg: "#36505a"}
		for _, x := range ogBranded[label] {
			if foldText(x.name) == foldText(b) || foldText(x.badge) == foldText(b) || x.alias != nil && x.alias.MatchString(foldText(b)) {
				hit = x
			}
		}
		c.badges = append(c.badges, hit)
	}
	return c
}

var (
	ogFontsOnce       sync.Once
	ogBold, ogRegular *opentype.Font
	ogFaces           = map[string]font.Face{}
	ogFacesMu         sync.Mutex
)

func ogFace(bold bool, size float64) font.Face {
	ogFontsOnce.Do(func() {
		ogBold, _ = opentype.Parse(gobold.TTF)
		ogRegular, _ = opentype.Parse(goregular.TTF)
	})
	key := fmt.Sprint(bold, size)
	ogFacesMu.Lock()
	defer ogFacesMu.Unlock()
	if f, ok := ogFaces[key]; ok {
		return f
	}
	f := ogRegular
	if bold {
		f = ogBold
	}
	face, _ := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	ogFaces[key] = face
	return face
}

// Colores de la tarjeta de la app.
const (
	ogBg     = "#f2f6f5"
	ogLine   = "#dfe8e7"
	ogInk    = "#173c46"
	ogText2  = "#1f3a46"
	ogMuted  = "#5d7680"
	ogTeal   = "#2d7f86"
	ogLink   = "#126e7c"
	ogChipBd = "#d5e0e2"
	ogRed    = "#c62828"
)

// ogIcon dibuja un ícono de 24×24 escalado a size con su esquina en (x, y).
func ogIcon(dc *gg.Context, name string, x, y, size float64, col string) {
	ogIconW(dc, name, x, y, size, col, 2)
}

// ogIconW: como ogIcon con otro grosor de trazo (en la escala de 24; la app usa 2).
func ogIconW(dc *gg.Context, name string, x, y, size float64, col string, stroke float64) {
	s := size / 24
	dc.SetHexColor(col)
	dc.SetLineWidth(stroke * s)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	for _, el := range ogElemRe.FindAllStringSubmatch(ogIcons[name], -1) {
		attr := map[string]float64{}
		d := ""
		for _, a := range ogAttrRe.FindAllStringSubmatch(el[2], -1) {
			if a[1] == "d" {
				d = a[2]
			} else {
				attr[a[1]], _ = strconv.ParseFloat(a[2], 64)
			}
		}
		dc.NewSubPath()
		switch el[1] {
		case "circle":
			dc.DrawCircle(x+attr["cx"]*s, y+attr["cy"]*s, attr["r"]*s)
		case "rect":
			dc.DrawRoundedRectangle(x+attr["x"]*s, y+attr["y"]*s, attr["width"]*s, attr["height"]*s, attr["rx"]*s)
		case "path":
			ogPath(dc, d, x, y, s)
		}
		dc.Stroke()
	}
}

var (
	ogElemRe = regexp.MustCompile(`<(path|circle|rect)\s([^>]*?)/>`)
	ogAttrRe = regexp.MustCompile(`([a-z]+)="([^"]*)"`)
	ogNumRe  = regexp.MustCompile(`[MmLlHhVvCcSsQqAaZz]|-?(?:\d+\.?\d*|\.\d+)(?:e-?\d+)?`)
)

// ogPath traza un «d» de SVG (M L H V C S Q A Z, absolutos y relativos) en (ox, oy) con escala s.
func ogPath(dc *gg.Context, d string, ox, oy, s float64) {
	toks := ogNumRe.FindAllString(d, -1)
	i, cmd := 0, byte('M')
	var cx, cy, sx, sy, lcx, lcy float64
	num := func() float64 { v, _ := strconv.ParseFloat(toks[i], 64); i++; return v }
	isNum := func() bool { return i < len(toks) && !strings.ContainsAny(toks[i][:1], "MmLlHhVvCcSsQqAaZz") }
	P := func(x, y float64) (float64, float64) { return ox + x*s, oy + y*s }
	for i < len(toks) {
		if !isNum() {
			cmd = toks[i][0]
			i++
		}
		rel := cmd >= 'a'
		bx, by := 0.0, 0.0
		if rel {
			bx, by = cx, cy
		}
		switch cmd | 0x20 {
		case 'm':
			cx, cy = bx+num(), by+num()
			sx, sy = cx, cy
			dc.MoveTo(P(cx, cy))
			if rel {
				cmd = 'l'
			} else {
				cmd = 'L'
			}
			lcx, lcy = cx, cy
		case 'l':
			cx, cy = bx+num(), by+num()
			dc.LineTo(P(cx, cy))
			lcx, lcy = cx, cy
		case 'h':
			cx = bx + num()
			dc.LineTo(P(cx, cy))
			lcx, lcy = cx, cy
		case 'v':
			cy = by + num()
			dc.LineTo(P(cx, cy))
			lcx, lcy = cx, cy
		case 'c':
			x1, y1, x2, y2, x, y := bx+num(), by+num(), bx+num(), by+num(), bx+num(), by+num()
			ax, ay := P(x1, y1)
			bx2, by2 := P(x2, y2)
			ex, ey := P(x, y)
			dc.CubicTo(ax, ay, bx2, by2, ex, ey)
			lcx, lcy, cx, cy = x2, y2, x, y
		case 's':
			x1, y1 := 2*cx-lcx, 2*cy-lcy
			x2, y2, x, y := bx+num(), by+num(), bx+num(), by+num()
			ax, ay := P(x1, y1)
			bx2, by2 := P(x2, y2)
			ex, ey := P(x, y)
			dc.CubicTo(ax, ay, bx2, by2, ex, ey)
			lcx, lcy, cx, cy = x2, y2, x, y
		case 'q':
			x1, y1, x, y := bx+num(), by+num(), bx+num(), by+num()
			ax, ay := P(x1, y1)
			ex, ey := P(x, y)
			dc.QuadraticTo(ax, ay, ex, ey)
			lcx, lcy, cx, cy = x1, y1, x, y
		case 'a':
			rx, ry, _, large, sweep := num(), num(), num(), num(), num()
			x, y := bx+num(), by+num()
			ogArc(dc, cx, cy, rx, ry, large != 0, sweep != 0, x, y, P)
			cx, cy, lcx, lcy = x, y, x, y
		case 'z':
			dc.ClosePath()
			cx, cy, lcx, lcy = sx, sy, sx, sy
			if isNum() {
				cmd = 'L'
			}
		default:
			i++
		}
	}
}

// ogArc: arco elíptico de SVG (sin rotación, como en los íconos) pasado a su centro y ángulos.
func ogArc(dc *gg.Context, x1, y1, rx, ry float64, large, sweep bool, x2, y2 float64, P func(x, y float64) (float64, float64)) {
	if rx == 0 || ry == 0 || x1 == x2 && y1 == y2 {
		dc.LineTo(P(x2, y2))
		return
	}
	rx, ry = math.Abs(rx), math.Abs(ry)
	dx, dy := (x1-x2)/2, (y1-y2)/2
	if l := dx*dx/(rx*rx) + dy*dy/(ry*ry); l > 1 {
		rx, ry = rx*math.Sqrt(l), ry*math.Sqrt(l)
	}
	num := rx*rx*ry*ry - rx*rx*dy*dy - ry*ry*dx*dx
	den := rx*rx*dy*dy + ry*ry*dx*dx
	k := math.Sqrt(math.Max(0, num/den))
	if large == sweep {
		k = -k
	}
	ccx, ccy := k*rx*dy/ry+(x1+x2)/2, -k*ry*dx/rx+(y1+y2)/2
	a1 := math.Atan2((y1-ccy)/ry, (x1-ccx)/rx)
	a2 := math.Atan2((y2-ccy)/ry, (x2-ccx)/rx)
	if sweep && a2 < a1 {
		a2 += 2 * math.Pi
	} else if !sweep && a2 > a1 {
		a2 -= 2 * math.Pi
	}
	// En pantalla: el arco con el centro y los radios ya escalados.
	px, py := P(ccx, ccy)
	ex, _ := P(ccx+rx, ccy)
	_, ey := P(ccx, ccy+ry)
	dc.DrawEllipticalArc(px, py, ex-px, ey-py, a1, a2)
}

// ogFit corta s con «…» para que quepa en w.
func ogFit(dc *gg.Context, s string, w float64) string {
	if tw, _ := dc.MeasureString(s); tw <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 {
		r = r[:len(r)-1]
		if tw, _ := dc.MeasureString(strings.TrimSpace(string(r)) + "…"); tw <= w {
			break
		}
	}
	return strings.TrimSpace(string(r)) + "…"
}

func ogWidth(dc *gg.Context, face font.Face, s string) float64 {
	dc.SetFontFace(face)
	w, _ := dc.MeasureString(s)
	return w
}

// ogTextAt escribe s con su línea media en y; devuelve el ancho.
func ogTextAt(dc *gg.Context, face font.Face, col, s string, x, y float64) float64 {
	dc.SetFontFace(face)
	dc.SetHexColor(col)
	dc.DrawStringAnchored(s, x, y, 0, 0.35)
	w, _ := dc.MeasureString(s)
	return w
}

// ogImage: la tarjeta del punto o de la necesidad más reciente; sin ninguna necesidad abierta, una tarjeta general.
func ogImage(n ogNeed) ([]byte, error) {
	if n.ID == "" {
		return ogImageEmpty()
	}
	return ogImageBig(n)
}

// ogImageEmpty: sin necesidades abiertas, el mismo estilo con un mensaje general.
func ogImageEmpty() ([]byte, error) {
	dc := gg.NewContext(ogW, ogH)
	dc.SetHexColor("#ffffff")
	dc.Clear()
	const takeH, left = 118.0, 44.0
	tf := ogFace(true, 104)
	ogTextAt(dc, tf, ogInk, "Puntos de apoyo", left, 170)
	ogTextAt(dc, tf, ogInk, "en Azuero", left, 290)
	ogTextAt(dc, ogFace(false, 52), ogMuted, "Sin necesidades abiertas ahora", left, 410)
	dc.DrawRectangle(0, ogH-takeH, ogW, takeH)
	dc.SetHexColor("#ed7654")
	dc.Fill()
	cf := ogFace(true, 52)
	ogTextAt(dc, cf, "#ffffff", "Reporta un lugar que necesita ayuda", (ogW-ogWidth(dc, cf, "Reporta un lugar que necesita ayuda"))/2, ogH-takeH/2)
	return ogEncode(dc)
}

func ogEncode(dc *gg.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ogImageBig(n ogNeed) ([]byte, error) {
	dc := gg.NewContext(ogW, ogH)
	dc.SetHexColor("#ffffff")
	dc.Clear()
	// Tres líneas grandes y juntas a todo el ancho, llenando el alto sobre la barra de «¡Yo lo llevo!». Cada línea achica
	// su letra solo lo necesario para caber.
	const takeH = 118.0
	const line1, line2, line3 = 182.0, 140.0, 158.0
	const tMax, nMax, pMax = 168.0, 116.0, 150.0
	const left, right = 44.0, ogW - 44.0
	top := (ogH - takeH - line1 - line2 - line3) / 2
	fit := func(max, min float64, bold bool, width func(face font.Face, k float64) float64) float64 {
		size := max
		for ; size > min; size -= 2 {
			if width(ogFace(bold, size), size/max) <= right-left {
				break
			}
		}
		return size
	}
	urgent := normalizeUrgency(n.Urgency) == urgencyUrgente
	// 1: lugar (e «Urgente»).
	y := top + line1/2
	title := n.Area
	if title == "" {
		title = n.Place
	}
	uf := ogFace(true, 40)
	urgW := 0.0
	if urgent {
		urgW = ogWidth(dc, uf, "Urgente") + 48 + 18
	}
	ts := fit(tMax, 72, true, func(f font.Face, _ float64) float64 { return ogWidth(dc, f, title) + urgW })
	tf := ogFace(true, ts)
	dc.SetFontFace(tf)
	x := left + ogTextAt(dc, tf, ogInk, ogFit(dc, title, right-left-urgW), left, y)
	if urgent {
		w := ogWidth(dc, uf, "Urgente") + 48
		dc.DrawRoundedRectangle(x+18, y-34, w, 68, 34)
		dc.SetHexColor(ogRed)
		dc.Fill()
		ogTextAt(dc, uf, "#ffffff", "Urgente", x+42, y)
	}
	// 2: qué necesita, con su ícono y el nombre corto (sin el detalle entre paréntesis); lo que no cabe, «+N».
	y += line1/2 + line2/2
	type need struct{ icon, label string }
	needs := []need{}
	for _, it := range n.Items {
		c := ogChipOf(it)
		label := c.label
		if k := strings.Index(label, " ("); k > 0 {
			label = label[:k]
		}
		needs = append(needs, need{c.icon, label})
	}
	needW := func(f font.Face, k float64, upto int) float64 {
		w := 0.0
		for _, nd := range needs[:upto] {
			if nd.icon != "" {
				w += nMax*1.04*k + 10
			}
			w += ogWidth(dc, f, nd.label) + nMax*0.48*k
		}
		if upto < len(needs) {
			w += ogWidth(dc, f, fmt.Sprintf("+%d", len(needs)-upto))
		}
		return w
	}
	// Al menos dos cosas a la vista (o todas si son pocas), con la letra más grande que permita.
	show := min(2, len(needs))
	// Si ni así caben dos (un texto libre largo), se muestra una y su letra vuelve a crecer lo que pueda.
	ns := 0.0
	for {
		ns = fit(nMax, 50, true, func(f font.Face, k float64) float64 { return needW(f, k, show) })
		if show <= 1 || needW(ogFace(true, ns), ns/nMax, show) <= right-left {
			break
		}
		show--
	}
	nf, k := ogFace(true, ns), ns/nMax
	for show < len(needs) && needW(nf, k, show+1) <= right-left {
		show++
	}
	x = left
	more := 0.0
	if show < len(needs) {
		more = ogWidth(dc, nf, fmt.Sprintf("+%d", len(needs)-show)) + nMax*0.48*k
	}
	dc.SetFontFace(nf)
	for _, nd := range needs[:show] {
		if nd.icon != "" {
			ogIconW(dc, nd.icon, x, y-nMax*0.52*k, nMax*1.04*k, "#ed7654", 2.4)
			x += nMax*1.04*k + 10
		}
		// Un texto libre largo se corta con «…».
		x += ogTextAt(dc, nf, ogInk, ogFit(dc, nd.label, right-x-more), x, y) + nMax*0.48*k
	}
	if show < len(needs) {
		ogTextAt(dc, nf, ogMuted, fmt.Sprintf("+%d", len(needs)-show), x, y)
	}
	// 3: personas: solo íconos y números (personas, adultos, niños, bebés, mayores…), lo más grandes que quepan.
	y += line2/2 + line3/2
	counts := [][3]string{}
	if n.People != "" && n.People != "0" {
		counts = append(counts, [3]string{"people", n.People, ""})
	}
	counts = append(counts, n.Counts...)
	peopleW := func(f font.Face, size float64) float64 {
		w := 0.0
		for i, c := range counts {
			if i > 0 {
				w += size * 0.5
			}
			w += size + 8 + ogWidth(dc, f, c[1])
		}
		return w
	}
	ps := fit(pMax, 40, true, func(f font.Face, k float64) float64 { return peopleW(f, pMax*k) })
	bf := ogFace(true, ps)
	x = left
	for _, c := range counts {
		if x+ps+8+ogWidth(dc, bf, c[1]) > right {
			break
		}
		ogIconW(dc, c[0], x, y-ps/2, ps, ogText2, 2.1)
		x += ps + 8
		x += ogTextAt(dc, bf, ogText2, c[1], x, y) + ps*0.5
	}
	// «¡Yo lo llevo!» abajo, a todo lo ancho.
	dc.DrawRectangle(0, ogH-takeH, ogW, takeH)
	dc.SetHexColor("#ed7654")
	dc.Fill()
	cf := ogFace(true, 64)
	ogTextAt(dc, cf, "#ffffff", "¡Yo lo llevo!", (ogW-ogWidth(dc, cf, "¡Yo lo llevo!"))/2, ogH-takeH/2)
	return ogEncode(dc)
}
