package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/lib/pq"

	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
)

const (
	FkEventBelongsToSource = "fk_event_belongs_to_source"
)

type EventsPsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventsPsqlRepository(db boil.ContextExecutor) EventsPsqlRepository {
	return EventsPsqlRepository{
		db: db,
	}
}

func (r EventsPsqlRepository) Save(ctx context.Context, e domain.Event) error {
	row, err := mapDomainEventToModelEvent(e)
	if err != nil {
		return err
	}

	err = row.Insert(ctx, r.db, boil.Infer())
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == FkEventBelongsToSource {
		return domain.ErrSourceNotFoundWhileSavingEvent
	}

	if err != nil {
		return fmt.Errorf("error saving event: %v", err)
	}

	targetRows, err := mapDomainEventTargetsToModelEventTargets(e.ID(), e.Targets())
	if err != nil {
		return err
	}

	err = row.AddEventTargets(ctx, r.db, true, targetRows...)
	if err != nil {
		return fmt.Errorf("error saving event_targets: %v", err)
	}

	return nil
}

func (r EventsPsqlRepository) QueryAll(
	ctx context.Context,
	sourceID domain.ID,
	params query.CursorPaginationParams,
) (query.CursorPaginationResult[domain.Event], error) {
	queryOptions := []qm.QueryMod{
		qm.Limit(mapToLimit(params.Limit)),
		qm.OrderBy("occurred_at DESC, id DESC"),
		qm.Load(models.EventRels.EventTargets),
		models.EventWhere.SourceID.EQ(sourceID.String()),
	}

	if params.StartFromCursor != nil {
		cursor, err := unmarshallCursor(*params.StartFromCursor)
		if err != nil {
			return query.CursorPaginationResult[domain.Event]{}, err
		}

		queryOptions = append(queryOptions, qm.Where("(occurred_at, id) < (?, ?)", cursor.OccurredAt, cursor.EventID))
	}

	rows, err := models.Events(queryOptions...).All(ctx, r.db)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, fmt.Errorf("error querying events for source_id '%s': %v", sourceID, err)
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

type Cursor struct {
	OccurredAt time.Time `json:"occurred_at"`
	EventID    string    `json:"event_id"`
}
