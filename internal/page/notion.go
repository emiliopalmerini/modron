package page

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emiliopalmerini/notion-mcp/internal/shared/notion"
)

type NotionRepository struct {
	client *notion.Client
}

func NewNotionRepository(client *notion.Client) *NotionRepository {
	return &NotionRepository{client: client}
}

func (r *NotionRepository) Get(ctx context.Context, pageID string) (*Page, error) {
	raw, err := r.client.GetPage(ctx, pageID)
	if err != nil {
		return nil, err
	}
	return rawPageToPage(raw), nil
}

func (r *NotionRepository) Update(ctx context.Context, params UpdateParams) (*Page, error) {
	body := map[string]any{
		"properties": params.PropertyUpdates,
	}
	raw, err := r.client.UpdatePage(ctx, params.PageID, body)
	if err != nil {
		return nil, err
	}
	return rawPageToPage(raw), nil
}

func rawPageToPage(raw *notion.Page) *Page {
	p := &Page{
		ID:             raw.ID,
		URL:            raw.URL,
		CreatedTime:    raw.CreatedTime,
		LastEditedTime: raw.LastEditedTime,
	}

	for name, propRaw := range raw.Properties {
		prop := Property{Name: name}
		prop.Type = notion.GetPropertyType(propRaw)
		prop.Value = extractValue(prop.Type, propRaw)
		p.Properties = append(p.Properties, prop)
	}
	return p
}

func extractValue(propType string, raw json.RawMessage) string {
	switch propType {
	case "title":
		return notion.ExtractTitle(raw)
	case "rich_text":
		return notion.ExtractRichText(raw)
	case "select":
		return notion.ExtractSelect(raw)
	case "status":
		return notion.ExtractStatus(raw)
	case "multi_select":
		return strings.Join(notion.ExtractMultiSelect(raw), ", ")
	case "date":
		if d := notion.ExtractDate(raw); d != nil {
			if d.End != "" {
				return fmt.Sprintf("%s to %s", d.Start, d.End)
			}
			return d.Start
		}
		return ""
	case "checkbox":
		if notion.ExtractCheckbox(raw) {
			return "true"
		}
		return "false"
	case "url":
		return notion.ExtractURL(raw)
	case "number":
		if n := notion.ExtractNumber(raw); n != nil {
			return fmt.Sprintf("%g", *n)
		}
		return ""
	case "relation":
		ids := notion.ExtractRelation(raw)
		return strings.Join(ids, ", ")
	default:
		return ""
	}
}
