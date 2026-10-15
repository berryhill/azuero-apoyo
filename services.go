package main

import (
	"encoding/json"
	"strings"
	"time"
)

// Servicios del punto: recolección de basura, cambio de baño portátil, fumigación, ronda médica… Son necesidades de tipo
// «service» que se repiten cada RepeatDays días. Al completarlas (corredor, agente o administración) quedan abiertas otra
// vez pero programadas para NextDueAt, sin asignar; hasta esa fecha no cuentan como necesidad abierta ni aparecen en la
// lista de corredores ni en la vista pública. DoneLog guarda cuándo se hizo cada vez.

const serviceMaxDays = 365
const serviceLogMax = 100

func isService(n CampNeed) bool { return n.Kind == "service" }

// serviceScheduled: servicio que todavía no toca (abierto con fecha futura).
func serviceScheduled(n CampNeed, now time.Time) bool {
	if !isService(n) || n.Status != "OPEN" || n.NextDueAt == "" {
		return false
	}
	due, err := time.Parse(time.RFC3339, n.NextDueAt)
	return err == nil && now.Before(due)
}

// rescheduleService: un servicio que se repite y se completa queda listo para la próxima vez.
func rescheduleService(n *CampNeed) {
	if !isService(*n) || n.RepeatDays <= 0 || n.Status != "COMPLETED" {
		return
	}
	done := n.CompletedAt
	if done == "" {
		done = nowStamp()
	}
	at, err := time.Parse(time.RFC3339, done)
	if err != nil {
		at = time.Now().UTC()
	}
	n.DoneLog = append(n.DoneLog, done)
	if len(n.DoneLog) > serviceLogMax {
		n.DoneLog = n.DoneLog[len(n.DoneLog)-serviceLogMax:]
	}
	n.Status, n.RunnerID, n.EnRouteAt = "OPEN", "", ""
	n.NextDueAt = at.Add(time.Duration(n.RepeatDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
}

// applyService lee repeatDays (0–365) y nextDueAt (fecha AAAA-MM-DD o RFC3339) al crear o editar una necesidad.
func applyService(n *CampNeed, f map[string]json.RawMessage) (string, string) {
	if raw, has := f["repeatDays"]; has {
		var days int
		if json.Unmarshal(raw, &days) != nil || days < 0 || days > serviceMaxDays {
			return "repeatDays", "Cada cuántos días: de 0 a 365."
		}
		n.RepeatDays = days
	}
	if raw, has := f["nextDueAt"]; has {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return "nextDueAt", "Fecha inválida."
		}
		text = strings.TrimSpace(text)
		if text == "" {
			n.NextDueAt = ""
		} else if t, err := time.Parse(time.RFC3339, text); err == nil {
			n.NextDueAt = t.UTC().Format(time.RFC3339)
		} else if d, err := time.ParseInLocation("2006-01-02", text, panamaTime()); err == nil {
			n.NextDueAt = d.UTC().Format(time.RFC3339)
		} else {
			return "nextDueAt", "Fecha inválida (usa AAAA-MM-DD)."
		}
	}
	if !isService(*n) {
		n.RepeatDays, n.NextDueAt = 0, ""
	}
	return "", ""
}

// panamaTime: las fechas sin hora se toman como medianoche en Panamá (UTC−5, sin horario de verano).
func panamaTime() *time.Location { return time.FixedZone("America/Panama", -5*3600) }
