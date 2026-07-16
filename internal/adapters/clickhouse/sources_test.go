package clickhouse_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSources_Save(t *testing.T) {
	t.Run("save_new_source", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(db)
		source, err := domain.NewSource(gofakeit.AppName())
		require.NoError(t, err)

		err = repo.Save(ctx, source)
		require.NoError(t, err)
	})

	t.Run("error_saving_source", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(dbError)
		source, err := domain.NewSource(gofakeit.AppName())
		require.NoError(t, err)

		err = repo.Save(ctx, source)
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func TestSources_FindByID(t *testing.T) {
	t.Run("find_source_by_id", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(db)
		source, err := domain.NewSource(gofakeit.AppName())
		require.NoError(t, err)
		err = repo.Save(ctx, source)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, source.ID().String())
		require.NoError(t, err)
		require.Equal(t, source.ID(), found.ID())
		require.Equal(t, source.Name(), found.Name())
		require.WithinDuration(t, source.CreatedAt(), found.CreatedAt(), time.Second)
		require.WithinDuration(t, source.CreatedAt(), found.UpdatedAt(), time.Second)
	})

	t.Run("error_source_not_found", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(db)

		found, err := repo.FindByID(ctx, "non-existent-source-id")
		require.ErrorAs(t, err, &domain.ErrSourceNotFound)
		require.Nil(t, found)
	})

	t.Run("error_querying_source", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(dbError)

		found, err := repo.FindByID(ctx, "non-existent-source-id")
		require.ErrorAs(t, err, &errMockedClickhouse)
		require.Nil(t, found)
	})
}

func TestSources_QueryAll(t *testing.T) {
	t.Run("query_all", func(t *testing.T) {
		t.Parallel()

		limit := 40
		seedCount := limit / 2

		repo := chadapters.NewSourcesClickhouseRepository(db)
		_, seededSources := seedSources(t, repo, seedCount)

		result, err := repo.QueryAll(ctx, query.AllSources{
			Pagination: query.PaginationParams{
				Limit: limit,
				Page:  1,
			},
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result.Data), seedCount)
		require.Equal(t, limit, result.PerPage)
		require.Equal(t, 1, result.CurrentPage)

		var totalFound int
		for _, source := range result.Data {
			_, found := seededSources[source.ID().String()]
			if found {
				totalFound++
			}
		}

		require.Equal(t, seedCount, totalFound)
	})

	t.Run("query_all_with_name_filter", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(db)
		seededSources, _ := seedSources(t, repo, 5)

		result, err := repo.QueryAll(ctx, query.AllSources{
			Name: new(seededSources[0].Name()),
			Pagination: query.PaginationParams{
				Page:  1,
				Limit: 20,
			},
		})
		require.NoError(t, err)
		require.Len(t, result.Data, 1)
		require.Equal(t, 1, result.CurrentPage)
	})

	t.Run("error_query_all", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewSourcesClickhouseRepository(dbError)

		_, err := repo.QueryAll(ctx, query.AllSources{})
		require.ErrorAs(t, err, &errMockedClickhouse)
	})
}

func seedSources(t *testing.T, repo chadapters.SourcesClickhouseRepository, count int) ([]*domain.Source, map[string]*domain.Source) {
	sources := make([]*domain.Source, count)
	sourcesByID := map[string]*domain.Source{}

	for i := 0; i < count; i++ {
		source, err := domain.NewSource(fmt.Sprintf("Testing-%s-%s", gofakeit.AppName(), domain.NewID().String()))
		require.NoError(t, err)
		err = repo.Save(ctx, source)
		require.NoError(t, err)

		sources[i] = source
		sourcesByID[source.ID().String()] = source
	}

	return sources, sourcesByID
}
