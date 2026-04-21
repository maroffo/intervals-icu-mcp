// ABOUTME: MCP tool registrations and handlers for the wellness domain (get, list, update).
// ABOUTME: Handlers are exported as plain functions to keep them unit-testable without a server.

package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/maroffo/intervals-icu-mcp/internal/icu"
)

// RegisterWellness registers all wellness-related MCP tools on the given server.
func RegisterWellness(s *server.MCPServer, c *icu.Client) {
	s.AddTool(
		mcp.NewTool(
			"get_wellness",
			mcp.WithDescription("Get wellness data for a single day (yyyy-MM-dd)."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("date",
				mcp.Required(),
				mcp.Description("Target day in yyyy-MM-dd format."),
			),
		),
		HandleGetWellness(c),
	)

	s.AddTool(
		mcp.NewTool(
			"list_wellness",
			mcp.WithDescription("List wellness entries in a date range."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("oldest", mcp.Description("Inclusive start date (yyyy-MM-dd). Optional.")),
			mcp.WithString("newest", mcp.Description("Inclusive end date (yyyy-MM-dd). Optional.")),
		),
		HandleListWellness(c),
	)

	s.AddTool(
		mcp.NewTool(
			"update_wellness",
			mcp.WithDescription(
				"MUTATES DATA. Update wellness entry for a day. "+
					"Accepts any subset of fields (weight, restingHR, hrv, sleepSecs, fatigue, "+
					"soreness, mood, motivation, injury, sickness, etc.).",
			),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("date",
				mcp.Required(),
				mcp.Description("Target day in yyyy-MM-dd format."),
			),
			mcp.WithObject("fields",
				mcp.Required(),
				mcp.Description(
					"JSON object with wellness fields to update. "+
						"Keys are field names (string), values depend on the field. "+
						"Unset fields are left untouched server-side.",
				),
			),
		),
		HandleUpdateWellness(c),
	)
}

// HandleGetWellness returns an MCP handler for the get_wellness tool.
func HandleGetWellness(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		date, err := req.RequireString("date")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_wellness", err), nil
		}
		raw, err := c.GetWellness(ctx, date)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_wellness", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleListWellness returns an MCP handler for the list_wellness tool.
func HandleListWellness(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		oldest := req.GetString("oldest", "")
		newest := req.GetString("newest", "")
		raw, err := c.ListWellness(ctx, icu.GetWellnessRangeParams{
			Oldest: oldest,
			Newest: newest,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list_wellness", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleUpdateWellness returns an MCP handler for the update_wellness tool.
func HandleUpdateWellness(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		date, err := req.RequireString("date")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_wellness", err), nil
		}
		fields, err := extractObjectArg(req, "fields")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update_wellness: %s", err.Error())), nil
		}
		if len(fields) == 0 {
			return mcp.NewToolResultError(`update_wellness: argument "fields" must not be empty`), nil
		}
		raw, err := c.UpdateWellness(ctx, date, fields)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_wellness", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}
