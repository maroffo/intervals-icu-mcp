# ABOUTME: Changelog for intervals-icu-mcp releases
# ABOUTME: Follows Keep-a-Changelog format (https://keepachangelog.com/)

# Changelog

All notable changes to this project are documented here.
Format: Keep-a-Changelog; project follows SemVer.

## [Unreleased]

### Added
- `update_activity` tool (MUTATES DATA): partial `PUT /activity/{id}` limited to
  `type, name, description, icu_rpe, feel, commute, trainer, icu_ignore_hr,
  icu_ignore_power, icu_ignore_time`. Other fields, wrong JSON types and
  activity types outside the API spec enum are rejected before any request.
  Returns the id, the updated values and the new `icu_training_load` /
  `icu_ctl` / `icu_atl`. See ADR-0005.
- Strava-sourced activities get a clear error instead of a raw HTTP error.
  Detection is heuristic (the API spec does not document the error shape) and
  has not been verified against a live Strava activity.

### Changed
- All activity tools reject `.` and `..` as `activity_id`: `url.PathEscape`
  leaves them as dot segments, which could resolve to another endpoint.

## [0.1.0] - 2026-04-21

### Added
- Initial MCP server with 14 tools: activities (list/get/streams/intervals),
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
