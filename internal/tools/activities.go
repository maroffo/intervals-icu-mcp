// ABOUTME: MCP tool handlers for the activities domain (list, get, streams, intervals).
// ABOUTME: Exported handler factories return ToolHandlerFunc values used by RegisterActivities.

package tools

import (
	"context"

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
