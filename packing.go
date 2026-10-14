package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Lista de empaque: corredores y agentes anotan qué artículos y cuántos se preparan para cumplir una necesidad.
// Los artículos usados quedan en un catálogo para elegirlos la próxima vez.
var packingBucket = []byte("packing_items")

// Artículos iniciales; sus íconos están en pointform.js (PointForm.PACKING) con estos mismos nombres.
var defaultPackingItems = []string{
	"Bolsas alimentación", "Kit aseo personal", "Kit limpieza", "Kit bebé", "Linternas", "Carpas / toldas",
	"Repelente", "Sillas", "Hamacas", "Mosquiteros", "Extensiones", "Camping", "Toallas", "Capotes / ponchos",
	"Paraguas", "Medicamentos", "Utensilios", "Desechables", "Colchonetas",
}

type PackingItem struct {
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
}

type PackingLine struct {
	Item        string `json:"item"`
	Quantity    int    `json:"quantity"`
	AddedBy     string `json:"addedBy,omitempty"`
	AddedByName string `json:"addedByName,omitempty"`
	AddedAt     string `json:"addedAt,omitempty"`
}

const packingMaxQuantity = 100000

// seedPackingItems crea el catálogo y agrega los artículos iniciales que falten. La clave es el nombre normalizado.
func seedPackingItems(tx *bolt.Tx) error {
	b, err := tx.CreateBucketIfNotExists(packingBucket)
	if err != nil {
		return err
	}
	for _, name := range defaultPackingItems {
		if b.Get([]byte(foldText(name))) == nil {
			if e := putJSON(b, foldText(name), PackingItem{Name: name}); e != nil {
				return e
			}
		}
	}
	return nil
}

func packingItems(tx *bolt.Tx) ([]PackingItem, error) {
	list := []PackingItem{}
	err := tx.Bucket(packingBucket).ForEach(func(_, v []byte) error {
		var item PackingItem
		if e := json.Unmarshal(v, &item); e != nil {
			return e
		}
		list = append(list, item)
		return nil
	})
	sort.SliceStable(list, func(i, j int) bool { return foldText(list[i].Name) < foldText(list[j].Name) })
	return list, err
}

type packingInput struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

// parsePacking valida los artículos enviados y junta los repetidos sumando cantidades.
func parsePacking(raw json.RawMessage) ([]packingInput, error) {
	var in []packingInput
	if json.Unmarshal(raw, &in) != nil || len(in) == 0 || len(in) > 50 {
		return nil, userFieldError{"items", "Agrega entre 1 y 50 artículos."}
	}
	out := []packingInput{}
	index := map[string]int{}
	for _, line := range in {
		line.Item = strings.Join(strings.Fields(line.Item), " ")
		if line.Item == "" || len(line.Item) > 80 {
			return nil, userFieldError{"items", "Cada artículo necesita un nombre de hasta 80 caracteres."}
		}
		if line.Quantity < 1 || line.Quantity > packingMaxQuantity {
			return nil, userFieldError{"items", "La cantidad de «" + line.Item + "» debe ser un número entero mayor que cero."}
		}
		if i, ok := index[foldText(line.Item)]; ok {
			if out[i].Quantity += line.Quantity; out[i].Quantity > packingMaxQuantity {
				return nil, userFieldError{"items", "La cantidad de «" + line.Item + "» es demasiado grande."}
			}
			continue
		}
		index[foldText(line.Item)] = len(out)
		out = append(out, line)
	}
	return out, nil
}

// addPacking suma las líneas a la lista de empaque de la necesidad y guarda en el catálogo los artículos nuevos.
func addPacking(tx *bolt.Tx, n *CampNeed, lines []packingInput, byID, byName string) error {
	catalog := tx.Bucket(packingBucket)
	now := nowStamp()
	for _, line := range lines {
		if catalog.Get([]byte(foldText(line.Item))) == nil {
			if e := putJSON(catalog, foldText(line.Item), PackingItem{Name: line.Item, CreatedAt: now, CreatedBy: byID}); e != nil {
				return e
			}
		}
		n.Packing = append(n.Packing, PackingLine{Item: line.Item, Quantity: line.Quantity, AddedBy: byID, AddedByName: byName, AddedAt: now})
	}
	return putNeed(tx.Bucket(needsBucket), n)
}

var errPackingClosed = errors.New("Esta necesidad ya no está abierta ni en ruta; no se le puede agregar a la lista de empaque.")

func (a *App) packingRoutes(mux *http.ServeMux, runners *crew) {
	mux.HandleFunc("GET /api/runner/packing-items", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.crewSession(runners, w, r); !ok {
			return
		}
		var list []PackingItem
		err := a.db.View(func(tx *bolt.Tx) error {
			var e error
			list, e = packingItems(tx)
			return e
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	// Agregar a la lista de empaque: corredores solo en lo que tienen asignado o tomaron; agentes en todo lo abierto o en ruta.
	mux.HandleFunc("POST /api/runner/needs/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"items": 0})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		var n CampNeed
		err = a.db.Update(func(tx *bolt.Tx) error {
			u, found := getUser(tx, s.runnerID)
			if !found || u.Disabled || !(u.has(roleRunner) || s.agent && u.has(roleAgent)) {
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
			if !a.campExists(tx, n.CampID) || isRepair(n) || !s.agent && n.RunnerID != s.runnerID {
				return errMissing
			}
			if n.Status != "OPEN" && n.Status != "IN_PROGRESS" {
				return errPackingClosed
			}
			return addPacking(tx, &n, lines, u.ID, u.Name)
		})
		if errors.Is(err, errUserDisabled) {
			fail(w, 403, err.Error())
			return
		}
		if errors.Is(err, errPackingClosed) {
			fail(w, 409, err.Error())
			return
		}
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
	// Administración: mismo catálogo; puede agregar a cualquier necesidad de suministros que no esté eliminada.
	mux.HandleFunc("GET /api/admin/packing-items", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		var list []PackingItem
		err := a.db.View(func(tx *bolt.Tx) error {
			var e error
			list, e = packingItems(tx)
			return e
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, list)
	})
	mux.HandleFunc("POST /api/admin/needs/{id}/packing", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"items": 0})
		if !ok {
			return
		}
		lines, err := parsePacking(f["items"])
		if userError(w, err) {
			return
		}
		var n CampNeed
		err = a.db.Update(func(tx *bolt.Tx) error {
			v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("id")))
			if v == nil {
				return errMissing
			}
			if e := json.Unmarshal(v, &n); e != nil {
				return e
			}
			if isRepair(n) || n.Status == "DELETED" {
				return errMissing
			}
			return addPacking(tx, &n, lines, "admin", "Administración")
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, map[string]any{"id": n.ID, "packing": n.Packing})
	})
}
