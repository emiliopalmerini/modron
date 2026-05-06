package page_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/modron/internal/page"
)

type mockRepo struct {
	getPage    *page.Page
	getErr     error
	updatePage *page.Page
	updateErr  error
	lastGet    string
	lastUpdate page.UpdateParams
}

func (m *mockRepo) Get(_ context.Context, pageID string) (*page.Page, error) {
	m.lastGet = pageID
	return m.getPage, m.getErr
}

func (m *mockRepo) Update(_ context.Context, p page.UpdateParams) (*page.Page, error) {
	m.lastUpdate = p
	return m.updatePage, m.updateErr
}

func runCmd(t *testing.T, repo page.Repository, args ...string) (string, string, error) {
	t.Helper()
	cmd := page.NewCommand(repo)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestPageGet_RequiresPositionalArg(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "get")
	if err == nil {
		t.Fatal("expected error when page_id positional missing")
	}
}

func TestPageGet_Success(t *testing.T) {
	mock := &mockRepo{
		getPage: &page.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
			Properties: []page.Property{
				{Name: "Status", Type: "status", Value: "Done"},
			},
		},
	}
	out, _, err := runCmd(t, mock, "get", "page-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastGet != "page-123" {
		t.Errorf("expected pageID 'page-123', got '%s'", mock.lastGet)
	}
	if !strings.Contains(out, "page-123") || !strings.Contains(out, "Status") {
		t.Errorf("expected output to contain page id and Status property, got: %s", out)
	}
}

func TestPageUpdate_RequiresPositionalAndProperties(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "update")
	if err == nil {
		t.Fatal("expected error when positional page_id missing")
	}
	_, _, err = runCmd(t, mock, "update", "page-123")
	if err == nil {
		t.Fatal("expected error when --properties flag missing")
	}
}

func TestPageUpdate_RejectsInvalidJSON(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "update", "page-123", "--properties", "{not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid properties JSON") {
		t.Errorf("expected JSON error message, got: %s / %v", errOut, err)
	}
}

func TestPageUpdate_Success(t *testing.T) {
	mock := &mockRepo{
		updatePage: &page.Page{ID: "page-123", URL: "https://notion.so/page-123"},
	}
	out, _, err := runCmd(t, mock, "update", "page-123",
		"--properties", `{"Status":{"status":{"name":"Done"}}}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastUpdate.PageID != "page-123" {
		t.Errorf("expected pageID 'page-123', got '%s'", mock.lastUpdate.PageID)
	}
	statusProp, ok := mock.lastUpdate.PropertyUpdates["Status"]
	if !ok {
		t.Fatalf("expected Status in PropertyUpdates, got %+v", mock.lastUpdate.PropertyUpdates)
	}
	_ = statusProp
	if !strings.Contains(out, "page-123") {
		t.Errorf("expected output to contain page id, got: %s", out)
	}
}
