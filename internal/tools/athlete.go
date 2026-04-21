// ABOUTME: MCP tool registrations for the athlete domain (profile, fitness, folders).
// ABOUTME: Thin adapters over icu.Client; all tools are read-only and return raw JSON as text.

package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/maroffo/intervals-icu-mcp/internal/icu"
)

// RegisterAthlete registers all athlete-related MCP tools on s, delegating to c.
func RegisterAthlete(s *server.MCPServer, c *icu.Client) {
	s.AddTool(
		mcp.NewTool("get_athlete",
			mcp.WithDescription("Get athlete profile info (name, timezone, zones, sport settings)."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
		HandleGetAthlete(c),
	)

	s.AddTool(
		mcp.NewTool("get_fitness",
			mcp.WithDescription("Get CTL/ATL/TSB fitness time series for a date range."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("oldest", mcp.Description("Start date, yyyy-MM-dd.")),
			mcp.WithString("newest", mcp.Description("End date, yyyy-MM-dd.")),
			mcp.WithString("sport", mcp.Description("Optional sport filter (e.g. Ride, Run).")),
		),
		HandleGetFitness(c),
	)

	s.AddTool(
		mcp.NewTool("list_folders",
			mcp.WithDescription("List workout folders in the athlete's library."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
		HandleListFolders(c),
	)
}

// HandleGetAthlete returns the MCP handler for the get_athlete tool.
func HandleGetAthlete(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body, err := c.GetAthlete(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_athlete", err), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}
}

// HandleGetFitness returns the MCP handler for the get_fitness tool.
func HandleGetFitness(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		p := icu.GetFitnessParams{
			Oldest: req.GetString("oldest", ""),
			Newest: req.GetString("newest", ""),
			Sport:  req.GetString("sport", ""),
		}
		body, err := c.GetFitness(ctx, p)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get_fitness", err), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}
}

// HandleListFolders returns the MCP handler for the list_folders tool.
func HandleListFolders(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body, err := c.ListFolders(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list_folders", err), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}
}
