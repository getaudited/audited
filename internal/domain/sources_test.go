package domain_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewSource(t *testing.T) {
	t.Run("create_new_source", func(t *testing.T) {
		t.Parallel()

		sourceName := gofakeit.AppName()
		source, err := domain.NewSource(sourceName)
		require.NoError(t, err)
		require.NotNil(t, source)
		require.Equal(t, sourceName, source.Name())
		require.NotEmpty(t, sourceName, source.ID())
		require.True(t, source.CreatedAt().Before(time.Now()))
		require.True(t, source.UpdatedAt().Before(time.Now()))
	})

	t.Run("error_missing_name", func(t *testing.T) {
		t.Parallel()

		source, err := domain.NewSource("")
		require.Error(t, err)
		require.Nil(t, source)
	})
}

func TestMarshallToSource(t *testing.T) {
	id := domain.NewID().String()
	name := gofakeit.AppName()
	createdAt := time.Now()
	updatedAt := time.Now()
	source := domain.MarshallToSource(id, name, createdAt, updatedAt)
	require.Equal(t, id, source.ID().String())
	require.Equal(t, name, source.Name())
	require.WithinDuration(t, createdAt, source.CreatedAt(), time.Second)
	require.WithinDuration(t, updatedAt, source.UpdatedAt(), time.Second)
}
