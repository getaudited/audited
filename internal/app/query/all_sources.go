package query

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type AllSources struct {
	Name       *string
	Pagination PaginationParams
}

type sourcesFinder interface {
	QueryAll(ctx context.Context, params AllSources) (Pagination[domain.Source], error)
}

type AllSourcesHandler struct {
	finder sourcesFinder
}

func NewAllSourcesHandler(finder sourcesFinder) AllSourcesHandler {
	return AllSourcesHandler{
		finder: finder,
	}
}

func (e AllSourcesHandler) Execute(ctx context.Context, q AllSources) (Pagination[domain.Source], error) {
	return e.finder.QueryAll(ctx, q)
}
