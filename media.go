package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Fotos y videos que documentan una carrera: «partida» (al recoger) y «entrega» (al completar).
// Son privados: solo los ve el corredor asignado y administración, nunca el público.
//
// Almacenamiento: con S3_BUCKET, S3_ACCESS_KEY_ID y S3_SECRET_ACCESS_KEY (y S3_ENDPOINT/S3_REGION para
// proveedores compatibles con S3) se guardan en almacenamiento de objetos. Las credenciales vienen del
// entorno (en el clúster, del Secret de existingEnvSecret); nunca de la base de datos ni del panel.
// Sin esas variables se guardan en disco junto a la base de datos (para pruebas locales).

const (
	stagePickup   = "pickup"
	stageDelivery = "delivery"
	stageBefore   = "before"
	stageProgress = "progress"
	stageAfter    = "after"
	stageInspect  = "inspect"
	mediaPerNeed  = 60
)

type MediaItem struct {
	ID          string `json:"id"`
	StoredKey   string `json:"key,omitempty"`
	Stage       string `json:"stage"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	Name        string `json:"name,omitempty"`
	AddedBy     string `json:"addedBy,omitempty"`
	AddedByName string `json:"addedByName,omitempty"`
	AddedAt     string `json:"addedAt,omitempty"`
	// Ref: a qué evaluación de obra pertenece la foto (etapa «inspect»).
	Ref string `json:"ref,omitempty"`
}

// Tipos aceptados: fotos y videos comunes de teléfonos. SVG queda fuera porque puede llevar código.
var mediaTypes = map[string]string{
	"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp", "image/gif": "gif", "image/heic": "heic", "image/heif": "heif",
	"video/mp4": "mp4", "video/quicktime": "mov", "video/webm": "webm", "video/3gpp": "3gp",
}

type mediaStore interface {
	Put(ctx context.Context, key, contentType string, size int64, body io.Reader) error
	Delete(ctx context.Context, key string) error
	Serve(w http.ResponseWriter, r *http.Request, key, contentType string)
}

// newMediaStore elige almacenamiento de objetos si está configurado; si no, disco junto a la base de datos.
func newMediaStore(dbPath string) mediaStore {
	bucket, access, secret := os.Getenv("S3_BUCKET"), os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_SECRET_ACCESS_KEY")
	if bucket != "" && access != "" && secret != "" {
		endpoint := strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/")
		region := os.Getenv("S3_REGION")
		if region == "" {
			region = "us-east-1"
		}
		if endpoint == "" {
			endpoint = "https://s3." + region + ".amazonaws.com"
		}
		return &s3Store{endpoint: endpoint, region: region, bucket: bucket, access: access, secret: secret, client: &http.Client{Timeout: 10 * time.Minute}}
	}
	return &diskStore{dir: filepath.Join(filepath.Dir(dbPath), "media")}
}

func mediaMaxBytes() int64 {
	if mb, err := strconv.Atoi(os.Getenv("MEDIA_MAX_MB")); err == nil && mb > 0 {
		return int64(mb) << 20
	}
	return 250 << 20
}

// --- Disco ---

type diskStore struct{ dir string }

func (d *diskStore) path(key string) string { return filepath.Join(d.dir, filepath.FromSlash(key)) }
func (d *diskStore) Put(_ context.Context, key, _ string, size int64, body io.Reader) error {
	p := d.path(key)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	tmp := p + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && size >= 0 && n != size {
		err = fmt.Errorf("short upload: %d of %d bytes", n, size)
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, p)
}
func (d *diskStore) Delete(_ context.Context, key string) error {
	if err := os.Remove(d.path(key)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func (d *diskStore) Serve(w http.ResponseWriter, r *http.Request, key, contentType string) {
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, d.path(key))
}

// --- Almacenamiento de objetos compatible con S3 (firma AWS Signature V4, rutas estilo path) ---

type s3Store struct {
	endpoint, region, bucket, access, secret string
	client                                   *http.Client
}

// awsEscape codifica como pide SigV4: solo A-Z a-z 0-9 - _ . ~ quedan igual; «/» se conserva en rutas.
func awsEscape(s string, keepSlash bool) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_', c == '.', c == '~', keepSlash && c == '/':
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}

// sigV4 firma una solicitud: devuelve la firma para el método, ruta, consulta canónica y encabezados firmados.
func sigV4(secret, region, amzDate, method, path, query string, headers map[string]string, payloadHash string) (signature, signedHeaders, scope string) {
	names := make([]string, 0, len(headers))
	for k := range headers {
		names = append(names, strings.ToLower(k))
	}
	sort.Strings(names)
	var canonHeaders strings.Builder
	for _, k := range names {
		for hk, v := range headers {
			if strings.ToLower(hk) == k {
				canonHeaders.WriteString(k + ":" + strings.TrimSpace(v) + "\n")
			}
		}
	}
	signedHeaders = strings.Join(names, ";")
	canonical := strings.Join([]string{method, path, query, canonHeaders.String(), signedHeaders, payloadHash}, "\n")
	day := amzDate[:8]
	scope = day + "/" + region + "/s3/aws4_request"
	sum := sha256.Sum256([]byte(canonical))
	toSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(sum[:])
	key := hmacSHA256(hmacSHA256(hmacSHA256(hmacSHA256([]byte("AWS4"+secret), day), region), "s3"), "aws4_request")
	return hex.EncodeToString(hmacSHA256(key, toSign)), signedHeaders, scope
}

func (s *s3Store) objectURL(key string) (*url.URL, string, error) {
	u, err := url.Parse(s.endpoint)
	if err != nil {
		return nil, "", err
	}
	path := strings.TrimRight(u.Path, "/") + "/" + awsEscape(s.bucket, false) + "/" + awsEscape(key, true)
	return u, path, nil
}

func (s *s3Store) do(ctx context.Context, method, key, contentType string, size int64, body io.Reader) (*http.Response, error) {
	u, path, err := s.objectURL(key)
	if err != nil {
		return nil, err
	}
	amzDate := time.Now().UTC().Format("20060102T150405Z")
	headers := map[string]string{"host": u.Host, "x-amz-content-sha256": "UNSIGNED-PAYLOAD", "x-amz-date": amzDate}
	sig, signed, scope := sigV4(s.secret, s.region, amzDate, method, path, "", headers, "UNSIGNED-PAYLOAD")
	req, err := http.NewRequestWithContext(ctx, method, u.Scheme+"://"+u.Host+path, body)
	if err != nil {
		return nil, err
	}
	req.URL.RawPath = path
	if size >= 0 {
		req.ContentLength = size
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.access+"/"+scope+", SignedHeaders="+signed+", Signature="+sig)
	return s.client.Do(req)
}

func (s *s3Store) Put(ctx context.Context, key, contentType string, size int64, body io.Reader) error {
	res, err := s.do(ctx, http.MethodPut, key, contentType, size, body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("object storage put: %d %s", res.StatusCode, msg)
	}
	return nil
}
func (s *s3Store) Delete(ctx context.Context, key string) error {
	res, err := s.do(ctx, http.MethodDelete, key, "", 0, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("object storage delete: %d", res.StatusCode)
	}
	return nil
}

// presign arma una URL temporal de lectura (como la documenta AWS para «query string authentication»).
func (s *s3Store) presign(key string, now time.Time, expires time.Duration) (string, error) {
	u, path, err := s.objectURL(key)
	if err != nil {
		return "", err
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	q := []string{
		"X-Amz-Algorithm=AWS4-HMAC-SHA256",
		"X-Amz-Credential=" + awsEscape(s.access+"/"+amzDate[:8]+"/"+s.region+"/s3/aws4_request", false),
		"X-Amz-Date=" + amzDate,
		"X-Amz-Expires=" + strconv.Itoa(int(expires.Seconds())),
		"X-Amz-SignedHeaders=host",
	}
	sort.Strings(q)
	query := strings.Join(q, "&")
	sig, _, _ := sigV4(s.secret, s.region, amzDate, http.MethodGet, path, query, map[string]string{"host": u.Host}, "UNSIGNED-PAYLOAD")
	return u.Scheme + "://" + u.Host + path + "?" + query + "&X-Amz-Signature=" + sig, nil
}

// Serve redirige a una URL firmada de 10 minutos: el navegador descarga directo del almacenamiento (con Range para videos).
func (s *s3Store) Serve(w http.ResponseWriter, r *http.Request, key, _ string) {
	link, err := s.presign(key, time.Now(), 10*time.Minute)
	if err != nil {
		http.Error(w, "storage error", 500)
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=300")
	http.Redirect(w, r, link, http.StatusFound)
}

// --- Rutas ---

var errMediaFull = userFieldError{"media", "Esta carrera ya tiene el máximo de fotos y videos."}

// Corredores: partida y entrega. Reparaciones (constructores): antes, avance y final.
func validStage(s string) bool {
	return s == stagePickup || s == stageDelivery || s == stageBefore || s == stageProgress || s == stageAfter || s == stageInspect
}

const errMediaOff = "Fotos y videos no disponibles por ahora."

// receiveMedia guarda el cuerpo de la solicitud (el archivo tal cual, sin multipart) y lo agrega a la necesidad.
// allowed decide, dentro de la transacción, si quien sube puede documentar esa necesidad.
func (a *App) receiveMedia(w http.ResponseWriter, r *http.Request, byID, byName string, allowed func(tx *bolt.Tx, n CampNeed) error) {
	if !a.mediaOn {
		failField(w, 503, "media", errMediaOff)
		return
	}
	stage := r.URL.Query().Get("stage")
	if !validStage(stage) {
		failField(w, 400, "stage", "Indica si es foto de partida o de entrega.")
		return
	}
	ctype := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	ext, ok := mediaTypes[ctype]
	if !ok {
		failField(w, 415, "media", "Solo se aceptan fotos (JPG, PNG, WEBP, HEIC) o videos (MP4, MOV, WEBM).")
		return
	}
	if r.ContentLength <= 0 {
		failField(w, 411, "media", "El archivo está vacío o no indica su tamaño.")
		return
	}
	if max := mediaMaxBytes(); r.ContentLength > max {
		failField(w, 413, "media", fmt.Sprintf("El archivo pesa más de %d MB.", max>>20))
		return
	}
	needID := r.PathValue("id")
	check := func() error {
		return a.db.View(func(tx *bolt.Tx) error {
			var n CampNeed
			v := tx.Bucket(needsBucket).Get([]byte(needID))
			if v == nil || json.Unmarshal(v, &n) != nil {
				return errMissing
			}
			if len(n.Media) >= mediaPerNeed {
				return errMediaFull
			}
			return allowed(tx, n)
		})
	}
	if err := check(); userError(w, err) {
		return
	}
	name := strings.TrimSpace(r.Header.Get("X-File-Name"))
	if decoded, err := url.QueryUnescape(name); err == nil {
		name = decoded
	}
	if len(name) > 120 {
		name = name[:120]
	}
	item := MediaItem{ID: randomToken()[:24], Stage: stage, Type: ctype, Size: r.ContentLength, Name: name, AddedBy: byID, AddedByName: byName, AddedAt: nowStamp()}
	if ref := r.URL.Query().Get("ref"); ref != "" && len(ref) <= 40 {
		item.Ref = ref
	}
	item.StoredKey = "needs/" + needID + "/" + item.ID + "." + ext
	if err := a.media.Put(r.Context(), item.StoredKey, ctype, r.ContentLength, http.MaxBytesReader(w, r.Body, r.ContentLength)); err != nil {
		log.Printf("media upload failed: %v", err)
		fail(w, 502, "No se pudo guardar el archivo. Revisa tu conexión e intenta de nuevo.")
		return
	}
	var n CampNeed
	err := a.write(r, func(tx *bolt.Tx) error {
		b := tx.Bucket(needsBucket)
		v := b.Get([]byte(needID))
		if v == nil {
			return errMissing
		}
		if e := json.Unmarshal(v, &n); e != nil {
			return e
		}
		if e := allowed(tx, n); e != nil {
			return e
		}
		if len(n.Media) >= mediaPerNeed {
			return errMediaFull
		}
		n.Media = append(n.Media, item)
		return putNeed(b, &n)
	})
	if err != nil {
		a.media.Delete(context.Background(), item.StoredKey)
		userError(w, err)
		return
	}
	response(w, 201, item)
}

// serveMedia entrega un archivo de la necesidad si allowed lo permite.
func (a *App) serveMedia(w http.ResponseWriter, r *http.Request, allowed func(tx *bolt.Tx, n CampNeed) error) {
	var item MediaItem
	err := a.db.View(func(tx *bolt.Tx) error {
		var n CampNeed
		v := tx.Bucket(needsBucket).Get([]byte(r.PathValue("need")))
		if v == nil || json.Unmarshal(v, &n) != nil {
			return errMissing
		}
		if e := allowed(tx, n); e != nil {
			return e
		}
		for _, m := range n.Media {
			if m.ID == r.PathValue("media") {
				item = m
				return nil
			}
		}
		return errMissing
	})
	if userError(w, err) {
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	a.media.Serve(w, r, item.StoredKey, item.Type)
}

// sameOrigin rechaza solicitudes de otros sitios (las subidas no son JSON, así que no pasan por decodeFields).
func sameOrigin(w http.ResponseWriter, r *http.Request) bool {
	if origin := r.Header.Get("Origin"); origin != "" {
		if u, e := url.Parse(origin); e != nil || u.Host != r.Host {
			fail(w, 403, "origin mismatch")
			return false
		}
	}
	return true
}

func (a *App) mediaRoutes(mux *http.ServeMux, runners *crew) {
	// Si se pueden subir fotos: sin almacenamiento de objetos las páginas ocultan los botones y avisan.
	mux.HandleFunc("GET /api/media/status", func(w http.ResponseWriter, r *http.Request) {
		response(w, 200, map[string]bool{"enabled": a.mediaOn})
	})
	// Corredor: documenta sus propias carreras (en ruta o ya completadas); un team manager, también las de su equipo.
	runnerOwns := func(s runnerSession) func(tx *bolt.Tx, n CampNeed) error {
		return func(tx *bolt.Tx, n CampNeed) error {
			if s.agent || n.RunnerID != s.runnerID && !managesRunner(tx, s.runnerID, n.RunnerID) || isRepair(n) || n.Status != "IN_PROGRESS" && n.Status != "COMPLETED" {
				return errMissing
			}
			return nil
		}
	}
	mux.HandleFunc("POST /api/runner/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
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
		a.receiveMedia(w, r, s.runnerID, name, runnerOwns(s))
	})
	mux.HandleFunc("GET /api/runner/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		s, ok := a.crewSession(runners, w, r)
		if !ok {
			return
		}
		a.serveMedia(w, r, runnerOwns(s))
	})

	// Administración: ver, agregar, quitar y cambiar entre partida y entrega.
	anyNeed := func(*bolt.Tx, CampNeed) error { return nil }
	mux.HandleFunc("GET /api/admin/media/{need}/{media}", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) {
			return
		}
		a.serveMedia(w, r, anyNeed)
	})
	mux.HandleFunc("POST /api/admin/needs/{id}/media", func(w http.ResponseWriter, r *http.Request) {
		if !a.admin(w, r) || !sameOrigin(w, r) {
			return
		}
		byID, byName := actor(r)
		a.receiveMedia(w, r, byID, byName, anyNeed)
	})
	mux.HandleFunc("PATCH /api/admin/needs/{id}/media/{media}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		f, ok := decodeFields(w, r, map[string]int{"stage": 16})
		if !ok {
			return
		}
		stage := fieldText(f, "stage")
		if !validStage(stage) {
			failField(w, 400, "stage", "Etapa inválida.")
			return
		}
		var item MediaItem
		err := a.write(r, func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			var n CampNeed
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil {
				return errMissing
			}
			for i := range n.Media {
				if n.Media[i].ID == r.PathValue("media") {
					if e := mediaGuard(tx, r, n.Media[i]); e != nil {
						return e
					}
					n.Media[i].Stage = stage
					item = n.Media[i]
					return putNeed(b, &n)
				}
			}
			return errMissing
		})
		if serviceError(w, err) {
			return
		}
		response(w, 200, item)
	})
	mux.HandleFunc("DELETE /api/admin/needs/{id}/media/{media}", func(w http.ResponseWriter, r *http.Request) {
		if !a.adminWrite(w, r) {
			return
		}
		var removed MediaItem
		err := a.write(r, func(tx *bolt.Tx) error {
			b := tx.Bucket(needsBucket)
			var n CampNeed
			v := b.Get([]byte(r.PathValue("id")))
			if v == nil || json.Unmarshal(v, &n) != nil {
				return errMissing
			}
			for i, m := range n.Media {
				if m.ID == r.PathValue("media") {
					if e := mediaGuard(tx, r, m); e != nil {
						return e
					}
					removed = m
					n.Media = append(n.Media[:i], n.Media[i+1:]...)
					return putNeed(b, &n)
				}
			}
			return errMissing
		})
		if serviceError(w, err) {
			return
		}
		if e := a.media.Delete(r.Context(), removed.StoredKey); e != nil {
			log.Printf("media delete failed (record removed): %v", e)
		}
		response(w, 200, map[string]string{"id": removed.ID, "status": "deleted"})
	})
}
