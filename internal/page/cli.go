package page

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(repo Repository) *cobra.Command {
	root := &cobra.Command{
		Use:   "page",
		Short: "Get and update Notion pages",
	}
	root.AddCommand(newGetCmd(repo), newUpdateCmd(repo))
	return root
}

func newGetCmd(repo Repository) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <page_id>",
		Short: "Get a Notion page by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := repo.Get(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("Failed to get page: %w", err)
			}
			return writePage(cmd, p, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	return cmd
}

func newUpdateCmd(repo Repository) *cobra.Command {
	var (
		propsJSON string
		output    string
	)
	cmd := &cobra.Command{
		Use:   "update <page_id> --properties <json>",
		Short: "Update a Notion page's properties",
		Long: `Update a Notion page's properties.

The --properties flag takes a JSON object mapping property names to Notion API property values.
Examples:
  - Title:    {"Task name": {"title": [{"text": {"content": "New name"}}]}}
  - Status:   {"Status": {"status": {"name": "Done"}}}
  - Select:   {"Tag": {"select": {"name": "Dev"}}}
  - Date:     {"Due": {"date": {"start": "2026-04-15"}}}
  - Checkbox: {"Processed": {"checkbox": true}}
  - Relation: {"Project": {"relation": [{"id": "page-uuid"}]}}
  - Rich text:{"Summary": {"rich_text": [{"text": {"content": "text"}}]}}
  - URL:      {"userDefined:URL": {"url": "https://example.com"}}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var propUpdates map[string]any
			if err := json.Unmarshal([]byte(propsJSON), &propUpdates); err != nil {
				return fmt.Errorf("Invalid properties JSON: %w", err)
			}
			p, err := repo.Update(cmd.Context(), UpdateParams{
				PageID:          args[0],
				PropertyUpdates: propUpdates,
			})
			if err != nil {
				return fmt.Errorf("Failed to update page: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated page: %s\nURL: %s\n\n", p.ID, p.URL)
			return writePage(cmd, p, output)
		},
	}
	cmd.Flags().StringVar(&propsJSON, "properties", "", "JSON object of property updates")
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	_ = cmd.MarkFlagRequired("properties")
	return cmd
}

func writePage(cmd *cobra.Command, p *Page, output string) error {
	switch output {
	case "json":
		return json.NewEncoder(cmd.OutOrStdout()).Encode(p)
	default:
		fmt.Fprint(cmd.OutOrStdout(), formatPage(p))
		return nil
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
