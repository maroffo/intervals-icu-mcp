# ABOUTME: Living plan for the update_activity MCP tool (whitelisted partial PUT /activity/{id})
# ABOUTME: Decisions, progress and retrospective for the feat/update-activity branch

# Plan: `update_activity` tool

Branch: `feat/update-activity` (from `main` @ 41eb4c1). Complexity: moderate (known stack, prior art in repo: `update_event`, `update_wellness`).

## Goal

New MCP tool `update_activity` mapping to `PUT /api/v1/activity/{id}`: partial update restricted to a 10-field whitelist, `type` validated against the spec enum, Strava-sourced activities reported with a clear message, shaped response (id + changed fields + load/CTL/ATL).

## Spec facts (from https://intervals.icu/api/v1/docs, OpenAPI 3.0.1, fetched 2026-10-01)

- `PUT /api/v1/activity/{id}`, body `Activity`, 200 → `Activity`. Description: "Strava activities cannot be updated". Error body shape NOT documented.
- `Activity.type` is a bare `string`; the 60-value activity-type enum is shared by `SportInfo.type`, `SportSettings.types`, `Folder.activity_types`, etc. (exactly one distinct enum containing Ride/Run in the whole spec).
- Field types: `type`/`name`/`description` string; `icu_rpe`/`feel` int32; `commute`/`trainer`/`icu_ignore_hr`/`icu_ignore_power`/`icu_ignore_time` boolean; `icu_training_load` int32; `icu_ctl`/`icu_atl` float.
- `GET /activity/{id}`: "An empty stub object is returned for Strava activities"; `Activity.source` enum includes `STRAVA`.

## Files

| File | Change |
|------|--------|
| `internal/icu/activities.go` | `AllActivityTypes()`, `UpdateActivity`, whitelist + JSON-type + enum validation, `ErrStravaActivity`, `ErrActivityNotFound` |
| `internal/icu/activities_test.go` | domain tests |
| `internal/tools/activities.go` | register `update_activity`, handler, response shaping |
| `internal/tools/activities_test.go` | handler tests (success, invalid field, invalid type, Strava, 404, body = only sent keys) |
| `docs/adr/0005-whitelisted-activity-update.md` | new ADR: exception to ADR-0004 |
| `README.md` | tool catalog, counts 13→14, ADR list |
| `CHANGELOG.md` | `[Unreleased]` entry |

Excluded: `internal/icu/client.go` (no transport change), `client_integration_test.go` (integration tests are read-only by project rule; a PUT cannot be).

## Decisions

| # | Decision | Choice | Rationale | Revisit if |
|---|----------|--------|-----------|------------|
| 1 | Strava detection | Heuristic: 4xx body mentions "strava" (case-insensitive) OR 2xx body has `source=STRAVA` → `ErrStravaActivity` | Spec documents no error shape; no API key in session to probe. Flagged unverified in CHANGELOG/PR | Real error captured from a live Strava activity |
| 2 | Value validation | JSON-type check per spec (string/int/bool), no ranges | Clear local error for `icu_rpe:"7"`; spec has no ranges | Spec adds min/max |
| 3 | ADR | New ADR-0005 | ADR-0004 says handlers pass raw JSON through; this tool reshapes output and whitelists input | - |
| 4 | Validation layer | In `icu.UpdateActivity` (lowest layer), handler reuses it | No caller of `Client` can send a non-whitelisted field | - |
| 5 | Integration test | None | `client_integration_test.go`: "All integration tests must be read-only" | Max allows opt-in mutating tests |
| 6 | `updated` values | Taken from the response; a sent key absent from the response is omitted | Report server truth, never echo our own input as if confirmed | - |
| 7 | Lint | `go run honnef.co/go/tools/cmd/staticcheck@latest ./...` | staticcheck not installed (pre-existing: `make lint` fails at baseline); avoid installing binaries | - |
| 8 | Error identity (review #1) | `activityError` with `Unwrap() []error{sentinel, *APIError}` | Keeps the package's `errors.As(*APIError)` contract and a clean message | - |
| 9 | Not-found hint (review #13) | Added in the handler, not in `icu` | `icu` must not name MCP tools | - |
| 10 | Dot-segment ids (review #8) | Rejected in `requireActivityID`, so all activity tools | Single id validator; CWE-23 | An intervals.icu id format legitimately contains dots |
| 11 | Integer bounds (review #9) | int32 bounds, still no ranges | Spec declares `format: int32`; ranges stay out per Decision 2 | Spec adds min/max |
| 12 | Re-review after fix | Not launched; re-ran 15 targeted mutants + full gates | Sub-agent budget 5, 4 used | Max wants a fresh re-review |

## Budget

| Limit | Value |
|-------|-------|
| Fix rounds | 5, then escalate |
| Concurrent write agents | 0 (implemented directly: single interdependent scope) |
| Sub-agents for the run | 5 (review fleet) |
| Evidence to finalize | `make check` + staticcheck + `go vet -tags=integration` + `go build` green after last edit |

## Progress

- [x] 2026-10-01 Baseline: `make check` green; `make lint` fails (staticcheck not installed)
- [x] 2026-10-01 Requirements refined (Decisions 1-3)
- [x] 2026-10-01 Domain layer: tests red (build fail: UpdateActivity undefined) → green
- [x] 2026-10-01 Tool layer: tests red (HandleUpdateActivity undefined) → green
- [x] 2026-10-01 ADR-0005, README, CHANGELOG
- [x] 2026-10-01 VERIFY: tidy, vet, `go test -race`, build, integration vet, test-e2e green; staticcheck clean on new code, 1 pre-existing finding (main.go:54 ST1005)
- [x] 2026-10-01 REVIEW round 1: 4/4 reviewers; consolidated C0/M7/m11 → `quality_reports/reviews/2026-10-01_update-activity/001-findings.md`
- [x] 2026-10-01 FIX round 1: 7 Major + 10 Minor fixed, 1 Minor accepted; 15/15 targeted mutants killed; gates green after last edit
- [x] 2026-10-01 BLAST-RADIUS: 1 Minor (CHANGELOG missing read-tool id change), fixed
- [x] 2026-10-01 SCORE 97/100 (gate: pr); approval record `quality_reports/approvals/2026-10-01_update-activity.md`
- [x] 2026-10-01 UAT: Max approved ("go ahead, apply the fixes"); Strava truth accepted as unverified-live
- [x] 2026-10-01 Commit A (feature) on `feat/update-activity`, no push
- [x] 2026-10-01 Commit B (chore, separate): pre-existing staticcheck ST1005 `main.go:54`, gofmt drift in two test files, v0.1.0 tool count (13 → 14) in CHANGELOG/LEARNING
- [x] 2026-10-01 Plan closed, moved to `completed/`

## Surprises & Discoveries

- `Activity.type` has no enum in the spec; the activity-type enum lives on sibling schemas (see Spec facts).
- Pre-existing: `make lint` → staticcheck ST1005 at `main.go:54`; gofmt drift in `internal/icu/client_test.go`, `internal/icu/wellness_test.go`. Not touched (unrelated).
- mcp-go v0.48.0 does expose `MCPServer.GetTool`/`ListTools`; the comment on `TestRegisterActivities_InvokesAllHandlers` says otherwise (pre-existing, left as is).

## Outcomes & Retrospective

**Shipped:** `update_activity` (whitelist + JSON-type/int32 + enum checks in `icu.UpdateActivity`; shaped output; Strava/404 sentinels that keep `*APIError` in the chain), ADR-0005, README/CHANGELOG, stricter `activity_id` for all activity tools. Review: 1 round, 4/4 reviewers, C0/M7/m11, all fixed but 1 accepted Minor; SCORE 97 (gate pr).

**Gaps:** Strava detection unverified against a live Strava activity; no rpe/feel ranges (spec has none); activity-type enum is a hand copy of the spec (see `quality_reports/plans/tech-debt.md`).

**Lessons:**
- A resource schema can omit an enum that lives on sibling schemas (`Activity.type` vs `SportInfo.type`): search the whole spec for the enum before concluding there is none.
- mcp-go `NewTool` defaults `DestructiveHint` to true, so asserting it is tautological; assert the absence of `ReadOnlyHint` instead.
- Reviewer worktrees materialize at the default branch: ship uncommitted work to them as a patch file.
- The README tool count was already off by one before this change: count registrations (`grep -c 'mcp.NewTool('`) instead of incrementing the documented number.
