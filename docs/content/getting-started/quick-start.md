---
title: "Quick start"
description: "Query speedrun.com from the command line."
weight: 30
---

Once `speedrun` is on your `PATH`, search for a game:

```bash
speedrun games --search=mario
```

By default you get an aligned table. Ask for JSON when you want to pipe it:

```bash
$ speedrun games --search=mario -o json
[
  {
    "id": "pd0wq31e",
    "name": "Super Mario 64",
    "weblink": "https://www.speedrun.com/sm64",
    "released": 1996
  }
]
```

## Browse a leaderboard

Use the game ID from `games` to find categories, then pull the leaderboard:

```bash
speedrun categories pd0wq31e
speedrun leaderboard pd0wq31e w20e9lpd --top=10
```

## Browse recent runs

```bash
speedrun runs                          # recent verified runs site-wide
speedrun runs --game=pd0wq31e          # runs for Super Mario 64 only
speedrun runs --game=pd0wq31e -o jsonl | jq .primary_time
```

## Shape the output

The same flags work on every command:

```bash
speedrun games --search=mario --fields id,name   # keep only these columns
speedrun leaderboard pd0wq31e w20e9lpd -o jsonl  # one object per line
```

`-o` takes `table`, `json`, `jsonl`, `csv`, `tsv`, or `raw`. Left to
`auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream. See
[output formats](/reference/output/) for the full contract.

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
speedrun serve --addr :7777 &
curl -s 'localhost:7777/v1/games?search=mario'   # NDJSON, one record per line
speedrun mcp                                     # MCP over stdio
```
