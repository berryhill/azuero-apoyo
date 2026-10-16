package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Reparaciones de casa (página de constructores). Quien tiene la obra asignada la documenta: fotos de antes, de avance
// y final; qué falta para terminarla; los materiales necesarios (con cantidad, para armar pedidos de donación) y el
// porcentaje de avance. Administración ve todo en la ficha del punto.

// Materiales comunes en obras de la zona; se suman los que ya escribieron otros constructores.
var repairMaterials = []string{
	"Bolsas de cemento", "Carriolas", "Láminas de zinc", "Varilla de 3/8", "Varilla de 1/2", "Electrodos de soldar",
	"Bloques", "Arena (m³)", "Piedra (m³)", "Clavos", "Clavos de techo", "Tornillos de techo", "Alambre de amarre",
	"Tablas / madera", "Plywood", "Tubos (PVC)", "Pintura (galón)", "Cumbrera", "Malla", "Mano de obra (días)",
}

const repairScopeMax = 4000
const repairMaterialsMax = 80

var errRepairClosed = errors.New("Esta obra ya no está en curso a tu nombre.")

// Assessment: evaluación de una obra (antes de tomarla): qué trabajos lleva, materiales, días y personas estimadas y notas.
// Las fotos van en Media con etapa «inspect» y Ref = ID de la evaluación.
type Assessment struct {
	ID        string        `json:"id"`
	By        string        `json:"by"`
	ByName    string        `json:"byName"`
	At        string        `json:"at"`
	UpdatedAt string        `json:"updatedAt,omitempty"`
	Work      []string      `json:"work"`
	Materials []PackingLine `json:"materials"`
	Days      int           `json:"days,omitempty"`
	Workers   int           `json:"workers,omitempty"`
	Notes     string        `json:"notes,omitempty"`
}

const assessmentsMax = 20

var errAssessClosed = errors.New("Esta obra ya no se puede evaluar.")

func inspectMedia(list []MediaItem) []MediaItem {
	out := []MediaItem{}
	for _, m := range list {
		if m.Stage == stageInspect {
			out = append(out, m)
		}
	}
	return out
}

// parseMaterials valida una lista de materiales (nombre corto y cantidad 1–100000).
func parseMaterials(raw json.RawMessage) ([]PackingLine, string) {
	var in []PackingLine
	if json.Unmarshal(raw, &in) != nil || len(in) > repairMaterialsMax {
		return nil, "Lista de materiales inválida."
	}
	out := []PackingLine{}
	for _, m := range in {
		item := strings.Join(strings.Fields(m.Item), " ")
		if item == "" {
			continue
		}
		if len([]rune(item)) > 80 || m.Quantity < 1 || m.Quantity > packingMaxQuantity {
			return nil, "Cada material necesita un nombre corto y una cantidad de 1 a 100000."
		}
		out = append(out, PackingLine{Item: item, Quantity: m.Quantity})
	}
	return out, ""
}

// assessmentFrom lee una evaluación enviada por un constructor.
func assessmentFrom(f map[string]json.RawMessage) (Assessment, string, string) {
	var x Assessment
	if raw, has := f["work"]; has {
		var work []string
		if json.Unmarshal(raw, &work) != nil || len(work) > 30 {
			return x, "work", "Tipos de trabajo inválidos."
		}
		for _, w := range work {
			if w = strings.Join(strings.Fields(w), " "); w != "" && len([]rune(w)) <= 80 {
				x.Work = append(x.Work, w)
			}
		}
	}
	if raw, has := f["materials"]; has {
		var msg string
		if x.Materials, msg = parseMaterials(raw); msg != "" {
			return x, "materials", msg
		}
	}
	for key, dst := range map[string]*int{"days": &x.Days, "workers": &x.Workers} {
		if raw, has := f[key]; has {
			if json.Unmarshal(raw, dst) != nil || *dst < 0 || *dst > 365 {
				return x, key, "Número fuera de rango."
			}
		}
	}
	x.Notes = strings.TrimSpace(fieldText(f, "notes"))
	if x.Work == nil {
		x.Work = []string{}
	}
	if x.Materials == nil {
		x.Materials = []PackingLine{}
	}
	return x, "", ""
}

func (a *App) repairRoutes(mux *http.ServeMux, builders *crew) {
	// La obra es de quien la tiene asignada y está en curso (o recién completada, para fotos finales).
	owns := func(s runnerSession, needDone bool) func(tx *bolt.Tx, n CampNeed) error {
		return func(tx *bolt.Tx, n CampNeed) error {
			if !isRepair(n) || n.BuilderID != s.runnerID || n.Status != "IN_PROGRESS" && !(needDone && n.Status == "COMPLETED") {
				return errMissing
			}
			return nil
		}
	}
	mux.HandleFunc("POST /api/builder/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(builders, w, r)
		if !ok || !sameOrigin(w, r) {
			return
		}
		var name string
		a.db.View(func(tx *bolt.Tx) error {
			if u, ok := getUser(tx, s.runnerID); ok {
				name = u.Name
			}
			return nil
		})
		allowed := owns(s, true)
		if r.URL.Query().Get("stage") == stageInspect {
			ref := r.URL.Query().Get("ref")
			allowed = func(tx *bolt.Tx, n CampNeed) error {
				for _, x := range n.Assessments {
					if x.ID == ref && x.By == s.runnerID && isRepair(n) && n.Status != "DELETED" {
						return nil
					}
				}
				return errMissing
			}
		}
		a.receiveMedia(w, r, s.runnerID, name, allowed)
	})
	mux.HandleFunc("GET /api/builder/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(builders, w, r)
		if !ok {
			return
		}
		a.serveMedia(w, r, func(tx *bolt.Tx, n CampNeed) error {
			if owns(s, true)(tx, n) == nil {
				return nil
			}
			for _, m := range n.Media {
				if m.ID == r.PathValue("media") && m.Stage == stageInspect && isRepair(n) {
					return nil
				}
			}
			return errMissing
		})
	})
	mux.HandleFunc("GET /api/builder/materials", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.crewSession(builders, w, r); !ok {
			return
		}
		seen, list := map[string]bool{}, []PackingItem{}
		add := func(name string) {
			if k := foldText(name); k != "" && !seen[k] {
				seen[k] = true
				list = append(list, PackingItem{Name: name})
			}
		}
		for _, m := range repairMaterials {
			add(m)
		}
		a.db.View(func(tx *bolt.Tx) error {
			return tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
				var n CampNeed
				if json.Unmarshal(v, &n) == nil && isRepair(n) {
					for _, m := range n.Materials {
						add(m.Item)
					}
				}
				return nil
			})
		})
		sort.SliceStable(list[len(repairMaterials):], func(i, j int) bool {
			rest := list[len(repairMaterials):]
			return foldText(rest[i].Name) < foldText(rest[j].Name)
		})
		response(w, 200, list)
	})
	// Datos de la obra: cualquiera de scope, materials (lista completa, reemplaza la anterior) y progress (0–100).
	mux.HandleFunc("POST /api/builder/needs/{id}/project", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(builders, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"scope": repairScopeMax, "materials": 0, "progress": 0})
		if !ok {
			return
		}
		var materials []PackingLine
		if raw, has := f["materials"]; has {
			var msg string
			if materials, msg = parseMaterials(raw); msg != "" {
				failField(w, 400, "materials", msg)
				return
			}
		}
		progress := -1
		if raw, has := f["progress"]; has {
			if json.Unmarshal(raw, &progress) != nil || progress < 0 || progress > 100 {
				failField(w, 400, "progress", "El avance va de 0 a 100.")
				return
			}
		}
		var n CampNeed
		err := a.write(r, func(tx *bolt.Tx) error {
			u, found := getUser(tx, s.runnerID)
			if !found || u.Disabled || !u.has(roleBuilder) {
				return errUserDisabled
			}
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if e := owns(s, false)(tx, n); e != nil || !a.campExists(tx, n.CampID) {
				return errRepairClosed
			}
			if _, has := f["scope"]; has {
				n.Scope = strings.TrimSpace(fieldText(f, "scope"))
			}
			if materials != nil {
				for i := range materials {
					materials[i].AddedBy, materials[i].AddedByName, materials[i].AddedAt = u.ID, u.Name, nowStamp()
				}
				n.Materials = materials
			}
			if progress >= 0 {
				n.Progress, n.ProgressAt = progress, nowStamp()
			}
			return putNeed(b, &n)
		})
		if errors.Is(err, errUserDisabled) {
			fail(w, 403, err.Error())
			return
		}
		if errors.Is(err, errRepairClosed) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "scope": n.Scope, "materials": n.Materials, "progress": n.Progress})
	})
	// Evaluar una obra: cualquier constructor, aunque no la tome; cada uno edita solo la suya.
	saveAssessment := func(w http.ResponseWriter, r *http.Request, aid string) {
		s, ok := a.crewSession(builders, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"work": 0, "materials": 0, "days": 0, "workers": 0, "notes": repairScopeMax})
		if !ok {
			return
		}
		x, field, msg := assessmentFrom(f)
		if field != "" {
			failField(w, 400, field, msg)
			return
		}
		var saved Assessment
		err := a.write(r, func(tx *bolt.Tx) error {
			u, found := getUser(tx, s.runnerID)
			if !found || u.Disabled || !u.has(roleBuilder) {
				return errUserDisabled
			}
			b := tx.Bucket(needsBucket)
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			var n CampNeed
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if !isRepair(n) || !a.campExists(tx, n.CampID) || n.Status == "DELETED" || n.Status == "COMPLETED" {
				return errAssessClosed
			}
			if aid == "" {
				if len(n.Assessments) >= assessmentsMax {
					return errAssessClosed
				}
				x.ID, x.By, x.ByName, x.At = randomToken()[:16], u.ID, u.Name, nowStamp()
				n.Assessments = append(n.Assessments, x)
				saved = x
			} else {
				found := false
				for i := range n.Assessments {
					if n.Assessments[i].ID == aid && n.Assessments[i].By == u.ID {
						x.ID, x.By, x.ByName, x.At, x.UpdatedAt = aid, u.ID, u.Name, n.Assessments[i].At, nowStamp()
						n.Assessments[i], saved, found = x, x, true
					}
				}
				if !found {
					return errMissing
				}
			}
			return putNeed(b, &n)
		})
		if errors.Is(err, errUserDisabled) {
			fail(w, 403, err.Error())
			return
		}
		if errors.Is(err, errAssessClosed) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, saved)
	}
	mux.HandleFunc("POST /api/builder/needs/{id}/assessments", func(w http.ResponseWriter, r *http.Request) { saveAssessment(w, r, "") })
	mux.HandleFunc("POST /api/builder/needs/{id}/assessments/{aid}", func(w http.ResponseWriter, r *http.Request) { saveAssessment(w, r, r.PathValue("aid")) })
}
