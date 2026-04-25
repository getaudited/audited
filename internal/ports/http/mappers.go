package http

import (
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
)

func mapToQueryPaginationParams(page, limit *int) query.PaginationParams {
	r := query.PaginationParams{
		Limit: 25,
		Page:  1,
	}

	if page != nil && *page > 0 {
		r.Page = *page
	}

	if limit != nil && *limit > 0 {
		r.Limit = *limit
	}

	return r
}

func mapToSources(sources []domain.Source) []Source {
	result := make([]Source, len(sources))

	for i, s := range sources {
		result[i] = mapToSource(s)
	}

	return result
}

func mapToSource(s domain.Source) Source {
	return Source{
		Id:        s.ID().String(),
		Name:      s.Name(),
		CreatedAt: s.CreatedAt(),
		UpdatedAt: s.UpdatedAt(),
	}
}
