# ABOUTME: ADR-0004 - Tolerant JSON pass-through, no strict DTOs
# ABOUTME: intervals.icu schema evolves without notice; avoid typed DTO brittleness

# ADR-0004: Tolerant JSON pass-through, no strict DTOs

- **Status**: Accepted
- **Date**: 2026-04-21
- **Author**: Max
- **Amended by**: [ADR-0005](0005-whitelisted-activity-update.md) for `update_activity` (whitelisted input, shaped output)

## Context

intervals.icu backend adds/removes fields without versioning. A typed Go DTO
that decodes into `struct Activity { ... }` would break every time a field is
renamed or added, and silently ignore unknown fields (data loss for the LLM).

Reference: `rday/py-intervalsicu` added a `strict=False` flag for the same
reason.

## Decision

All `internal/icu/*.go` methods return `json.RawMessage`. MCP handlers forward
the raw JSON as a text result; the LLM reads the full payload.

Input mutating payloads (`create_event`, `update_event`, `update_wellness`) are
accepted as `map[string]any` and validated only for required top-level fields.
Everything else passes through.

## Rationale

- Zero decode cost.
- Future-proof against schema additions.
- The LLM handles JSON natively; a typed intermediate is noise.

## Consequences

- No compile-time type safety on intervals.icu fields.
- Manual refactor if we ever need Go code to operate on the data (rare for MCP
  servers, their job is pass-through).
