# ABOUTME: Local MCP server in Go for intervals.icu REST API
# ABOUTME: Exposes 13 tools (activities, wellness, events, athlete) over stdio

# intervals-icu-mcp

[![CI](https://github.com/maroffo/intervals-icu-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/maroffo/intervals-icu-mcp/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/maroffo/intervals-icu-mcp)](https://goreportcard.com/report/github.com/maroffo/intervals-icu-mcp)
[![Go Reference](https://pkg.go.dev/badge/github.com/maroffo/intervals-icu-mcp.svg)](https://pkg.go.dev/github.com/maroffo/intervals-icu-mcp)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-2025--11--25-informational)](https://modelcontextprotocol.io/)

A [Model Context Protocol](https://modelcontextprotocol.io/) server that lets
Claude Code, Claude Desktop, and any other MCP client query and update your
[intervals.icu](https://intervals.icu) training data directly from a
conversation.

Ask your assistant things like _"what did I run last week?"_, _"what's my CTL
trend this month?"_, _"log today's HRV at 62 and fatigue 3"_, _"schedule a 90
min endurance ride for Saturday"_, without leaving the chat.

## Table of Contents

- [Features](#features)
- [Quick start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Tool catalog](#tool-catalog)
- [Example prompts](#example-prompts)
- [Development](#development)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features

- **13 tools** covering activities, wellness, calendar events, and athlete fitness
- **Safe by default**: rate-limited (3 req/s) with retry on 429/5xx, bounded Retry-After
- **Read + write**: get activities, read/update wellness, full CRUD on calendar events
- **Graceful shutdown** on SIGINT/SIGTERM, structured JSON logging on stderr
- **Tolerant parsing**: raw JSON pass-through so schema changes on intervals.icu don't break the server
- **Zero config beyond your API key**: sensible defaults, works out of the box with a coach or personal account
- **stdio transport**: local-only, no network listener, no open ports

## Quick start

```bash
# 1. Install
go install github.com/maroffo/intervals-icu-mcp@latest

# 2. Register with Claude Code (or your MCP client)
#    Edit ~/.claude.json and add:
{
  "mcpServers": {
    "intervals-icu": {
      "command": "/absolute/path/to/intervals-icu-mcp",
      "env": {
        "INTERVALS_API_KEY": "your-key-from-intervals.icu/settings"
      }
    }
  }
}

# 3. Restart your MCP client. The 13 tools are now available.
```

Get your API key at [intervals.icu Settings → Developer Settings](https://intervals.icu/settings).

## Installation

### `go install` (recommended)

```bash
go install github.com/maroffo/intervals-icu-mcp@latest
# Binary ends up in $(go env GOPATH)/bin
```

Requires Go 1.25+.

### Build from source

```bash
git clone https://github.com/maroffo/intervals-icu-mcp.git
cd intervals-icu-mcp
make build          # ./bin/intervals-icu-mcp
# or
make install        # to $GOPATH/bin
```

### Versioning

The binary stamps its version via `-ldflags`:

```bash
go build -ldflags "-X main.serverVersion=0.1.0" -o bin/intervals-icu-mcp .
```

Without `-ldflags`, the version falls back to `dev+<git-revision>` when built
from a git working copy, else `dev`.

## Configuration

### Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `INTERVALS_API_KEY` | yes | - | Your API key from intervals.icu Developer Settings. Paste the raw key; the `API_KEY:` Basic-auth username is prepended automatically. |
| `INTERVALS_ATHLETE_ID` | no | `0` | Leave unset for normal use (resolves to the key's own athlete). Set to an explicit id only for coach accounts. |
| `INTERVALS_BASE_URL` | no | `https://intervals.icu/api/v1` | Override only for staging, proxies, or record/replay testing. |

### Client integration

#### Claude Code (`~/.claude.json`)

```json
{
  "mcpServers": {
    "intervals-icu": {
      "command": "/absolute/path/to/intervals-icu-mcp",
      "env": {
        "INTERVALS_API_KEY": "..."
      }
    }
  }
}
```

#### Claude Desktop

Same `mcpServers` block, placed in `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS, `%APPDATA%\Claude\claude_desktop_config.json` on Windows.

#### Other clients (Cursor, Continue, Windsurf, Zed, …)

Any MCP client that supports **stdio** servers uses the same shape. The path to
the client config file varies; the `mcpServers` block is identical.

## Tool catalog

All tools return raw JSON as text. Mutating tools are tagged with `MUTATES DATA`
in their description and carry the MCP `destructive` annotation.

### Activities (read)

| Tool | Description | Required args |
|---|---|---|
| `list_activities` | List activities in a date range | _(none; oldest/newest/limit optional)_ |
| `get_activity` | Get a single activity by id | `activity_id` |
| `get_activity_streams` | Per-second data streams (watts, HR, cadence…) | `activity_id`; `types` optional |
| `get_activity_intervals` | Detected or manual intervals for an activity | `activity_id` |

### Wellness

| Tool | Description | Required args |
|---|---|---|
| `get_wellness` | Single day wellness (sleep, HRV, weight, fatigue…) | `date` (`yyyy-MM-dd`) |
| `list_wellness` | Wellness entries in a date range | _(none)_ |
| `update_wellness` | Mutate a day's wellness fields | `date`, `fields` (object) |

### Events / calendar

| Tool | Description | Required args |
|---|---|---|
| `list_events` | Calendar events (workouts, races, notes, fitness targets) | _(none)_ |
| `create_event` | Create a new calendar event | `event` (must include `start_date_local` + `category`) |
| `update_event` | Update a calendar event | `event_id`, `event` |
| `delete_event` | Delete a calendar event | `event_id` |

Event categories: `WORKOUT`, `RACE_A`, `RACE_B`, `RACE_C`, `NOTE`, `FITNESS`, `TARGET`.

### Athlete

| Tool | Description | Required args |
|---|---|---|
| `get_athlete` | Athlete profile (name, timezone, zones, sport settings) | _(none)_ |
| `get_fitness` | CTL/ATL/TSB time series for a date range | _(none)_ |
| `list_folders` | Workout library folders | _(none)_ |

## Example prompts

Once registered, you can drive intervals.icu through natural language:

> **Training review**
> "Summarize my rides from last week: total TSS, peak power, and how each session compared to the planned workout."

> **Wellness logging**
> "Log today's wellness: slept 7h15m, HRV 58, resting HR 48, fatigue 2, motivation 4."

> **Fitness trend**
> "What's my CTL/ATL/TSB over the last 30 days? Any ramp-rate warnings?"

> **Calendar planning**
> "Schedule a 2-hour Z2 endurance ride on Saturday morning and a 45-min recovery spin on Sunday."

> **Activity deep-dive**
> "Look at my last threshold interval session: pull the intervals, compute average power and drift, and flag anything unusual vs my FTP."

> **Race prep**
> "What races do I have scheduled in the next 8 weeks? Pull my fitness curve and tell me if I'm on track."

## Development

```bash
make help           # list all targets
make check          # go vet + unit tests
make test           # unit tests only
make integration    # live API tests (requires INTERVALS_API_KEY)
make test-e2e       # same as integration; exits 0 on skip for CI gates
make lint           # staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
make build          # produces bin/intervals-icu-mcp
make install        # installs to $GOPATH/bin
make clean
```

Integration tests are gated behind the `integration` build tag and hit the real
intervals.icu API; they skip cleanly when `INTERVALS_API_KEY` is not set, so they
are safe for CI pre-flight.

### Repository layout

```
intervals-icu-mcp/
├── main.go                       # stdio server entrypoint
├── internal/
│   ├── config/                   # env var loader
│   ├── icu/                      # HTTP client + domain methods
│   │   ├── client.go             # shared transport, auth, retry, rate limit
│   │   ├── activities.go         # list/get/streams/intervals
│   │   ├── wellness.go           # get/list/update
│   │   ├── events.go             # CRUD
│   │   └── athlete.go            # profile/fitness/folders
│   └── tools/                    # MCP tool handlers per domain
├── docs/adr/                     # architecture decision records
├── Makefile
├── CHANGELOG.md
└── LEARNING.md                   # session-level retrospective notes
```

## Architecture

See [`docs/adr/`](docs/adr/) for the decision records:

- [ADR-0001](docs/adr/0001-sdk-mark3labs.md) — Use `mark3labs/mcp-go` as the Go MCP SDK
- [ADR-0002](docs/adr/0002-transport-stdio.md) — stdio-only transport
- [ADR-0003](docs/adr/0003-client-rate-limit-3rps.md) — Client-side rate limit at 3 req/s
- [ADR-0004](docs/adr/0004-tolerant-json-passthrough.md) — Tolerant JSON pass-through, no strict DTOs

Behaviour notes:

- **Authentication**: HTTP Basic, username literal `API_KEY`, password = your API key (intervals.icu convention).
- **Retry**: automatic retry on `429` and `5xx` up to 3 attempts with exponential backoff, honouring `Retry-After` (seconds or HTTP-date), capped at 60s.
- **Typing**: responses pass through as raw JSON so new or renamed fields don't break anything.
- **Errors**: upstream error bodies are truncated to 200 chars in the error message; the full body remains accessible programmatically via `APIError.Body`.
- **Shutdown**: `SIGINT`/`SIGTERM` cancel the server context and close stdio cleanly.

## Troubleshooting

### Server does not appear in Claude Code / Claude Desktop

- The `command` path must be **absolute** (no `~`, no relative paths).
- The binary must be executable: `chmod +x /absolute/path/to/intervals-icu-mcp`.
- Restart the client after editing the MCP config; stdio servers are spawned at client startup.
- Check the client log (Claude Desktop: Developer → Open log) for spawn errors.

### `401 Unauthorized`

- Wrong or revoked API key, or you pasted the Basic-auth username into `INTERVALS_API_KEY`. Set the env var to the **raw key only**; the server adds the `API_KEY:` prefix itself.
- Regenerate the key at intervals.icu → Settings → Developer Settings if in doubt.

### `429 Too Many Requests`

- The server rate-limits itself at 3 req/s and retries on 429, so normal interactive use rarely hits this.
- If it persists, the backend is under pressure: reduce bulk queries (`list_activities` with very wide date ranges is the usual culprit).
- The client honours `Retry-After` up to 60s per attempt, then surfaces the error.

### Coach account querying another athlete

- Set `INTERVALS_ATHLETE_ID` to the target athlete id. Your API key must be authorised for that athlete (otherwise you'll see `403` instead of `401`).

### Version debugging

- The server logs the resolved version at boot on stderr as structured JSON. If you're on `dev` without a VCS revision, rebuild with `-ldflags "-X main.serverVersion=..."` or from a git working copy.

## Contributing

Issues and pull requests welcome. For larger changes, please open an issue first to discuss the approach.

Development workflow:

1. Fork and create a feature branch
2. `make check` must pass locally
3. Add tests for behaviour you change (unit tests in `*_test.go`, integration tests behind the `integration` build tag if relevant)
4. Open a PR against `main`

The repo follows [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

## License

Apache License 2.0. See [LICENSE](LICENSE).

## Acknowledgements

- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) — Go SDK for MCP
- [intervals.icu](https://intervals.icu) — the best training platform for analytical cyclists
- [Model Context Protocol](https://modelcontextprotocol.io/) — the standard behind this server
