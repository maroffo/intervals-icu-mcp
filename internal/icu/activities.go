// ABOUTME: intervals.icu activities domain: list, get, streams, intervals, whitelisted update.
// ABOUTME: Thin wrappers over Client.Do that return json.RawMessage for pass-through.

package icu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// ListActivitiesParams holds the filters supported by the list endpoint.
type ListActivitiesParams struct {
	Oldest string // yyyy-MM-dd (optional)
	Newest string // yyyy-MM-dd (optional)
	Limit  int    // optional; when <= 0 the server default is used
}

// ListActivities returns the raw JSON array of activities for the authenticated
// athlete. Filtering is done server-side via the `oldest`, `newest` and `limit`
// query parameters. Empty filters are omitted from the request.
func (c *Client) ListActivities(ctx context.Context, p ListActivitiesParams) (json.RawMessage, error) {
	q := url.Values{}
	if s := strings.TrimSpace(p.Oldest); s != "" {
		q.Set("oldest", s)
	}
	if s := strings.TrimSpace(p.Newest); s != "" {
		q.Set("newest", s)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}

	path := "/athlete/" + url.PathEscape(c.athleteID) + "/activities"
	raw, err := c.Do(ctx, "GET", path, q, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// GetActivity returns the full activity document for the given id.
func (c *Client) GetActivity(ctx context.Context, activityID string) (json.RawMessage, error) {
	id, err := requireActivityID(activityID)
	if err != nil {
		return nil, err
	}
	path := "/activity/" + url.PathEscape(id)
	raw, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// GetActivityStreams returns the per-second data streams for an activity.
// If types is empty, no `types` query param is sent and the server returns all
// available streams.
func (c *Client) GetActivityStreams(ctx context.Context, activityID string, types []string) (json.RawMessage, error) {
	id, err := requireActivityID(activityID)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	if len(types) > 0 {
		q.Set("types", strings.Join(types, ","))
	}

	path := "/activity/" + url.PathEscape(id) + "/streams"
	raw, err := c.Do(ctx, "GET", path, q, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// GetActivityIntervals returns the detected or manual intervals for an activity.
func (c *Client) GetActivityIntervals(ctx context.Context, activityID string) (json.RawMessage, error) {
	id, err := requireActivityID(activityID)
	if err != nil {
		return nil, err
	}
	path := "/activity/" + url.PathEscape(id) + "/intervals"
	raw, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func requireActivityID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "", fmt.Errorf("activity_id is required")
	}
	// url.PathEscape leaves dot segments alone, so a path-normalising proxy or
	// server would resolve them to a different endpoint.
	if trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("invalid activity_id %q", trimmed)
	}
	return trimmed, nil
}

// activityTypes is the activity-type enum from the intervals.icu OpenAPI spec.
// Activity.type itself is a bare string there; this enum is the one shared by
// SportInfo.type, SportSettings.types and Folder.activity_types. Order matches
// the spec.
var activityTypes = []string{
	"Ride",
	"Run",
	"Swim",
	"WeightTraining",
	"Hike",
	"Walk",
	"AlpineSki",
	"BackcountrySki",
	"Badminton",
	"Canoeing",
	"Crossfit",
	"EBikeRide",
	"EMountainBikeRide",
	"Elliptical",
	"Golf",
	"GravelRide",
	"TrackRide",
	"Handcycle",
	"HighIntensityIntervalTraining",
	"Hockey",
	"IceSkate",
	"InlineSkate",
	"Kayaking",
	"Kitesurf",
	"MountainBikeRide",
	"Cyclocross",
	"NordicSki",
	"OpenWaterSwim",
	"Padel",
	"Pilates",
	"Pickleball",
	"Racquetball",
	"Rugby",
	"RockClimbing",
	"RollerSki",
	"Rowing",
	"Sail",
	"Skateboard",
	"Snowboard",
	"Snowshoe",
	"Soccer",
	"Squash",
	"StairStepper",
	"StandUpPaddling",
	"Surfing",
	"TableTennis",
	"Tennis",
	"TrailRun",
	"Transition",
	"Velomobile",
	"VirtualRide",
	"VirtualRow",
	"VirtualRun",
	"VirtualSki",
	"WaterSport",
	"Wheelchair",
	"Windsurf",
	"Workout",
	"Yoga",
	"Other",
}

// AllActivityTypes returns a copy of the activity types accepted by
// intervals.icu. Suitable for MCP enum/schema generation.
func AllActivityTypes() []string {
	return append([]string(nil), activityTypes...)
}

// activityUpdateFields is the whitelist of fields UpdateActivity may send,
// mapped to their JSON Schema type in the OpenAPI Activity schema.
var activityUpdateFields = map[string]string{
	"type":             "string",
	"name":             "string",
	"description":      "string",
	"icu_rpe":          "integer",
	"feel":             "integer",
	"commute":          "boolean",
	"trainer":          "boolean",
	"icu_ignore_hr":    "boolean",
	"icu_ignore_power": "boolean",
	"icu_ignore_time":  "boolean",
}

// ActivityUpdateFieldTypes returns a copy of the UpdateActivity whitelist:
// field name to JSON Schema type ("string", "integer" or "boolean").
func ActivityUpdateFieldTypes() map[string]string {
	out := make(map[string]string, len(activityUpdateFields))
	for k, v := range activityUpdateFields {
		out[k] = v
	}
	return out
}

// ErrStravaActivity reports an activity imported from Strava: intervals.icu
// does not allow reading or editing those through its API.
var ErrStravaActivity = errors.New("imported from Strava, which intervals.icu does not allow reading or editing through its API")

// ErrActivityNotFound reports a 404 from the activity endpoint.
var ErrActivityNotFound = errors.New("not found")

// activityError attaches the activity id, and the upstream *APIError when
// there is one, to a sentinel. Error() stays free of raw HTTP detail while
// errors.Is (sentinel) and errors.As (*APIError) both keep working.
type activityError struct {
	id       string
	sentinel error
	apiErr   *APIError
}

func (e *activityError) Error() string {
	return fmt.Sprintf("activity %s: %v", e.id, e.sentinel)
}

func (e *activityError) Unwrap() []error {
	if e.apiErr == nil {
		return []error{e.sentinel}
	}
	return []error{e.sentinel, e.apiErr}
}

// UpdateActivity performs a partial update of an activity. Only whitelisted
// fields (see ActivityUpdateFieldTypes) are accepted; anything else is rejected
// before any HTTP call, so the full activity is never sent back.
// PUT /activity/{id}
//
// The spec documents that Strava activities cannot be updated but not the
// error shape, so a 4xx whose body mentions Strava, or a 2xx carrying a
// source=STRAVA stub, maps to ErrStravaActivity. A 404 maps to
// ErrActivityNotFound. In both cases the *APIError stays in the chain.
func (c *Client) UpdateActivity(ctx context.Context, activityID string, fields map[string]any) (json.RawMessage, error) {
	id, err := requireActivityID(activityID)
	if err != nil {
		return nil, err
	}
	if err := validateActivityUpdate(fields); err != nil {
		return nil, err
	}

	path := "/activity/" + url.PathEscape(id)
	raw, err := c.Do(ctx, "PUT", path, nil, fields)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// Checked before 404 so that a 404 mentioning Strava gets the
			// Strava message. 5xx is a server failure, never a Strava refusal.
			if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 &&
				strings.Contains(strings.ToLower(apiErr.Body), "strava") {
				return nil, &activityError{id: id, sentinel: ErrStravaActivity, apiErr: apiErr}
			}
			if apiErr.StatusCode == http.StatusNotFound {
				return nil, &activityError{id: id, sentinel: ErrActivityNotFound, apiErr: apiErr}
			}
		}
		return nil, err
	}

	var stub struct {
		Source string `json:"source"`
	}
	if json.Unmarshal(raw, &stub) == nil && stub.Source == "STRAVA" {
		return nil, &activityError{id: id, sentinel: ErrStravaActivity}
	}
	return json.RawMessage(raw), nil
}

// validateActivityUpdate checks fields against the whitelist, the JSON types
// in the spec, and the activity-type enum. Keys are checked in sorted order so
// error messages are deterministic.
func validateActivityUpdate(fields map[string]any) error {
	if len(fields) == 0 {
		return errors.New("activity: at least one field is required")
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var unknown []string
	for _, k := range keys {
		if _, ok := activityUpdateFields[k]; !ok {
			unknown = append(unknown, strconv.Quote(k))
		}
	}
	if len(unknown) > 0 {
		allowed := make([]string, 0, len(activityUpdateFields))
		for k := range activityUpdateFields {
			allowed = append(allowed, k)
		}
		sort.Strings(allowed)
		return fmt.Errorf("activity: unsupported field(s) %s; only these fields can be updated: %s",
			strings.Join(unknown, ", "), strings.Join(allowed, ", "))
	}

	for _, k := range keys {
		want := activityUpdateFields[k]
		if !hasJSONType(fields[k], want) {
			if want == "integer" {
				want = "integer (int32)"
			}
			return fmt.Errorf("activity: field %q: expected %s, got %s", k, want, describeJSONValue(fields[k]))
		}
	}

	if t, ok := fields["type"]; ok {
		return validateActivityType(t.(string))
	}
	return nil
}

// validateActivityType checks t against the spec enum. Matching is exact; a
// case-only mismatch gets a suggestion.
func validateActivityType(t string) error {
	if slices.Contains(activityTypes, t) {
		return nil
	}
	for _, known := range activityTypes {
		if strings.EqualFold(known, t) {
			return fmt.Errorf("activity: unsupported activity type %q (did you mean %q?)", t, known)
		}
	}
	return fmt.Errorf("activity: unsupported activity type %q; see the enum in the update_activity schema", t)
}

// hasJSONType reports whether v, as decoded from JSON (or passed in-process),
// matches the JSON Schema type want. Integers are int32 in the spec.
func hasJSONType(v any, want string) bool {
	switch want {
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "integer":
		switch n := v.(type) {
		case float64:
			return n == math.Trunc(n) && n >= math.MinInt32 && n <= math.MaxInt32
		case int:
			return n >= math.MinInt32 && n <= math.MaxInt32
		case int64:
			return n >= math.MinInt32 && n <= math.MaxInt32
		}
	}
	return false
}

// describeJSONValue names v in JSON terms for error messages.
func describeJSONValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("string %q", x)
	case bool:
		return fmt.Sprintf("boolean %t", x)
	case float64, int, int64:
		return fmt.Sprintf("number %v", x)
	case map[string]any:
		return "object"
	case []any:
		return "array"
	}
	return fmt.Sprintf("%T", v)
}
