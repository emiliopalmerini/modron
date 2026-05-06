package projects

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(repo Repository) *cobra.Command {
	root := &cobra.Command{
		Use:   "projects",
		Short: "Query and create projects",
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
		Short: "Query projects from Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if f.Status != "" && !isValid(f.Status, ValidStatuses) {
				return fmt.Errorf("Invalid status '%s'. Valid options: %s", f.Status, strings.Join(ValidStatuses, ", "))
			}
			if f.Tag != "" && !isValid(f.Tag, ValidTags) {
				return fmt.Errorf("Invalid tag '%s'. Valid options: %s", f.Tag, strings.Join(ValidTags, ", "))
			}
			result, err := repo.Query(cmd.Context(), f)
			if err != nil {
				return fmt.Errorf("Query failed: %w", err)
			}
			return writeQueryResult(cmd, result, output)
		},
	}
	cmd.Flags().StringVar(&f.Status, "status", "", "Filter by status: "+strings.Join(ValidStatuses, ", "))
	cmd.Flags().StringVar(&f.Tag, "tag", "", "Filter by tag: "+strings.Join(ValidTags, ", "))
	cmd.Flags().StringVar(&f.Name, "name", "", "Filter by name (contains match)")
	cmd.Flags().StringVar(&f.DateStartAfter, "date-start-after", "", "Filter projects starting on or after this date (ISO-8601)")
	cmd.Flags().StringVar(&f.DateStartBefore, "date-start-before", "", "Filter projects starting on or before this date (ISO-8601)")
	cmd.Flags().BoolVar(&f.HasLaunchDate, "has-launch-date", false, "Filter projects that have a launch date set")
	cmd.Flags().StringVar(&f.SortBy, "sort-by", "", "Property to sort by (default: Dates)")
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
		Short: "Create a new project in Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if p.Status != "" && !isValid(p.Status, ValidStatuses) {
				return fmt.Errorf("Invalid status '%s'. Valid options: %s", p.Status, strings.Join(ValidStatuses, ", "))
			}
			if p.Tag != "" && !isValid(p.Tag, ValidTags) {
				return fmt.Errorf("Invalid tag '%s'. Valid options: %s", p.Tag, strings.Join(ValidTags, ", "))
			}
			project, err := repo.Create(cmd.Context(), p)
			if err != nil {
				return fmt.Errorf("Create failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created project: **%s**\nID: %s\nURL: %s\n", project.Name, project.ID, project.URL)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Name, "name", "", "Project name")
	cmd.Flags().StringVar(&p.Status, "status", "", "Initial status (default: Planning)")
	cmd.Flags().StringVar(&p.Tag, "tag", "", "Project tag: "+strings.Join(ValidTags, ", "))
	cmd.Flags().StringVar(&p.Summary, "summary", "", "Brief project description")
	cmd.Flags().StringVar(&p.DatesStart, "dates-start", "", "Project start date (ISO-8601)")
	cmd.Flags().StringVar(&p.DatesEnd, "dates-end", "", "Project end date (ISO-8601)")
	cmd.Flags().StringVar(&p.LaunchDate, "launch-date", "", "Target launch/publish date (ISO-8601)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func writeQueryResult(cmd *cobra.Command, r *QueryResult, output string) error {
	switch output {
	case "json":
		payload := map[string]any{
			"results":     r.Projects,
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
	if len(r.Projects) == 0 {
		return "No projects found.\n"
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
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_\n", r.NextCursor)
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
