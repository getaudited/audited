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
		body.Action,
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

func mapToEventType(et query.EventType) EventType {
	versions := make([]EventTypeVersion, len(et.Versions))
	for i, eventTypeVersion := range et.Versions {
		versions[i] = EventTypeVersion{
			Version:     eventTypeVersion.Version,
			Schema:      new(string(eventTypeVersion.Schema)),
			TargetTypes: eventTypeVersion.TargetTypes,
			CreatedAt:   eventTypeVersion.CreatedAt,
		}
	}

	return EventType{
		Action:                       et.Action,
		Versions:                     versions,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		CreatedAt:                    et.CreatedAt,
	}
}

func mapToEventTypeList(result query.Pagination[query.EventType]) EventTypeList {
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

func mapDomainEventsToEvents(events []domain.Event) []Event {
	r := make([]Event, len(events))

	for i, e := range events {
		r[i] = mapDomainEventToEvent(e)
	}

	return r
}

func mapDomainEventToEvent(e domain.Event) Event {
	targets := make([]Target, len(e.Targets()))

	for i, t := range e.Targets() {
		targets[i] = Target{
			Id:         t.ID,
			Name:       t.Name,
			Metadata:   t.Metadata,
			TargetType: t.TargetType,
		}
	}

	return Event{
		Id:       e.ID().String(),
		Metadata: e.Metadata(),
		SourceId: e.SourceID().String(),
		Action:   e.Action(),
		Actor: Actor{
			Id:        e.Actor().ID,
			Name:      e.Actor().Name,
			ActorType: e.Actor().ActorType,
			Metadata:  e.Actor().Metadata,
		},
		Context: Context{
			Location:  e.Context().Location,
			UserAgent: e.Context().UserAgent,
		},
		Targets:    targets,
		Version:    e.Version(),
		OccurredAt: e.OccurredAt().String(),
	}
}
