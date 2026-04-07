package page

import "context"

type Repository interface {
	Get(ctx context.Context, pageID string) (*Page, error)
	Update(ctx context.Context, params UpdateParams) (*Page, error)
}
