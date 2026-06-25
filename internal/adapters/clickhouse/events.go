package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	sq "github.com/Masterminds/squirrel"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

type EventsClickhouseRepository struct {
	db clickhousedb.Conn
}

func NewEventsClickhouseRepository(conn clickhousedb.Conn) *EventsClickhouseRepository {
	return &EventsClickhouseRepository{
		db: conn,
	}
}

func (r EventsClickhouseRepository) QueryAll(
	ctx context.Context,
	params query.AllEventsParams,
	pagination query.CursorPaginationParams,
) (query.CursorPaginationResult[domain.Event], error) {
	queryAll := sq.Select(`id,
		source_id,
		version,
		action,
		actor_id,
		actor_type,
		actor_name,
		actor_metadata,
		context_location,
		context_user_agent,
		metadata,
		occurred_at,
		targets.internal_id,
		targets.id,
		targets.name,
		targets.type,
		targets.metadata`).From("events")

	q, _, err := queryAll.ToSql()
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, fmt.Errorf("error parsing query: %w", err)
	}

	rows, err := r.db.Query(ctx, q, params.SourceID, pagination.Limit)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, fmt.Errorf("error querying events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	data, err := mapRowsToEvents(rows)
	if err != nil {
		return query.CursorPaginationResult[domain.Event]{}, fmt.Errorf("error querying events: %w", err)
	}

	return query.CursorPaginationResult[domain.Event]{
		LastItemCursor: "",
		Data:           data,
	}, nil
}

func (r EventsClickhouseRepository) FindByID(ctx context.Context, id domain.ID) (domain.Event, error) {
	row := r.db.QueryRow(ctx, queryFindByID, id.String())
	return mapRowToDomainEvent(row)
}

func (r EventsClickhouseRepository) Save(ctx context.Context, e domain.Event, token domain.TokenValue) error {
	targetsLen := len(e.Targets())
	internalIDs := make([]string, targetsLen)
	ids := make([]string, targetsLen)
	names := make([]*string, targetsLen)
	types := make([]string, targetsLen)
	metadatas := make([]string, targetsLen)

	for i, et := range e.Targets() {
		internalIDs[i] = domain.NewID().String()
		ids[i] = et.ID
		names[i] = et.Name
		types[i] = et.TargetType
		if et.Metadata != nil {
			metadata, err := json.Marshal(et.Metadata)
			if err != nil {
				return err
			}
			metadatas[i] = string(metadata)
		}
	}

	var actorMetadata string
	var eventMetadata string
	if e.Actor().Metadata != nil {
		m, err := json.Marshal(e.Actor().Metadata)
		if err != nil {
			return err
		}
		actorMetadata = string(m)
	}

	if e.Metadata() != nil {
		m, err := json.Marshal(e.Metadata())
		if err != nil {
			return err
		}
		eventMetadata = string(m)
	}

	err := r.db.Exec(ctx, querySaveEvent,
		e.ID().String(),
		e.SourceID().String(),
		e.Version(),
		e.Action(),
		e.Actor().ID,
		e.Actor().ActorType,
		e.Actor().Name,
		actorMetadata,
		e.Context().Location,
		e.Context().UserAgent,
		eventMetadata,
		e.OccurredAt(),
		internalIDs,
		ids,
		names,
		types,
		metadatas,
	)
	if err != nil {
		return fmt.Errorf("unable to save event: %w", err)
	}

	return nil
}
