package events

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(repo Repository) *cobra.Command {
	root := &cobra.Command{
		Use:   "events",
		Short: "Query and create events",
	}
	root.AddCommand(newQueryCmd(repo), newCreateCmd(repo))
	return root
}

func newQueryCmd(repo Repository) *cobra.Command {
	var (
		f      Filter
		output string
	)
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query events from Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := repo.Query(cmd.Context(), f)
			if err != nil {
				return fmt.Errorf("Query failed: %w", err)
			}
			return writeQueryResult(cmd, result, output)
		},
	}
	cmd.Flags().StringVar(&f.Name, "name", "", "Filter by name (contains match)")
	cmd.Flags().StringVar(&f.DateAfter, "date-after", "", "Filter events on or after this date (ISO-8601)")
	cmd.Flags().StringVar(&f.DateBefore, "date-before", "", "Filter events on or before this date (ISO-8601)")
	cmd.Flags().StringVar(&f.SortBy, "sort-by", "", "Property to sort by (default: Date)")
	cmd.Flags().StringVar(&f.SortDir, "sort-direction", "", "Sort direction: ascending or descending (default: ascending)")
	cmd.Flags().IntVar(&f.PageSize, "page-size", 0, "Number of results (1-100, default 50)")
	cmd.Flags().StringVar(&f.StartCursor, "start-cursor", "", "Pagination cursor from previous query")
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	return cmd
}

func newCreateCmd(repo Repository) *cobra.Command {
	var p CreateParams
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new event in Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			event, err := repo.Create(cmd.Context(), p)
			if err != nil {
				return fmt.Errorf("Create failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created event: **%s**\nID: %s\nURL: %s\n", event.Name, event.ID, event.URL)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Name, "name", "", "Event name")
	cmd.Flags().StringVar(&p.DateStart, "date-start", "", "Event start date/datetime (ISO-8601)")
	cmd.Flags().StringVar(&p.DateEnd, "date-end", "", "Event end date/datetime (ISO-8601)")
	cmd.Flags().BoolVar(&p.IsDatetime, "is-datetime", false, "Whether the date includes time")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("date-start")
	return cmd
}

func writeQueryResult(cmd *cobra.Command, r *QueryResult, output string) error {
	switch output {
	case "json":
		payload := map[string]any{
			"results":     r.Events,
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
	if len(r.Events) == 0 {
		return "No events found.\n"
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
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_\n", r.NextCursor)
	}
	return sb.String()
}
