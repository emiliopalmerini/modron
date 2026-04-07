package events

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, repo Repository) {
	s.AddTool(mcp.NewTool("query_events",
		mcp.WithDescription("Query events from Notion. Returns filtered, sorted results."),
		mcp.WithString("name", mcp.Description("Filter by name (contains match)")),
		mcp.WithString("date_after", mcp.Description("Filter events on or after this date (ISO-8601)")),
		mcp.WithString("date_before", mcp.Description("Filter events on or before this date (ISO-8601)")),
		mcp.WithString("sort_by", mcp.Description("Property to sort by (default: Date)")),
		mcp.WithString("sort_direction", mcp.Description("Sort direction: ascending or descending (default: ascending)")),
		mcp.WithNumber("page_size", mcp.Description("Number of results (1-100, default 50)")),
		mcp.WithString("start_cursor", mcp.Description("Pagination cursor from previous query")),
	), queryHandler(repo))

	s.AddTool(mcp.NewTool("create_event",
		mcp.WithDescription("Create a new event in Notion."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Event name")),
		mcp.WithString("date_start", mcp.Required(), mcp.Description("Event start date/datetime (ISO-8601)")),
		mcp.WithString("date_end", mcp.Description("Event end date/datetime (ISO-8601)")),
		mcp.WithBoolean("is_datetime", mcp.Description("Whether the date includes time (default: false)")),
	), createHandler(repo))
}

func queryHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		f := Filter{
			Name:        argString(args, "name"),
			DateAfter:   argString(args, "date_after"),
			DateBefore:  argString(args, "date_before"),
			SortBy:      argString(args, "sort_by"),
			SortDir:     argString(args, "sort_direction"),
			PageSize:    argInt(args, "page_size"),
			StartCursor: argString(args, "start_cursor"),
		}

		result, err := repo.Query(ctx, f)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Query failed: %v", err)), nil
		}

		return mcp.NewToolResultText(formatQueryResult(result)), nil
	}
}

func createHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		p := CreateParams{
			Name:       argString(args, "name"),
			DateStart:  argString(args, "date_start"),
			DateEnd:    argString(args, "date_end"),
			IsDatetime: argBool(args, "is_datetime"),
		}

		event, err := repo.Create(ctx, p)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Create failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Created event: **%s**\nID: %s\nURL: %s", event.Name, event.ID, event.URL)), nil
	}
}

func formatQueryResult(r *QueryResult) string {
	if len(r.Events) == 0 {
		return "No events found."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d event(s):\n\n", len(r.Events))
	for _, e := range r.Events {
		fmt.Fprintf(&sb, "- **%s** [%s]", e.Name, e.ID)
		if e.DateStart != "" {
			fmt.Fprintf(&sb, " — Date: %s", e.DateStart)
			if e.DateEnd != "" {
				fmt.Fprintf(&sb, " to %s", e.DateEnd)
			}
		}
		sb.WriteString("\n")
	}
	if r.HasMore {
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_", r.NextCursor)
	}
	return sb.String()
}

func argString(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}

func argInt(args map[string]any, key string) int {
	v, _ := args[key].(float64)
	return int(v)
}

func argBool(args map[string]any, key string) bool {
	v, _ := args[key].(bool)
	return v
}
