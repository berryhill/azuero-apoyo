package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Respaldo: «Descargar respaldo» (menú de administración) baja una copia completa de la base, comprimida con gzip.
// «Importar respaldo» la carga en otra instalación para probar con datos reales; solo funciona si el servidor arrancó
// con ALLOW_IMPORT=1 (servidor local de pruebas), así nadie reemplaza por error los datos del sitio en vivo.
// Al importar se conserva la contraseña maestra de esta instalación; todo lo demás se reemplaza por lo del respaldo.
// Las fotos y videos no van en el respaldo (están en el almacenamiento de archivos, no en la base).
const (
	backupUploadLimit = 1 << 30 // comprimido
	backupDataLimit   = 4 << 30 // descomprimido
)

func importEnabled() bool { return os.Getenv("ALLOW_IMPORT") == "1" }

func (a *App) backupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/backup", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		name := "azuero-respaldo-" + time.Now().UTC().Format("2006-01-02-1504") + ".db.gz"
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
		w.Header().Set("Cache-Control", "no-store")
		gz := gzip.NewWriter(w)
		// Una transacción de lectura da una copia coherente mientras el sitio sigue recibiendo cambios.
		if err := a.db.View(func(tx *bolt.Tx) error { _, e := tx.WriteTo(gz); return e }); err != nil {
			return
		}
		gz.Close()
	})
	mux.HandleFunc("POST /api/admin/restore", func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if u, e := url.Parse(origin); e != nil || u.Host != r.Host {
				fail(w, 403, "origin mismatch")
				return
			}
		}
		if !a.admin(w, r) {
			return
		}
		if !importEnabled() {
			fail(w, 403, "Importar respaldos solo está permitido en el servidor local de pruebas (ALLOW_IMPORT=1).")
			return
		}
		counts, err := a.restore(http.MaxBytesReader(w, r.Body, backupUploadLimit))
		if err != nil {
			fail(w, 400, err.Error())
			return
		}
		response(w, 200, counts)
	})
}

// restore reemplaza todos los datos por los del respaldo (gzip de una base bbolt), en una sola transacción.
func (a *App) restore(body io.Reader) (map[string]int, error) {
	invalid := errors.New("El archivo no es un respaldo válido de Azuero Apoyo (.db.gz).")
	gz, err := gzip.NewReader(body)
	if err != nil {
		return nil, invalid
	}
	tmp, err := os.CreateTemp("", "azuero-import-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	n, err := io.Copy(tmp, io.LimitReader(gz, backupDataLimit+1))
	tmp.Close()
	if err != nil {
		return nil, invalid
	}
	if n > backupDataLimit {
		return nil, errors.New("El respaldo es demasiado grande.")
	}
	src, err := bolt.Open(tmp.Name(), 0600, &bolt.Options{ReadOnly: true, Timeout: 2 * time.Second})
	if err != nil {
		return nil, invalid
	}
	defer src.Close()
	counts := map[string]int{}
	err = src.View(func(in *bolt.Tx) error {
		if m := in.Bucket(meta); m == nil || m.Get([]byte("password_hash")) == nil || in.Bucket(submissions) == nil {
			return invalid
		}
		for key, bucket := range map[string][]byte{"points": submissions, "needs": needsBucket, "users": usersBucket} {
			if b := in.Bucket(bucket); b != nil {
				counts[key] = b.Stats().KeyN
			}
		}
		return a.db.Update(func(out *bolt.Tx) error {
			password := append([]byte(nil), out.Bucket(meta).Get([]byte("password_hash"))...)
			var old [][]byte
			out.ForEach(func(name []byte, _ *bolt.Bucket) error { old = append(old, append([]byte(nil), name...)); return nil })
			for _, name := range old {
				if e := out.DeleteBucket(name); e != nil {
					return e
				}
			}
			if e := in.ForEach(func(name []byte, b *bolt.Bucket) error {
				dst, e := out.CreateBucket(name)
				if e != nil {
					return e
				}
				return copyBucket(dst, b)
			}); e != nil {
				return e
			}
			return out.Bucket(meta).Put([]byte("password_hash"), password)
		})
	})
	if err != nil {
		if err == invalid {
			return nil, err
		}
		return nil, fmt.Errorf("No se pudo importar: %v", err)
	}
	// Las sesiones de corredores, constructores y quienes entraron con código eran de los datos anteriores.
	// La de administración sigue: la contraseña no cambió.
	a.mu.Lock()
	clear(a.runnerSessions)
	clear(a.builderSessions)
	clear(a.viewerSessions)
	clear(a.lastSeen)
	clear(a.locations)
	a.mu.Unlock()
	if err := a.migrate(); err != nil {
		return nil, fmt.Errorf("Respaldo importado, pero falló la actualización de datos: %v", err)
	}
	return counts, nil
}

func copyBucket(dst, src *bolt.Bucket) error {
	if e := dst.SetSequence(src.Sequence()); e != nil {
		return e
	}
	return src.ForEach(func(k, v []byte) error {
		if v != nil {
			return dst.Put(k, v)
		}
		child, e := dst.CreateBucket(k)
		if e != nil {
			return e
		}
		return copyBucket(child, src.Bucket(k))
	})
}
