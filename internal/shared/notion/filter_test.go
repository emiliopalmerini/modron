package notion_test

import (
	"encoding/json"
	"testing"

	"github.com/emiliopalmerini/modron/internal/shared/notion"
)

func TestFilterBuilder_SingleCondition(t *testing.T) {
	f := notion.NewFilter().StatusEquals("Status", "In Progress").Build()

	data, _ := json.Marshal(f)
	expected := `{"property":"Status","status":{"equals":"In Progress"}}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestFilterBuilder_MultipleConditions(t *testing.T) {
	f := notion.NewFilter().
		StatusEquals("Status", "In Progress").
		DateOnOrBefore("Due", "2026-04-07").
		Build()

	data, _ := json.Marshal(f)
	var result map[string]any
	json.Unmarshal(data, &result)

	and, ok := result["and"].([]any)
	if !ok {
		t.Fatalf("expected 'and' array, got %v", result)
	}
	if len(and) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(and))
	}
}

func TestFilterBuilder_Empty(t *testing.T) {
	f := notion.NewFilter().Build()
	if f != nil {
		t.Errorf("expected nil filter, got %v", f)
	}
}

func TestBuildQueryBody(t *testing.T) {
	filter := notion.NewFilter().StatusEquals("Status", "Done").Build()
	sorts := []notion.Sort{{Property: "Due", Direction: "ascending"}}
	body := notion.BuildQueryBody(filter, sorts, 50, "cursor-abc")

	if body["page_size"] != 50 {
		t.Errorf("expected page_size 50, got %v", body["page_size"])
	}
	if body["start_cursor"] != "cursor-abc" {
		t.Errorf("expected cursor-abc, got %v", body["start_cursor"])
	}
	if body["filter"] == nil {
		t.Error("expected filter to be set")
	}
	if body["sorts"] == nil {
		t.Error("expected sorts to be set")
	}
}
