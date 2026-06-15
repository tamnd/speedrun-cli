// Package speedrun is the library behind the speedrun command line:
// the HTTP client, request shaping, and typed data models for speedrun.com.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite, and retries the
// transient failures (429 and 5xx) that any public site throws under load.
package speedrun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Host is the site this client talks to.
const Host = "www.speedrun.com"

// BaseURL is the root every API request is built from.
const BaseURL = "https://www.speedrun.com/api/v1"

// DefaultUserAgent identifies this client honestly.
const DefaultUserAgent = "speedrun-cli/0.1 (tamnd87@gmail.com)"

// Client talks to speedrun.com API over HTTP.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	BaseURL   string
	// Rate is the minimum gap between requests.
	Rate    time.Duration
	Retries int

	last time.Time
}

// NewClient returns a Client with sensible defaults.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 20 * time.Second},
		UserAgent: DefaultUserAgent,
		BaseURL:   BaseURL,
		Rate:      500 * time.Millisecond,
		Retries:   3,
	}
}

// Get fetches a URL and returns the body bytes. Paces and retries automatically.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// --- Wire types (raw API response shapes) ---

type apiGameNames struct {
	International string `json:"international"`
}

type apiGame struct {
	ID       string       `json:"id"`
	Names    apiGameNames `json:"names"`
	WebLink  string       `json:"weblink"`
	Released int          `json:"released"`
}

type apiGamesResponse struct {
	Data       []apiGame     `json:"data"`
	Pagination apiPagination `json:"pagination"`
}

type apiPagination struct {
	Offset int `json:"offset"`
	Max    int `json:"max"`
	Size   int `json:"size"`
}

type apiRunStatus struct {
	Status     string `json:"status"`
	Examiner   string `json:"examiner"`
	VerifyDate string `json:"verify-date"`
}

type apiRunTimes struct {
	Primary   string  `json:"primary"`
	PrimaryT  float64 `json:"primary_t"`
	Realtime  string  `json:"realtime"`
	RealtimeT float64 `json:"realtime_t"`
}

type apiRunPlayer struct {
	Rel string `json:"rel"`
	ID  string `json:"id"`
	URI string `json:"uri"`
}

type apiRun struct {
	ID       string         `json:"id"`
	WebLink  string         `json:"weblink"`
	Game     string         `json:"game"`
	Category string         `json:"category"`
	Status   apiRunStatus   `json:"status"`
	Times    apiRunTimes    `json:"times"`
	Players  []apiRunPlayer `json:"players"`
	Date     string         `json:"date"`
}

type apiRunsResponse struct {
	Data []apiRun `json:"data"`
}

type apiLeaderboardRun struct {
	Place int    `json:"place"`
	Run   apiRun `json:"run"`
}

type apiLeaderboard struct {
	Game     string              `json:"game"`
	Category string              `json:"category"`
	Runs     []apiLeaderboardRun `json:"runs"`
}

type apiLeaderboardResponse struct {
	Data apiLeaderboard `json:"data"`
}

type apiCategory struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	WebLink string `json:"weblink"`
	Type    string `json:"type"`
}

type apiCategoriesResponse struct {
	Data []apiCategory `json:"data"`
}

// --- Output types ---

// Game is a single speedrun.com game record.
type Game struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	WebLink  string `json:"weblink"`
	Released int    `json:"released"`
}

// Run is a single speedrun record.
type Run struct {
	ID          string  `json:"id"`
	WebLink     string  `json:"weblink"`
	Game        string  `json:"game"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	PrimaryTime float64 `json:"primary_time"`
	Date        string  `json:"date"`
}

// LeaderboardEntry is a single entry in a game leaderboard.
type LeaderboardEntry struct {
	Place       int     `json:"place"`
	RunID       string  `json:"run_id"`
	PrimaryTime float64 `json:"primary_time"`
	PlayerID    string  `json:"player_id"`
}

// Category is a single run category for a game.
type Category struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	WebLink string `json:"weblink"`
	Type    string `json:"type"`
}

// --- Client methods ---

// ListGames searches for games by name (empty search returns recent games).
func (c *Client) ListGames(ctx context.Context, search string, limit int) ([]*Game, error) {
	if limit <= 0 {
		limit = 20
	}
	url := fmt.Sprintf("%s/games?limit=%d", c.BaseURL, limit)
	if search != "" {
		url += "&name=" + search
	}
	body, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp apiGamesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse games: %w", err)
	}
	out := make([]*Game, 0, len(resp.Data))
	for _, g := range resp.Data {
		out = append(out, &Game{
			ID:       g.ID,
			Name:     g.Names.International,
			WebLink:  g.WebLink,
			Released: g.Released,
		})
	}
	return out, nil
}

// ListRuns fetches recent verified runs, optionally filtered by game ID.
func (c *Client) ListRuns(ctx context.Context, gameID string, limit int) ([]*Run, error) {
	if limit <= 0 {
		limit = 20
	}
	url := fmt.Sprintf("%s/runs?status=verified&max=%d", c.BaseURL, limit)
	if gameID != "" {
		url += "&game=" + gameID
	}
	body, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp apiRunsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse runs: %w", err)
	}
	out := make([]*Run, 0, len(resp.Data))
	for _, r := range resp.Data {
		out = append(out, &Run{
			ID:          r.ID,
			WebLink:     r.WebLink,
			Game:        r.Game,
			Category:    r.Category,
			Status:      r.Status.Status,
			PrimaryTime: r.Times.PrimaryT,
			Date:        r.Date,
		})
	}
	return out, nil
}

// GetLeaderboard fetches the leaderboard for a game and category.
func (c *Client) GetLeaderboard(ctx context.Context, gameID, categoryID string, top int) ([]*LeaderboardEntry, error) {
	if top <= 0 {
		top = 10
	}
	url := fmt.Sprintf("%s/leaderboards/%s/category/%s?top=%d", c.BaseURL, gameID, categoryID, top)
	body, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp apiLeaderboardResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse leaderboard: %w", err)
	}
	out := make([]*LeaderboardEntry, 0, len(resp.Data.Runs))
	for _, e := range resp.Data.Runs {
		playerID := ""
		if len(e.Run.Players) > 0 {
			playerID = e.Run.Players[0].ID
		}
		out = append(out, &LeaderboardEntry{
			Place:       e.Place,
			RunID:       e.Run.ID,
			PrimaryTime: e.Run.Times.PrimaryT,
			PlayerID:    playerID,
		})
	}
	return out, nil
}

// ListCategories fetches all categories for a game.
func (c *Client) ListCategories(ctx context.Context, gameID string) ([]*Category, error) {
	url := fmt.Sprintf("%s/games/%s/categories", c.BaseURL, gameID)
	body, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp apiCategoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse categories: %w", err)
	}
	out := make([]*Category, 0, len(resp.Data))
	for _, cat := range resp.Data {
		out = append(out, &Category{
			ID:      cat.ID,
			Name:    cat.Name,
			WebLink: cat.WebLink,
			Type:    cat.Type,
		})
	}
	return out, nil
}

// formatTime converts seconds to "H:MM:SS.mmm" or "M:SS.mmm" format.
func formatTime(secs float64) string {
	h := int(secs) / 3600
	m := (int(secs) % 3600) / 60
	s := int(secs) % 60
	ms := int((secs - float64(int(secs))) * 1000)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", h, m, s, ms)
	}
	return fmt.Sprintf("%d:%02d.%03d", m, s, ms)
}
