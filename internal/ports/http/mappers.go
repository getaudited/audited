package http

import (
	"context"

	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/labstack/echo/v4"
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
		"users.created",
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

func mapToTokens(tokens []*domain.Token) []Token {
	r := make([]Token, len(tokens))

	for i, token := range tokens {
		r[i] = Token{
			Id:        token.ID().String(),
			Value:     token.MaskedValue(),
			Name:      token.Name(),
			SourceId:  token.SourceID().String(),
			CreatedAt: token.CreatedAt(),
		}
	}

	return r
}

func mapToEventType(et *domain.EventType) EventType {
	var schema *string
	if len(et.Schema) > 0 {
		s := string(et.Schema)
		schema = new(s)
	}

	return EventType{
		Id:                           et.Id,
		Version:                      et.Version,
		Action:                       et.Action,
		TargetTypes:                  et.TargetTypes,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		Schema:                       schema,
		CreatedAt:                    et.CreatedAt,
		UpdatedAt:                    et.UpdatedAt,
	}
}

func mapToEventTypeList(result query.Pagination[*domain.EventType]) EventTypeList {
	data := make([]EventType, len(result.Data))
	for i, et := range result.Data {
		data[i] = mapToEventType(et)
	}

	return EventTypeList{
		Data: data,
		Pagination: Pagination{
			Total:       result.Total,
			PerPage:     result.PerPage,
			CurrentPage: result.CurrentPage,
			TotalPages:  result.TotalPages,
		},
	}
}

func mapEchoCtxToCtx(c echo.Context) context.Context {
	return c.Request().Context()
}
