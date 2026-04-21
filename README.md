# ABOUTME: Local MCP server in Go for intervals.icu REST API
# ABOUTME: Exposes 13 tools (activities, wellness, events, athlete) over stdio

# intervals-icu-mcp

Local [Model Context Protocol](https://modelcontextprotocol.io/) server that exposes
[intervals.icu](https://intervals.icu) API endpoints to MCP clients (Claude Code,
Claude Desktop, etc.).

## What it does

13 tools, grouped by domain:

| Domain | Tools |
|---|---|
| Activities (read) | `list_activities`, `get_activity`, `get_activity_streams`, `get_activity_intervals` |
| Wellness | `get_wellness`, `list_wellness`, `update_wellness` (mutates) |
| Events / Calendar | `list_events`, `create_event`, `update_event`, `delete_event` (mutate) |
| Athlete / Fitness | `get_athlete`, `get_fitness`, `list_folders` |

All tools return raw JSON as text. Mutating tools are tagged in their description.

## Setup

### 1. Get your intervals.icu API key

Go to [intervals.icu Settings](https://intervals.icu/settings), open the
Developer Settings section, generate a key.

### 2. Build

```bash
make build         # produces ./bin/intervals-icu-mcp
# or
make install       # installs to $GOPATH/bin
```

Requires Go >= 1.25.

### 3. Configure in your MCP client

#### Claude Code (project or user scope, `~/.claude.json` or `.mcp.json`)

```json
{
  "mcpServers": {
    "intervals-icu": {
      "command": "/absolute/path/to/intervals-icu-mcp",
      "env": {
        "INTERVALS_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

#### Claude Desktop (`claude_desktop_config.json`)

Same `mcpServers` block, placed in `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS.

Any MCP client supporting stdio servers uses the same `mcpServers` block; path varies by client.

### Environment variables

| Var | Required | Default | Purpose |
|---|---|---|---|
| `INTERVALS_API_KEY` | yes | - | API key from intervals.icu Developer Settings. Paste the raw key only; the `API_KEY:` Basic-auth username is added automatically. |
| `INTERVALS_ATHLETE_ID` | no | `0` | Leave unset for normal use. Set to an explicit id only for coach accounts managing other athletes. |
| `INTERVALS_BASE_URL` | no | `https://intervals.icu/api/v1` | Override only for staging or when behind a proxy. |

### `INTERVALS_BASE_URL`

Optional override. Defaults to `https://intervals.icu/api/v1`; point it at a
staging endpoint or a local reverse proxy if you need to intercept traffic for
debugging. Unset in production.

## Development

```bash
make check         # go vet + unit tests
make test          # unit tests only
make lint          # staticcheck (install first: go install honnef.co/go/tools/cmd/staticcheck@latest)
make integration   # opt-in live API test, requires INTERVALS_API_KEY
make clean
```

Integration tests are gated behind the `integration` build tag and hit the real
intervals.icu API; they skip if `INTERVALS_API_KEY` is not set.

## Behavior notes

- **Authentication**: HTTP Basic with literal username `API_KEY` and the key as password (intervals.icu convention).
- **Rate limiting**: client-side limiter at 3 req/s with a burst of 3.
- **Retry**: automatic retry on `429` and `5xx` up to 3 attempts with exponential backoff, honoring `Retry-After` when present.
- **Typing**: responses are passed through as raw JSON. The backend schema is stable
  in practice but occasionally grows new fields, so the server stays tolerant.
- **Transport**: stdio only. No HTTP/SSE mode.

## Troubleshooting

### Server does not appear in Claude Code / Claude Desktop

- Config must use an **absolute** path to the binary, not a relative or `~`-relative one.
- The binary must be executable: `chmod +x /absolute/path/to/intervals-icu-mcp`.
- Restart the client after editing the MCP config; stdio servers are spawned at client startup.
- Inspect the client log (Claude Desktop: Developer menu -> Open log) for spawn errors.

### `401 Unauthorized`

- The API key is wrong, revoked, or you accidentally pasted the Basic-auth
  username: set `INTERVALS_API_KEY` to the **raw key only**, no `API_KEY:`
  prefix. The server constructs the Basic auth header itself.
- Regenerate the key in intervals.icu Developer Settings if unsure.

### `429 Too Many Requests`

- The server rate-limits itself at 3 req/s with retry on `429`, so normal
  interactive use rarely hits this. If it persists, the backend is under
  pressure: reduce how often your prompts ask for bulk data (`list_activities`
  with very wide date ranges is the usual culprit).
- Respect any `Retry-After` guidance; the client honors it up to 60s per
  attempt and then gives up.

## Architecture Decisions

See `docs/adr/` for the short-form decision records:

- [ADR-0001: SDK choice (`mark3labs/mcp-go`)](docs/adr/0001-sdk-mark3labs.md)
- [ADR-0002: Transport — stdio only](docs/adr/0002-transport-stdio.md)
- [ADR-0003: Client rate limit at 3 rps](docs/adr/0003-client-rate-limit-3rps.md)
- [ADR-0004: Tolerant JSON pass-through, no strict DTOs](docs/adr/0004-tolerant-json-passthrough.md)

## License

Personal project. Use at your own risk.
