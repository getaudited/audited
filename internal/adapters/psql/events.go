package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/lib/pq"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

const (
	FkEventBelongsToSource    = "fk_event_belongs_to_source"
	FkEventHasEventTypeAction = "fk_event_has_action"
)

type Cursor struct {
	OccurredAt time.Time `json:"occurred_at"`
	EventID    string    `json:"event_id"`
}

type EventsPsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventsPsqlRepository(db boil.ContextExecutor) EventsPsqlRepository {
	return EventsPsqlRepository{
		db: db,
	}
}

func (r EventsPsqlRepository) Save(ctx context.Context, e domain.Event, token domain.TokenValue) error {
	err := r.validateToken(ctx, token, e.SourceID())
	if err != nil {
		return err
	}

	err = r.checkEventType(ctx, e.Action())
	if err != nil {
		return err
	}

	row, err := mapDomainEventToModelEvent(e)
	if err != nil {
		return err
	}

	err = row.Insert(ctx, r.db, boil.Infer())
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == FkEventBelongsToSource {
		return domain.ErrSourceNotFoundWhileSavingEvent
	}
	if err != nil {
		return fmt.Errorf("error saving event: %w", err)
	}

	targetRows, err := mapDomainEventTargetsToModelEventTargets(e.ID(), e.Targets())
	if err != nil {
		return err
	}

	err = row.AddEventTargets(ctx, r.db, true, targetRows...)
	if err != nil {
		return fmt.Errorf("error saving event_targets: %w", err)
	}

	return nil
}

func (r EventsPsqlRepository) QueryAll(
	ctx context.Context,
	params query.AllEventsParams,
	pagination query.CursorPaginationParams,
) (query.CursorPaginationResult[domain.Event], error) {
	opts := []qm.QueryMod{
		qm.Limit(mapToLimit(pagination.Limit)),
		qm.OrderBy("occurred_at DESC, id DESC"),
		models.EventWhere.SourceID.EQ(params.SourceID.String()),
	}

	if !params.ActorID.Empty() {
		opts = append(opts, models.EventWhere.ActorID.EQ(params.ActorID.String()))
	}

	if params.ActorName != nil {
		opts = append(opts, models.EventWhere.ActorName.ILIKE(null.StringFromPtr(params.ActorName)))
	}

	if params.TargetID.Empty() {
		opts = append(opts, qm.Load(models.EventRels.EventTargets))
	} else {
		opts = append(
			opts,
			// Fetch only events by the given target_id;
			qm.InnerJoin("event_targets et ON et.event_id = events.id AND et.id = ?", params.TargetID.String()),
			// Fetch the events targets in a separate SQL call by providing ids of the previously fetched events;
			qm.Load(models.EventRels.EventTargets, models.EventTargetWhere.ID.EQ(params.TargetID.String())),
		)
	}

	if params.StartDate != nil {
		opts = append(opts, models.EventWhere.OccurredAt.GTE(*params.StartDate))
	}

	if params.EndDate != nil {
		opts = append(opts, models.EventWhere.OccurredAt.LTE(*params.EndDate))
	}

	if pagination.StartFromCursor != nil {
		cursor, err := unmarshallCursor(*pagination.StartFromCursor)
		if err != nil {
			return query.CursorPaginationResult[domain.Event]{}, err
		}

		opts = append(opts, qm.Where("(occurred_at, id) < (?, ?)", cursor.OccurredAt, cursor.EventID))
	}

	rows, err := models.Events(opts...).All(ctx, r.db)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, fmt.Errorf("error querying events for source_id '%s': %w", params.SourceID, err)
	}

	lastItemCursor, err := mapLastItemCursor(rows)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, err
	}

	events, err := mapRowsToDomainEvents(rows)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, err
	}

	return query.CursorPaginationResult[domain.Event]{
		Data:           events,
		LastItemCursor: lastItemCursor,
	}, nil
}

func (r EventsPsqlRepository) validateToken(ctx context.Context, token domain.TokenValue, sourceID domain.ID) error {
	exists, err := models.Tokens(
		models.TokenWhere.Value.EQ(token.String()),
		models.TokenWhere.SourceID.EQ(sourceID.String()),
	).Exists(ctx, r.db)
	if err != nil {
		return fmt.Errorf("error validating token: %w", err)
	}

	if !exists {
		return domain.ErrTokenNotFound
	}

	return nil
}

func (r EventsPsqlRepository) checkEventType(ctx context.Context, eventTypeAction string) error {
	exists, err := models.EventTypes(
		models.EventTypeWhere.Action.EQ(eventTypeAction),
	).Exists(ctx, r.db)
	if err != nil {
		return fmt.Errorf("error checking for event_type: '%s': %w", eventTypeAction, err)
	}

	if !exists {
		return domain.ErrEventTypeActionNotFound
	}

	return nil
}
