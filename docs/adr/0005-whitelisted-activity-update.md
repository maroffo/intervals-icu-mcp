# ABOUTME: ADR-0005 - update_activity whitelists its input and shapes its output
# ABOUTME: Scoped exception to ADR-0004 pass-through for the one activity write tool

# ADR-0005: Whitelisted input and shaped output for `update_activity`

- **Status**: Accepted
- **Date**: 2026-10-01
- **Author**: Max

## Context

ADR-0004 makes every tool pass raw JSON through: mutating payloads are
`map[string]any` checked only for required fields, and responses are
forwarded as-is.

`PUT /api/v1/activity/{id}` takes the full `Activity` schema (184 fields in
the spec, many computed: power curves, load, zones). An LLM that echoes back a
fetched activity, or invents a field, could overwrite computed data on a
recorded workout, which is harder to repair than a calendar event or a
wellness entry. The response is also the full activity document, far more
than the caller needs to confirm an edit.

The spec states "Strava activities cannot be updated" but does not document
the error shape.

## Decision

For `update_activity` only:

- **Input whitelist** enforced in `icu.UpdateActivity` (the lowest layer):
  `type, name, description, icu_rpe, feel, commute, trainer, icu_ignore_hr,
  icu_ignore_power, icu_ignore_time`. Any other key is rejected before the
  HTTP call. Values are checked against the JSON types in the spec, and `type`
  against the spec's activity-type enum (taken from `SportInfo.type`, since
  `Activity.type` is a bare string).
- **Shaped output**: `{id, updated, icu_training_load, icu_ctl, icu_atl}`,
  where `updated` carries the server's values for the keys sent and the load
  metrics appear only when present and non-null.
- **Strava detection** is a heuristic: a 4xx whose body mentions "strava", or a
  2xx returning a `source=STRAVA` stub, maps to `icu.ErrStravaActivity`.

## Rationale

- A whitelist in the client makes "never send the full activity back" a
  property of the code, not of the prompt.
- The shaped result tells the caller what changed and what it did to training
  load, without a multi-kilobyte document.

## Consequences

- New editable fields need a code change (whitelist + test), by design.
- The activity-type enum is a copy of the spec; a new sport on intervals.icu
  is rejected locally until the list is updated.
- Strava detection is unverified against a live Strava-sourced activity; if
  the real error does not mention Strava, callers see the plain HTTP error
  (or "not found" on a 404) instead of the Strava message.
- ADR-0004 still governs every other tool.
