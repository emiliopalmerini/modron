package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, repo Repository) {
	s.AddTool(mcp.NewTool("query_tasks",
		mcp.WithDescription("Query tasks from Notion. Returns filtered, sorted results."),
		mcp.WithString("status", mcp.Description("Filter by status: Not Started, In progress, Done, Archived")),
		mcp.WithString("name", mcp.Description("Filter by name (contains match)")),
		mcp.WithString("due_before", mcp.Description("Filter tasks due on or before this date (ISO-8601)")),
		mcp.WithString("due_after", mcp.Description("Filter tasks due on or after this date (ISO-8601)")),
		mcp.WithString("project_id", mcp.Description("Filter by project page ID")),
		mcp.WithString("sort_by", mcp.Description("Property to sort by (default: Due)")),
		mcp.WithString("sort_direction", mcp.Description("Sort direction: ascending or descending (default: ascending)")),
		mcp.WithNumber("page_size", mcp.Description("Number of results (1-100, default 50)")),
		mcp.WithString("start_cursor", mcp.Description("Pagination cursor from previous query")),
	), queryHandler(repo))

	s.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task in Notion."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Task name")),
		mcp.WithString("due", mcp.Required(), mcp.Description("Due date (ISO-8601)")),
		mcp.WithString("status", mcp.Description("Initial status (default: Not Started)")),
		mcp.WithString("project_id", mcp.Description("Parent project page ID")),
		mcp.WithString("parent_task_id", mcp.Description("Parent task page ID")),
	), createHandler(repo))
}

func queryHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		f := Filter{
			Status:      argString(args, "status"),
			Name:        argString(args, "name"),
			DueBefore:   argString(args, "due_before"),
			DueAfter:    argString(args, "due_after"),
			ProjectID:   argString(args, "project_id"),
			SortBy:      argString(args, "sort_by"),
			SortDir:     argString(args, "sort_direction"),
			PageSize:    argInt(args, "page_size"),
			StartCursor: argString(args, "start_cursor"),
		}

		if f.Status != "" && !isValidStatus(f.Status) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid status '%s'. Valid options: %s", f.Status, strings.Join(ValidStatuses, ", "),
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
			Name:         argString(args, "name"),
			Due:          argString(args, "due"),
			Status:       argString(args, "status"),
			ProjectID:    argString(args, "project_id"),
			ParentTaskID: argString(args, "parent_task_id"),
		}

		if p.Status != "" && !isValidStatus(p.Status) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"Invalid status '%s'. Valid options: %s", p.Status, strings.Join(ValidStatuses, ", "),
			)), nil
		}

		task, err := repo.Create(ctx, p)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Create failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Created task: **%s**\nID: %s\nURL: %s", task.Name, task.ID, task.URL)), nil
	}
}

func formatQueryResult(r *QueryResult) string {
	if len(r.Tasks) == 0 {
		return "No tasks found."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d task(s):\n\n", len(r.Tasks))
	for _, t := range r.Tasks {
		fmt.Fprintf(&sb, "- **%s** [%s] — Status: %s", t.Name, t.ID, t.Status)
		if t.Due != "" {
			fmt.Fprintf(&sb, ", Due: %s", t.Due)
		}
		sb.WriteString("\n")
	}
	if r.HasMore {
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_", r.NextCursor)
	}
	return sb.String()
}

func isValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
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
