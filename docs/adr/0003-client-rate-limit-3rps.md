# ABOUTME: ADR-0003 - Client-side rate limit at 3 req/s
# ABOUTME: Conservative default, revisit based on observed 429 rate

# ADR-0003: Client rate limit at 3 req/s

- **Status**: Accepted
- **Date**: 2026-04-21
- **Author**: Max

## Context

intervals.icu does not publish an official API rate limit. Forum anecdotes
suggest ~10 rps triggers 429s under burst conditions.

## Decision

Client-side `golang.org/x/time/rate.Limiter` at 3 rps with burst 3, plus
retry-on-429 with exponential backoff (respecting `Retry-After`, capped at 60s).

## Rationale

- Conservative: well below observed break-point.
- Bursts of 3 allow interactive "list + get detail" flows without blocking.
- Cap on `Retry-After` protects against upstream misbehaviour forcing hour-long
  stalls.

## Consequences

- Bulk operations feel slow (3 rps = ~20 min for 3000 activities).
- Revisit if a legitimate bulk use case needs higher throughput; make the rate
  configurable via env var if/when.
