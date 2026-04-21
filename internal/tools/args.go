// ABOUTME: Shared argument extraction helpers for MCP tool handlers.
// ABOUTME: Centralizes the "extract JSON object from CallToolRequest args" pattern.

package tools

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// extractObjectArg reads argName from the request arguments and returns it as
// map[string]any. It accepts map[string]any, json.RawMessage, []byte, or a
// JSON-encoded string (some MCP transports re-marshal nested objects).
// Returns an error if missing, null, or not a JSON object.
func extractObjectArg(req mcp.CallToolRequest, argName string) (map[string]any, error) {
	args := req.GetArguments()
	raw, ok := args[argName]
	if !ok || raw == nil {
		return nil, fmt.Errorf("missing required argument %q", argName)
	}
	switch v := raw.(type) {
	case map[string]any:
		return v, nil
	case json.RawMessage:
		var out map[string]any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, fmt.Errorf("argument %q: decode: %w", argName, err)
		}
		return out, nil
	case []byte:
		var out map[string]any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, fmt.Errorf("argument %q: decode: %w", argName, err)
		}
		return out, nil
	case string:
		var out map[string]any
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return nil, fmt.Errorf("argument %q: decode string: %w", argName, err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("argument %q: expected JSON object, got %T", argName, raw)
	}
}
