package projects_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/modron/internal/projects"
)

type mockRepo struct {
	queryResult   *projects.QueryResult
	queryErr      error
	createProject *projects.Project
	createErr     error
	lastFilter    projects.Filter
	lastCreate    projects.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f projects.Filter) (*projects.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, m.queryErr
}

func (m *mockRepo) Create(_ context.Context, p projects.CreateParams) (*projects.Project, error) {
	m.lastCreate = p
	return m.createProject, m.createErr
}

func runCmd(t *testing.T, repo projects.Repository, args ...string) (string, string, error) {
	t.Helper()
	cmd := projects.NewCommand(repo)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestProjectsQuery_InvalidStatus(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "query", "--status", "Bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid status") {
		t.Errorf("expected 'Invalid status' in error output, got: %s / %v", errOut, err)
	}
}

func TestProjectsQuery_InvalidTag(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "query", "--tag", "Nonsense")
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid tag") {
		t.Errorf("expected 'Invalid tag' in error output, got: %s / %v", errOut, err)
	}
}

func TestProjectsQuery_PassesFilter(t *testing.T) {
	mock := &mockRepo{
		queryResult: &projects.QueryResult{
			Projects: []projects.Project{{ID: "p1", Name: "Alpha", Status: "In Progress", Tag: "Dev"}},
		},
	}
	_, _, err := runCmd(t, mock, "query",
		"--status", "In Progress",
		"--tag", "Dev",
		"--has-launch-date",
		"--date-start-after", "2026-01-01",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastFilter.Status != "In Progress" || mock.lastFilter.Tag != "Dev" {
		t.Errorf("filter mismatch: %+v", mock.lastFilter)
	}
	if !mock.lastFilter.HasLaunchDate {
		t.Error("expected HasLaunchDate true")
	}
	if mock.lastFilter.DateStartAfter != "2026-01-01" {
		t.Errorf("expected date-start-after '2026-01-01', got '%s'", mock.lastFilter.DateStartAfter)
	}
}

func TestProjectsCreate_Success(t *testing.T) {
	mock := &mockRepo{
		createProject: &projects.Project{ID: "p-new", URL: "https://notion.so/p-new", Name: "New Proj"},
	}
	out, _, err := runCmd(t, mock, "create", "--name", "New Proj", "--status", "Planning", "--tag", "Dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastCreate.Name != "New Proj" {
		t.Errorf("expected name 'New Proj', got '%s'", mock.lastCreate.Name)
	}
	if !strings.Contains(out, "p-new") {
		t.Errorf("expected output to contain new ID, got: %s", out)
	}
}

func TestProjectsCreate_RequiresName(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create")
	if err == nil {
		t.Fatal("expected error when name flag missing")
	}
}
