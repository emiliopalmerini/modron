package tasks

import "context"

type QueryResult struct {
	Tasks       []Task
	HasMore     bool
	NextCursor  string
}

type Repository interface {
	Query(ctx context.Context, filter Filter) (*QueryResult, error)
	Create(ctx context.Context, params CreateParams) (*Task, error)
}
