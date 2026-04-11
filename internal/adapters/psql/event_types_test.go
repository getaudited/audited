package psql_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"

	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/adapters/psql"
	"github.com/firminochangani/audited/internal/domain"
)

func TestEventTypePsqlRepository_Save(t *testing.T) {
	repo := psql.NewEventTypePsqlRepository(db)

	// GIVEN
	et := fixtureEventType()

	// WHEN
	err := repo.Save(ctx, et)
	require.NoError(t, err)

	// THEN
	stored := queryEventTypeByID(t, et.Id)
	require.NotNil(t, stored)

	require.Equal(t, et.Id, stored.ID)
	require.Equal(t, et.Version, stored.Version)
	require.Equal(t, et.Action, stored.Action)
	require.Equal(t, et.TargetTypes, []string(stored.TargetTypes))
	require.Equal(t, et.ShouldValidateMetadataSchema, stored.ShouldValidateMetadataSchema)
	require.False(t, stored.EventSchema.Valid)
	require.WithinDuration(t, et.CreatedAt, stored.CreatedAt, time.Millisecond)
	require.WithinDuration(t, et.UpdatedAt, stored.UpdatedAt, time.Millisecond)
}

func TestEventTypePsqlRepository_FindByAction(t *testing.T) {
	repo := psql.NewEventTypePsqlRepository(db)

	t.Run("found", func(t *testing.T) {
		// GIVEN
		et := fixtureEventType()
		require.NoError(t, repo.Save(ctx, et))

		// WHEN
		result, err := repo.FindByAction(ctx, et.Action)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, et.Id, result.Id)
		require.Equal(t, et.Version, result.Version)
		require.Equal(t, et.Action, result.Action)
		require.Equal(t, et.TargetTypes, result.TargetTypes)
		require.Equal(t, et.ShouldValidateMetadataSchema, result.ShouldValidateMetadataSchema)
		require.WithinDuration(t, et.CreatedAt, result.CreatedAt, time.Millisecond)
		require.WithinDuration(t, et.UpdatedAt, result.UpdatedAt, time.Millisecond)
	})

	t.Run("not found", func(t *testing.T) {
		// WHEN
		result, err := repo.FindByAction(ctx, gofakeit.Word())

		// THEN
		require.ErrorIs(t, err, domain.ErrEventTypeNotFound)
		require.Nil(t, result)
	})
}

func queryEventTypeByID(t *testing.T, id string) *models.EventType {
	t.Helper()

	row, err := models.EventTypes(
		models.EventTypeWhere.ID.EQ(id),
	).One(ctx, db)
	require.NoError(t, err)

	return row
}

func fixtureEventType() domain.EventType {
	now := time.Now()
	return domain.EventType{
		Id:                           ulid.Make().String(),
		Version:                      1,
		Action:                       gofakeit.Word(),
		TargetTypes:                  []string{"user", "team"},
		ShouldValidateMetadataSchema: false,
		Schema:                       nil,
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}
}
