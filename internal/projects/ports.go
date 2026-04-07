package projects

import "context"

type QueryResult struct {
	Projects    []Project
	HasMore     bool
	NextCursor  string
}

type Repository interface {
	Query(ctx context.Context, filter Filter) (*QueryResult, error)
	Create(ctx context.Context, params CreateParams) (*Project, error)
}
