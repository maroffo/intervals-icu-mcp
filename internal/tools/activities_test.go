// ABOUTME: Tests for activities MCP tool handlers: argument parsing, success, error mapping.
// ABOUTME: Reuses the package-level newToolClient, callReq and resultText helpers.

package tools

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestHandleListActivities_PassesFilters(t *testing.T) {
	var gotQuery, gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"id":"a1"}]`))
	}))

	res, err := HandleListActivities(c)(context.Background(), callReq(map[string]any{
		"oldest": "2025-01-01",
		"newest": "2025-04-30",
		"limit":  float64(25),
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotPath != "/athlete/0/activities" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{"oldest=2025-01-01", "newest=2025-04-30", "limit=25"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	if !strings.Contains(resultText(t, res), `"a1"`) {
		t.Fatalf("result text missing payload: %q", resultText(t, res))
	}
}

func TestHandleListActivities_NoArgsOmitsQuery(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	res, err := HandleListActivities(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotQuery != "" {
		t.Fatalf("query should be empty, got %q", gotQuery)
	}
}

func TestHandleListActivities_APIErrorBubbles(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`bad range`))
	}))

	res, err := HandleListActivities(c)(context.Background(), callReq(map[string]any{"oldest": "x"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got %s", resultText(t, res))
	}
	if !strings.Contains(resultText(t, res), "list_activities") {
		t.Fatalf("error text = %q", resultText(t, res))
	}
}

func TestHandleGetActivity_Success(t *testing.T) {
	var gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"i1"}`))
	}))

	res, err := HandleGetActivity(c)(context.Background(), callReq(map[string]any{"activity_id": "i1"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotPath != "/activity/i1" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(resultText(t, res), `"i1"`) {
		t.Fatalf("result text = %q", resultText(t, res))
	}
}

func TestHandleGetActivity_MissingIDReturnsToolError(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	res, err := HandleGetActivity(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got %s", resultText(t, res))
	}
}

func TestHandleGetActivityStreams_WithTypes(t *testing.T) {
	var gotPath, gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"type":"watts"}]`))
	}))

	res, err := HandleGetActivityStreams(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i9",
		"types":       []any{"watts", "heartrate"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotPath != "/activity/i9/streams" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "types=watts%2Cheartrate" {
		t.Fatalf("query = %q", gotQuery)
	}
	if !strings.Contains(resultText(t, res), `"watts"`) {
		t.Fatalf("result text = %q", resultText(t, res))
	}
}

func TestHandleGetActivityStreams_NoTypes(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	res, err := HandleGetActivityStreams(c)(context.Background(), callReq(map[string]any{"activity_id": "i9"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotQuery != "" {
		t.Fatalf("query should be empty, got %q", gotQuery)
	}
}

func TestHandleGetActivityStreams_MissingIDReturnsToolError(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	res, err := HandleGetActivityStreams(c)(context.Background(), callReq(map[string]any{"types": []any{"watts"}}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got %s", resultText(t, res))
	}
}

func TestHandleGetActivityIntervals_Success(t *testing.T) {
	var gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"i7","intervals":[]}`))
	}))

	res, err := HandleGetActivityIntervals(c)(context.Background(), callReq(map[string]any{"activity_id": "i7"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotPath != "/activity/i7/intervals" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(resultText(t, res), `"i7"`) {
		t.Fatalf("result text = %q", resultText(t, res))
	}
}

func TestHandleGetActivityIntervals_MissingIDReturnsToolError(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	res, err := HandleGetActivityIntervals(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got %s", resultText(t, res))
	}
}

// TestRegisterActivities_InvokesAllHandlers exercises RegisterActivities for
// its side-effects (wiring the four tool handlers) and then invokes each
// handler factory directly to confirm they each return a *mcp.CallToolResult
// without panicking when given minimal input. The mcp-go SDK (v0.48.0) does
// not expose ListTools introspection on an MCPServer, so we can't assert on
// the number of registered tools; invoking the handlers is the closest
// behavioural check available.
func TestRegisterActivities_InvokesAllHandlers(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal JSON that keeps every handler happy on the success path.
		_, _ = w.Write([]byte(`[]`))
	}))

	// Side-effect: must not panic.
	s := server.NewMCPServer("test", "0.0.0")
	RegisterActivities(s, c)

	type handlerCase struct {
		name    string
		handler server.ToolHandlerFunc
		args    map[string]any
	}
	cases := []handlerCase{
		{"list_activities", HandleListActivities(c), nil},
		{"get_activity", HandleGetActivity(c), map[string]any{"activity_id": "i1"}},
		{"get_activity_streams", HandleGetActivityStreams(c), map[string]any{"activity_id": "i1"}},
		{"get_activity_intervals", HandleGetActivityIntervals(c), map[string]any{"activity_id": "i1"}},
	}
	for _, tc := range cases {
		res, err := tc.handler(context.Background(), callReq(tc.args))
		if err != nil {
			t.Errorf("%s: handler err: %v", tc.name, err)
			continue
		}
		if res == nil {
			t.Errorf("%s: nil *mcp.CallToolResult", tc.name)
			continue
		}
		// Confirm static type even though non-nil already implies this at
		// compile time; guards against future signature drift.
		var _ *mcp.CallToolResult = res
	}
}
