package blackhole

import (
	"context"
	"strings"

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

	if f.Type != "" {
		fb.SelectEquals("Type", f.Type)
	}
	if f.Name != "" {
		fb.TitleContains("Name", f.Name)
	}
	if f.Tags != "" {
		for _, tag := range strings.Split(f.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				fb.MultiSelectContains("Tags", tag)
			}
		}
	}
	if f.Processed != nil {
		fb.CheckboxEquals("Processed", *f.Processed)
	}
	if f.HasURL {
		fb.URLIsNotEmpty("userDefined:URL")
	}

	sortDir := f.SortDir
	if sortDir == "" {
		sortDir = "descending"
	}

	var sorts []notion.Sort
	if f.SortBy != "" {
		sorts = []notion.Sort{{Property: f.SortBy, Direction: sortDir}}
	} else {
		sorts = []notion.Sort{{Timestamp: "created_time", Direction: sortDir}}
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	body := notion.BuildQueryBody(fb.Build(), sorts, pageSize, f.StartCursor)

	resp, err := r.client.QueryDatabase(ctx, DatabaseID, body)
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Entries: make([]Entry, len(resp.Results)),
		HasMore: resp.HasMore,
	}
	if resp.NextCursor != nil {
		result.NextCursor = *resp.NextCursor
	}

	for i, page := range resp.Results {
		result.Entries[i] = pageToEntry(page)
	}
	return result, nil
}

func (r *NotionRepository) Create(ctx context.Context, p CreateParams) (*Entry, error) {
	pb := notion.NewPropertyBuilder().
		Title("Name", p.Name).
		Select("Type", p.Type)

	if p.Summary != "" {
		pb.RichText("Summary", p.Summary)
	}
	if p.Tags != "" {
		tags := strings.Split(p.Tags, ",")
		trimmed := make([]string, 0, len(tags))
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				trimmed = append(trimmed, t)
			}
		}
		if len(trimmed) > 0 {
			pb.MultiSelect("Tags", trimmed)
		}
	}
	if p.URL != "" {
		pb.URL("userDefined:URL", p.URL)
	}
	if p.Processed {
		pb.Checkbox("Processed", true)
	}

	body := notion.CreatePageBody(DatabaseID, pb.Build())
	page, err := r.client.CreatePage(ctx, body)
	if err != nil {
		return nil, err
	}

	entry := pageToEntry(*page)
	return &entry, nil
}

func pageToEntry(page notion.Page) Entry {
	e := Entry{
		ID:  page.ID,
		URL: page.URL,
	}
	if raw, ok := page.Properties["Name"]; ok {
		e.Name = notion.ExtractTitle(raw)
	}
	if raw, ok := page.Properties["Type"]; ok {
		e.Type = notion.ExtractSelect(raw)
	}
	if raw, ok := page.Properties["Tags"]; ok {
		e.Tags = notion.ExtractMultiSelect(raw)
	}
	if raw, ok := page.Properties["Summary"]; ok {
		e.Summary = notion.ExtractRichText(raw)
	}
	if raw, ok := page.Properties["Processed"]; ok {
		e.Processed = notion.ExtractCheckbox(raw)
	}
	if raw, ok := page.Properties["userDefined:URL"]; ok {
		e.EntryURL = notion.ExtractURL(raw)
	}
	return e
}
