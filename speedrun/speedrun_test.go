package speedrun

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0 // no pacing in the test

	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	c := NewClient()
	c.Rate = 0
	c.Retries = 5

	start := time.Now()
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("body = %q after retries", body)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestListGames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/games" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":       "pd0wq31e",
					"names":    map[string]interface{}{"international": "Super Mario 64"},
					"weblink":  "https://www.speedrun.com/sm64",
					"released": 1996,
				},
			},
			"pagination": map[string]interface{}{"offset": 0, "max": 20, "size": 1},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.BaseURL = ts.URL
	c.Rate = 0

	games, err := c.ListGames(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("ListGames: %v", err)
	}
	if len(games) != 1 {
		t.Fatalf("want 1 game, got %d", len(games))
	}
	g := games[0]
	if g.ID != "pd0wq31e" {
		t.Errorf("ID: got %q, want %q", g.ID, "pd0wq31e")
	}
	if g.Name != "Super Mario 64" {
		t.Errorf("Name: got %q, want %q", g.Name, "Super Mario 64")
	}
	if g.Released != 1996 {
		t.Errorf("Released: got %d, want 1996", g.Released)
	}
}

func TestListRuns(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runs" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":       "ekz5onyr",
					"weblink":  "https://www.speedrun.com/run/ekz5onyr",
					"game":     "pd0wq31e",
					"category": "w20e9lpd",
					"status":   map[string]interface{}{"status": "verified", "examiner": "abc"},
					"times":    map[string]interface{}{"primary": "PT2H11M0S", "primary_t": 7860.0, "realtime": "PT2H11M0S", "realtime_t": 7860.0},
					"players":  []map[string]interface{}{{"rel": "user", "id": "userid1"}},
					"date":     "2024-01-01",
				},
			},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.BaseURL = ts.URL
	c.Rate = 0

	runs, err := c.ListRuns(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("want 1 run, got %d", len(runs))
	}
	r := runs[0]
	if r.ID != "ekz5onyr" {
		t.Errorf("ID: got %q, want %q", r.ID, "ekz5onyr")
	}
	if r.Status != "verified" {
		t.Errorf("Status: got %q, want %q", r.Status, "verified")
	}
	if r.PrimaryTime != 7860.0 {
		t.Errorf("PrimaryTime: got %v, want 7860.0", r.PrimaryTime)
	}
}

func TestGetLeaderboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"game":     "pd0wq31e",
				"category": "w20e9lpd",
				"runs": []map[string]interface{}{
					{
						"place": 1,
						"run": map[string]interface{}{
							"id":      "run1",
							"times":   map[string]interface{}{"primary_t": 6785.0},
							"players": []map[string]interface{}{{"rel": "user", "id": "player1"}},
						},
					},
					{
						"place": 2,
						"run": map[string]interface{}{
							"id":      "run2",
							"times":   map[string]interface{}{"primary_t": 6900.0},
							"players": []map[string]interface{}{{"rel": "user", "id": "player2"}},
						},
					},
				},
			},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.BaseURL = ts.URL
	c.Rate = 0

	entries, err := c.GetLeaderboard(context.Background(), "pd0wq31e", "w20e9lpd", 10)
	if err != nil {
		t.Fatalf("GetLeaderboard: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	if entries[0].Place != 1 {
		t.Errorf("Place: got %d, want 1", entries[0].Place)
	}
	if entries[0].PrimaryTime != 6785.0 {
		t.Errorf("PrimaryTime: got %v, want 6785.0", entries[0].PrimaryTime)
	}
	if entries[0].PlayerID != "player1" {
		t.Errorf("PlayerID: got %q, want %q", entries[0].PlayerID, "player1")
	}
}

func TestListCategories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":      "w20e9lpd",
					"name":    "120 Star",
					"weblink": "https://www.speedrun.com/sm64#120_Star",
					"type":    "per-game",
				},
				{
					"id":      "wkpoo02r",
					"name":    "70 Star",
					"weblink": "https://www.speedrun.com/sm64#70_Star",
					"type":    "per-game",
				},
			},
		})
	}))
	defer ts.Close()

	c := NewClient()
	c.BaseURL = ts.URL
	c.Rate = 0

	cats, err := c.ListCategories(context.Background(), "pd0wq31e")
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("want 2 categories, got %d", len(cats))
	}
	if cats[0].ID != "w20e9lpd" {
		t.Errorf("ID: got %q, want %q", cats[0].ID, "w20e9lpd")
	}
	if cats[0].Name != "120 Star" {
		t.Errorf("Name: got %q, want %q", cats[0].Name, "120 Star")
	}
}

func TestFormatTime(t *testing.T) {
	cases := []struct {
		secs float64
		want string
	}{
		{0, "0:00.000"},
		{61.5, "1:01.500"},
		{3661.123, "1:01:01.123"},
		{6785.0, "1:53:05.000"},
		{7860.0, "2:11:00.000"},
	}
	for _, tc := range cases {
		got := formatTime(tc.secs)
		if got != tc.want {
			t.Errorf("formatTime(%v) = %q, want %q", tc.secs, got, tc.want)
		}
	}
}
