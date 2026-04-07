package events_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/notion-mcp/internal/events"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockRepo struct {
	queryResult *events.QueryResult
	createEvent *events.Event
	lastFilter  events.Filter
	lastCreate  events.CreateParams
}

func (m *mockRepo) Query(_ context.Context, f events.Filter) (*events.QueryResult, error) {
	m.lastFilter = f
	return m.queryResult, nil
}

func (m *mockRepo) Create(_ context.Context, p events.CreateParams) (*events.Event, error) {
	m.lastCreate = p
	return m.createEvent, nil
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

func TestQueryEvents_Success(t *testing.T) {
	mock := &mockRepo{
		queryResult: &events.QueryResult{
			Events: []events.Event{
				{ID: "e1", Name: "Team meeting", DateStart: "2026-04-07T10:00:00"},
			},
		},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	events.RegisterTools(s, mock)

	result, err := callTool(s, "query_events", map[string]any{
		"date_after":  "2026-04-01",
		"date_before": "2026-04-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastFilter.DateAfter != "2026-04-01" {
		t.Errorf("expected '2026-04-01', got '%s'", mock.lastFilter.DateAfter)
	}
}

func TestCreateEvent_Success(t *testing.T) {
	mock := &mockRepo{
		createEvent: &events.Event{ID: "new-e", URL: "https://notion.so/new-e", Name: "Launch party"},
	}
	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	events.RegisterTools(s, mock)

	result, err := callTool(s, "create_event", map[string]any{
		"name":       "Launch party",
		"date_start": "2026-04-15T18:00:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("unexpected error result")
	}
	if mock.lastCreate.Name != "Launch party" {
		t.Errorf("expected 'Launch party', got '%s'", mock.lastCreate.Name)
	}
}
