package events

import "context"

type QueryResult struct {
	Events     []Event
	HasMore    bool
	NextCursor string
}

type Repository interface {
	Query(ctx context.Context, filter Filter) (*QueryResult, error)
	Create(ctx context.Context, params CreateParams) (*Event, error)
}
