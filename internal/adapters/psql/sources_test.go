package psql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/domain"
)

func TestSourcesPsqlRepository_Save(t *testing.T) {
	repo := psql.NewSourcesPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)

	// WHEN
	err := repo.Save(ctx, source)
	require.NoError(t, err)

	// THEN
	stored := querySourceByID(t, source.ID().String())
	require.NotNil(t, stored)

	require.Equal(t, source.ID().String(), stored.ID)
	require.Equal(t, source.Name(), stored.Name)
	require.WithinDuration(t, source.CreatedAt(), stored.CreatedAt, time.Millisecond)
	require.WithinDuration(t, source.UpdatedAt(), stored.UpdatedAt, time.Millisecond)
}

func querySourceByID(t *testing.T, id string) *models.Source {
	t.Helper()

	row, err := models.Sources(
		models.SourceWhere.ID.EQ(id),
	).One(ctx, db)
	require.NoError(t, err)

	return row
}

func fixtureSource(t *testing.T) *domain.Source {
	t.Helper()

	source, err := domain.NewSource(fmt.Sprintf("svc-%s-%s", gofakeit.AppName(), gofakeit.AppVersion()))
	require.NoError(t, err)

	return source
}
