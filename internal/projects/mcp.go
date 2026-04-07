package projects

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, repo Repository) {
	s.AddTool(mcp.NewTool("query_projects",
		mcp.WithDescription("Query projects from Notion. Returns filtered, sorted results."),
		mcp.WithString("status", mcp.Description("Filter by status: Planning, In Progress, Paused, Backlog, Done, Canceled")),
		mcp.WithString("tag", mcp.Description("Filter by tag: Content, Dev, Marketing, Community, Business, Work")),
		mcp.WithString("name", mcp.Description("Filter by name (contains match)")),
		mcp.WithString("date_start_after", mcp.Description("Filter projects starting on or after this date (ISO-8601)")),
		mcp.WithString("date_start_before", mcp.Description("Filter projects starting on or before this date (ISO-8601)")),
		mcp.WithBoolean("has_launch_date", mcp.Description("Filter projects that have a launch date set")),
		mcp.WithString("sort_by", mcp.Description("Property to sort by (default: Dates)")),
		mcp.WithString("sort_direction", mcp.Description("Sort direction: ascending or descending (default: descending)")),
		mcp.WithNumber("page_size", mcp.Description("Number of results (1-100, default 50)")),
		mcp.WithString("start_cursor", mcp.Description("Pagination cursor from previous query")),
	), queryHandler(repo))

	s.AddTool(mcp.NewTool("create_project",
		mcp.WithDescription("Create a new project in Notion."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
		mcp.WithString("status", mcp.Description("Initial status (default: Planning)")),
		mcp.WithString("tag", mcp.Description("Project tag: Content, Dev, Marketing, Community, Business, Work")),
		mcp.WithString("summary", mcp.Description("Brief project description")),
		mcp.WithString("dates_start", mcp.Description("Project start date (ISO-8601)")),
		mcp.WithString("dates_end", mcp.Description("Project end date (ISO-8601)")),
		mcp.WithString("launch_date", mcp.Description("Target launch/publish date (ISO-8601)")),
	), createHandler(repo))
}

func queryHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		f := Filter{
			Status:          argString(args, "status"),
			Tag:             argString(args, "tag"),
			Name:            argString(args, "name"),
			DateStartAfter:  argString(args, "date_start_after"),
			DateStartBefore: argString(args, "date_start_before"),
			HasLaunchDate:   argBool(args, "has_launch_date"),
			SortBy:          argString(args, "sort_by"),
			SortDir:         argString(args, "sort_direction"),
			PageSize:        argInt(args, "page_size"),
			StartCursor:     argString(args, "start_cursor"),
		}

		if f.Status != "" && !isValid(f.Status, ValidStatuses) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid status '%s'. Valid options: %s", f.Status, strings.Join(ValidStatuses, ", "),
			)), nil
		}
		if f.Tag != "" && !isValid(f.Tag, ValidTags) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid tag '%s'. Valid options: %s", f.Tag, strings.Join(ValidTags, ", "),
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
			Name:       argString(args, "name"),
			Status:     argString(args, "status"),
			Tag:        argString(args, "tag"),
			Summary:    argString(args, "summary"),
			DatesStart: argString(args, "dates_start"),
			DatesEnd:   argString(args, "dates_end"),
			LaunchDate: argString(args, "launch_date"),
		}

		if p.Status != "" && !isValid(p.Status, ValidStatuses) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid status '%s'. Valid options: %s", p.Status, strings.Join(ValidStatuses, ", "),
			)), nil
		}
		if p.Tag != "" && !isValid(p.Tag, ValidTags) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid tag '%s'. Valid options: %s", p.Tag, strings.Join(ValidTags, ", "),
			)), nil
		}

		project, err := repo.Create(ctx, p)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Create failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Created project: **%s**\nID: %s\nURL: %s", project.Name, project.ID, project.URL)), nil
	}
}

func formatQueryResult(r *QueryResult) string {
	if len(r.Projects) == 0 {
		return "No projects found."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d project(s):\n\n", len(r.Projects))
	for _, p := range r.Projects {
		fmt.Fprintf(&sb, "- **%s** [%s] — Status: %s", p.Name, p.ID, p.Status)
		if p.Tag != "" {
			fmt.Fprintf(&sb, ", Tag: %s", p.Tag)
		}
		if p.DatesStart != "" {
			fmt.Fprintf(&sb, ", Dates: %s", p.DatesStart)
			if p.DatesEnd != "" {
				fmt.Fprintf(&sb, " to %s", p.DatesEnd)
			}
		}
		sb.WriteString("\n")
	}
	if r.HasMore {
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_", r.NextCursor)
	}
	return sb.String()
}

func isValid(value string, valid []string) bool {
	for _, v := range valid {
		if v == value {
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
