package events_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/modron/internal/events"
)

type mockRepo struct {
	queryResult *events.QueryResult
	queryErr    error
	createEvent *events.Event
	createErr   error
	lastFilter  events.Filter
	lastCreate  events.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f events.Filter) (*events.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, m.queryErr
}

func (m *mockRepo) Create(_ context.Context, p events.CreateParams) (*events.Event, error) {
	m.lastCreate = p
	return m.createEvent, m.createErr
}

func runCmd(t *testing.T, repo events.Repository, args ...string) (string, string, error) {
	t.Helper()
	cmd := events.NewCommand(repo)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestEventsQuery_PassesFilter(t *testing.T) {
	mock := &mockRepo{
		queryResult: &events.QueryResult{
			Events: []events.Event{{ID: "e1", Name: "Demo", DateStart: "2026-05-10"}},
		},
	}
	out, _, err := runCmd(t, mock, "query", "--name", "Demo", "--date-after", "2026-05-01", "--page-size", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastFilter.Name != "Demo" || mock.lastFilter.DateAfter != "2026-05-01" || mock.lastFilter.PageSize != 10 {
		t.Errorf("filter mismatch: %+v", mock.lastFilter)
	}
	if !strings.Contains(out, "Demo") {
		t.Errorf("expected output to contain 'Demo', got: %s", out)
	}
}

func TestEventsCreate_Success(t *testing.T) {
	mock := &mockRepo{
		createEvent: &events.Event{ID: "e-new", URL: "https://notion.so/e-new", Name: "Launch"},
	}
	out, _, err := runCmd(t, mock, "create", "--name", "Launch", "--date-start", "2026-06-01", "--is-datetime")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastCreate.Name != "Launch" || mock.lastCreate.DateStart != "2026-06-01" {
		t.Errorf("create params mismatch: %+v", mock.lastCreate)
	}
	if !mock.lastCreate.IsDatetime {
		t.Error("expected IsDatetime true")
	}
	if !strings.Contains(out, "e-new") {
		t.Errorf("expected output to contain 'e-new', got: %s", out)
	}
}

func TestEventsCreate_RequiresName(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create", "--date-start", "2026-06-01")
	if err == nil {
		t.Fatal("expected error when name flag missing")
	}
}

func TestEventsCreate_RequiresDateStart(t *testing.T) {
	mock := &mockRepo{}
	_, _, err := runCmd(t, mock, "create", "--name", "X")
	if err == nil {
		t.Fatal("expected error when date-start flag missing")
	}
}
