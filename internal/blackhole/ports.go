package blackhole

import "context"

type QueryResult struct {
	Entries    []Entry
	HasMore    bool
	NextCursor string
}

type Repository interface {
	Query(ctx context.Context, filter Filter) (*QueryResult, error)
	Create(ctx context.Context, params CreateParams) (*Entry, error)
}
