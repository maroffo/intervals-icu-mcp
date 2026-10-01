// ABOUTME: MCP tool handlers for the activities domain (list, get, streams, intervals, update).
// ABOUTME: Exported handler factories return ToolHandlerFunc values used by RegisterActivities.

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/maroffo/intervals-icu-mcp/internal/icu"
)

// RegisterActivities registers all activities-related MCP tools on the server.
func RegisterActivities(s *server.MCPServer, c *icu.Client) {
	s.AddTool(
		mcp.NewTool("list_activities",
			mcp.WithDescription("List athlete activities in a date range."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("oldest", mcp.Description("Earliest date, yyyy-MM-dd (optional).")),
			mcp.WithString("newest", mcp.Description("Latest date, yyyy-MM-dd (optional).")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of activities to return (optional, server default when 0).")),
		),
		HandleListActivities(c),
	)

	s.AddTool(
		mcp.NewTool("get_activity",
			mcp.WithDescription("Get a single activity by id."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("activity_id", mcp.Required(), mcp.Description("Intervals.icu activity id, e.g. i1234567.")),
		),
		HandleGetActivity(c),
	)

	s.AddTool(
		mcp.NewTool("get_activity_streams",
			mcp.WithDescription("Get per-second data streams (power, HR, cadence, etc.) for an activity."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("activity_id", mcp.Required(), mcp.Description("Intervals.icu activity id.")),
			mcp.WithArray("types",
				mcp.Description("Optional list of stream names (e.g. watts, heartrate, cadence, time, distance, latlng)."),
				mcp.Items(map[string]any{"type": "string"}),
			),
		),
		HandleGetActivityStreams(c),
	)

	s.AddTool(
		mcp.NewTool("get_activity_intervals",
			mcp.WithDescription("Get detected or manual intervals for an activity."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("activity_id", mcp.Required(), mcp.Description("Intervals.icu activity id.")),
		),
		HandleGetActivityIntervals(c),
	)

	s.AddTool(
		mcp.NewTool("update_activity",
			mcp.WithDescription("MUTATES DATA. Update a recorded activity. Only these fields are sent, as a partial update "+
				"(anything else is rejected and the full activity is never sent back): type, name, description, icu_rpe, "+
				"feel, commute, trainer, icu_ignore_hr, icu_ignore_power, icu_ignore_time. Activities imported from Strava "+
				"cannot be edited through the API. Returns the activity id, the updated values, and the new "+
				"icu_training_load / icu_ctl / icu_atl when present."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("activity_id", mcp.Required(), mcp.Description("Intervals.icu activity id, e.g. i190900884.")),
			mcp.WithObject("activity",
				mcp.Required(),
				mcp.Description("Fields to change, as a JSON object. type must be one of the values in its enum."),
				mcp.Properties(activityUpdateProperties()),
				mcp.AdditionalProperties(false),
			),
		),
		HandleUpdateActivity(c),
	)
}

// activityUpdateProperties builds the JSON Schema properties of the
// update_activity "activity" argument from the icu whitelist.
func activityUpdateProperties() map[string]any {
	props := map[string]any{}
	for name, typ := range icu.ActivityUpdateFieldTypes() {
		props[name] = map[string]any{"type": typ}
	}
	props["type"] = map[string]any{"type": "string", "enum": icu.AllActivityTypes()}
	return props
}

// HandleListActivities returns a handler for the list_activities tool.
func HandleListActivities(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := icu.ListActivitiesParams{
			Oldest: req.GetString("oldest", ""),
			Newest: req.GetString("newest", ""),
			Limit:  int(req.GetFloat("limit", 0)),
		}
		raw, err := c.ListActivities(ctx, params)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list_activities", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleGetActivity returns a handler for the get_activity tool.
func HandleGetActivity(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("activity_id")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity", err), nil
		}
		raw, err := c.GetActivity(ctx, id)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleGetActivityStreams returns a handler for the get_activity_streams tool.
func HandleGetActivityStreams(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("activity_id")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity_streams", err), nil
		}
		types := req.GetStringSlice("types", nil)
		raw, err := c.GetActivityStreams(ctx, id, types)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity_streams", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleGetActivityIntervals returns a handler for the get_activity_intervals tool.
func HandleGetActivityIntervals(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("activity_id")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity_intervals", err), nil
		}
		raw, err := c.GetActivityIntervals(ctx, id)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_activity_intervals", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleUpdateActivity returns a handler for the update_activity tool.
func HandleUpdateActivity(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("activity_id")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_activity", err), nil
		}
		fields, err := extractObjectArg(req, "activity")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update_activity: %s", err.Error())), nil
		}
		raw, err := c.UpdateActivity(ctx, id, fields)
		if errors.Is(err, icu.ErrActivityNotFound) {
			return mcp.NewToolResultError(fmt.Sprintf("update_activity: %s (check the id with list_activities)", err.Error())), nil
		}
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_activity", err), nil
		}
		summary, err := summarizeActivityUpdate(id, fields, raw)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_activity: update sent but the response could not be decoded", err), nil
		}
		return mcp.NewToolResultText(string(summary)), nil
	}
}

// activityLoadFields are reported after an update when the response has them.
var activityLoadFields = []string{"icu_training_load", "icu_ctl", "icu_atl"}

// summarizeActivityUpdate builds the update_activity result: the activity id
// (from the response, falling back to the requested one), the server's values
// for the fields that were sent, and the training load metrics when present
// (non-null) in the response.
func summarizeActivityUpdate(id string, sent map[string]any, raw json.RawMessage) ([]byte, error) {
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	updated := make(map[string]json.RawMessage, len(sent))
	for k := range sent {
		if v, ok := resp[k]; ok {
			updated[k] = v
		}
	}
	out := map[string]any{"id": id, "updated": updated}
	if v, ok := resp["id"]; ok {
		out["id"] = v
	}
	for _, k := range activityLoadFields {
		if v, ok := resp[k]; ok && string(v) != "null" {
			out[k] = v
		}
	}
	return json.Marshal(out)
}
