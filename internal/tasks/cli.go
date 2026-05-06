package tasks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(repo Repository) *cobra.Command {
	root := &cobra.Command{
		Use:   "tasks",
		Short: "Query and create tasks",
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
		Short: "Query tasks from Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if f.Status != "" && !isValidStatus(f.Status) {
				return fmt.Errorf("Invalid status '%s'. Valid options: %s", f.Status, strings.Join(ValidStatuses, ", "))
			}
			result, err := repo.Query(cmd.Context(), f)
			if err != nil {
				return fmt.Errorf("Query failed: %w", err)
			}
			return writeQueryResult(cmd, result, output)
		},
	}
	cmd.Flags().StringVar(&f.Status, "status", "", "Filter by status: "+strings.Join(ValidStatuses, ", "))
	cmd.Flags().StringVar(&f.Name, "name", "", "Filter by name (contains match)")
	cmd.Flags().StringVar(&f.DueBefore, "due-before", "", "Filter tasks due on or before this date (ISO-8601)")
	cmd.Flags().StringVar(&f.DueAfter, "due-after", "", "Filter tasks due on or after this date (ISO-8601)")
	cmd.Flags().StringVar(&f.ProjectID, "project-id", "", "Filter by project page ID")
	cmd.Flags().StringVar(&f.SortBy, "sort-by", "", "Property to sort by (default: Due)")
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
		Short: "Create a new task in Notion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if p.Status != "" && !isValidStatus(p.Status) {
				return fmt.Errorf("Invalid status '%s'. Valid options: %s", p.Status, strings.Join(ValidStatuses, ", "))
			}
			task, err := repo.Create(cmd.Context(), p)
			if err != nil {
				return fmt.Errorf("Create failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created task: **%s**\nID: %s\nURL: %s\n", task.Name, task.ID, task.URL)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Name, "name", "", "Task name")
	cmd.Flags().StringVar(&p.Due, "due", "", "Due date (ISO-8601)")
	cmd.Flags().StringVar(&p.Status, "status", "", "Initial status (default: Not Started)")
	cmd.Flags().StringVar(&p.ProjectID, "project-id", "", "Parent project page ID")
	cmd.Flags().StringVar(&p.ParentTaskID, "parent-task-id", "", "Parent task page ID")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("due")
	return cmd
}

func writeQueryResult(cmd *cobra.Command, r *QueryResult, output string) error {
	switch output {
	case "json":
		payload := map[string]any{
			"results":     r.Tasks,
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
	if len(r.Tasks) == 0 {
		return "No tasks found.\n"
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
		fmt.Fprintf(&sb, "\n_More results available. Use start_cursor: `%s`_\n", r.NextCursor)
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
