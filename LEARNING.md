# ABOUTME: Retrospective log for intervals-icu-mcp - decisions, lessons, pitfalls
# ABOUTME: Grows over time; latest entries on top

# LEARNING.md

## Project Overview

Local MCP server in Go exposing intervals.icu REST API to MCP clients over stdio. Built from scratch in a single session with parallel software-engineer agents. v0.1.0 shipped 14 tools across 4 domains (activities, wellness, events, athlete); v0.2.0 adds `update_activity` (15 tools).

## Lessons Learned

### 2026-10-01: A resource schema can hide its enum on sibling schemas

**Context:** `update_activity` had to validate `type` against the intervals.icu enum "from the OpenAPI spec".

**Problem:** `Activity.type` is a bare `string` in the spec. The 60-value activity-type enum lives on sibling schemas (`SportInfo.type`, `SportSettings.types`, `Folder.activity_types`).

**Takeaway:** Before concluding a spec has no enum, search the whole document for one (`jq` over every `enum` containing a known value). Check it is the only candidate before copying it.

### 2026-10-01: The spec documents the rule, not the error

**Context:** The spec says "Strava activities cannot be updated" but documents no error status or body.

**Takeaway:** When the error shape is undocumented and no live account is at hand, ship a narrow heuristic (4xx + body mention), say "unverified" in CHANGELOG/README/ADR, and log the live check as tech debt. Don't present a guess as a guarantee in user-facing docs.

### 2026-10-01: SDK defaults can make assertions tautological

**Context:** A schema test asserted `DestructiveHint == true` on `update_activity`.

**Problem:** mcp-go v0.48.0 `NewTool` already defaults `DestructiveHint` to true, so the assertion could never fail (a hand mutant removing the option survived).

**Takeaway:** Assert what the default does not give you: for a mutating tool, that `ReadOnlyHint` is not set.

### 2026-10-01: Reviewer worktrees start from the default branch

**Context:** Worktree-isolated reviewers were launched on uncommitted work.

**Takeaway:** The worktree materializes at the default branch, not at the working tree. Ship uncommitted work to reviewers as a patch file and have them `git apply` it in their own copy.

### 2026-10-01: Documented counts drift; count from code

**Context:** README and CHANGELOG said v0.1.0 had 13 tools; the code registered 14. Adding one tool carried the error forward until a reviewer counted.

**Takeaway:** Derive documented counts from the code (`grep -c 'mcp.NewTool(' internal/tools/*.go`), never by incrementing the previous number.

### 2026-04-21: Parallel agents + go.mod = coordination blind spot

**Context:** Delegated Fase 2 to 4 software-engineer agents in parallel, each scoped to its own domain files (`activities`, `wellness`, `events`, `athlete`), told explicitly "don't touch go.mod".

**Problem:** All 4 tried to import `mark3labs/mcp-go` simultaneously. The package was listed as `// indirect` in go.mod but no transitive deps were in go.sum. Three agents produced code that couldn't compile; one agent unilaterally ran `go mod download all` as a workaround, mutating go.mod.

**Solution:** Pre-install every external dependency before fan-out. Also surface shared-integration-surfaces (go.mod, go.sum, DI config, barrel exports) explicitly in agent briefs as "leader-owned, don't touch, flag if blocked".

**Takeaway:** When orchestrating parallel write-agents, the "excluded" files are often exactly the files that matter most. Pre-flight the module graph and run one dry build at the leader level before fan-out.

### 2026-04-21: Shared test helpers drift between parallel agents

**Context:** Each of 4 parallel agents wrote its own `newToolClient`/`callReq`/`resultText` helpers in its test file, picking different names (`newEventsBackend`, `newWellnessFakeClient`) and different defaults.

**Problem:** `newEventsBackend` kept production retry settings (3 retries, 500ms base), which would block ~3.5s per flaky test in CI - a latent flake magnet. Only a review round caught it.

**Takeaway:** Inject a `testhelpers_test.go` in the initial scaffolding before fan-out. Parallel agents should consume shared helpers, not invent them.

### 2026-04-21: Dead parameters survive generous review

**Context:** `Client.Do(ctx, method, path, query, body, out any) ([]byte, error)` had `out any` as an unused tail parameter. It passed code review, type checks, and 70 unit tests before an architecture reviewer flagged it.

**Takeaway:** Explicitly ask reviewers to check for dead parameters. If a refactor would force all downstream agents to use the same dead argument correctly, simplify the signature as an orchestrator pre-step.

### 2026-04-21: Agent severity scales are inflated vs rubric

**Context:** Review agents reported 5 "CRITICAL" findings (missing coverage, doc contradictions). Per `quality-gates.md`, CRITICAL = auto-fail = score 0 and is reserved for tests-failing / build-broken / security vuln / data-loss / unplanned-stubs. None of the agent findings matched.

**Takeaway:** Always reconcile reviewer severity to the project rubric before scoring. Treat agent outputs as raw data.

### 2026-04-21: CLAUDE.md discipline wins over DX reviewer nudges

**Context:** DX reviewer suggested removing `ABOUTME` headers from test files as noise. Max's global CLAUDE.md says "All files: 2-line ABOUTME header" with no test-file exception.

**Takeaway:** When a reviewer finding conflicts with an explicit Max rule, the rule wins unless Max changes it. Flag the tension so Max can revisit, don't silently override.

## Best Practices Discovered

- **Ask via AskUserQuestion, always.** Max's preference, now a saved feedback memory. Prose-with-questions-at-the-end gets ignored.
- **Pre-commit hook as scaffolding forcing function.** The `make test-e2e` requirement surfaced mid-commit and forced a proper target (aliased to integration) rather than a workaround.
- **Plan file in vault + repo.** Vault copy for cross-project discoverability, repo copy for offline + PR context.
