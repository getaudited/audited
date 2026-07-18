package clickhouse_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

func TestEventTypes_FindByAction(t *testing.T) {
	t.Run("find_event_by_action", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]

		found, err := repo.FindByAction(ctx, eventType.Action())
		require.NoError(t, err)

		require.Equal(t, eventType.Action(), found.Action)
		require.Equal(t, eventType.TargetTypes(), found.TargetTypes)
		require.Equal(t, eventType.Version(), found.Version)
		require.Equal(t, eventType.ShouldValidateMetadataSchema(), found.ShouldValidateMetadataSchema)
		require.Equal(t, string(eventType.Schema()), found.Schema)
		require.WithinDuration(t, eventType.CreatedAt(), found.CreatedAt, time.Second)
	})

	t.Run("error_event_type_not_found", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		_, err := repo.FindByAction(ctx, "non-existent-event-type")
		require.ErrorIs(t, err, domain.ErrEventTypeNotFound)
	})

	t.Run("error_querying_event_type", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		_, err := repo.FindByAction(ctx, "some-action")
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_QueryAll(t *testing.T) {
	t.Run("query_all", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		seedCount := 10
		_, eventTypes := seedEventTypes(t, repo, seedCount)

		result, err := repo.QueryAll(ctx, query.AllEventTypes{
			PaginationParams: query.PaginationParams{
				Limit: 20,
				Page:  1,
			},
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result.Data), seedCount)

		var totalFound int
		for _, eventType := range result.Data {
			foundEventType, found := eventTypes[eventType.Action]
			if found {
				totalFound++
				requireEqualEventTypes(t, foundEventType, eventType)
			}
		}

		require.Equal(t, seedCount, totalFound)
	})

	t.Run("query_all_filter_by_action", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)

		result, err := repo.QueryAll(ctx, query.AllEventTypes{
			Action: new(eventTypes[0].Action()),
			PaginationParams: query.PaginationParams{
				Limit: 1,
				Page:  1,
			},
		})
		require.NoError(t, err)
		require.Len(t, result.Data, 1)
		requireEqualEventTypes(t, eventTypes[0], result.Data[0])
	})

	t.Run("error_querying", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		_, err := repo.QueryAll(ctx, query.AllEventTypes{})
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_Delete(t *testing.T) {
	t.Run("delete_event_type", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)

		err := repo.Delete(ctx, eventTypes[0].Action())
		require.NoError(t, err)

		_, err = repo.FindByAction(ctx, eventTypes[0].Action())
		require.ErrorIs(t, err, domain.ErrEventTypeNotFound)
	})

	t.Run("error_deleting", func(t *testing.T) {
		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		err := repo.Delete(ctx, "some-action")
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_Save(t *testing.T) {
	t.Run("save_event_type", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventType := fixtureEventType(t)

		err := repo.Save(ctx, eventType)
		require.NoError(t, err)

		_, err = repo.FindByAction(ctx, eventType.Action())
		require.NoError(t, err)
	})

	t.Run("error_event_type_exists", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventType := fixtureEventType(t)

		err := repo.Save(ctx, eventType)
		require.NoError(t, err)

		err = repo.Save(ctx, eventType)
		require.ErrorIs(t, err, domain.ErrEventTypeExists)
	})

	t.Run("error_saving_event_type", func(t *testing.T) {
		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		err := repo.Save(ctx, fixtureEventType(t))
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_SaveVersion(t *testing.T) {
	t.Run("save_event_type_version", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]

		err := repo.SaveVersion(ctx, eventType.Action(), eventType.TargetTypes(), eventType.Schema())
		require.NoError(t, err)

		eventTypeV2, err := repo.FindByAction(ctx, eventType.Action())
		require.NoError(t, err)

		require.Equal(t, 2, eventTypeV2.Version)
	})

	t.Run("error_saving_event_type_version", func(t *testing.T) {
		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		err := repo.SaveVersion(ctx, "user.created", []string{"user"}, []byte(""))
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_RollbackVersion(t *testing.T) {
	t.Run("rollback_event_type_version", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]

		for i := 0; i < 3; i++ {
			err := repo.SaveVersion(ctx, eventType.Action(), eventType.TargetTypes(), eventType.Schema())
			require.NoError(t, err)
		}

		eventTypeV4, err := repo.FindByAction(ctx, eventType.Action())
		require.NoError(t, err)
		require.Equal(t, 4, eventTypeV4.Version)

		for i := 0; i < 3; i++ {
			err := repo.RollbackVersion(ctx, eventType.Action())
			require.NoError(t, err)
		}

		eventTypeV1, err := repo.FindByAction(ctx, eventType.Action())
		require.NoError(t, err)
		require.Equal(t, 1, eventTypeV1.Version)
	})

	t.Run("error_rolling_back_event_type_version_1", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]

		err := repo.RollbackVersion(ctx, eventType.Action())
		require.ErrorIs(t, err, domain.ErrVersionOneOfEventTypeCannotBeRolledBack)
	})

	t.Run("error_rolling_back_event_type", func(t *testing.T) {
		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		err := repo.RollbackVersion(ctx, "user.created")
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestEventTypes_AllVersionsByAction(t *testing.T) {
	t.Run("get_all_versions_of_an_event_type", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes, _ := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]
		totalVersionsExpected := 4

		for i := 0; i < totalVersionsExpected-1; i++ {
			err := repo.SaveVersion(ctx, eventType.Action(), eventType.TargetTypes(), eventType.Schema())
			require.NoError(t, err)
		}

		allEventTypeVersions, err := repo.AllVersionsByAction(ctx, eventType.Action())
		require.NoError(t, err)
		require.Len(t, allEventTypeVersions, 4)

		version := totalVersionsExpected
		for _, et := range allEventTypeVersions {
			require.Equal(t, version, et.Version)
			version--
		}
	})

	t.Run("error_querying_all_event_type_versions", func(t *testing.T) {
		repo := chadapters.NewEventTypesClickhouseRepository(dbError)
		_, err := repo.AllVersionsByAction(ctx, "user.created")
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func seedEventTypes(
	t *testing.T,
	repo *chadapters.EventTypesClickhouseRepository,
	count int,
) ([]domain.EventType, map[string]domain.EventType) {
	eventTypes := make([]domain.EventType, count)
	eventTypesByAction := map[string]domain.EventType{}

	for i := 0; i < count; i++ {
		eventType := fixtureEventType(t)
		err := repo.Save(ctx, eventType)
		require.NoError(t, err)

		eventTypes[i] = eventType
		eventTypesByAction[eventType.Action()] = eventType
	}

	return eventTypes, eventTypesByAction
}

func fixtureEventType(t *testing.T) domain.EventType {
	eventType, err := domain.NewEventType(fmt.Sprintf("test.created.%s", domain.NewID()), gofakeit.Bool(), []string{"test"}, nil)
	require.NoError(t, err)
	return *eventType
}

func requireEqualEventTypes(t *testing.T, expected domain.EventType, got query.EventType) {
	t.Helper()

	require.Equal(t, expected.Action(), got.Action)
	require.Equal(t, expected.TargetTypes(), got.TargetTypes)
	require.Equal(t, expected.Version(), got.Version)
	require.Equal(t, string(expected.Schema()), got.Schema)
	require.Equal(t, expected.ShouldValidateMetadataSchema(), got.ShouldValidateMetadataSchema)
	require.WithinDuration(t, expected.CreatedAt(), got.CreatedAt, time.Second)
}
