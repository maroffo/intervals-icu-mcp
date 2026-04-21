# ABOUTME: Changelog for intervals-icu-mcp releases
# ABOUTME: Follows Keep-a-Changelog format (https://keepachangelog.com/)

# Changelog

All notable changes to this project are documented here.
Format: Keep-a-Changelog; project follows SemVer.

## [Unreleased]

## [0.1.0] - 2026-04-21

### Added
- Initial MCP server with 13 tools: activities (list/get/streams/intervals),
  wellness (get/list/update), events (list/create/update/delete), athlete
  (profile/fitness/folders).
- HTTP Basic auth against intervals.icu API v1.
- Client-side rate limiting (3 rps) + retry on 429/5xx with exponential backoff
  and Retry-After header support (RFC 7231 seconds + HTTP-date, capped at 60s).
- stdio transport via mark3labs/mcp-go v0.48.0.
- Graceful shutdown on SIGINT/SIGTERM.
- Structured logging on stderr via log/slog.
- Opt-in integration tests (build tag `integration`).
- Apache License 2.0.
- GitHub Actions CI (vet + race-enabled tests + build on Go 1.25 and 1.26).
