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

func mapRequestToDomainEvent(body CreateEventJSONBody) (domain.Event, error) {
	targets := make([]domain.Target, len(body.Targets))
	for i, target := range body.Targets {
		targets[i] = domain.Target{
			ID:         target.Id,
			Name:       target.Name,
			TargetType: target.Type,
			Metadata:   target.Metadata,
		}
	}

	return domain.NewEvent(
		domain.ID(body.SourceId),
		body.Version,
		domain.Actor{
			ID:        body.Actor.Id,
			ActorType: body.Actor.Type,
			Name:      body.Actor.Name,
			Metadata:  body.Actor.Metadata,
		},
		targets,
		domain.Context{
			Location:  body.Context.Location,
			UserAgent: body.Context.UserAgent,
		},
		body.Metadata,
		body.OccurredAt,
	)
}
