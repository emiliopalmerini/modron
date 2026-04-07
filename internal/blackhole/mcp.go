package blackhole

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, repo Repository) {
	s.AddTool(mcp.NewTool("query_blackhole",
		mcp.WithDescription("Query BlackholeDB entries from Notion. Stores ideas, references, and content to read/watch later."),
		mcp.WithString("type", mcp.Description("Filter by type: Idea, Reference, TBR")),
		mcp.WithString("tags", mcp.Description("Filter by tags (comma-separated, all must match)")),
		mcp.WithString("name", mcp.Description("Filter by name (contains match)")),
		mcp.WithBoolean("processed", mcp.Description("Filter by processed status")),
		mcp.WithBoolean("has_url", mcp.Description("Filter entries that have a URL")),
		mcp.WithString("sort_by", mcp.Description("Property to sort by (default: created_time)")),
		mcp.WithString("sort_direction", mcp.Description("Sort direction: ascending or descending (default: descending)")),
		mcp.WithNumber("page_size", mcp.Description("Number of results (1-100, default 50)")),
		mcp.WithString("start_cursor", mcp.Description("Pagination cursor from previous query")),
	), queryHandler(repo))

	s.AddTool(mcp.NewTool("create_blackhole_entry",
		mcp.WithDescription("Save an idea, reference, or content to BlackholeDB in Notion."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Entry name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Entry type: Idea, Reference, or TBR")),
		mcp.WithString("summary", mcp.Description("Brief description of the content")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		mcp.WithString("url", mcp.Description("URL of the resource")),
		mcp.WithBoolean("processed", mcp.Description("Whether already read/acted upon (default: false)")),
	), createHandler(repo))
}

func queryHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		f := Filter{
			Type:        argString(args, "type"),
			Tags:        argString(args, "tags"),
			Name:        argString(args, "name"),
			HasURL:      argBool(args, "has_url"),
			SortBy:      argString(args, "sort_by"),
			SortDir:     argString(args, "sort_direction"),
			PageSize:    argInt(args, "page_size"),
			StartCursor: argString(args, "start_cursor"),
		}

		if v, ok := args["processed"].(bool); ok {
			f.Processed = &v
		}

		if f.Type != "" && !isValidType(f.Type) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid type '%s'. Valid options: %s", f.Type, strings.Join(ValidTypes, ", "),
			)), nil
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
			Name:      argString(args, "name"),
			Type:      argString(args, "type"),
			Summary:   argString(args, "summary"),
			Tags:      argString(args, "tags"),
			URL:       argString(args, "url"),
			Processed: argBool(args, "processed"),
		}

		if !isValidType(p.Type) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid type '%s'. Valid options: %s", p.Type, strings.Join(ValidTypes, ", "),
			)), nil
		}

		entry, err := repo.Create(ctx, p)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Create failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Created entry: **%s** (%s)\nID: %s\nURL: %s", entry.Name, entry.Type, entry.ID, entry.URL)), nil
	}
}

func formatQueryResult(r *QueryResult) string {
	if len(r.Entries) == 0 {
		return "No entries found."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d entry(ies):\n\n", len(r.Entries))
	for _, e := range r.Entries {
		fmt.Fprintf(&sb, "- **%s** [%s] — Type: %s", e.Name, e.ID, e.Type)
		if len(e.Tags) > 0 {
			fmt.Fprintf(&sb, ", Tags: %s", strings.Join(e.Tags, ", "))
		}
		if e.EntryURL != "" {
			fmt.Fprintf(&sb, ", URL: %s", e.EntryURL)
		}
		if e.Processed {
			sb.WriteString(" [processed]")
		}
		sb.WriteString("\n")
	}
	if r.HasMore {
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_", r.NextCursor)
	}
	return sb.String()
}

func isValidType(t string) bool {
	for _, v := range ValidTypes {
		if v == t {
			return true
		}
	}
	return false
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
