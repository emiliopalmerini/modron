package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/notion-mcp/internal/projects"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockRepo struct {
	queryResult *projects.QueryResult
	createProj  *projects.Project
	lastFilter  projects.Filter
	lastCreate  projects.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f projects.Filter) (*projects.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, nil
}

func (m *mockRepo) Create(_ context.Context, p projects.CreateParams) (*projects.Project, error) {
	m.lastCreate = p
	return m.createProj, nil
}

func callTool(s *server.MCPServer, name string, args map[string]any) (*mcp.CallToolResult, error) {
	argsJSON, _ := json.Marshal(args)
	reqJSON, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "tools/call",
		"params": map[string]any{"name": name, "arguments": json.RawMessage(argsJSON)},
	})
	resp := s.HandleMessage(context.Background(), reqJSON)
	respJSON, _ := json.Marshal(resp)
	var rpcResp struct {
		Result mcp.CallToolResult `json:"result"`
		Error  *struct{ Message string } `json:"error"`
	}
	if err := json.Unmarshal(respJSON, &rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
	}
	return &rpcResp.Result, nil
}

func TestQueryProjects_InvalidStatus(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	projects.RegisterTools(s, mock)

	result, _ := callTool(s, "query_projects", map[string]any{"status": "Bad"})
	if !result.IsError {
		t.Error("expected error for invalid status")
	}
}

func TestQueryProjects_InvalidTag(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	projects.RegisterTools(s, mock)

	result, _ := callTool(s, "query_projects", map[string]any{"tag": "Bad"})
	if !result.IsError {
		t.Error("expected error for invalid tag")
	}
}

func TestQueryProjects_Success(t *testing.T) {
	mock := &mockRepo{
		queryResult: &projects.QueryResult{
			Projects: []projects.Project{
				{ID: "p1", Name: "My project", Status: "In Progress", Tag: "Dev"},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	projects.RegisterTools(s, mock)

	result, err := callTool(s, "query_projects", map[string]any{"status": "In Progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastFilter.Status != "In Progress" {
		t.Errorf("expected 'In Progress', got '%s'", mock.lastFilter.Status)
	}
}

func TestCreateProject_Success(t *testing.T) {
	mock := &mockRepo{
		createProj: &projects.Project{ID: "new-p", URL: "https://notion.so/new-p", Name: "New proj"},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	projects.RegisterTools(s, mock)

	result, err := callTool(s, "create_project", map[string]any{"name": "New proj", "tag": "Dev"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastCreate.Name != "New proj" {
		t.Errorf("expected 'New proj', got '%s'", mock.lastCreate.Name)
	}
	if mock.lastCreate.Tag != "Dev" {
		t.Errorf("expected tag 'Dev', got '%s'", mock.lastCreate.Tag)
	}
}
