package tasks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/notion-mcp/internal/tasks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

func callTool(s *server.MCPServer, name string, args map[string]any) (*mcp.CallToolResult, error) {
	argsJSON, _ := json.Marshal(args)
	reqJSON, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": json.RawMessage(argsJSON),
		},
	})

	resp := s.HandleMessage(context.Background(), reqJSON)
	respJSON, _ := json.Marshal(resp)

	var rpcResp struct {
		Result mcp.CallToolResult `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respJSON, &rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
	}
	return &rpcResp.Result, nil
}

func TestQueryTasks_InvalidStatus(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	tasks.RegisterTools(s, mock)

	result, err := callTool(s, "query_tasks", map[string]any{"status": "BadStatus"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid status")
	}
}

func TestQueryTasks_Success(t *testing.T) {
	mock := &mockRepo{
		queryResult: &tasks.QueryResult{
			Tasks: []tasks.Task{
				{ID: "t1", Name: "Test task", Status: "In progress", Due: "2026-04-07"},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	tasks.RegisterTools(s, mock)

	result, err := callTool(s, "query_tasks", map[string]any{
		"status":    "In progress",
		"due_after": "2026-04-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result")
	}

	if mock.lastFilter.Status != "In progress" {
		t.Errorf("expected filter status 'In progress', got '%s'", mock.lastFilter.Status)
	}
	if mock.lastFilter.DueAfter != "2026-04-01" {
		t.Errorf("expected filter due_after '2026-04-01', got '%s'", mock.lastFilter.DueAfter)
	}
}

func TestCreateTask_Success(t *testing.T) {
	mock := &mockRepo{
		createTask: &tasks.Task{
			ID:   "new-id",
			URL:  "https://notion.so/new-id",
			Name: "My task",
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	tasks.RegisterTools(s, mock)

	result, err := callTool(s, "create_task", map[string]any{
		"name": "My task",
		"due":  "2026-04-15",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result")
	}

	if mock.lastCreate.Name != "My task" {
		t.Errorf("expected name 'My task', got '%s'", mock.lastCreate.Name)
	}
	if mock.lastCreate.Due != "2026-04-15" {
		t.Errorf("expected due '2026-04-15', got '%s'", mock.lastCreate.Due)
	}
}
