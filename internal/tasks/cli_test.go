package tasks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/emiliopalmerini/modron/internal/tasks"
)

type mockRepo struct {
	queryResult *tasks.QueryResult
	queryErr    error
	createTask  *tasks.Task
	createErr   error
	lastFilter  tasks.Filter
	lastCreate  tasks.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f tasks.Filter) (*tasks.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, m.queryErr
}

func (m *mockRepo) Create(_ context.Context, p tasks.CreateParams) (*tasks.Task, error) {
	m.lastCreate = p
	return m.createTask, m.createErr
}

func runCmd(t *testing.T, repo tasks.Repository, args ...string) (string, string, error) {
	t.Helper()
	cmd := tasks.NewCommand(repo)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestTasksQuery_InvalidStatus(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "query", "--status", "BadStatus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid status") {
		t.Errorf("expected 'Invalid status' in error output, got: %s / %v", errOut, err)
	}
}

func TestTasksQuery_PassesFilter(t *testing.T) {
	mock := &mockRepo{
		queryResult: &tasks.QueryResult{
			Tasks: []tasks.Task{{ID: "t1", Name: "Test", Status: "In progress", Due: "2026-04-07"}},
		},
	}
	out, _, err := runCmd(t, mock, "query", "--status", "In progress", "--due-after", "2026-04-01", "--page-size", "25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastFilter.Status != "In progress" {
		t.Errorf("expected status 'In progress', got '%s'", mock.lastFilter.Status)
	}
	if mock.lastFilter.DueAfter != "2026-04-01" {
		t.Errorf("expected due-after '2026-04-01', got '%s'", mock.lastFilter.DueAfter)
	}
	if mock.lastFilter.PageSize != 25 {
		t.Errorf("expected page-size 25, got %d", mock.lastFilter.PageSize)
	}
	if !strings.Contains(out, "Test") {
		t.Errorf("expected output to contain task name 'Test', got: %s", out)
	}
}

func TestTasksQuery_JSONOutput(t *testing.T) {
	mock := &mockRepo{
		queryResult: &tasks.QueryResult{
			Tasks:      []tasks.Task{{ID: "t1", Name: "Test", Status: "Done"}},
			HasMore:    true,
			NextCursor: "cursor-xyz",
		},
	}
	out, _, err := runCmd(t, mock, "query", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\noutput: %s", err, out)
	}
	if got["has_more"] != true {
		t.Errorf("expected has_more=true, got %v", got["has_more"])
	}
	if got["next_cursor"] != "cursor-xyz" {
		t.Errorf("expected next_cursor='cursor-xyz', got %v", got["next_cursor"])
	}
}

func TestTasksCreate_RequiresName(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create", "--due", "2026-04-15")
	if err == nil {
		t.Fatal("expected error when name flag missing")
	}
}

func TestTasksCreate_RequiresDue(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create", "--name", "X")
	if err == nil {
		t.Fatal("expected error when due flag missing")
	}
}

func TestTasksCreate_Success(t *testing.T) {
	mock := &mockRepo{
		createTask: &tasks.Task{ID: "new-id", URL: "https://notion.so/new-id", Name: "My task"},
	}
	out, _, err := runCmd(t, mock, "create", "--name", "My task", "--due", "2026-04-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastCreate.Name != "My task" {
		t.Errorf("expected name 'My task', got '%s'", mock.lastCreate.Name)
	}
	if mock.lastCreate.Due != "2026-04-15" {
		t.Errorf("expected due '2026-04-15', got '%s'", mock.lastCreate.Due)
	}
	if !strings.Contains(out, "new-id") {
		t.Errorf("expected output to contain page ID, got: %s", out)
	}
}

func TestTasksCreate_InvalidStatus(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "create", "--name", "X", "--due", "2026-04-15", "--status", "Bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid status") {
		t.Errorf("expected 'Invalid status' in error output, got: %s / %v", errOut, err)
	}
}
