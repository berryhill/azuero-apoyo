package main

import bolt "go.etcd.io/bbolt"

// Regla: nadie ve el teléfono de un runner que no es de su equipo. Los contactos de un punto pueden ser runners (p. ej.
// quien lo reportó se pone a sí mismo como «Reportado por» con su teléfono); antes de mandarlos a Runners, Supply o
// Scout, el teléfono de un runner de otro equipo se quita (el nombre queda).
type runnerPhoneGuard struct {
	tv     *teamView
	owners map[string][]string // últimos 8 dígitos → runners con ese teléfono
}

func newRunnerPhoneGuard(tx *bolt.Tx, tv *teamView) *runnerPhoneGuard {
	g := &runnerPhoneGuard{tv: tv, owners: map[string][]string{}}
	eachUser(tx, func(u User) error {
		if d := last8(u.Telephone); u.has(roleRunner) && len(d) >= 7 {
			g.owners[d] = append(g.owners[d], u.ID)
		}
		return nil
	})
	return g
}

func last8(phone string) string {
	d := phoneDigits(phone)
	if len(d) > 8 {
		d = d[len(d)-8:]
	}
	return d
}

// hidden: el teléfono es de un runner y ninguno de los runners con ese teléfono es de mi equipo (ni soy yo).
func (g *runnerPhoneGuard) hidden(phone string) bool {
	ids := g.owners[last8(phone)]
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if g.tv.mate(id) {
			return false
		}
	}
	return true
}

func (g *runnerPhoneGuard) runnerContacts(list []runnerContact) []runnerContact {
	out := list[:0]
	for _, c := range list {
		if g.hidden(c.Phone) {
			continue // runnerContact exige teléfono: sin él no sirve como contacto
		}
		out = append(out, c)
	}
	return out
}

func (g *runnerPhoneGuard) scoutContacts(list []scoutContact) []scoutContact {
	for i := range list {
		if g.hidden(list[i].Phone) {
			list[i].Phone = ""
		}
	}
	return list
}
