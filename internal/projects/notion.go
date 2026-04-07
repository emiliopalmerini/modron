package projects

import (
	"context"

	"github.com/emiliopalmerini/notion-mcp/internal/shared/notion"
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
	if f.Tag != "" {
		fb.SelectEquals("Tag", f.Tag)
	}
	if f.Name != "" {
		fb.TitleContains("Project name", f.Name)
	}
	if f.DateStartAfter != "" {
		fb.DateOnOrAfter("Dates", f.DateStartAfter)
	}
	if f.DateStartBefore != "" {
		fb.DateOnOrBefore("Dates", f.DateStartBefore)
	}
	if f.HasLaunchDate {
		fb.DateIsNotEmpty("Launch date")
	}

	sortBy := f.SortBy
	if sortBy == "" {
		sortBy = "Dates"
	}
	sortDir := f.SortDir
	if sortDir == "" {
		sortDir = "descending"
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
		Projects: make([]Project, len(resp.Results)),
		HasMore:  resp.HasMore,
	}
	if resp.NextCursor != nil {
		result.NextCursor = *resp.NextCursor
	}

	for i, page := range resp.Results {
		result.Projects[i] = pageToProject(page)
	}
	return result, nil
}

func (r *NotionRepository) Create(ctx context.Context, p CreateParams) (*Project, error) {
	pb := notion.NewPropertyBuilder().
		Title("Project name", p.Name)

	status := p.Status
	if status == "" {
		status = "Planning"
	}
	pb.Status("Status", status)

	if p.Tag != "" {
		pb.Select("Tag", p.Tag)
	}
	if p.Summary != "" {
		pb.RichText("Summary", p.Summary)
	}
	if p.DatesStart != "" {
		if p.DatesEnd != "" {
			pb.DateRange("Dates", p.DatesStart, p.DatesEnd)
		} else {
			pb.Date("Dates", p.DatesStart)
		}
	}
	if p.LaunchDate != "" {
		pb.Date("Launch date", p.LaunchDate)
	}

	body := notion.CreatePageBody(DatabaseID, pb.Build())
	page, err := r.client.CreatePage(ctx, body)
	if err != nil {
		return nil, err
	}

	proj := pageToProject(*page)
	return &proj, nil
}

func pageToProject(page notion.Page) Project {
	p := Project{
		ID:  page.ID,
		URL: page.URL,
	}
	if raw, ok := page.Properties["Project name"]; ok {
		p.Name = notion.ExtractTitle(raw)
	}
	if raw, ok := page.Properties["Status"]; ok {
		p.Status = notion.ExtractStatus(raw)
	}
	if raw, ok := page.Properties["Tag"]; ok {
		p.Tag = notion.ExtractSelect(raw)
	}
	if raw, ok := page.Properties["Summary"]; ok {
		p.Summary = notion.ExtractRichText(raw)
	}
	if raw, ok := page.Properties["Dates"]; ok {
		if d := notion.ExtractDate(raw); d != nil {
			p.DatesStart = d.Start
			p.DatesEnd = d.End
		}
	}
	if raw, ok := page.Properties["Launch date"]; ok {
		if d := notion.ExtractDate(raw); d != nil {
			p.LaunchDate = d.Start
		}
	}
	if raw, ok := page.Properties["Tasks"]; ok {
		p.TaskIDs = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Is Blocking"]; ok {
		p.IsBlocking = notion.ExtractRelation(raw)
	}
	if raw, ok := page.Properties["Blocked By"]; ok {
		p.BlockedBy = notion.ExtractRelation(raw)
	}
	return p
}
