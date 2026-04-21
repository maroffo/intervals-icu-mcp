// ABOUTME: Shared test helpers for MCP tool handler tests across all domains.
// ABOUTME: Consolidates fake-client factory, request builder, and result text extraction.

package tools

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/maroffo/intervals-icu-mcp/internal/icu"
)

// newToolClient spins up an httptest.Server and returns an icu.Client pointed at it.
// Retries disabled, rate limiter effectively off: tests are fast and deterministic.
func newToolClient(t *testing.T, handler http.Handler) *icu.Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return icu.New(ts.URL, "test-key", "0",
		icu.WithMaxRetries(0),
		icu.WithRateLimit(1000, 1000),
		icu.WithRetryBase(time.Millisecond),
	)
}

// callReq builds a CallToolRequest with the given arguments map.
func callReq(args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Arguments = args
	return req
}

// resultText extracts concatenated text from a tool result (test helper).
// Fails the test on nil result so callers can assume non-nil.
func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil {
		t.Fatal("nil result")
	}
	var out string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			out += tc.Text
		}
	}
	return out
}

// textContent is a test-only helper that mirrors resultText but returns ""
// on nil without failing. Useful for older test styles that pass the result
// directly into strings.Contains without preflighting nil.
func textContent(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var out string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			out += tc.Text
		}
	}
	return out
}
