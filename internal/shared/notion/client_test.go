package notion_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emiliopalmerini/modron/internal/shared/notion"
)

func TestQueryDatabase(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/databases/test-db-id/query" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Notion-Version") != notion.APIVersion {
			t.Errorf("unexpected version header: %s", r.Header.Get("Notion-Version"))
		}

		resp := notion.QueryResponse{
			Results: []notion.Page{
				{ID: "page-1", URL: "https://notion.so/page-1"},
				{ID: "page-2", URL: "https://notion.so/page-2"},
			},
			HasMore: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := notion.NewClient("test-token").WithBaseURL(srv.URL + "/v1")
	result, err := client.QueryDatabase(context.Background(), "test-db-id", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].ID != "page-1" {
		t.Errorf("expected page-1, got %s", result.Results[0].ID)
	}
}

func TestCreatePage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/pages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		page := notion.Page{ID: "new-page", URL: "https://notion.so/new-page"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := notion.NewClient("test-token").WithBaseURL(srv.URL + "/v1")
	page, err := client.CreatePage(context.Background(), map[string]any{
		"parent":     map[string]any{"database_id": "db-id"},
		"properties": map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "new-page" {
		t.Errorf("expected new-page, got %s", page.ID)
	}
}

func TestAPIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    "validation_error",
			"message": "invalid property",
		})
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := notion.NewClient("test-token").WithBaseURL(srv.URL + "/v1")
	_, err := client.QueryDatabase(context.Background(), "db-id", map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*notion.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "validation_error" {
		t.Errorf("expected validation_error, got %s", apiErr.Code)
	}
}
