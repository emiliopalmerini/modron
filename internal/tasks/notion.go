package tasks

import (
	"context"

	"github.com/emiliopalmerini/modron/internal/shared/notion"
)

type NotionRepository struct {
	client *notion.Client
}

func NewNotionRepository(client *notion.Client) *NotionRepository {
	return &NotionRepository{client: client}
}

func (r *NotionRepository) Query(ctx context.Context, f Filter) (*QueryResult, error) {
	fb := notion.NewFilter()

	if f.Status != "" {
		fb.StatusEquals("Status", f.Status)
	}
	if f.Name != "" {
		fb.TitleContains("Task name", f.Name)
	}
	if f.DueBefore != "" {
		fb.DateOnOrBefore("Due", f.DueBefore)
	}
	if f.DueAfter != "" {
		fb.DateOnOrAfter("Due", f.DueAfter)
	}
	if f.ProjectID != "" {
		fb.RelationContains("Project", f.ProjectID)
	}

	sortBy := f.SortBy
	if sortBy == "" {
		sortBy = "Due"
	}
	sortDir := f.SortDir
	if sortDir == "" {
		sortDir = "ascending"
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	body := notion.BuildQueryBody(
		fb.Build(),
		[]notion.Sort{{Property: sortBy, Direction: sortDir}},
		pageSize,
		f.StartCursor,
	)

	resp, err := r.client.QueryDatabase(ctx, DatabaseID, body)
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Tasks:   make([]Task, len(resp.Results)),
		HasMore: resp.HasMore,
	}
	if resp.NextCursor != nil {
		result.NextCursor = *resp.NextCursor
	}

	for i, page := range resp.Results {
		result.Tasks[i] = pageToTask(page)
	}
	return result, nil
}

func (r *NotionRepository) Create(ctx context.Context, p CreateParams) (*Task, error) {
	pb := notion.NewPropertyBuilder().
		Title("Task name", p.Name).
		Date("Due", p.Due)

	status := p.Status
	if status == "" {
		status = "Not Started"
	}
	pb.Status("Status", status)

	if p.ProjectID != "" {
		pb.Relation("Project", []string{p.ProjectID})
	}
	if p.ParentTaskID != "" {
		pb.Relation("Parent-task", []string{p.ParentTaskID})
	}

	body := notion.CreatePageBody(DatabaseID, pb.Build())
	page, err := r.client.CreatePage(ctx, body)
	if err != nil {
		return nil, err
	}

	task := pageToTask(*page)
	return &task, nil
}

func pageToTask(page notion.Page) Task {
	t := Task{
		ID:  page.ID,
		URL: page.URL,
	}
	if raw, ok := page.Properties["Task name"]; ok {
		t.Name = notion.ExtractTitle(raw)
	}
	if raw, ok := page.Properties["Status"]; ok {
		t.Status = notion.ExtractStatus(raw)
	}
	if raw, ok := page.Properties["Due"]; ok {
		if d := notion.ExtractDate(raw); d != nil {
			t.Due = d.Start
		}
	}
	if raw, ok := page.Properties["Project"]; ok {
		t.ProjectIDs = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Parent-task"]; ok {
		t.ParentTask = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Sub-tasks"]; ok {
		t.SubTasks = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Is Blocking"]; ok {
		t.IsBlocking = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Blocked By"]; ok {
		t.BlockedBy = notion.ExtractRelation(raw)
	}
	return t
}
