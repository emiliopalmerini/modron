package page

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, repo Repository) {
	s.AddTool(mcp.NewTool("get_page",
		mcp.WithDescription("Get a Notion page by ID. Returns all properties with their values."),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Notion page ID (UUID)")),
	), getHandler(repo))

	s.AddTool(mcp.NewTool("update_page",
		mcp.WithDescription("Update a Notion page's properties. Pass property updates as a JSON object mapping property names to Notion API property values."),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Notion page ID (UUID)")),
		mcp.WithString("properties", mcp.Required(), mcp.Description(
			`JSON object of property updates. Examples:
- Title: {"Task name": {"title": [{"text": {"content": "New name"}}]}}
- Status: {"Status": {"status": {"name": "Done"}}}
- Select: {"Tag": {"select": {"name": "Dev"}}}
- Date: {"Due": {"date": {"start": "2026-04-15"}}}
- Checkbox: {"Processed": {"checkbox": true}}
- Relation: {"Project": {"relation": [{"id": "page-uuid"}]}}
- Rich text: {"Summary": {"rich_text": [{"text": {"content": "text"}}]}}
- URL: {"userDefined:URL": {"url": "https://example.com"}}`,
		)),
	), updateHandler(repo))
}

func getHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		pageID, _ := args["page_id"].(string)

		page, err := repo.Get(ctx, pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page: %v", err)), nil
		}

		return mcp.NewToolResultText(formatPage(page)), nil
	}
}

func updateHandler(repo Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		pageID, _ := args["page_id"].(string)
		propsJSON, _ := args["properties"].(string)

		var propUpdates map[string]any
		if err := json.Unmarshal([]byte(propsJSON), &propUpdates); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid properties JSON: %v", err)), nil
		}

		page, err := repo.Update(ctx, UpdateParams{
			PageID:          pageID,
			PropertyUpdates: propUpdates,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update page: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Updated page: %s\nURL: %s\n\n%s", page.ID, page.URL, formatPage(page))), nil
	}
}

func formatPage(p *Page) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "**Page** %s\n", p.ID)
	fmt.Fprintf(&sb, "URL: %s\n", p.URL)
	if p.CreatedTime != "" {
		fmt.Fprintf(&sb, "Created: %s\n", p.CreatedTime)
	}
	if p.LastEditedTime != "" {
		fmt.Fprintf(&sb, "Last edited: %s\n", p.LastEditedTime)
	}
	sb.WriteString("\n**Properties:**\n")

	// Sort properties by name for deterministic output
	sorted := make([]Property, len(p.Properties))
	copy(sorted, p.Properties)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, prop := range sorted {
		if prop.Value == "" {
			continue
		}
		fmt.Fprintf(&sb, "- %s (%s): %s\n", prop.Name, prop.Type, prop.Value)
	}
	return sb.String()
}
