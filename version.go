package main

import (
	_ "embed"
	"net/http"
	"strings"
)

// VERSION (major.minor.patch) es la versión de la app: se sube con scripts/bump-version.sh en cada commit que se publica
// (también la appVersion de helm/Chart.yaml). El menú de la persona la muestra abajo y avisa cuando hay una nueva.
//
//go:embed VERSION
var versionFile string

// errCSRF: la ficha CSRF no coincide (otra pestaña volvió a entrar, o la sesión se renovó después de una actualización).
// Las páginas toman la nueva y repiten lo que iban a guardar una vez (nav.js freshCsrf); si aun así llega, se ve esto.
const errCSRF = "Tu sesión se renovó (ficha CSRF): vuelve a tocar el botón; lo que llenaste sigue aquí. Si sigue pasando, recarga la página."

func appVersion() string { return strings.TrimSpace(versionFile) }

// withVersion: cada respuesta de /api/ lleva X-App-Version, así las páginas ven que hay versión nueva con las consultas
// que ya hacen (la revisión de acceso cada 45 s), sin una consulta más.
func withVersion(h http.Handler) http.Handler {
	v := appVersion()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("X-App-Version", v)
		}
		h.ServeHTTP(w, r)
	})
}

func (a *App) versionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		response(w, 200, map[string]string{"version": appVersion()})
	})
}
