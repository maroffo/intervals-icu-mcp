// ABOUTME: Tests for MCP event tool handlers: success, required-field validation, body propagation.
// ABOUTME: Uses httptest as the fake intervals.icu backend and calls handlers directly.

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHandleListEvents_Success(t *testing.T) {
	var gotPath, gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"id":"EVT1","name":"long run"}]`))
	}))

	h := HandleListEvents(c)
	res, err := h(context.Background(), callReq(map[string]any{
		"oldest":   "2026-01-01",
		"newest":   "2026-01-31",
		"category": "RACE_A",
		"resolve":  true,
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", resultText(t, res))
	}
	// newToolClient uses athleteID="0".
	if gotPath != "/athlete/0/events" {
		t.Errorf("path = %q", gotPath)
	}
	for _, want := range []string{"oldest=2026-01-01", "category=RACE_A", "resolve=true"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query = %q missing %q", gotQuery, want)
		}
	}
	if !strings.Contains(resultText(t, res), "EVT1") {
		t.Errorf("text = %q", resultText(t, res))
	}
}

func TestHandleListEvents_NoArgs(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	h := HandleListEvents(c)
	res, err := h(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", resultText(t, res))
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
}

func TestHandleCreateEvent_PostsBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"EVT77"}`))
	}))

	event := map[string]any{
		"start_date_local": "2026-04-01T06:30:00",
		"category":         "WORKOUT",
		"name":             "Intervals 5x1k",
		"type":             "Run",
	}
	h := HandleCreateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event": event}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", resultText(t, res))
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/athlete/0/events" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["name"] != "Intervals 5x1k" || gotBody["category"] != "WORKOUT" {
		t.Errorf("body = %v", gotBody)
	}
	if !strings.Contains(resultText(t, res), "EVT77") {
		t.Errorf("text = %q", resultText(t, res))
	}
}

func TestHandleCreateEvent_MissingEvent(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleCreateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
	if !strings.Contains(resultText(t, res), `"event"`) {
		t.Errorf("text = %q, want it to mention the missing event argument", resultText(t, res))
	}
}

func TestHandleCreateEvent_EventWrongType(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleCreateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event": "not-an-object"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
	// extractObjectArg reports either "JSON object" (for scalars other than
	// string) or a JSON decode failure (for strings that aren't JSON); either
	// is acceptable, what matters is that we rejected the input.
	txt := resultText(t, res)
	if !strings.Contains(txt, `"event"`) {
		t.Errorf("text = %q, want it to reference the event argument", txt)
	}
}

func TestHandleCreateEvent_MissingRequiredField(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleCreateEvent(c)
	// Missing category.
	res, err := h(context.Background(), callReq(map[string]any{
		"event": map[string]any{"start_date_local": "2026-04-01T06:30:00"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(t, res), "category") {
		t.Fatalf("want error about category, got %q (isErr=%v)", resultText(t, res), res.IsError)
	}
}

func TestHandleUpdateEvent_Success(t *testing.T) {
	var gotMethod, gotEscapedPath string
	var gotBody map[string]any
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotEscapedPath = r.URL.EscapedPath()
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"EVT1","name":"renamed"}`))
	}))

	h := HandleUpdateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{
		"event_id": "EVT1",
		"event":    map[string]any{"name": "renamed"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", resultText(t, res))
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q", gotMethod)
	}
	if gotEscapedPath != "/athlete/0/events/EVT1" {
		t.Errorf("path = %q", gotEscapedPath)
	}
	if gotBody["name"] != "renamed" {
		t.Errorf("body = %v", gotBody)
	}
}

func TestHandleUpdateEvent_MissingEventID(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleUpdateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event": map[string]any{"name": "x"}}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(t, res), "event_id") {
		t.Fatalf("want error about event_id, got %q (isErr=%v)", resultText(t, res), res.IsError)
	}
}

func TestHandleUpdateEvent_MissingEventObject(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleUpdateEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event_id": "EVT1"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(t, res), `"event"`) {
		t.Fatalf("want error about event, got %q (isErr=%v)", resultText(t, res), res.IsError)
	}
}

func TestHandleDeleteEvent_Success(t *testing.T) {
	var gotMethod, gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	h := HandleDeleteEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event_id": "EVT9"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", resultText(t, res))
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/athlete/0/events/EVT9" {
		t.Errorf("path = %q", gotPath)
	}
	txt := resultText(t, res)
	if !strings.Contains(strings.ToLower(txt), "deleted") || !strings.Contains(txt, "EVT9") {
		t.Errorf("text = %q, want 'deleted EVT9'", txt)
	}
}

func TestHandleDeleteEvent_MissingEventID(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not hit server")
	}))
	h := HandleDeleteEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(t, res), "event_id") {
		t.Fatalf("want error about event_id, got %q (isErr=%v)", resultText(t, res), res.IsError)
	}
}

func TestHandleDeleteEvent_BackendError(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"gone"}`))
	}))
	h := HandleDeleteEvent(c)
	res, err := h(context.Background(), callReq(map[string]any{"event_id": "GHOST"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result, got: %s", resultText(t, res))
	}
}
