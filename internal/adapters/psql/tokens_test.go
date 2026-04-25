package psql_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/adapters/psql"
	"github.com/firminochangani/audited/internal/domain"
)

func TestTokensPsqlRepository_Save(t *testing.T) {
	repo := psql.NewTokensPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)
	storeSource(t, source)

	token := fixtureToken(t, source.ID())

	// WHEN
	err := repo.Save(ctx, token)
	require.NoError(t, err)

	// THEN
	stored := queryTokenByID(t, token.ID().String())
	require.NotNil(t, stored)

	require.Equal(t, token.ID().String(), stored.ID)
	require.Equal(t, token.Value().String(), stored.Value)
	require.Equal(t, token.SourceID().String(), stored.SourceID)
	require.Equal(t, token.Name(), stored.Name)
	require.WithinDuration(t, token.CreatedAt(), stored.CreatedAt, time.Millisecond)
}

func queryTokenByID(t *testing.T, id string) *models.Token {
	t.Helper()

	row, err := models.Tokens(
		models.TokenWhere.ID.EQ(id),
	).One(ctx, db)
	require.NoError(t, err)

	return row
}

func TestTokensPsqlRepository_Delete(t *testing.T) {
	repo := psql.NewTokensPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)
	storeSource(t, source)

	token := fixtureToken(t, source.ID())
	require.NoError(t, repo.Save(ctx, token))

	// WHEN
	err := repo.Delete(ctx, token.ID(), token.SourceID())
	require.NoError(t, err)

	// THEN
	exists, err := models.TokenExists(ctx, db, token.ID().String())
	require.NoError(t, err)
	require.False(t, exists)
}

func fixtureToken(t *testing.T, sourceID domain.ID) domain.Token {
	t.Helper()

	token, err := domain.NewToken(sourceID, gofakeit.Word())
	require.NoError(t, err)

	return *token
}
