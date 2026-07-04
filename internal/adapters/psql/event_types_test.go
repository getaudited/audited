package psql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/domain"
)

/*
	func TestEventTypePsqlRepository_Save(t *testing.T) {
		repo := psql.NewEventTypePsqlRepository(db)

		// GIVEN
		et := fixtureEventType()

		// WHEN
		err := repo.Save(ctx, et)
		require.NoError(t, err)

		// THEN
		stored, storedVersion := queryEventTypeByAction(t, et.Action)
		require.NotNil(t, stored)
		require.NotNil(t, storedVersion)

		require.Equal(t, et.Action, stored.Action)
		require.Equal(t, et.ShouldValidateMetadataSchema, stored.ShouldValidateMetadataSchema)
		require.WithinDuration(t, et.CreatedAt, stored.CreatedAt, time.Millisecond)

		require.Equal(t, et.LastVersion.Version, storedVersion.Version)
		require.Equal(t, et.LastVersion.TargetTypes, []string(storedVersion.TargetTypes))
		require.False(t, storedVersion.EventSchema.Valid)
		require.WithinDuration(t, et.LastVersion.CreatedAt, storedVersion.CreatedAt, time.Millisecond)
	}

	func TestEventTypePsqlRepository_FindByAction(t *testing.T) {
		repo := psql.NewEventTypePsqlRepository(db)

		t.Run("found", func(t *testing.T) {
			// GIVEN
			et := fixtureEventType()
			require.NoError(t, repo.Save(ctx, et))

			// WHEN
			found, err := repo.FindByAction(ctx, et.Action)

			// THEN
			require.NoError(t, err)

			require.Equal(t, et.Action, found.Action)
			require.Equal(t, et.ShouldValidateMetadataSchema, found.ShouldValidateMetadataSchema)
			require.WithinDuration(t, et.CreatedAt, found.CreatedAt, time.Millisecond)

			require.Equal(t, et.LastVersion.Version, found.Version)
			require.Equal(t, et.LastVersion.TargetTypes, found.TargetTypes)
			require.WithinDuration(t, et.LastVersion.CreatedAt, found.CreatedAt, time.Millisecond)
		})

		t.Run("not found", func(t *testing.T) {
			// WHEN
			result, err := repo.FindByAction(ctx, gofakeit.Word())

			// THEN
			require.ErrorIs(t, err, domain.ErrEventTypeNotFound)
			require.Empty(t, result.Action)
		})
	}

	func TestEventTypePsqlRepository_Delete(t *testing.T) {
		repo := psql.NewEventTypePsqlRepository(db)

		t.Run("deletes existing event type", func(t *testing.T) {
			// GIVEN
			et := fixtureEventType()
			require.NoError(t, repo.Save(ctx, et))

			// WHEN
			err := repo.Delete(ctx, et.Action)
			require.NoError(t, err)

			// THEN
			_, err = repo.FindByAction(ctx, et.Action)
			require.ErrorIs(t, err, domain.ErrEventTypeNotFound)
		})

		t.Run("no error when action does not exist", func(t *testing.T) {
			// WHEN
			err := repo.Delete(ctx, gofakeit.Word())

			// THEN
			require.NoError(t, err)
		})
	}

	func TestEventTypePsqlRepository_QueryAll(t *testing.T) {
		repo := psql.NewEventTypePsqlRepository(db)

		et1 := fixtureEventType()
		et2 := fixtureEventType()
		et3 := fixtureEventType()
		storeEventType(t, et1)
		storeEventType(t, et2)
		storeEventType(t, et3)

		t.Run("returns all event types ordered by created_at desc", func(t *testing.T) {
			result, err := repo.QueryAll(ctx, query.AllEventTypes{
				PaginationParams: query.PaginationParams{Limit: 10, Page: 1},
			})
			require.NoError(t, err)

			actions := make([]string, len(result.Data))
			for i, et := range result.Data {
				actions[i] = et.Action
			}
			assert.Contains(t, actions, et1.Action)
			assert.Contains(t, actions, et2.Action)
			assert.Contains(t, actions, et3.Action)

			for i := 1; i < len(result.Data); i++ {
				assert.False(t, result.Data[i].CreatedAt.After(result.Data[i-1].CreatedAt))
			}
		})

		t.Run("paginates correctly", func(t *testing.T) {
			page1, err := repo.QueryAll(ctx, query.AllEventTypes{
				PaginationParams: query.PaginationParams{Limit: 2, Page: 1},
			})
			require.NoError(t, err)
			assert.Len(t, page1.Data, 2)
			assert.Equal(t, 1, page1.CurrentPage)
			assert.Equal(t, 2, page1.PerPage)
			assert.GreaterOrEqual(t, page1.Total, 3)

			page2, err := repo.QueryAll(ctx, query.AllEventTypes{
				PaginationParams: query.PaginationParams{Limit: 2, Page: 2},
			})
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(page2.Data), 1)
			assert.Equal(t, 2, page2.CurrentPage)

			page1Actions := make([]string, len(page1.Data))
			for i, et := range page1.Data {
				page1Actions[i] = et.Action
			}
			for _, et := range page2.Data {
				assert.NotContains(t, page1Actions, et.Action)
			}
		})

		t.Run("filters by action (case-insensitive partial match)", func(t *testing.T) {
			action := fmt.Sprintf("zz_unique_needle_%d_zz", time.Now().UnixNano())
			et := fixtureEventType()
			et.Action = action + "_suffix"
			storeEventType(t, et)

			result, err := repo.QueryAll(ctx, query.AllEventTypes{
				Action:           &action,
				PaginationParams: query.PaginationParams{Limit: 10, Page: 1},
			})
			require.NoError(t, err)
			require.Len(t, result.Data, 1)
			assert.Equal(t, et.Action, result.Data[0].Action)
		})

		t.Run("returns empty result when no event types match the action filter", func(t *testing.T) {
			action := "this-action-does-not-exist-anywhere"
			result, err := repo.QueryAll(ctx, query.AllEventTypes{
				Action:           &action,
				PaginationParams: query.PaginationParams{Limit: 10, Page: 1},
			})
			require.NoError(t, err)
			assert.Empty(t, result.Data)
		})
	}

	func queryEventTypeByAction(t *testing.T, action string) (*models.EventType, *models.EventTypeVersion) {
		t.Helper()

		row, err := models.EventTypes(
			models.EventTypeWhere.Action.EQ(action),
		).One(ctx, db)
		require.NoError(t, err)

		version, err := models.EventTypeVersions(
			models.EventTypeVersionWhere.EventTypeAction.EQ(action),
		).One(ctx, db)
		require.NoError(t, err)

		return row, version
	}
*/
func fixtureEventType() domain.EventType {
	now := time.Now()
	return domain.EventType{
		Action:                       fmt.Sprintf("%s_%d", gofakeit.Word(), time.Now().UnixMilli()),
		ShouldValidateMetadataSchema: false,
		LastVersion: domain.EventTypeVersion{
			Version:     1,
			TargetTypes: []string{"user", "team"},
			Schema:      nil,
			CreatedAt:   now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func storeEventType(t *testing.T, eventType domain.EventType) {
	repo := psql.NewEventTypePsqlRepository(db)
	require.NoError(t, repo.Save(ctx, eventType))
}
