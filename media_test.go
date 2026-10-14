package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Vector oficial de AWS para URLs firmadas (documentación «Authenticating Requests: Using Query Parameters»).
func TestSigV4MatchesAWSPresignExample(t *testing.T) {
	query := "X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host"
	sig, signed, scope := sigV4("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "us-east-1", "20130524T000000Z", "GET", "/test.txt", query, map[string]string{"host": "examplebucket.s3.amazonaws.com"}, "UNSIGNED-PAYLOAD")
	if sig != "aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404" || signed != "host" || scope != "20130524/us-east-1/s3/aws4_request" {
		t.Fatalf("signature %s signed %s scope %s", sig, signed, scope)
	}
	s := &s3Store{endpoint: "https://s3.example.com", region: "us-east-1", bucket: "fotos", access: "AKIAIOSFODNN7EXAMPLE", secret: "x"}
	link, _ := s.presign("needs/n1/a b.jpg", time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC), 10*time.Minute)
	if !strings.HasPrefix(link, "https://s3.example.com/fotos/needs/n1/a%20b.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=600&X-Amz-SignedHeaders=host&X-Amz-Signature=") {
		t.Fatalf("presigned link: %s", link)
	}
}

func TestS3StorePutSignsPathStyleRequest(t *testing.T) {
	var got *http.Request
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, body = r, nil
		body, _ = io.ReadAll(r.Body)
	}))
	defer srv.Close()
	s := &s3Store{endpoint: srv.URL, region: "auto", bucket: "fotos", access: "AK", secret: "SK", client: srv.Client()}
	if err := s.Put(context.Background(), "needs/n1/m1.jpg", "image/jpeg", 5, strings.NewReader("hello")); err != nil {
		t.Fatal(err)
	}
	auth := got.Header.Get("Authorization")
	if got.Method != "PUT" || got.URL.Path != "/fotos/needs/n1/m1.jpg" || string(body) != "hello" || got.ContentLength != 5 || got.Header.Get("Content-Type") != "image/jpeg" ||
		!strings.HasPrefix(auth, "AWS4-HMAC-SHA256 Credential=AK/") || !strings.Contains(auth, "/auto/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=") {
		t.Fatalf("put: %s %s %q %d %q", got.Method, got.URL.Path, body, got.ContentLength, auth)
	}
}

func TestNewMediaStoreUsesObjectStorageOnlyWithCredentials(t *testing.T) {
	t.Setenv("S3_BUCKET", "fotos")
	t.Setenv("S3_ACCESS_KEY_ID", "")
	t.Setenv("S3_SECRET_ACCESS_KEY", "")
	if _, ok := newMediaStore("/data/relief.db").(*diskStore); !ok {
		t.Fatal("without credentials media stays on disk")
	}
	t.Setenv("S3_ACCESS_KEY_ID", "AK")
	t.Setenv("S3_SECRET_ACCESS_KEY", "SK")
	t.Setenv("S3_ENDPOINT", "https://abc.r2.cloudflarestorage.com/")
	s, ok := newMediaStore("/data/relief.db").(*s3Store)
	if !ok || s.endpoint != "https://abc.r2.cloudflarestorage.com" || s.region != "us-east-1" {
		t.Fatalf("object storage: %+v", s)
	}
}

func rawUpload(s *serviceHarness, path string, cookie *http.Cookie, csrf, ctype string, data []byte) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", path, bytes.NewReader(data))
	r.Header.Set("Content-Type", ctype)
	r.Header.Set("X-File-Name", "foto%20uno.jpg")
	r.Header.Set("X-CSRF-Token", csrf)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w
}

func TestRunMediaPickupDeliveryAndAdminCuration(t *testing.T) {
	s := newServiceHarness(t)
	id := s.expect("POST", "/api/admin/camps/"+s.camp+"/needs", `{"need":"Agua","urgency":"normal"}`, 201)["id"].(string)
	cookie, csrf := s.crewLogin("/api/runner", "Rosa")
	jpeg := []byte("\xff\xd8\xff fake jpeg")

	if w := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=pickup", cookie, csrf, "image/jpeg", jpeg); w.Code != 404 {
		t.Fatalf("runner can't document a run they haven't taken: %d %s", w.Code, w.Body.String())
	}
	s.crewCall("POST", "/api/runner/needs/"+id+"/take", "{}", cookie, csrf)
	for ctype, want := range map[string]int{"image/svg+xml": 415, "text/html": 415} {
		if w := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=pickup", cookie, csrf, ctype, jpeg); w.Code != want {
			t.Fatalf("%s: %d", ctype, w.Code)
		}
	}
	if w := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=other", cookie, csrf, "image/jpeg", jpeg); w.Code != 400 {
		t.Fatalf("bad stage: %d", w.Code)
	}
	if w := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=pickup", cookie, "wrong", "image/jpeg", jpeg); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	w := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=pickup", cookie, csrf, "image/jpeg", jpeg)
	if w.Code != 201 {
		t.Fatalf("pickup upload: %d %s", w.Code, w.Body.String())
	}
	var pickup MediaItem
	json.Unmarshal(w.Body.Bytes(), &pickup)
	if pickup.Stage != "pickup" || pickup.Name != "foto uno.jpg" || pickup.AddedByName != "Rosa" || pickup.Size != int64(len(jpeg)) {
		t.Fatalf("pickup item: %+v", pickup)
	}
	stored := filepath.Join(filepath.Dir(s.app.db.Path()), "media", "needs", id, pickup.ID+".jpg")
	if data, err := os.ReadFile(stored); err != nil || !bytes.Equal(data, jpeg) {
		t.Fatalf("stored on disk without credentials: %v", err)
	}
	video := rawUpload(s, "/api/runner/needs/"+id+"/media?stage=delivery", cookie, csrf, "video/mp4", []byte("mp4 data"))
	if video.Code != 201 {
		t.Fatalf("delivery video: %d %s", video.Code, video.Body.String())
	}
	if body := s.crewCall("GET", "/api/runner/needs", "", cookie, "").Body.String(); !strings.Contains(body, `"stage":"pickup"`) || !strings.Contains(body, `"stage":"delivery"`) {
		t.Fatalf("runner list shows their media: %s", body)
	}
	if got := s.crewCall("GET", "/api/runner/media/"+id+"/"+pickup.ID, "", cookie, ""); got.Code != 200 || got.Body.String() != string(jpeg) || got.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("runner views own photo: %d", got.Code)
	}
	other, _ := s.crewLogin("/api/runner", "Otra")
	if got := s.crewCall("GET", "/api/runner/media/"+id+"/"+pickup.ID, "", other, ""); got.Code != 404 {
		t.Fatalf("other runner must not see it: %d", got.Code)
	}
	if body := s.call("GET", "/api/camps/"+s.camp+"/needs", "", false).Body.String(); strings.Contains(body, "media") || strings.Contains(body, pickup.ID) {
		t.Fatalf("public needs must not include media: %s", body)
	}

	// Admin: sees, moves between pickup and delivery, deletes, and adds.
	if got := s.call("GET", "/api/admin/media/"+id+"/"+pickup.ID, "", true); got.Code != 200 || got.Body.String() != string(jpeg) {
		t.Fatalf("admin views photo: %d", got.Code)
	}
	if m := s.expect("PATCH", "/api/admin/needs/"+id+"/media/"+pickup.ID, `{"stage":"delivery"}`, 200); m["stage"] != "delivery" {
		t.Fatalf("move to delivery: %v", m)
	}
	req := httptest.NewRequest("DELETE", "/api/admin/needs/"+id+"/media/"+pickup.ID, nil)
	req.AddCookie(s.cookie)
	rec := httptest.NewRecorder()
	s.h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("delete requires CSRF: %d", rec.Code)
	}
	if got := s.call("DELETE", "/api/admin/needs/"+id+"/media/"+pickup.ID, "", true); got.Code != 200 {
		t.Fatalf("delete: %d %s", got.Code, got.Body.String())
	}
	if _, err := os.Stat(stored); !os.IsNotExist(err) {
		t.Fatalf("deleted file must be removed: %v", err)
	}
	if got := rawUpload(s, "/api/admin/needs/"+id+"/media?stage=pickup", s.cookie, s.csrf, "image/png", []byte("png")); got.Code != 201 || !strings.Contains(got.Body.String(), "Administración") {
		t.Fatalf("admin upload: %d %s", got.Code, got.Body.String())
	}
	var needs []CampNeed
	json.Unmarshal(s.call("GET", "/api/admin/camps/"+s.camp+"/needs", "", true).Body.Bytes(), &needs)
	if len(needs) != 1 || len(needs[0].Media) != 2 {
		t.Fatalf("media after curation: %+v", needs)
	}
}
