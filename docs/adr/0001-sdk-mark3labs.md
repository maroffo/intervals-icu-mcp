# ABOUTME: ADR-0001 - Choice of Go MCP SDK
# ABOUTME: Rationale for mark3labs/mcp-go over the official modelcontextprotocol/go-sdk

# ADR-0001: Use mark3labs/mcp-go as the Go MCP SDK

- **Status**: Accepted
- **Date**: 2026-04-21
- **Author**: Max

## Context

Two Go SDKs are available to implement an MCP server:
1. `github.com/mark3labs/mcp-go` (community, v0.48.0, 8.6k stars, very active).
2. `github.com/modelcontextprotocol/go-sdk` (official, newer, API still evolving).

## Decision

Use `mark3labs/mcp-go`.

## Rationale

- Mature, many examples, stable-enough API (pre-1.0 but low churn).
- stdio / SSE / streamable-HTTP transports all supported.
- Community active (weekly commits).
- Official SDK still settling its shape; migration cost later will be lower if
  and when it clearly wins (the protocol boundary is small).

## Consequences

- We depend on a community SDK; if abandoned, migration to the official SDK is a
  bounded refactor (~1 day) because our usage is limited to `server.NewMCPServer`,
  `s.AddTool`, `mcp.NewTool`, `mcp.NewToolResultText/Error*`.
- Revisit if the official SDK hits v1.0 and mark3labs lags on spec updates.
