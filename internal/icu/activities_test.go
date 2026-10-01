// ABOUTME: Tests for activities domain API: paths, query params, input validation, raw decode.
// ABOUTME: Uses httptest.Server and an in-package client with retries disabled.

package icu

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newActivitiesClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)
	c := New(ts.URL, "k", "42",
		WithMaxRetries(0),
		WithRetryBase(time.Millisecond),
		WithRateLimit(1000, 1000),
	)
	return c, ts
}

func TestListActivities_PathAndAllQueryParams(t *testing.T) {
	var gotPath, gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"id":"abc"}]`))
	}))

	raw, err := c.ListActivities(context.Background(), ListActivitiesParams{
		Oldest: "2025-01-01",
		Newest: "2025-04-30",
		Limit:  42,
	})
	if err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if gotPath != "/athlete/42/activities" {
		t.Fatalf("path = %q", gotPath)
	}
	// query param order is deterministic (alphabetical) with url.Values.Encode.
	for _, want := range []string{"oldest=2025-01-01", "newest=2025-04-30", "limit=42"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	var decoded []map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	if len(decoded) != 1 || decoded[0]["id"] != "abc" {
		t.Fatalf("unexpected decoded payload: %v", decoded)
	}
}

func TestListActivities_OmitsEmptyFilters(t *testing.T) {
	var gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	if _, err := c.ListActivities(context.Background(), ListActivitiesParams{}); err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("query should be empty when no filters set, got %q", gotQuery)
	}
}

func TestListActivities_LimitZeroOmitted(t *testing.T) {
	var gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	if _, err := c.ListActivities(context.Background(), ListActivitiesParams{Oldest: "2025-01-01", Limit: 0}); err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if strings.Contains(gotQuery, "limit=") {
		t.Fatalf("limit=0 should not be sent, query=%q", gotQuery)
	}
	if !strings.Contains(gotQuery, "oldest=2025-01-01") {
		t.Fatalf("oldest should still be present, query=%q", gotQuery)
	}
}

func TestListActivities_TrimsWhitespaceFilters(t *testing.T) {
	var gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	if _, err := c.ListActivities(context.Background(), ListActivitiesParams{Oldest: "   ", Newest: "\t"}); err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("whitespace-only filters should be omitted, query=%q", gotQuery)
	}
}

func TestGetActivity_PathEscapeAndDecode(t *testing.T) {
	var gotPath, gotEscaped string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotEscaped = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"id":"i 123","type":"Ride"}`))
	}))

	raw, err := c.GetActivity(context.Background(), "i 123")
	if err != nil {
		t.Fatalf("GetActivity: %v", err)
	}
	// Decoded path has the literal space; escaped path proves the client sent %20.
	if gotPath != "/activity/i 123" {
		t.Fatalf("decoded path = %q", gotPath)
	}
	if gotEscaped != "/activity/i%20123" {
		t.Fatalf("escaped path = %q, want /activity/i%%20123", gotEscaped)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	if decoded["id"] != "i 123" {
		t.Fatalf("decoded id = %v", decoded["id"])
	}
}

func TestGetActivity_EmptyIDRejected(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called, got %s %s", r.Method, r.URL.Path)
	}))
	for _, id := range []string{"", "   ", "\t"} {
		if _, err := c.GetActivity(context.Background(), id); err == nil {
			t.Fatalf("expected error for id=%q", id)
		}
	}
}

func TestGetActivityStreams_WithTypes(t *testing.T) {
	var gotPath, gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"type":"watts","data":[1,2,3]}]`))
	}))

	types := []string{"watts", "heartrate", "cadence"}
	raw, err := c.GetActivityStreams(context.Background(), "99", types)
	if err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if gotPath != "/activity/99/streams" {
		t.Fatalf("path = %q", gotPath)
	}
	// url.Values.Encode escapes commas to %2C.
	if gotQuery != "types=watts%2Cheartrate%2Ccadence" {
		t.Fatalf("query = %q", gotQuery)
	}
	if !strings.Contains(string(raw), `"watts"`) {
		t.Fatalf("raw body missing expected stream, got %s", raw)
	}
}

func TestGetActivityStreams_WithoutTypes(t *testing.T) {
	var gotQuery string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))

	if _, err := c.GetActivityStreams(context.Background(), "99", nil); err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("query should be empty when types is nil, got %q", gotQuery)
	}

	// Empty slice should also omit the param.
	if _, err := c.GetActivityStreams(context.Background(), "99", []string{}); err != nil {
		t.Fatalf("GetActivityStreams (empty slice): %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("query should be empty when types is [], got %q", gotQuery)
	}
}

func TestGetActivityStreams_EmptyIDRejected(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	if _, err := c.GetActivityStreams(context.Background(), "  ", []string{"watts"}); err == nil {
		t.Fatal("expected error for blank activity id")
	}
}

func TestGetActivityIntervals_Path(t *testing.T) {
	var gotPath string
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"7","intervals":[]}`))
	}))

	raw, err := c.GetActivityIntervals(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetActivityIntervals: %v", err)
	}
	if gotPath != "/activity/7/intervals" {
		t.Fatalf("path = %q", gotPath)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded["id"] != "7" {
		t.Fatalf("unexpected decoded payload: %v", decoded)
	}
}

func TestGetActivityIntervals_EmptyIDRejected(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	if _, err := c.GetActivityIntervals(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty activity id")
	}
}

func TestUpdateActivity_PutsOnlyProvidedFields(t *testing.T) {
	var gotMethod, gotEscaped, gotCT string
	var gotBody map[string]any
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotEscaped = r.URL.EscapedPath()
		gotCT = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"i190900884","name":"Tempo","icu_rpe":7,"source":"GARMIN_CONNECT"}`))
	}))

	raw, err := c.UpdateActivity(context.Background(), " i190900884 ", map[string]any{
		"name":    "Tempo",
		"icu_rpe": float64(7),
	})
	if err != nil {
		t.Fatalf("UpdateActivity: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q", gotMethod)
	}
	if gotEscaped != "/activity/i190900884" {
		t.Errorf("escaped path = %q", gotEscaped)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}
	// Partial update: exactly the provided keys, nothing else.
	if len(gotBody) != 2 || gotBody["name"] != "Tempo" || gotBody["icu_rpe"] != float64(7) {
		t.Errorf("body = %v, want exactly {name, icu_rpe}", gotBody)
	}
	if !strings.Contains(string(raw), `"icu_rpe":7`) {
		t.Errorf("raw = %s", raw)
	}
}

func TestUpdateActivity_AcceptsEveryWhitelistedField(t *testing.T) {
	var gotBody map[string]any
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"i1"}`))
	}))

	fields := map[string]any{
		"type":             "VirtualRide",
		"name":             "Zwift",
		"description":      "",
		"feel":             2, // Go int, as an in-process caller would pass it
		"icu_rpe":          int64(6),
		"commute":          false,
		"trainer":          true,
		"icu_ignore_hr":    true,
		"icu_ignore_power": false,
		"icu_ignore_time":  false,
	}
	if _, err := c.UpdateActivity(context.Background(), "i1", fields); err != nil {
		t.Fatalf("UpdateActivity: %v", err)
	}
	if len(gotBody) != len(fields) {
		t.Errorf("body has %d keys, want %d: %v", len(gotBody), len(fields), gotBody)
	}
}

func TestUpdateActivity_RejectsNonWhitelistedFields(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	_, err := c.UpdateActivity(context.Background(), "i1", map[string]any{
		"name":        "ok",
		"moving_time": float64(3600),
		"icu_ftp":     float64(300),
	})
	if err == nil {
		t.Fatal("expected error for non-whitelisted fields")
	}
	msg := err.Error()
	// Every offending field is named (sorted), and the allowed set is listed.
	if !strings.Contains(msg, `"icu_ftp", "moving_time"`) {
		t.Errorf("err = %q, want both offending fields in sorted order", msg)
	}
	allowed := "commute, description, feel, icu_ignore_hr, icu_ignore_power, icu_ignore_time, icu_rpe, name, trainer, type"
	if !strings.Contains(msg, "only these fields can be updated: "+allowed) {
		t.Errorf("err = %q, want the sorted allowed fields", msg)
	}
}

func TestUpdateActivity_RejectsUnknownActivityType(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	for _, bad := range []string{"Bike", "ride", ""} {
		_, err := c.UpdateActivity(context.Background(), "i1", map[string]any{"type": bad})
		if err == nil || !strings.Contains(err.Error(), "unsupported activity type") {
			t.Errorf("type %q: err = %v, want unsupported activity type", bad, err)
		}
	}
}

func TestUpdateActivity_RejectsWrongJSONTypes(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	cases := []struct {
		field string
		value any
		want  string
	}{
		{"icu_rpe", "7", `expected integer (int32), got string "7"`},
		{"icu_rpe", 7.5, "expected integer (int32), got number 7.5"},
		{"icu_rpe", 1e300, "expected integer (int32), got number 1e+300"},
		{"icu_rpe", float64(2147483648), "expected integer (int32)"},
		{"feel", int64(-2147483649), "expected integer (int32)"},
		{"feel", true, "expected integer (int32), got boolean true"},
		{"commute", "yes", `expected boolean, got string "yes"`},
		{"trainer", float64(1), "expected boolean, got number 1"},
		{"name", map[string]any{"x": 1}, "expected string, got object"},
		{"description", nil, "expected string, got null"},
		{"type", []any{"Ride"}, "expected string, got array"},
	}
	for _, tc := range cases {
		_, err := c.UpdateActivity(context.Background(), "i1", map[string]any{tc.field: tc.value})
		if err == nil || !strings.Contains(err.Error(), tc.field) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s=%#v: err = %v, want %q", tc.field, tc.value, err, tc.want)
		}
	}
}

func TestUpdateActivity_RejectsEmptyInput(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	if _, err := c.UpdateActivity(context.Background(), "  ", map[string]any{"name": "x"}); err == nil {
		t.Error("expected error for empty activity id")
	}
	if _, err := c.UpdateActivity(context.Background(), "i1", nil); err == nil {
		t.Error("expected error for nil fields")
	}
	if _, err := c.UpdateActivity(context.Background(), "i1", map[string]any{}); err == nil {
		t.Error("expected error for empty fields")
	}
}

func TestUpdateActivity_StravaRejectionMapsToSentinel(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Undocumented upstream shape; the detection keys only on 4xx + "strava".
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"status":422,"error":"Strava activities cannot be updated"}`))
	}))

	_, err := c.UpdateActivity(context.Background(), "12345678", map[string]any{"name": "x"})
	if !errors.Is(err, ErrStravaActivity) {
		t.Fatalf("err = %v, want ErrStravaActivity", err)
	}
	if strings.Contains(err.Error(), "422") || strings.Contains(err.Error(), "PUT") {
		t.Errorf("err = %q, should not leak the raw HTTP error", err)
	}
	if !strings.Contains(err.Error(), "12345678") {
		t.Errorf("err = %q, want the activity id", err)
	}
}

func TestUpdateActivity_StravaStubResponseMapsToSentinel(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"12345678","source":"STRAVA","_note":"STRAVA activities are not available via the API"}`))
	}))

	_, err := c.UpdateActivity(context.Background(), "12345678", map[string]any{"name": "x"})
	if !errors.Is(err, ErrStravaActivity) {
		t.Fatalf("err = %v, want ErrStravaActivity", err)
	}
}

func TestUpdateActivity_NotFoundMapsToSentinel(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"error":"Not Found"}`))
	}))

	_, err := c.UpdateActivity(context.Background(), "i404", map[string]any{"name": "x"})
	if !errors.Is(err, ErrActivityNotFound) {
		t.Fatalf("err = %v, want ErrActivityNotFound", err)
	}
	if !strings.Contains(err.Error(), "i404") {
		t.Errorf("err = %q, want the activity id", err)
	}
}

func TestUpdateActivity_OtherAPIErrorsPassThrough(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"Access denied"}`))
	}))

	_, err := c.UpdateActivity(context.Background(), "i1", map[string]any{"name": "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("err = %v, want *APIError 403", err)
	}
	if errors.Is(err, ErrStravaActivity) || errors.Is(err, ErrActivityNotFound) {
		t.Fatalf("err = %v, should not map to a sentinel", err)
	}
}

func TestAllActivityTypes_MatchesSpecEnum(t *testing.T) {
	types := AllActivityTypes()
	// 60 values in the intervals.icu OpenAPI activity-type enum (SportInfo.type).
	if len(types) != 60 {
		t.Fatalf("len = %d, want 60", len(types))
	}
	seen := map[string]bool{}
	for _, ty := range types {
		if seen[ty] {
			t.Errorf("duplicate type %q", ty)
		}
		seen[ty] = true
	}
	for _, want := range []string{"Ride", "Run", "Swim", "VirtualRide", "TrailRun", "Other"} {
		if !seen[want] {
			t.Errorf("missing %q", want)
		}
	}
	// Callers get a copy: mutating it must not affect validation.
	types[0] = "Bike"
	if AllActivityTypes()[0] != "Ride" {
		t.Error("AllActivityTypes returned shared backing array")
	}
}

func TestUpdateActivity_CaseMismatchSuggestsType(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	_, err := c.UpdateActivity(context.Background(), "i1", map[string]any{"type": "virtualride"})
	if err == nil || !strings.Contains(err.Error(), `did you mean "VirtualRide"?`) {
		t.Fatalf("err = %v, want a VirtualRide suggestion", err)
	}
}

func TestUpdateActivity_RejectsDotSegmentIDs(t *testing.T) {
	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called, got %s", r.URL.EscapedPath())
	}))

	// url.PathEscape leaves "." and ".." untouched, so they would reach the
	// server as dot segments of the PUT path.
	for _, id := range []string{".", "..", " .. "} {
		if _, err := c.UpdateActivity(context.Background(), id, map[string]any{"name": "x"}); err == nil {
			t.Errorf("id %q: expected error", id)
		}
	}
}

func TestUpdateActivity_ErrorClassification(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error // nil: plain *APIError, no sentinel
	}{
		{"400 mentioning Strava", http.StatusBadRequest, `{"error":"Strava activities cannot be updated"}`, ErrStravaActivity},
		{"499 mentioning STRAVA", 499, `STRAVA`, ErrStravaActivity},
		{"404 mentioning Strava wins over not-found", http.StatusNotFound, `{"error":"Strava activity"}`, ErrStravaActivity},
		{"plain 404", http.StatusNotFound, `{"error":"Not Found"}`, ErrActivityNotFound},
		{"500 mentioning Strava is a server error", http.StatusInternalServerError, `strava sync down`, nil},
		{"403 without Strava", http.StatusForbidden, `{"error":"Access denied"}`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))

			_, err := c.UpdateActivity(context.Background(), "i7", map[string]any{"name": "x"})
			// Every HTTP failure keeps the upstream *APIError reachable.
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
				t.Fatalf("err = %v, want *APIError %d in the chain", err, tc.status)
			}
			for _, sentinel := range []error{ErrStravaActivity, ErrActivityNotFound} {
				if got := errors.Is(err, sentinel); got != (sentinel == tc.want) {
					t.Errorf("errors.Is(err, %v) = %v; err = %v", sentinel, got, err)
				}
			}
			// Mapped errors carry a clean message, never the raw HTTP error.
			if tc.want != nil && strings.Contains(err.Error(), "PUT") {
				t.Errorf("err = %q, should not leak the raw HTTP error", err)
			}
		})
	}
}

func TestActivityUpdateFieldTypes_ReturnsCopy(t *testing.T) {
	types := ActivityUpdateFieldTypes()
	if len(types) != 10 || types["icu_rpe"] != "integer" || types["commute"] != "boolean" {
		t.Fatalf("whitelist = %v", types)
	}
	types["name"] = "integer"
	delete(types, "feel")

	c, _ := newActivitiesClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"i1"}`))
	}))
	if _, err := c.UpdateActivity(context.Background(), "i1", map[string]any{"name": "x", "feel": float64(3)}); err != nil {
		t.Fatalf("mutating the returned map changed validation: %v", err)
	}
}
