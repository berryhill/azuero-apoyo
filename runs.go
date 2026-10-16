package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Varios runners en la misma necesidad: cada runner lleva su propio registro (copia ligada con SameAs, como «Otro
// runner a lo mismo» de Supply), con su lista de empaque, sus fotos, su despacho y su «Completé». Poner a otro runner
// en una necesidad que ya lleva alguien no se la quita: le crea su propio run. Y si otra persona cambia lo que pide una
// carrera que ya lleva un runner, el run queda como estaba y lo nuevo se guarda como otra necesidad abierta del punto.

var errAlreadyRunning = userFieldError{"runnerId", "Ese runner ya lleva esta necesidad."}
var errRunLocked = errors.New("Lo que lleva un runner no se cambia desde aquí: solo se pueden agregar cosas, que quedan como otra necesidad.")

// runCopy: registro nuevo del mismo pedido, en el grupo de orig.
func runCopy(orig CampNeed) CampNeed {
	return CampNeed{ID: randomToken(), CampID: orig.CampID, Need: orig.Need, Urgency: orig.Urgency, Kind: orig.Kind, RequesterName: orig.RequesterName, RequesterPhone: orig.RequesterPhone, CreatedAt: nowStamp(), SameAs: groupOf(orig)}
}

// otherRunner: a una necesidad de suministros que ya lleva (o llevó) un runner se le pone otro.
func otherRunner(before CampNeed, runnerID string) bool {
	return before.Kind == "" && before.RunnerID != "" && runnerID != "" && runnerID != before.RunnerID
}

// runOf: el run abierto o en camino de runnerID en el grupo de n, si ya tiene uno.
func runOf(tx *bolt.Tx, n CampNeed, runnerID string) (CampNeed, bool, error) {
	var found CampNeed
	ok := false
	err := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
		var o CampNeed
		if e := json.Unmarshal(v, &o); e != nil {
			return e
		}
		if !ok && o.RunnerID == runnerID && groupOf(o) == groupOf(n) && (o.Status == "OPEN" || o.Status == "IN_PROGRESS") {
			found, ok = o, true
		}
		return nil
	})
	return found, ok, err
}

// newRun prepara (sin guardar) el run propio de runnerID en el grupo de orig, con el estado pedido (en camino si
// venía Abierta). Revisa que sea un runner activo y que no lleve ya 10 runs; los en camino van al final de su orden.
func newRun(tx *bolt.Tx, orig CampNeed, runnerID, status string) (CampNeed, error) {
	runner, found := getUser(tx, runnerID)
	if !found || runner.Disabled || !runner.has(roleRunner) {
		return CampNeed{}, errInvalid
	}
	n := runCopy(orig)
	n.RunnerID, n.Status = runnerID, status
	if n.Status == "OPEN" || n.Status == "" {
		n.Status = "IN_PROGRESS"
	}
	switch n.Status {
	case "IN_PROGRESS":
		busy := 0
		if e := tx.Bucket(needsBucket).ForEach(func(_, v []byte) error {
			var o CampNeed
			if e := json.Unmarshal(v, &o); e != nil {
				return e
			}
			if o.RunnerID == runnerID && o.Status == "IN_PROGRESS" && !isRepair(o) {
				busy++
			}
			return nil
		}); e != nil {
			return n, e
		}
		if busy >= 10 {
			return n, errRunnerBusy
		}
		n.EnRouteAt = nowStamp()
		runner.RunOrder = append(runner.RunOrder, n.ID)
		if e := putJSON(tx.Bucket(usersBucket), runner.ID, runner); e != nil {
			return n, e
		}
	case "COMPLETED":
		n.CompletedAt = nowStamp()
	}
	return n, nil
}

// addedItems: lo que el texto nuevo pide y el anterior no (por partes separadas por comas, sin acentos ni mayúsculas).
func addedItems(before, after string) []string {
	had := map[string]bool{}
	for _, p := range strings.Split(before, ",") {
		had[foldText(p)] = true
	}
	var out []string
	for _, p := range strings.Split(after, ",") {
		if p = strings.TrimSpace(p); p != "" && !had[foldText(p)] {
			had[foldText(p)] = true
			out = append(out, p)
		}
	}
	return out
}

// editOfRun: quien no es el runner cambia lo que pide su carrera (en camino o completada). El run conserva su pedido;
// lo agregado se devuelve como necesidad abierta nueva del punto, ligada al run y anotada con quien la agregó.
func editOfRun(orig CampNeed, text, byID, byName string) (*CampNeed, error) {
	added := addedItems(orig.Need, text)
	if len(added) == 0 {
		return nil, errRunLocked
	}
	n := runCopy(orig)
	n.Need, n.Status, n.CreatedBy, n.CreatedByName = strings.Join(added, ", "), "OPEN", byID, byName
	return &n, nil
}

// runError: 409 si el runner ya lleva 10 runs o si se intenta cambiar lo que lleva un runner; si no, como userError.
func runError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, errRunnerBusy) || errors.Is(err, errRunLocked) {
		fail(w, 409, err.Error())
		return true
	}
	return userError(w, err)
}
