package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestGPSTrackIsRecordedThinnedAndAdminOnly(t *testing.T) {
	s := newServiceHarness(t)
	fufo := s.expect("POST", "/api/admin/users", `{"name":"Fufo","telephone":"6999-1111","roles":["runner"],"accessCode":"FUFO12"}`, 201)["id"].(string)
	viewer := s.viewerCookie("FUFO12")
	s.as(viewer, "POST", "/api/full/location/consent", `{"accept":true}`)
	for _, body := range []string{
		`{"lat":7.5,"lon":-80.3,"accuracy":9}`,
		`{"lat":7.50003,"lon":-80.3}`, // ~3 m and seconds later: skipped
		`{"lat":7.501,"lon":-80.3}`,   // ~110 m: recorded
	} {
		if w := s.as(viewer, "POST", "/api/full/location", body); w.Code != 200 {
			t.Fatalf("share: %d %s", w.Code, w.Body.String())
		}
	}
	// Stopping hides the live pin but keeps the history.
	s.as(viewer, "POST", "/api/full/location/stop", "{}")
	var got struct {
		Points    []trackPoint
		Truncated bool
	}
	json.Unmarshal(s.call("GET", "/api/admin/users/"+fufo+"/track", "", true).Body.Bytes(), &got)
	if len(got.Points) != 2 || got.Points[0].Lat != 7.5 || got.Points[0].Accuracy != 9 || got.Points[1].Lat != 7.501 || got.Truncated {
		t.Fatalf("thinned history of the last 24 h: %+v", got)
	}
	if w := s.call("GET", "/api/admin/users/"+fufo+"/track", "", false); w.Code != 401 {
		t.Fatalf("admin only: %d", w.Code)
	}
	if w := s.as(viewer, "GET", "/api/admin/users/"+fufo+"/track", ""); w.Code != 401 {
		t.Fatalf("not for code users: %d", w.Code)
	}
	if w := s.call("GET", "/api/admin/users/nadie/track", "", true); w.Code != 404 {
		t.Fatalf("unknown user: %d", w.Code)
	}
	if w := s.call("GET", "/api/admin/users/"+fufo+"/track?from=ayer", "", true); w.Code != 400 {
		t.Fatalf("bad date: %d", w.Code)
	}
	past := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	json.Unmarshal(s.call("GET", "/api/admin/users/"+fufo+"/track?from="+past+"&to="+time.Now().Add(-25*time.Hour).UTC().Format(time.RFC3339), "", true).Body.Bytes(), &got)
	if len(got.Points) != 0 {
		t.Fatalf("range filter: %+v", got.Points)
	}
	if strings.Contains(s.call("GET", "/api/points", "", false).Body.String(), "7.501") {
		t.Fatal("never public")
	}
}

func TestGPSTrackPrunesOldPoints(t *testing.T) {
	s := newServiceHarness(t)
	now := time.Now().UTC()
	s.app.db.Update(func(tx *bolt.Tx) error {
		for _, p := range []struct {
			user string
			at   time.Time
		}{{"a", now.Add(-40 * 24 * time.Hour)}, {"a", now.Add(-time.Hour)}, {"b", now.Add(-31 * 24 * time.Hour)}} {
			if e := recordTrack(tx, p.user, liveLocation{Lat: 7.5, Lon: -80.3, At: p.at}); e != nil {
				t.Fatal(e)
			}
		}
		return pruneTracks(tx, now.Add(-trackKeep))
	})
	s.app.db.View(func(tx *bolt.Tx) error {
		if pts, _ := userTrack(tx, "a", now.Add(-60*24*time.Hour), now); len(pts) != 1 {
			t.Fatalf("keeps the recent point: %+v", pts)
		}
		if tx.Bucket(trackBucket).Bucket([]byte("b")) != nil {
			t.Fatal("empty user history removed")
		}
		return nil
	})
}
