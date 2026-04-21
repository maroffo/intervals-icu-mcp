// ABOUTME: MCP tool registration and handlers for the calendar events domain.
// ABOUTME: Exposes list/create/update/delete events; mutating tools are clearly marked.

package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/maroffo/intervals-icu-mcp/internal/icu"
)

// RegisterEvents wires all calendar events MCP tools onto the given server.
func RegisterEvents(s *server.MCPServer, c *icu.Client) {
	categories := icu.AllEventCategories()

	s.AddTool(
		mcp.NewTool("list_events",
			mcp.WithDescription("List calendar events (workouts, races, notes, fitness targets) in a date range."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("oldest", mcp.Description("Oldest date (yyyy-MM-dd), inclusive.")),
			mcp.WithString("newest", mcp.Description("Newest date (yyyy-MM-dd), inclusive.")),
			mcp.WithString("category",
				mcp.Description("Filter by category."),
				mcp.Enum(categories...),
			),
			mcp.WithBoolean("resolve", mcp.Description("If true, expand referenced workouts inline.")),
		),
		HandleListEvents(c),
	)

	s.AddTool(
		mcp.NewTool("create_event",
			mcp.WithDescription("MUTATES DATA. Create a new calendar event. Required fields in event: start_date_local (ISO datetime), category (WORKOUT/RACE_A/RACE_B/RACE_C/NOTE/FITNESS/TARGET). Common fields: name, description, end_date_local, type, indoor, color, moving_time, load_target, workout_doc (structured workout steps)."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithObject("event",
				mcp.Required(),
				mcp.Description("Event fields as a JSON object. Must include start_date_local and category."),
			),
		),
		HandleCreateEvent(c),
	)

	s.AddTool(
		mcp.NewTool("update_event",
			mcp.WithDescription("MUTATES DATA. Update an existing calendar event. Pass any subset of fields to modify."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("Id of the event to update.")),
			mcp.WithObject("event",
				mcp.Required(),
				mcp.Description("Subset of fields to modify, as a JSON object."),
			),
		),
		HandleUpdateEvent(c),
	)

	s.AddTool(
		mcp.NewTool("delete_event",
			mcp.WithDescription("MUTATES DATA. Delete a calendar event by id."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("Id of the event to delete.")),
		),
		HandleDeleteEvent(c),
	)
}

// HandleListEvents returns the MCP handler for the list_events tool.
func HandleListEvents(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := icu.ListEventsParams{
			Oldest:   req.GetString("oldest", ""),
			Newest:   req.GetString("newest", ""),
			Category: req.GetString("category", ""),
			Resolve:  req.GetBool("resolve", false),
		}
		raw, err := c.ListEvents(ctx, params)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list_events", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleCreateEvent returns the MCP handler for the create_event tool.
func HandleCreateEvent(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		payload, err := extractObjectArg(req, "event")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("create_event: %s", err.Error())), nil
		}
		if len(payload) == 0 {
			return mcp.NewToolResultError(`create_event: argument "event" must not be empty`), nil
		}
		if _, ok := payload["start_date_local"]; !ok {
			return mcp.NewToolResultError(`create_event: event must include "start_date_local"`), nil
		}
		if _, ok := payload["category"]; !ok {
			return mcp.NewToolResultError(`create_event: event must include "category"`), nil
		}
		raw, err := c.CreateEvent(ctx, payload)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create_event", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleUpdateEvent returns the MCP handler for the update_event tool.
func HandleUpdateEvent(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		eventID, err := req.RequireString("event_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update_event: %s", err.Error())), nil
		}
		payload, err := extractObjectArg(req, "event")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update_event: %s", err.Error())), nil
		}
		if len(payload) == 0 {
			return mcp.NewToolResultError(`update_event: argument "event" must not be empty`), nil
		}
		raw, err := c.UpdateEvent(ctx, eventID, payload)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update_event", err), nil
		}
		return mcp.NewToolResultText(string(raw)), nil
	}
}

// HandleDeleteEvent returns the MCP handler for the delete_event tool.
func HandleDeleteEvent(c *icu.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		eventID, err := req.RequireString("event_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("delete_event: %s", err.Error())), nil
		}
		if err := c.DeleteEvent(ctx, eventID); err != nil {
			return mcp.NewToolResultErrorFromErr("delete_event", err), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("deleted event %s", eventID)), nil
	}
}
