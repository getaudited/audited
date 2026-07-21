package clickhouse_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/common/testhelpers"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestEvents_Save(t *testing.T) {
	t.Run("save_event", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventsClickhouseRepository(db)
		event, err := domain.NewEvent(
			domain.NewID(),
			1,
			"user.created",
			testhelpers.FixtureActor(),
			[]domain.Target{testhelpers.FixtureTarget()},
			testhelpers.FixtureContext(),
			testhelpers.FixtureMetadata(),
			time.Now(),
		)
		require.NoError(t, err)

		token := seedToken(t)

		err = repo.Save(ctx, event, token.Value())
		require.NoError(t, err)
	})

	t.Run("error_saving_event", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventsClickhouseRepository(dbError)
		err := repo.Save(ctx, domain.Event{}, "token")
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEvents_QueryAll(t *testing.T) {
	t.Run("query_all", func(t *testing.T) {
		t.Parallel()

		seedCount := 20
		repo := chadapters.NewEventsClickhouseRepository(db)
		events := seedEvents(t, repo, seedToken(t).Value(), seedCount)
		sourceID := events[0].SourceID()
		limit := 2

		result, err := repo.QueryAll(ctx, query.AllEvents{
			SourceID: sourceID,
			Limit:    new(limit),
		})
		require.NoError(t, err)
		require.Len(t, result.Data, limit)
	})

	t.Run("error_querying_all", func(t *testing.T) {
		t.Parallel()

		repoErr := chadapters.NewEventsClickhouseRepository(dbError)
		_, err := repoErr.QueryAll(ctx, query.AllEvents{})
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEvents_QueryAll_WithFilters(t *testing.T) {
	seedCount := 20
	repo := chadapters.NewEventsClickhouseRepository(db)
	token := seedToken(t).Value()
	events := seedEvents(t, repo, token, seedCount)
	sourceID := events[0].SourceID()
	oldEvents := make([]*domain.Event, 10)
	for i := 0; i < 10; i++ {
		occurredAt := gofakeit.DateRange(
			time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		)
		event := fixtureEventWithOccurredAt(t, sourceID, occurredAt)
		seedEvent(t, repo, token, event)
		oldEvents[i] = event
	}

	testCases := []struct {
		name           string
		query          query.AllEvents
		expectedEvents []*domain.Event
	}{
		{
			name: "filter_by_actor_name",
			query: query.AllEvents{
				SourceID:  sourceID,
				ActorName: events[0].Actor().Name,
			},
			expectedEvents: []*domain.Event{events[0]},
		},
		{
			name: "filter_by_actor_id",
			query: query.AllEvents{
				SourceID: sourceID,
				ActorID:  domain.ID(events[0].Actor().ID),
			},
			expectedEvents: []*domain.Event{events[0]},
		},
		{
			name: "filter_by_target_id",
			query: query.AllEvents{
				SourceID: sourceID,
				TargetID: domain.ID(events[0].Targets()[0].ID),
			},
			expectedEvents: []*domain.Event{events[0]},
		},
		{
			name: "filter_by_start_date",
			query: query.AllEvents{
				SourceID:  sourceID,
				StartDate: new(time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)),
			},
			expectedEvents: filterEvents(oldEvents, func(event *domain.Event) bool {
				return event.OccurredAt().After(time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC))
			}),
		},
		{
			name: "filter_by_end_date",
			query: query.AllEvents{
				SourceID: sourceID,
				EndDate:  new(time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)),
			},
			expectedEvents: filterEvents(oldEvents, func(event *domain.Event) bool {
				return event.OccurredAt().Before(time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC))
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := repo.QueryAll(ctx, tc.query)
			require.NoError(t, err)
			for _, expected := range tc.expectedEvents {
				i := slices.IndexFunc(result.Data, func(event domain.Event) bool {
					return event.ID() == expected.ID()
				})
				require.GreaterOrEqual(t, i, 0)
			}
		})
	}
}

func TestEvents_QueryAll_CursorPagination(t *testing.T) {
	seedCount := 20
	limit := 2
	repo := chadapters.NewEventsClickhouseRepository(db)
	token := seedToken(t).Value()
	events := seedEvents(t, repo, token, seedCount)
	sourceID := events[0].SourceID()

	slices.SortFunc(events, func(a, b *domain.Event) int {
		return b.OccurredAt().Compare(a.OccurredAt())
	})

	var totalPages int
	var startingAfter *string
	previousEventIDs := make(map[domain.ID]struct{})

	for {
		result, err := repo.QueryAll(ctx, query.AllEvents{
			SourceID:      sourceID,
			Limit:         new(limit),
			StartingAfter: startingAfter,
		})
		require.NoError(t, err)
		require.Len(t, result.Data, limit)
		startingAfter = new(result.Data[1].ID().String())

		for _, event := range result.Data {
			_, exists := previousEventIDs[event.ID()]
			require.False(t, exists)
			previousEventIDs[event.ID()] = struct{}{}
		}

		totalPages++

		if !result.HasMore {
			break
		}
	}

	clear(previousEventIDs)

	var endingBefore = events[len(events)-1].ID().String()
	for {
		result, err := repo.QueryAll(ctx, query.AllEvents{
			SourceID:     sourceID,
			Limit:        new(limit),
			EndingBefore: new(endingBefore),
		})
		require.NoError(t, err)
		require.Len(t, result.Data, limit)

		for _, event := range result.Data {
			_, exists := previousEventIDs[event.ID()]
			require.False(t, exists)
			previousEventIDs[event.ID()] = struct{}{}
		}

		if events[0].ID() == result.Data[0].ID() {
			break
		}
	}
}

func TestEvents_FindByID(t *testing.T) {
	t.Run("find_by_id", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventsClickhouseRepository(db)
		events := seedEvents(t, repo, seedToken(t).Value(), 1)

		found, err := repo.FindByID(ctx, events[0].ID())
		require.NoError(t, err)
		requireEventEqual(t, events[0], &found)
	})

	t.Run("error_event_not_found", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventsClickhouseRepository(db)

		_, err := repo.FindByID(ctx, domain.NewID())
		require.ErrorIs(t, err, domain.ErrEventNotFound)
	})

	t.Run("error_querying_event", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventsClickhouseRepository(dbError)

		_, err := repo.FindByID(ctx, domain.NewID())
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func seedToken(t *testing.T) *domain.Token {
	token, err := domain.NewToken(
		domain.NewID(),
		fmt.Sprintf("test-%s", domain.NewID()),
	)
	require.NoError(t, err)

	repo := chadapters.NewTokenChRepository(db)
	err = repo.Save(ctx, token)
	require.NoError(t, err)

	return token
}

func seedEvents(
	t *testing.T,
	repo *chadapters.EventsClickhouseRepository,
	token domain.TokenValue,
	count int,
) []*domain.Event {
	events := make([]*domain.Event, count)

	sourceID := domain.NewID()
	for i := 0; i < count; i++ {
		event := fixtureEvent(t, sourceID)
		seedEvent(t, repo, token, event)
		events[i] = event
	}

	return events
}

func seedEvent(t *testing.T, repo *chadapters.EventsClickhouseRepository, token domain.TokenValue, event *domain.Event) {
	err := repo.Save(ctx, *event, token)
	require.NoError(t, err)
}

func fixtureEvent(t *testing.T, sourceID domain.ID) *domain.Event {
	event, err := domain.NewEvent(
		sourceID,
		1,
		fmt.Sprintf("user.created.test-%s", domain.NewID()),
		testhelpers.FixtureActor(),
		[]domain.Target{testhelpers.FixtureTarget()},
		testhelpers.FixtureContext(),
		testhelpers.FixtureMetadata(),
		time.Now(),
	)
	require.NoError(t, err)

	return &event
}

func fixtureEventWithOccurredAt(t *testing.T, sourceID domain.ID, occurredAt time.Time) *domain.Event {
	event, err := domain.NewEvent(
		sourceID,
		1,
		fmt.Sprintf("user.created.test-%s", domain.NewID()),
		testhelpers.FixtureActor(),
		[]domain.Target{testhelpers.FixtureTarget()},
		testhelpers.FixtureContext(),
		testhelpers.FixtureMetadata(),
		occurredAt,
	)
	require.NoError(t, err)

	return &event
}

func requireEventEqual(t *testing.T, expected, got *domain.Event) {
	t.Helper()

	require.Equal(t, expected.ID(), got.ID())
	require.Equal(t, expected.SourceID(), got.SourceID())
	require.Equal(t, expected.Action(), got.Action())
	require.Equal(t, expected.Actor(), got.Actor())
	require.Equal(t, expected.Version(), got.Version())
	require.Equal(t, expected.Targets(), got.Targets())
	require.Equal(t, expected.Context(), got.Context())
	require.Equal(t, expected.Metadata(), got.Metadata())
}

func filterEvents(events []*domain.Event, filter func(event *domain.Event) bool) []*domain.Event {
	var result []*domain.Event
	for _, event := range events {
		if filter(event) {
			result = append(result, event)
		}
	}

	return result
}
