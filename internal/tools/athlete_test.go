// ABOUTME: Unit tests for athlete MCP tool handlers: exercises wiring over httptest icu.Client.
// ABOUTME: Verifies happy path text output and query-param propagation for get_fitness.

package tools

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestHandleGetAthlete_OK(t *testing.T) {
	var gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"0","name":"Max"}`))
	}))

	res, err := HandleGetAthlete(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("result reports error: %s", resultText(t, res))
	}
	if gotPath != "/athlete/0" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(resultText(t, res), `"name":"Max"`) {
		t.Errorf("text = %q", resultText(t, res))
	}
}

func TestHandleGetFitness_WithFilters(t *testing.T) {
	var gotPath, gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"ctl":50}]`))
	}))

	args := map[string]any{
		"oldest": "2024-01-01",
		"newest": "2024-01-31",
		"sport":  "Ride",
	}
	res, err := HandleGetFitness(c)(context.Background(), callReq(args))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("result reports error: %s", resultText(t, res))
	}
	if gotPath != "/athlete/0/fitness" {
		t.Errorf("path = %q", gotPath)
	}
	// Assert each param independently instead of string-matching the full query:
	// key order is an implementation detail of url.Values.Encode.
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query %q: %v", gotQuery, err)
	}
	for k, want := range map[string]string{
		"oldest": "2024-01-01",
		"newest": "2024-01-31",
		"sport":  "Ride",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("query %s = %q, want %q", k, got, want)
		}
	}
	if !strings.Contains(resultText(t, res), `"ctl":50`) {
		t.Errorf("text = %q", resultText(t, res))
	}
}

func TestHandleGetFitness_NoParams(t *testing.T) {
	var gotQuery string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	res, err := HandleGetFitness(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("result reports error: %s", resultText(t, res))
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
}

func TestHandleListFolders_OK(t *testing.T) {
	var gotPath string
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`[{"id":"f1","name":"Base"}]`))
	}))

	res, err := HandleListFolders(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("result reports error: %s", resultText(t, res))
	}
	if gotPath != "/athlete/0/folders" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(resultText(t, res), `"name":"Base"`) {
		t.Errorf("text = %q", resultText(t, res))
	}
}

func TestHandleGetAthlete_APIErrorReturnsToolError(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`nope`))
	}))

	res, err := HandleGetAthlete(c)(context.Background(), callReq(nil))
	if err != nil {
		t.Fatalf("handler should not return error; got %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true, got text=%q", resultText(t, res))
	}
}
