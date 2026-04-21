// ABOUTME: Tests for wellness MCP tool handlers: required args, success payloads, PUT body round-trip.
// ABOUTME: Relies on package-level shared helpers (callReq, newToolClient, textContent).

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// --- get_wellness ---

func TestHandleGetWellness_Success(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/0/wellness/2026-04-21" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"2026-04-21","weight":72.5}`))
	}))

	res, err := HandleGetWellness(c)(context.Background(), callReq(map[string]any{
		"date": "2026-04-21",
	}))
	if err != nil {
		t.Fatalf("handler returned err: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError=true: %q", textContent(res))
	}
	body := textContent(res)
	if !strings.Contains(body, `"weight":72.5`) {
		t.Errorf("body = %q", body)
	}
}

func TestHandleGetWellness_MissingDate(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))

	res, err := HandleGetWellness(c)(context.Background(), callReq(map[string]any{}))
	if err != nil {
		t.Fatalf("handler returned err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got content=%q", textContent(res))
	}
}

func TestHandleGetWellness_InvalidDatePropagates(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))

	res, err := HandleGetWellness(c)(context.Background(), callReq(map[string]any{
		"date": "nope",
	}))
	if err != nil {
		t.Fatalf("handler returned err: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true on invalid date")
	}
}

// --- list_wellness ---

func TestHandleListWellness_WithRange(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"id":"2026-04-20"}]`))
	}))

	res, err := HandleListWellness(c)(context.Background(), callReq(map[string]any{
		"oldest": "2026-04-01",
		"newest": "2026-04-21",
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError: %q", textContent(res))
	}
	if !strings.Contains(gotQuery, "oldest=2026-04-01") || !strings.Contains(gotQuery, "newest=2026-04-21") {
		t.Errorf("query = %q", gotQuery)
	}
	if !strings.Contains(textContent(res), `"2026-04-20"`) {
		t.Errorf("body = %q", textContent(res))
	}
}

func TestHandleListWellness_NoArgs(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	res, err := HandleListWellness(c)(context.Background(), callReq(map[string]any{}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError: %q", textContent(res))
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
}

// --- update_wellness ---

func TestHandleUpdateWellness_PutBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))

	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"date": "2026-04-21",
		"fields": map[string]any{
			"weight":  72.9,
			"fatigue": 2,
			"mood":    4,
		},
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError: %q", textContent(res))
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/athlete/0/wellness/2026-04-21" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["weight"].(float64) != 72.9 {
		t.Errorf("body.weight = %v", gotBody["weight"])
	}
	if gotBody["fatigue"].(float64) != 2 {
		t.Errorf("body.fatigue = %v", gotBody["fatigue"])
	}
	if !strings.Contains(textContent(res), `"updated":true`) {
		t.Errorf("response body = %q", textContent(res))
	}
}

func TestHandleUpdateWellness_MissingDate(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))
	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"fields": map[string]any{"weight": 70},
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true on missing date")
	}
}

func TestHandleUpdateWellness_MissingFields(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))
	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"date": "2026-04-21",
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true on missing fields")
	}
}

func TestHandleUpdateWellness_EmptyFields(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))
	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"date":   "2026-04-21",
		"fields": map[string]any{},
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true on empty fields")
	}
}

func TestHandleUpdateWellness_FieldsAsJSONString(t *testing.T) {
	// Some MCP transports deliver tool args re-marshaled, so `fields` may arrive
	// as a JSON-encoded string rather than a native map. Handler must accept both.
	var gotBody map[string]any
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))

	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"date":   "2026-04-21",
		"fields": `{"weight": 71.2, "sleepSecs": 26000}`,
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError: %q", textContent(res))
	}
	if gotBody["weight"].(float64) != 71.2 {
		t.Errorf("body.weight = %v", gotBody["weight"])
	}
}

func TestHandleUpdateWellness_FieldsWrongType(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))
	res, err := HandleUpdateWellness(c)(context.Background(), callReq(map[string]any{
		"date":   "2026-04-21",
		"fields": 42, // number is not a valid object
	}))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true on wrong fields type")
	}
}
