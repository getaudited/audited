package psql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/app/query"
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

func TestSourcesPsqlRepository_QueryAll(t *testing.T) {
	repo := psql.NewSourcesPsqlRepository(db)

	source1 := fixtureSource(t)
	source2 := fixtureSource(t)
	source3 := fixtureSource(t)
	storeSource(t, source1)
	storeSource(t, source2)
	storeSource(t, source3)

	t.Run("returns all sources ordered by created_at desc", func(t *testing.T) {
		result, err := repo.QueryAll(ctx, query.AllSources{
			Pagination: query.PaginationParams{Limit: 10, Page: 1},
		})
		require.NoError(t, err)

		ids := make([]string, len(result.Data))
		for i, s := range result.Data {
			ids[i] = s.ID().String()
		}
		assert.Contains(t, ids, source1.ID().String())
		assert.Contains(t, ids, source2.ID().String())
		assert.Contains(t, ids, source3.ID().String())
	})

	t.Run("paginates correctly", func(t *testing.T) {
		page1, err := repo.QueryAll(ctx, query.AllSources{
			Pagination: query.PaginationParams{Limit: 2, Page: 1},
		})
		require.NoError(t, err)
		assert.Len(t, page1.Data, 2)
		assert.Equal(t, 1, page1.CurrentPage)
		assert.Equal(t, 2, page1.PerPage)
		assert.GreaterOrEqual(t, page1.Total, 3)

		page2, err := repo.QueryAll(ctx, query.AllSources{
			Pagination: query.PaginationParams{Limit: 2, Page: 2},
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(page2.Data), 1)
		assert.Equal(t, 2, page2.CurrentPage)

		page1IDs := make([]string, len(page1.Data))
		for i, s := range page1.Data {
			page1IDs[i] = s.ID().String()
		}
		for _, s := range page2.Data {
			assert.NotContains(t, page1IDs, s.ID().String())
		}
	})

	t.Run("filters by name (case-insensitive partial match)", func(t *testing.T) {
		needle := "zz-unique-needle-zz"
		named, err := domain.NewSource("svc-" + needle + "-svc")
		require.NoError(t, err)
		storeSource(t, named)

		result, err := repo.QueryAll(ctx, query.AllSources{
			Name:       new(needle),
			Pagination: query.PaginationParams{Limit: 10, Page: 1},
		})
		require.NoError(t, err)
		require.Len(t, result.Data, 1)
		assert.Equal(t, named.ID().String(), result.Data[0].ID().String())
	})

	t.Run("returns empty result when no sources match the name filter", func(t *testing.T) {
		name := "this-name-does-not-exist-anywhere"
		result, err := repo.QueryAll(ctx, query.AllSources{
			Name:       new(name),
			Pagination: query.PaginationParams{Limit: 10, Page: 1},
		})
		require.NoError(t, err)
		assert.Empty(t, result.Data)
	})
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
