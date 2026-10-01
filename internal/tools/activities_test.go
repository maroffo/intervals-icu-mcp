// ABOUTME: Tests for activities MCP tool handlers: argument parsing, success, error mapping.
// ABOUTME: Reuses the package-level newToolClient, callReq and resultText helpers.

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
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
// its side-effects (wiring the five tool handlers) and then invokes each
// handler factory directly to confirm they each return a *mcp.CallToolResult
// without panicking when given minimal input. The mcp-go SDK (v0.48.0) does
// not expose ListTools introspection on an MCPServer, so we can't assert on
// the number of registered tools; invoking the handlers is the closest
// behavioural check available.
func TestRegisterActivities_InvokesAllHandlers(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal JSON that keeps every handler happy on the success path
		// (update_activity takes its decode-error path on it, still non-nil).
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
		{"update_activity", HandleUpdateActivity(c), map[string]any{"activity_id": "i1", "activity": map[string]any{"name": "x"}}},
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

func TestHandleUpdateActivity_SuccessShapesResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		// Full activity document, as intervals.icu returns it.
		_, _ = w.Write([]byte(`{"id":"i190900884","name":"Threshold 3x10","icu_rpe":8,"type":"Ride",` +
			`"moving_time":3600,"icu_training_load":92,"icu_ctl":61.4,"icu_atl":70.2,"source":"GARMIN_CONNECT"}`))
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": " i190900884 ",
		"activity":    map[string]any{"name": "Threshold 3x10", "icu_rpe": float64(8)},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	if gotMethod != http.MethodPut || gotPath != "/activity/i190900884" {
		t.Errorf("request = %s %s", gotMethod, gotPath)
	}
	if len(gotBody) != 2 {
		t.Errorf("body = %v, want only the two provided fields", gotBody)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(resultText(t, res)), &out); err != nil {
		t.Fatalf("result is not JSON: %v (%s)", err, resultText(t, res))
	}
	want := map[string]any{
		"id":                "i190900884",
		"updated":           map[string]any{"name": "Threshold 3x10", "icu_rpe": float64(8)},
		"icu_training_load": float64(92),
		"icu_ctl":           61.4,
		"icu_atl":           70.2,
	}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("result = %v\nwant    %v", out, want)
	}
}

func TestHandleUpdateActivity_OmitsLoadFieldsWhenAbsent(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"i1","commute":true,"icu_ctl":null}`))
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i1",
		"activity":    map[string]any{"commute": true},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %s", resultText(t, res))
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(resultText(t, res)), &out); err != nil {
		t.Fatalf("result is not JSON: %v", err)
	}
	for _, k := range []string{"icu_training_load", "icu_ctl", "icu_atl"} {
		if _, ok := out[k]; ok {
			t.Errorf("result has %q, want it omitted: %v", k, out)
		}
	}
	if !reflect.DeepEqual(out["updated"], map[string]any{"commute": true}) {
		t.Errorf("updated = %v", out["updated"])
	}
}

func TestHandleUpdateActivity_InvalidFieldRejected(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i1",
		"activity":    map[string]any{"name": "x", "icu_weight": float64(70)},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	txt := resultText(t, res)
	if !res.IsError || !strings.Contains(txt, "update_activity") || !strings.Contains(txt, `"icu_weight"`) {
		t.Fatalf("want error naming icu_weight, got %q (isErr=%v)", txt, res.IsError)
	}
}

func TestHandleUpdateActivity_InvalidTypeRejected(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i1",
		"activity":    map[string]any{"type": "Cycling"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	txt := resultText(t, res)
	if !res.IsError || !strings.Contains(txt, `"Cycling"`) {
		t.Fatalf("want error naming Cycling, got %q (isErr=%v)", txt, res.IsError)
	}
}

func TestHandleUpdateActivity_StravaActivityClearMessage(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"status":422,"error":"Strava activities cannot be updated"}`))
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "12345678",
		"activity":    map[string]any{"name": "x"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	txt := resultText(t, res)
	if !res.IsError || !strings.Contains(txt, "Strava") || !strings.Contains(txt, "12345678") {
		t.Fatalf("want clear Strava error, got %q (isErr=%v)", txt, res.IsError)
	}
	if strings.Contains(txt, "intervals.icu PUT") || strings.Contains(txt, "422") {
		t.Errorf("text = %q, should not be the raw HTTP error", txt)
	}
}

func TestHandleUpdateActivity_NotFound(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"error":"Not Found"}`))
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i404",
		"activity":    map[string]any{"name": "x"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	txt := resultText(t, res)
	if !res.IsError || !strings.Contains(txt, "not found") || !strings.Contains(txt, "i404") {
		t.Fatalf("want not-found error for i404, got %q (isErr=%v)", txt, res.IsError)
	}
	if !strings.Contains(txt, "list_activities") {
		t.Errorf("text = %q, want a recovery hint pointing at list_activities", txt)
	}
}

func TestHandleUpdateActivity_MissingArgs(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))

	cases := map[string]struct {
		args map[string]any
		want string
	}{
		"no activity_id":  {map[string]any{"activity": map[string]any{"name": "x"}}, "activity_id"},
		"no activity":     {map[string]any{"activity_id": "i1"}, `"activity"`},
		"empty activity":  {map[string]any{"activity_id": "i1", "activity": map[string]any{}}, "at least one field"},
		"activity scalar": {map[string]any{"activity_id": "i1", "activity": float64(3)}, `"activity"`},
	}
	for name, tc := range cases {
		res, err := HandleUpdateActivity(c)(context.Background(), callReq(tc.args))
		if err != nil {
			t.Fatalf("%s: handler err: %v", name, err)
		}
		if !res.IsError || !strings.Contains(resultText(t, res), tc.want) {
			t.Errorf("%s: got %q (isErr=%v), want error mentioning %s", name, resultText(t, res), res.IsError, tc.want)
		}
	}
}

func TestRegisterActivities_UpdateActivitySchema(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be called")
	}))
	s := server.NewMCPServer("test", "0.0.0")
	RegisterActivities(s, c)

	st := s.GetTool("update_activity")
	if st == nil {
		t.Fatal("update_activity not registered")
	}
	tool := st.Tool
	if !strings.HasPrefix(tool.Description, "MUTATES DATA.") {
		t.Errorf("description = %q, want MUTATES DATA. prefix", tool.Description)
	}
	if tool.Annotations.DestructiveHint == nil || !*tool.Annotations.DestructiveHint {
		t.Error("want destructive hint annotation")
	}
	// mcp-go already defaults DestructiveHint to true; the regression that
	// matters for a mutating tool is being marked read-only.
	if h := tool.Annotations.ReadOnlyHint; h != nil && *h {
		t.Error("update_activity must not carry a read-only hint")
	}
	for _, f := range []string{"type", "name", "description", "icu_rpe", "feel", "commute", "trainer",
		"icu_ignore_hr", "icu_ignore_power", "icu_ignore_time"} {
		if !strings.Contains(tool.Description, f) {
			t.Errorf("description does not list whitelisted field %q", f)
		}
	}
	if !slices.Equal(tool.InputSchema.Required, []string{"activity_id", "activity"}) {
		t.Errorf("required = %v", tool.InputSchema.Required)
	}

	obj, ok := tool.InputSchema.Properties["activity"].(map[string]any)
	if !ok {
		t.Fatalf("activity schema = %#v", tool.InputSchema.Properties["activity"])
	}
	if obj["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", obj["additionalProperties"])
	}
	props, _ := obj["properties"].(map[string]any)
	if len(props) != 10 {
		t.Errorf("activity has %d properties, want 10: %v", len(props), props)
	}
	typeProp, _ := props["type"].(map[string]any)
	if enum, _ := typeProp["enum"].([]string); len(enum) != 60 || !slices.Contains(enum, "Ride") {
		t.Errorf("type enum = %v, want the 60 spec values", typeProp["enum"])
	}
	if rpe, _ := props["icu_rpe"].(map[string]any); rpe["type"] != "integer" {
		t.Errorf("icu_rpe schema = %v", props["icu_rpe"])
	}
}

func TestHandleUpdateActivity_UndecodableResponse(t *testing.T) {
	c := newToolClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))

	res, err := HandleUpdateActivity(c)(context.Background(), callReq(map[string]any{
		"activity_id": "i1",
		"activity":    map[string]any{"name": "x"},
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	// The PUT already went out: the error must say so, so the caller does not
	// assume nothing changed.
	txt := resultText(t, res)
	if !res.IsError || !strings.Contains(txt, "update sent") {
		t.Fatalf("want error saying the update was sent, got %q (isErr=%v)", txt, res.IsError)
	}
}
