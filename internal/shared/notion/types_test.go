package notion_test

import (
	"encoding/json"
	"testing"

	"github.com/emiliopalmerini/modron/internal/shared/notion"
)

func TestExtractTitle(t *testing.T) {
	raw := json.RawMessage(`{"type":"title","title":[{"plain_text":"My Task"}]}`)
	got := notion.ExtractTitle(raw)
	if got != "My Task" {
		t.Errorf("expected 'My Task', got '%s'", got)
	}
}

func TestExtractStatus(t *testing.T) {
	raw := json.RawMessage(`{"type":"status","status":{"name":"In Progress"}}`)
	got := notion.ExtractStatus(raw)
	if got != "In Progress" {
		t.Errorf("expected 'In Progress', got '%s'", got)
	}
}

func TestExtractStatus_Nil(t *testing.T) {
	raw := json.RawMessage(`{"type":"status","status":null}`)
	got := notion.ExtractStatus(raw)
	if got != "" {
		t.Errorf("expected empty string, got '%s'", got)
	}
}

func TestExtractSelect(t *testing.T) {
	raw := json.RawMessage(`{"type":"select","select":{"name":"Dev"}}`)
	got := notion.ExtractSelect(raw)
	if got != "Dev" {
		t.Errorf("expected 'Dev', got '%s'", got)
	}
}

func TestExtractMultiSelect(t *testing.T) {
	raw := json.RawMessage(`{"type":"multi_select","multi_select":[{"name":"AI"},{"name":"Dev"}]}`)
	got := notion.ExtractMultiSelect(raw)
	if len(got) != 2 || got[0] != "AI" || got[1] != "Dev" {
		t.Errorf("expected [AI Dev], got %v", got)
	}
}

func TestExtractDate(t *testing.T) {
	raw := json.RawMessage(`{"type":"date","date":{"start":"2026-04-07","end":"2026-04-10"}}`)
	got := notion.ExtractDate(raw)
	if got == nil {
		t.Fatal("expected date, got nil")
	}
	if got.Start != "2026-04-07" {
		t.Errorf("expected start 2026-04-07, got %s", got.Start)
	}
	if got.End != "2026-04-10" {
		t.Errorf("expected end 2026-04-10, got %s", got.End)
	}
}

func TestExtractDate_Nil(t *testing.T) {
	raw := json.RawMessage(`{"type":"date","date":null}`)
	got := notion.ExtractDate(raw)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestExtractCheckbox(t *testing.T) {
	raw := json.RawMessage(`{"type":"checkbox","checkbox":true}`)
	got := notion.ExtractCheckbox(raw)
	if !got {
		t.Error("expected true")
	}
}

func TestExtractURL(t *testing.T) {
	raw := json.RawMessage(`{"type":"url","url":"https://example.com"}`)
	got := notion.ExtractURL(raw)
	if got != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", got)
	}
}

func TestExtractRelation(t *testing.T) {
	raw := json.RawMessage(`{"type":"relation","relation":[{"id":"abc-123"},{"id":"def-456"}]}`)
	got := notion.ExtractRelation(raw)
	if len(got) != 2 || got[0] != "abc-123" {
		t.Errorf("expected [abc-123 def-456], got %v", got)
	}
}

func TestExtractRichText(t *testing.T) {
	raw := json.RawMessage(`{"type":"rich_text","rich_text":[{"plain_text":"Hello "},{"plain_text":"world"}]}`)
	got := notion.ExtractRichText(raw)
	if got != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", got)
	}
}
