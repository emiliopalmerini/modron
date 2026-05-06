package tasks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emiliopalmerini/modron/internal/shared/notion"
	"github.com/emiliopalmerini/modron/internal/tasks"
)

func newTestPage(id, name, status, due string) notion.Page {
	props := map[string]json.RawMessage{
		"Task name": json.RawMessage(`{"type":"title","title":[{"plain_text":"` + name + `"}]}`),
		"Status":    json.RawMessage(`{"type":"status","status":{"name":"` + status + `"}}`),
	}
	if due != "" {
		props["Due"] = json.RawMessage(`{"type":"date","date":{"start":"` + due + `"}}`)
	}
	return notion.Page{
		ID:         id,
		URL:        "https://notion.so/" + id,
		Properties: props,
	}
}

func TestNotionRepository_Query(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		// Verify filter is sent
		filter, hasFilter := body["filter"]
		if !hasFilter {
			t.Error("expected filter in request body")
		}
		filterMap := filter.(map[string]any)
		if filterMap["property"] != "Status" {
			t.Errorf("expected filter on Status, got %v", filterMap["property"])
		}

		resp := notion.QueryResponse{
			Results: []notion.Page{
				newTestPage("t1", "Fix bug", "In progress", "2026-04-07"),
				newTestPage("t2", "Write docs", "In progress", "2026-04-10"),
			},
			HasMore: false,
		}
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := notion.NewClient("token").WithBaseURL(srv.URL)
	repo := tasks.NewNotionRepository(client)

	result, err := repo.Query(context.Background(), tasks.Filter{Status: "In progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}
	if result.Tasks[0].Name != "Fix bug" {
		t.Errorf("expected 'Fix bug', got '%s'", result.Tasks[0].Name)
	}
	if result.Tasks[0].Status != "In progress" {
		t.Errorf("expected 'In progress', got '%s'", result.Tasks[0].Status)
	}
	if result.Tasks[0].Due != "2026-04-07" {
		t.Errorf("expected '2026-04-07', got '%s'", result.Tasks[0].Due)
	}
}

func TestNotionRepository_Create(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		parent := body["parent"].(map[string]any)
		if parent["database_id"] != tasks.DatabaseID {
			t.Errorf("expected database_id %s, got %v", tasks.DatabaseID, parent["database_id"])
		}

		page := newTestPage("new-task", "New task", "Not Started", "2026-04-15")
		json.NewEncoder(w).Encode(page)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := notion.NewClient("token").WithBaseURL(srv.URL)
	repo := tasks.NewNotionRepository(client)

	task, err := repo.Create(context.Background(), tasks.CreateParams{
		Name: "New task",
		Due:  "2026-04-15",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Name != "New task" {
		t.Errorf("expected 'New task', got '%s'", task.Name)
	}
	if task.ID != "new-task" {
		t.Errorf("expected 'new-task', got '%s'", task.ID)
	}
}
