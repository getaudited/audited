package clickhouse_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestEventTypes_FindByAction(t *testing.T) {
	t.Run("find_event_by_action", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewEventTypesClickhouseRepository(db)
		eventTypes := seedEventTypes(t, repo, 1)
		eventType := eventTypes[0]

		found, err := repo.FindByAction(ctx, eventType.Action)
		require.NoError(t, err)

		require.Equal(t, eventType.Action, found.Action)
		require.Equal(t, eventType.TargetTypes, found.TargetTypes)
		require.Equal(t, eventType.Version, found.Version)
		require.Equal(t, eventType.ShouldValidateMetadataSchema, found.ShouldValidateMetadataSchema)
		require.Equal(t, string(eventType.Schema), found.Schema)
		require.WithinDuration(t, eventType.CreatedAt, found.CreatedAt, time.Second)
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
func TestEventTypes_QueryAll(t *testing.T)            {}
func TestEventTypes_Delete(t *testing.T)              {}
func TestEventTypes_Save(t *testing.T)                {}
func TestEventTypes_RollbackVersion(t *testing.T)     {}
func TestEventTypes_SaveVersion(t *testing.T)         {}
func TestEventTypes_AllVersionsByAction(t *testing.T) {}

func seedEventTypes(t *testing.T, repo *chadapters.EventTypesClickhouseRepository, count int) []domain.EventType {
	eventTypes := make([]domain.EventType, count)
	for i := 0; i < count; i++ {
		eventType := domain.EventType{
			Action:                       fmt.Sprintf("test.created.%s", domain.NewID()),
			ShouldValidateMetadataSchema: gofakeit.Bool(),
			Version:                      1,
			TargetTypes:                  []string{"test"},
			Schema:                       nil,
			LastVersion: domain.EventTypeVersion{
				Version:     1,
				TargetTypes: []string{"test"},
				Schema:      nil,
				CreatedAt:   time.Now(),
			},
			CreatedAt: time.Now(),
		}
		err := repo.Save(ctx, eventType)
		require.NoError(t, err)

		eventTypes[i] = eventType
	}

	return eventTypes
}
