---
title: "speedrun"
description: "A command line for speedrun.com — search games, browse leaderboards, and stream verified runs. No API key required."
heroTitle: "speedrun.com, from the command line"
heroLead: "Search 36,000+ games, browse leaderboards, and stream verified runs from speedrun.com. One pure-Go binary, no API key, output that pipes into the rest of your tools."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`speedrun` reads public speedrun.com data over plain HTTPS, shapes it into
clean records, and gets out of your way.

```bash
speedrun games --search=mario          # search games by name
speedrun categories pd0wq31e           # list categories for Super Mario 64
speedrun leaderboard pd0wq31e w20e9lpd # top runs in 120 Star category
speedrun runs --game=pd0wq31e          # recent verified runs for a game
speedrun serve --addr :7777            # the same operations over HTTP
```

There is nothing to sign up for and nothing to run alongside it. Output adapts
to where it goes: an aligned table on your terminal, JSONL the moment you pipe
it somewhere.

## What you can do

- **Search games** across 36,000+ titles tracked on speedrun.com
- **Browse leaderboards** for any game and category combination
- **Stream recent runs** filtered by game or across the whole site
- **List categories** to discover what run types a game supports

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Doing a specific job? The [guides](/guides/) are task-first.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
