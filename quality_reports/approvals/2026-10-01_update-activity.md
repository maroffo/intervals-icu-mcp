# ABOUTME: Redacted approval record for the update_activity change (committed with the change)
# ABOUTME: Counts, CWE ids, score and residual risks only; findings detail stays in the local reviews dir

# Approval: update_activity

| Field | Value |
|-------|-------|
| Branch | `feat/update-activity` (base `41eb4c1`) |
| Commit | the commit that adds this file (feature commit; a separate chore commit fixes pre-existing lint/format issues) |
| Rounds run | 1 review round (4/4 reviewers: architecture, security, test, dx), 1 fix round |
| Counts (consolidated, before fix) | Critical 0 / Major 7 / Minor 11 |
| Counts (after fix) | Critical 0 / Major 0 / Minor 1 (accepted) |
| CWE ids | CWE-23, CWE-20 (both Minor, fixed) |
| Final score | SCORE: 97/100 (threshold: 90, gate: pr) |
| Findings path | `quality_reports/reviews/2026-10-01_update-activity/` (local, gitignored) |

## Residual risks

- Strava detection is a heuristic: the API spec does not document the error, and it was not verified against a live Strava activity.
- `icu_rpe` / `feel` have no range checks: the spec defines none (Decision 2 in the plan).
- The activity-type enum is a copy of the spec; a sport added upstream is rejected until the list is updated.
- No integration test for the PUT: the project's integration tests are read-only by rule.
