package blackhole

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(repo Repository) *cobra.Command {
	root := &cobra.Command{
		Use:   "blackhole",
		Short: "Query and create BlackholeDB entries (ideas, references, TBR)",
	}
	root.AddCommand(newQueryCmd(repo), newCreateCmd(repo))
	return root
}

func newQueryCmd(repo Repository) *cobra.Command {
	var (
		f         Filter
		output    string
		processed bool
	)
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query BlackholeDB entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("processed") {
				v := processed
				f.Processed = &v
			}
			if f.Type != "" && !isValidType(f.Type) {
				return fmt.Errorf("Invalid type '%s'. Valid options: %s", f.Type, strings.Join(ValidTypes, ", "))
			}
			result, err := repo.Query(cmd.Context(), f)
			if err != nil {
				return fmt.Errorf("Query failed: %w", err)
			}
			return writeQueryResult(cmd, result, output)
		},
	}
	cmd.Flags().StringVar(&f.Type, "type", "", "Filter by type: "+strings.Join(ValidTypes, ", "))
	cmd.Flags().StringVar(&f.Tags, "tags", "", "Filter by tags (comma-separated, any match)")
	cmd.Flags().StringVar(&f.Name, "name", "", "Filter by name (contains match)")
	cmd.Flags().BoolVar(&processed, "processed", false, "Filter by processed status")
	cmd.Flags().BoolVar(&f.HasURL, "has-url", false, "Filter entries that have a URL")
	cmd.Flags().StringVar(&f.SortBy, "sort-by", "", "Property to sort by (default: created_time)")
	cmd.Flags().StringVar(&f.SortDir, "sort-direction", "", "Sort direction: ascending or descending (default: descending)")
	cmd.Flags().IntVar(&f.PageSize, "page-size", 0, "Number of results (1-100, default 50)")
	cmd.Flags().StringVar(&f.StartCursor, "start-cursor", "", "Pagination cursor from previous query")
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	return cmd
}

func newCreateCmd(repo Repository) *cobra.Command {
	var p CreateParams
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Save an idea, reference, or content to BlackholeDB",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !isValidType(p.Type) {
				return fmt.Errorf("Invalid type '%s'. Valid options: %s", p.Type, strings.Join(ValidTypes, ", "))
			}
			entry, err := repo.Create(cmd.Context(), p)
			if err != nil {
				return fmt.Errorf("Create failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created entry: **%s** (%s)\nID: %s\nURL: %s\n", entry.Name, entry.Type, entry.ID, entry.URL)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Name, "name", "", "Entry name")
	cmd.Flags().StringVar(&p.Type, "type", "", "Entry type: "+strings.Join(ValidTypes, ", "))
	cmd.Flags().StringVar(&p.Summary, "summary", "", "Brief description of the content")
	cmd.Flags().StringVar(&p.Tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&p.URL, "url", "", "URL of the resource")
	cmd.Flags().BoolVar(&p.Processed, "processed", false, "Whether already read/acted upon")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func writeQueryResult(cmd *cobra.Command, r *QueryResult, output string) error {
	switch output {
	case "json":
		payload := map[string]any{
			"results":     r.Entries,
			"has_more":    r.HasMore,
			"next_cursor": r.NextCursor,
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(payload)
	default:
		fmt.Fprint(cmd.OutOrStdout(), formatQueryResult(r))
		return nil
	}
}

func formatQueryResult(r *QueryResult) string {
	if len(r.Entries) == 0 {
		return "No entries found.\n"
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
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_\n", r.NextCursor)
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
