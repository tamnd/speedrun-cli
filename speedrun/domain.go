package speedrun

import (
	"context"

	"github.com/tamnd/any-cli/kit"
)

func init() { kit.Register(Domain{}) }

// Domain is the speedrun.com driver.
type Domain struct{}

// Info describes the scheme, hostnames, and binary identity.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "speedrun",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "speedrun",
			Short:  "A command line for speedrun.com",
			Long: `speedrun reads public speedrun.com data over plain HTTPS, shapes it into
clean records, and prints output that pipes into the rest of your tools.

No API key required. Speedrun.com tracks 36,000+ games and 5.7M+ submitted runs.`,
			Site: Host,
			Repo: "https://github.com/tamnd/speedrun-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "games", Group: "read", List: true,
		Summary: "List or search games on speedrun.com (--search, --limit)"}, listGames)

	kit.Handle(app, kit.OpMeta{Name: "runs", Group: "read", List: true,
		Summary: "List recent verified speedruns (--game for specific game)"}, listRuns)

	kit.Handle(app, kit.OpMeta{Name: "leaderboard", Group: "read", List: true,
		Summary: "Get leaderboard for a game and category",
		Args: []kit.Arg{
			{Name: "game-id", Help: "game ID (e.g. pd0wq31e for Super Mario 64)"},
			{Name: "category-id", Help: "category ID"},
		}}, getLeaderboard)

	kit.Handle(app, kit.OpMeta{Name: "categories", Group: "read", List: true,
		Summary: "List categories for a game",
		Args:    []kit.Arg{{Name: "game-id", Help: "game ID"}}}, listCategories)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// --- inputs ---

type gamesInput struct {
	Search string  `kit:"flag" help:"search by name"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type runsInput struct {
	Game   string  `kit:"flag" help:"game ID filter"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type leaderboardInput struct {
	GameID     string  `kit:"arg" help:"game ID"`
	CategoryID string  `kit:"arg" help:"category ID"`
	Top        int     `kit:"flag" help:"top N results"`
	Client     *Client `kit:"inject"`
}

type categoriesInput struct {
	GameID string  `kit:"arg" help:"game ID"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listGames(ctx context.Context, in gamesInput, emit func(*Game) error) error {
	games, err := in.Client.ListGames(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, g := range games {
		if err := emit(g); err != nil {
			return err
		}
	}
	return nil
}

func listRuns(ctx context.Context, in runsInput, emit func(*Run) error) error {
	runs, err := in.Client.ListRuns(ctx, in.Game, in.Limit)
	if err != nil {
		return err
	}
	for _, r := range runs {
		if err := emit(r); err != nil {
			return err
		}
	}
	return nil
}

func getLeaderboard(ctx context.Context, in leaderboardInput, emit func(*LeaderboardEntry) error) error {
	entries, err := in.Client.GetLeaderboard(ctx, in.GameID, in.CategoryID, in.Top)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := emit(e); err != nil {
			return err
		}
	}
	return nil
}

func listCategories(ctx context.Context, in categoriesInput, emit func(*Category) error) error {
	cats, err := in.Client.ListCategories(ctx, in.GameID)
	if err != nil {
		return err
	}
	for _, cat := range cats {
		if err := emit(cat); err != nil {
			return err
		}
	}
	return nil
}

// Classify turns any accepted input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	return "game", input, nil
}

// Locate is the inverse: the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	return "https://" + Host + "/" + id, nil
}
