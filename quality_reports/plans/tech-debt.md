# ABOUTME: Tech debt discovered during planned work but not addressed there
# ABOUTME: One line per item, each pointing back to the plan that found it

- Verify the Strava error shape for `PUT /activity/{id}` against a live Strava-sourced activity and pin the test fixture to it (completed/2026-10-01_update-activity.md, Decision 1).
- Activity-type enum in `internal/icu/activities.go` is a hand copy of the OpenAPI spec with no drift check (completed/2026-10-01_update-activity.md, ADR-0005 Consequences).
- `make lint` (staticcheck) is not run in CI (completed/2026-10-01_update-activity.md, Surprises).
