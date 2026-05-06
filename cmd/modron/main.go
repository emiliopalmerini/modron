package main

import (
	"fmt"
	"os"

	"github.com/emiliopalmerini/modron/internal/blackhole"
	"github.com/emiliopalmerini/modron/internal/events"
	"github.com/emiliopalmerini/modron/internal/page"
	"github.com/emiliopalmerini/modron/internal/projects"
	"github.com/emiliopalmerini/modron/internal/shared/notion"
	"github.com/emiliopalmerini/modron/internal/tasks"
	"github.com/spf13/cobra"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		return fmt.Errorf("NOTION_TOKEN environment variable is required")
	}

	client := notion.NewClient(token)

	root := &cobra.Command{
		Use:           "modron",
		Short:         "Deterministic CLI for Notion databases",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.AddCommand(
		tasks.NewCommand(tasks.NewNotionRepository(client)),
		projects.NewCommand(projects.NewNotionRepository(client)),
		events.NewCommand(events.NewNotionRepository(client)),
		blackhole.NewCommand(blackhole.NewNotionRepository(client)),
		page.NewCommand(page.NewNotionRepository(client)),
	)

	return root.Execute()
}
