package clickhouse_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestTokens_Save(t *testing.T) {
	token, err := domain.NewToken(domain.NewID(), gofakeit.AppName())
	require.NoError(t, err)

	t.Run("save_token", func(t *testing.T) {
		repo := chadapters.NewTokenChRepository(db)
		err = repo.Save(ctx, token)
		require.NoError(t, err)
	})

	t.Run("error_saving_token", func(t *testing.T) {
		repo := chadapters.NewTokenChRepository(dbError)
		err = repo.Save(ctx, token)
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestTokens_QueryAll(t *testing.T) {
	t.Run("query_all", func(t *testing.T) {
		seedCount := 20
		repo := chadapters.NewTokenChRepository(db)
		tokens, tokensByID := seedTokens(t, repo, seedCount)

		result, err := repo.QueryAll(ctx, tokens[0].SourceID())
		require.NoError(t, err)
		require.Len(t, result, seedCount)

		for _, token := range result {
			found, exists := tokensByID[token.ID()]
			require.True(t, exists)
			requireEqualToken(t, found, token)
		}
	})

	t.Run("error_querying_all", func(t *testing.T) {
		repo := chadapters.NewTokenChRepository(dbError)
		_, err := repo.QueryAll(ctx, domain.NewID())
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestTokens_Delete(t *testing.T) {
	t.Run("delete_token", func(t *testing.T) {
		repo := chadapters.NewTokenChRepository(db)
		tokens, _ := seedTokens(t, repo, 1)

		err := repo.Delete(ctx, tokens[0].ID(), tokens[0].SourceID())
		require.NoError(t, err)
	})

	t.Run("error_delete_token", func(t *testing.T) {
		repo := chadapters.NewTokenChRepository(dbError)
		err := repo.Delete(ctx, domain.NewID(), domain.NewID())
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func seedTokens(t *testing.T, repo *chadapters.TokenChRepository, count int) ([]*domain.Token, map[domain.ID]*domain.Token) {
	tokens := make([]*domain.Token, count)
	tokensByID := make(map[domain.ID]*domain.Token)

	sourceID := domain.NewID()
	for i := 0; i < count; i++ {
		token, err := domain.NewToken(sourceID, gofakeit.AppName())
		require.NoError(t, err)

		err = repo.Save(ctx, token)
		require.NoError(t, err)

		tokens[i] = token
		tokensByID[token.ID()] = token
	}

	return tokens, tokensByID
}

func requireEqualToken(t *testing.T, expected, got *domain.Token) {
	t.Helper()

	require.Equal(t, expected.ID(), got.ID())
	require.Equal(t, expected.Name(), got.Name())
	require.Equal(t, expected.SourceID(), got.SourceID())
	require.Equal(t, expected.Value(), got.Value())
	require.WithinDuration(t, expected.CreatedAt(), got.CreatedAt(), time.Second)
}
