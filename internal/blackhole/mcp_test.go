package blackhole_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/notion-mcp/internal/blackhole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockRepo struct {
	queryResult *blackhole.QueryResult
	createEntry *blackhole.Entry
	lastFilter  blackhole.Filter
	lastCreate  blackhole.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f blackhole.Filter) (*blackhole.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, nil
}

func (m *mockRepo) Create(_ context.Context, p blackhole.CreateParams) (*blackhole.Entry, error) {
	m.lastCreate = p
	return m.createEntry, nil
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

func TestQueryBlackhole_InvalidType(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	blackhole.RegisterTools(s, mock)

	result, _ := callTool(s, "query_blackhole", map[string]any{"type": "Bad"})
	if !result.IsError {
		t.Error("expected error for invalid type")
	}
}

func TestQueryBlackhole_Success(t *testing.T) {
	mock := &mockRepo{
		queryResult: &blackhole.QueryResult{
			Entries: []blackhole.Entry{
				{ID: "b1", Name: "Cool article", Type: "Reference", Tags: []string{"AI", "Dev"}},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	blackhole.RegisterTools(s, mock)

	processed := false
	_ = processed
	result, err := callTool(s, "query_blackhole", map[string]any{
		"type":      "Reference",
		"tags":      "AI",
		"processed": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastFilter.Type != "Reference" {
		t.Errorf("expected 'Reference', got '%s'", mock.lastFilter.Type)
	}
	if mock.lastFilter.Tags != "AI" {
		t.Errorf("expected 'AI', got '%s'", mock.lastFilter.Tags)
	}
	if mock.lastFilter.Processed == nil || *mock.lastFilter.Processed != false {
		t.Error("expected Processed to be false")
	}
}

func TestCreateBlackhole_InvalidType(t *testing.T) {
	mock := &mockRepo{}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	blackhole.RegisterTools(s, mock)

	result, _ := callTool(s, "create_blackhole_entry", map[string]any{
		"name": "test", "type": "Bad",
	})
	if !result.IsError {
		t.Error("expected error for invalid type")
	}
}

func TestCreateBlackhole_Success(t *testing.T) {
	mock := &mockRepo{
		createEntry: &blackhole.Entry{
			ID: "new-b", URL: "https://notion.so/new-b", Name: "Great article", Type: "Reference",
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	blackhole.RegisterTools(s, mock)

	result, err := callTool(s, "create_blackhole_entry", map[string]any{
		"name":    "Great article",
		"type":    "Reference",
		"tags":    "AI,Dev",
		"url":     "https://example.com/article",
		"summary": "An insightful article about AI",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastCreate.Name != "Great article" {
		t.Errorf("expected 'Great article', got '%s'", mock.lastCreate.Name)
	}
	if mock.lastCreate.Tags != "AI,Dev" {
		t.Errorf("expected 'AI,Dev', got '%s'", mock.lastCreate.Tags)
	}
	if mock.lastCreate.URL != "https://example.com/article" {
		t.Errorf("expected URL, got '%s'", mock.lastCreate.URL)
	}
}
