# ABOUTME: ADR-0002 - Transport: stdio only
# ABOUTME: Rationale for not supporting SSE/HTTP transport in v0.1

# ADR-0002: stdio only transport

- **Status**: Accepted
- **Date**: 2026-04-21
- **Author**: Max

## Context

MCP supports stdio, SSE, and streamable HTTP transports. Our use case is a
local server spawned by Claude Code / Claude Desktop.

## Decision

stdio only, no HTTP/SSE exposure.

## Rationale

- Local-only use: no authentication/authorization layer needed.
- Eliminates a whole class of network security concerns (TLS, CORS, listener).
- Simpler operational model: no port management, no reverse proxy.

## Consequences

- Remote use requires a different deployment (tunneling or a rewrite with a
  proper auth layer). Accept this trade-off for v0.1.
- Revisit if a concrete need for remote access emerges.
