package events

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

	if f.Name != "" {
		fb.TitleContains("Name", f.Name)
	}
	if f.DateAfter != "" {
		fb.DateOnOrAfter("Date", f.DateAfter)
	}
	if f.DateBefore != "" {
		fb.DateOnOrBefore("Date", f.DateBefore)
	}

	sortBy := f.SortBy
	if sortBy == "" {
		sortBy = "Date"
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
		Events:  make([]Event, len(resp.Results)),
		HasMore: resp.HasMore,
	}
	if resp.NextCursor != nil {
		result.NextCursor = *resp.NextCursor
	}

	for i, page := range resp.Results {
		result.Events[i] = pageToEvent(page)
	}
	return result, nil
}

func (r *NotionRepository) Create(ctx context.Context, p CreateParams) (*Event, error) {
	pb := notion.NewPropertyBuilder().
		Title("Name", p.Name)

	if p.DateEnd != "" {
		pb.DateRange("Date", p.DateStart, p.DateEnd)
	} else {
		pb.Date("Date", p.DateStart)
	}

	body := notion.CreatePageBody(DatabaseID, pb.Build())
	page, err := r.client.CreatePage(ctx, body)
	if err != nil {
		return nil, err
	}

	event := pageToEvent(*page)
	return &event, nil
}

func pageToEvent(page notion.Page) Event {
	e := Event{
		ID:  page.ID,
		URL: page.URL,
	}
	if raw, ok := page.Properties["Name"]; ok {
		e.Name = notion.ExtractTitle(raw)
	}
	if raw, ok := page.Properties["Date"]; ok {
		if d := notion.ExtractDate(raw); d != nil {
			e.DateStart = d.Start
			e.DateEnd = d.End
		}
	}
	return e
}
