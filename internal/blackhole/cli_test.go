package blackhole_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/modron/internal/blackhole"
)

type mockRepo struct {
	queryResult *blackhole.QueryResult
	queryErr    error
	createEntry *blackhole.Entry
	createErr   error
	lastFilter  blackhole.Filter
	lastCreate  blackhole.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f blackhole.Filter) (*blackhole.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, m.queryErr
}

func (m *mockRepo) Create(_ context.Context, p blackhole.CreateParams) (*blackhole.Entry, error) {
	m.lastCreate = p
	return m.createEntry, m.createErr
}

func runCmd(t *testing.T, repo blackhole.Repository, args ...string) (string, string, error) {
	t.Helper()
	cmd := blackhole.NewCommand(repo)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestBlackholeQuery_InvalidType(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "query", "--type", "Bogus")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid type") {
		t.Errorf("expected 'Invalid type' in error output, got: %s / %v", errOut, err)
	}
}

func TestBlackholeQuery_PassesFilterAndProcessedFlag(t *testing.T) {
	mock := &mockRepo{
		queryResult: &blackhole.QueryResult{
			Entries: []blackhole.Entry{{ID: "b1", Name: "Idea1", Type: "Idea"}},
		},
	}
	_, _, err := runCmd(t, mock, "query",
		"--type", "Idea",
		"--tags", "Dev,AI",
		"--processed=false",
		"--has-url",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastFilter.Type != "Idea" {
		t.Errorf("expected type 'Idea', got '%s'", mock.lastFilter.Type)
	}
	if mock.lastFilter.Tags != "Dev,AI" {
		t.Errorf("expected tags 'Dev,AI', got '%s'", mock.lastFilter.Tags)
	}
	if mock.lastFilter.Processed == nil || *mock.lastFilter.Processed != false {
		t.Errorf("expected Processed=&false, got %v", mock.lastFilter.Processed)
	}
	if !mock.lastFilter.HasURL {
		t.Error("expected HasURL true")
	}
}

func TestBlackholeQuery_ProcessedNilWhenNotSet(t *testing.T) {
	mock := &mockRepo{queryResult: &blackhole.QueryResult{}}
	_, _, err := runCmd(t, mock, "query", "--type", "Idea")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastFilter.Processed != nil {
		t.Errorf("expected Processed nil when --processed not provided, got %v", *mock.lastFilter.Processed)
	}
}

func TestBlackholeCreate_InvalidType(t *testing.T) {
	mock := &mockRepo{}
	_, errOut, err := runCmd(t, mock, "create", "--name", "X", "--type", "Whatever")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(errOut+err.Error(), "Invalid type") {
		t.Errorf("expected 'Invalid type' in error output, got: %s / %v", errOut, err)
	}
}

func TestBlackholeCreate_Success(t *testing.T) {
	mock := &mockRepo{
		createEntry: &blackhole.Entry{ID: "b-new", URL: "https://notion.so/b-new", Name: "Read me", Type: "Reference"},
	}
	out, _, err := runCmd(t, mock, "create",
		"--name", "Read me",
		"--type", "Reference",
		"--url", "https://example.com",
		"--tags", "Dev",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastCreate.Type != "Reference" || mock.lastCreate.Name != "Read me" {
		t.Errorf("create params mismatch: %+v", mock.lastCreate)
	}
	if !strings.Contains(out, "b-new") {
		t.Errorf("expected output to contain 'b-new', got: %s", out)
	}
}

func TestBlackholeCreate_RequiresNameAndType(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create", "--type", "Idea")
	if err == nil {
		t.Fatal("expected error when name flag missing")
	}
	_, _, err = runCmd(t, mock, "create", "--name", "X")
	if err == nil {
		t.Fatal("expected error when type flag missing")
	}
}
