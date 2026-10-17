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
