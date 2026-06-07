package psql_test

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

func TestEventsPsqlRepository_Save(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)
	token := fixtureToken(t, source.ID())
	eventType := fixtureEventType()
	storeSource(t, source)
	storeToken(t, token)
	storeEventType(t, eventType)

	event := fixtureEvent(t, source.ID(), eventType.Action)

	// WHEN
	err := repo.Save(context.Background(), event, token.Value())
	require.NoError(t, err)

	// THEN
	storedEvent := queryEventByID(t, event.ID())
	require.NotNil(t, storedEvent)

	require.Equal(t, event.ID().String(), storedEvent.ID)
	require.Equal(t, event.Version(), storedEvent.Version)
	require.WithinDuration(t, event.OccurredAt(), storedEvent.OccurredAt, time.Millisecond)

	require.Equal(t, event.Actor().ID, storedEvent.ActorID)
	require.Equal(t, event.Actor().ActorType, storedEvent.ActorType)
	require.Equal(t, event.Actor().Name, storedEvent.ActorName.Ptr())
	requireEqualMetadata(t, event.Actor().Metadata, storedEvent.ActorMetadata)

	require.Equal(t, event.Context().Location, storedEvent.ContextLocation)
	require.Equal(t, event.Context().UserAgent, storedEvent.ContextUserAgent.Ptr())

	requireEqualMetadata(t, event.Metadata(), storedEvent.Metadata)

	require.Len(t, storedEvent.R.EventTargets, len(event.Targets()))

	for _, target := range event.Targets() {
		storedTarget := findStoredTarget(storedEvent.R.EventTargets, target.ID)
		require.NotNilf(t, storedTarget, "target %s not found in stored event", target.ID)
		require.Equal(t, target.ID, storedTarget.ID)
		require.Equal(t, event.ID().String(), storedTarget.EventID)
		require.Equal(t, target.TargetType, storedTarget.Type)
		require.Equal(t, target.Name, storedTarget.Name.Ptr())
		requireEqualMetadata(t, target.Metadata, storedTarget.Metadata)
	}
}

func TestEventsPsqlRepository_QueryAll(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// dedicated source + token
	source := fixtureSource(t)
	token := fixtureToken(t, source.ID())
	eventType := fixtureEventType()
	storeSource(t, source)
	storeToken(t, token)
	storeEventType(t, eventType)

	// separate source to verify SourceID isolation
	otherSource := fixtureSource(t)
	otherToken := fixtureToken(t, otherSource.ID())
	storeSource(t, otherSource)
	storeToken(t, otherToken)

	knownActorID := ulid.Make().String()
	knownActorName := "Alice"
	knownTargetID := ulid.Make().String()

	baseTime := time.Now().UTC().Truncate(time.Millisecond)

	// event1: oldest, with known actor + target
	event1 := fixtureEventWith(
		t,
		eventType.Action,
		source.ID(),
		domain.Actor{ID: knownActorID, ActorType: "user", Name: &knownActorName},
		[]domain.Target{{ID: knownTargetID, TargetType: "resource"}},
		baseTime.Add(-2*time.Hour),
	)

	// event2: middle timestamp, random actor + target
	event2 := fixtureEventWith(
		t,
		eventType.Action,
		source.ID(),
		domain.Actor{ID: ulid.Make().String(), ActorType: "user"},
		[]domain.Target{{ID: ulid.Make().String(), TargetType: "resource"}},
		baseTime.Add(-time.Hour),
	)

	// event3: newest, random actor + target
	event3 := fixtureEventWith(
		t,
		eventType.Action,
		source.ID(),
		domain.Actor{ID: ulid.Make().String(), ActorType: "system"},
		[]domain.Target{{ID: ulid.Make().String(), TargetType: "resource"}},
		baseTime,
	)

	// event for another source — must never appear in source-scoped queries
	eventOtherSource := fixtureEventWith(
		t,
		eventType.Action,
		otherSource.ID(),
		domain.Actor{ID: ulid.Make().String(), ActorType: "user"},
		[]domain.Target{{ID: ulid.Make().String(), TargetType: "resource"}},
		baseTime,
	)

	storeEvent(t, event1, token.Value())
	storeEvent(t, event2, token.Value())
	storeEvent(t, event3, token.Value())
	storeEvent(t, eventOtherSource, otherToken.Value())

	// pre-fetch first page (limit=2) so its cursor can be used in a table case
	firstPage, err := repo.QueryAll(ctx,
		query.AllEventsParams{SourceID: source.ID()},
		query.CursorPaginationParams{Limit: new(2)},
	)
	require.NoError(t, err)
	require.NotEmpty(t, firstPage.LastItemCursor)
	cursorAfterFirstPage := firstPage.LastItemCursor

	testCases := []struct {
		name        string
		params      query.AllEventsParams
		pagination  query.CursorPaginationParams
		expectedIDs []domain.ID
	}{
		{
			name:   "returns all events for source ordered by occurred_at desc",
			params: query.AllEventsParams{SourceID: source.ID()},
			expectedIDs: []domain.ID{
				event3.ID(),
				event2.ID(),
				event1.ID(),
			},
		},
		{
			name:        "returns only events belonging to the given source",
			params:      query.AllEventsParams{SourceID: otherSource.ID()},
			expectedIDs: []domain.ID{eventOtherSource.ID()},
		},
		{
			name: "filters by actor_id",
			params: query.AllEventsParams{
				SourceID: source.ID(),
				ActorID:  domain.ID(knownActorID),
			},
			expectedIDs: []domain.ID{event1.ID()},
		},
		{
			name: "filters by actor_name case-insensitively",
			params: query.AllEventsParams{
				SourceID:  source.ID(),
				ActorName: new("alice"),
			},
			expectedIDs: []domain.ID{event1.ID()},
		},
		{
			name: "filters by target_id",
			params: query.AllEventsParams{
				SourceID: source.ID(),
				TargetID: domain.ID(knownTargetID),
			},
			expectedIDs: []domain.ID{event1.ID()},
		},
		{
			name: "filters by start_date",
			params: query.AllEventsParams{
				SourceID:  source.ID(),
				StartDate: new(baseTime.Add(-90 * time.Minute)),
			},
			expectedIDs: []domain.ID{event3.ID(), event2.ID()},
		},
		{
			name: "filters by end_date",
			params: query.AllEventsParams{
				SourceID: source.ID(),
				EndDate:  new(baseTime.Add(-90 * time.Minute)),
			},
			expectedIDs: []domain.ID{event1.ID()},
		},
		{
			name: "filters by date range",
			params: query.AllEventsParams{
				SourceID:  source.ID(),
				StartDate: new(baseTime.Add(-90 * time.Minute)),
				EndDate:   new(baseTime.Add(-30 * time.Minute)),
			},
			expectedIDs: []domain.ID{event2.ID()},
		},
		{
			name:       "respects pagination limit",
			params:     query.AllEventsParams{SourceID: source.ID()},
			pagination: query.CursorPaginationParams{Limit: new(1)},
			// only the most-recent event
			expectedIDs: []domain.ID{event3.ID()},
		},
		{
			name:   "uses cursor to fetch subsequent page",
			params: query.AllEventsParams{SourceID: source.ID()},
			pagination: query.CursorPaginationParams{
				Limit:           new(10),
				StartFromCursor: new(cursorAfterFirstPage),
			},
			// first page (limit=2) returned event3 + event2; cursor continues from event2
			expectedIDs: []domain.ID{event1.ID()},
		},
		{
			name:        "returns empty list when source has no events",
			params:      query.AllEventsParams{SourceID: domain.NewID()},
			expectedIDs: []domain.ID{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := repo.QueryAll(ctx, tc.params, tc.pagination)
			require.NoError(t, err)
			require.Len(t, result.Data, len(tc.expectedIDs), "unexpected number of events returned")

			for i, expectedID := range tc.expectedIDs {
				require.Equal(t, expectedID, result.Data[i].ID())
			}
		})
	}
}

func TestEventsPsqlRepository_Save_ErrTokenNotFound(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// GIVEN
	nonExistentSourceID := domain.NewID()
	nonExistentToken := domain.TokenValue("some-value")

	event := fixtureEvent(t, nonExistentSourceID, "users.created")

	// WHEN
	err := repo.Save(context.Background(), event, nonExistentToken)

	// THEN
	require.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func TestEventsPsqlRepository_Save_ErrEventTypeNotFound(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)
	token := fixtureToken(t, source.ID())
	eventType := fixtureEventType()
	storeSource(t, source)
	storeToken(t, token)
	storeEventType(t, eventType)

	event := fixtureEvent(t, source.ID(), "dummy.action")

	// WHEN
	err := repo.Save(context.Background(), event, token.Value())

	// THEN
	require.ErrorIs(t, err, domain.ErrEventTypeActionNotFound)
}
