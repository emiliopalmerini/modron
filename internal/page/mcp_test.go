package page_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/notion-mcp/internal/page"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockRepo struct {
	getPage    *page.Page
	getErr     error
	updatePage *page.Page
	updateErr  error
	lastUpdate page.UpdateParams
}

func (m *mockRepo) Get(_ context.Context, pageID string) (*page.Page, error) {
	return m.getPage, m.getErr
}

func (m *mockRepo) Update(_ context.Context, params page.UpdateParams) (*page.Page, error) {
	m.lastUpdate = params
	return m.updatePage, m.updateErr
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

func TestGetPage_Success(t *testing.T) {
	mock := &mockRepo{
		getPage: &page.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
			Properties: []page.Property{
				{Name: "Task name", Type: "title", Value: "Fix bug"},
				{Name: "Status", Type: "status", Value: "In progress"},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	page.RegisterTools(s, mock)

	result, err := callTool(s, "get_page", map[string]any{"page_id": "page-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
}

func TestGetPage_Error(t *testing.T) {
	mock := &mockRepo{
		getErr: fmt.Errorf("not found"),
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	page.RegisterTools(s, mock)

	result, err := callTool(s, "get_page", map[string]any{"page_id": "bad-id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result")
	}
}

func TestUpdatePage_Success(t *testing.T) {
	mock := &mockRepo{
		updatePage: &page.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
			Properties: []page.Property{
				{Name: "Status", Type: "status", Value: "Done"},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	page.RegisterTools(s, mock)

	result, err := callTool(s, "update_page", map[string]any{
		"page_id":    "page-123",
		"properties": `{"Status": {"status": {"name": "Done"}}}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastUpdate.PageID != "page-123" {
		t.Errorf("expected page-123, got %s", mock.lastUpdate.PageID)
	}
	statusProp, ok := mock.lastUpdate.PropertyUpdates["Status"]
	if !ok {
		t.Fatal("expected Status in property updates")
	}
	statusMap := statusProp.(map[string]any)
	statusInner := statusMap["status"].(map[string]any)
	if statusInner["name"] != "Done" {
		t.Errorf("expected Done, got %v", statusInner["name"])
	}
}

func TestUpdatePage_InvalidJSON(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	page.RegisterTools(s, mock)

	result, err := callTool(s, "update_page", map[string]any{
		"page_id":    "page-123",
		"properties": "not json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid JSON")
	}
}
