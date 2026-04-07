package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emiliopalmerini/notion-mcp/internal/blackhole"
	"github.com/emiliopalmerini/notion-mcp/internal/events"
	"github.com/emiliopalmerini/notion-mcp/internal/page"
	"github.com/emiliopalmerini/notion-mcp/internal/projects"
	"github.com/emiliopalmerini/notion-mcp/internal/shared/notion"
	"github.com/emiliopalmerini/notion-mcp/internal/tasks"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		log.Fatal("NOTION_TOKEN environment variable is required")
	}

	client := notion.NewClient(token)

	mcpServer := server.NewMCPServer(
		"notion-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Register all slices
	tasks.RegisterTools(mcpServer, tasks.NewNotionRepository(client))
	projects.RegisterTools(mcpServer, projects.NewNotionRepository(client))
	events.RegisterTools(mcpServer, events.NewNotionRepository(client))
	blackhole.RegisterTools(mcpServer, blackhole.NewNotionRepository(client))
	page.RegisterTools(mcpServer, page.NewNotionRepository(client))

	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
