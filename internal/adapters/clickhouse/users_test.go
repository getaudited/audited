package clickhouse_test

import (
	"testing"
	"time"

	"github.com/getaudited/audited/internal/common/testhelpers"
	"github.com/stretchr/testify/require"

	chadapters "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/domain"
)

func TestUsers_Save(t *testing.T) {
	t.Run("save_user", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)
		user, err := domain.NewUser(testhelpers.MustEmail(t), testhelpers.MustPassword(t), domain.UserRoleAdmin)
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByEmail(ctx, user.Email())
		require.NoError(t, err)

		requireEqualUsers(t, user, found)
	})

	t.Run("error_user_exists", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)
		user, err := domain.NewUser(testhelpers.MustEmail(t), testhelpers.MustPassword(t), domain.UserRoleAdmin)
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.NoError(t, err)

		// Attempt to save the same user
		err = repo.Save(ctx, user)
		require.ErrorIs(t, err, domain.ErrUserExists)
	})

	t.Run("error_saving_user", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(dbError)
		user, err := domain.NewUser(testhelpers.MustEmail(t), testhelpers.MustPassword(t), domain.UserRoleAdmin)
		require.NoError(t, err)

		err = repo.Save(ctx, user)
		require.ErrorIs(t, err, errMockedClickhouse)
	})
}

func TestUsers_FindByEmail(t *testing.T) {
	t.Run("find_user_by_email", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)
		user := seedUser(t, repo)

		found, err := repo.FindByEmail(ctx, user.Email())
		require.NoError(t, err)

		requireEqualUsers(t, user, found)
	})

	t.Run("error_user_not_found", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)

		found, err := repo.FindByEmail(ctx, testhelpers.MustEmail(t))
		require.ErrorIs(t, err, domain.ErrUserNotFound)
		require.Nil(t, found)
	})

	t.Run("error_querying_user", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(dbError)

		found, err := repo.FindByEmail(ctx, testhelpers.MustEmail(t))
		require.ErrorIs(t, err, errMockedClickhouse)
		require.Nil(t, found)
	})
}

func TestUsers_FindByID(t *testing.T) {
	t.Run("find_user_by_id", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)
		user := seedUser(t, repo)

		found, err := repo.FindByID(ctx, user.ID())
		require.NoError(t, err)
		requireEqualUsers(t, user, found)
	})

	t.Run("user_not_found", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(db)

		found, err := repo.FindByID(ctx, domain.NewID())
		require.ErrorIs(t, err, domain.ErrUserNotFound)
		require.Nil(t, found)
	})

	t.Run("error_querying_user", func(t *testing.T) {
		t.Parallel()

		repo := chadapters.NewUsersClickhouseRepository(dbError)

		found, err := repo.FindByID(ctx, domain.NewID())
		require.ErrorIs(t, err, errMockedClickhouse)
		require.Nil(t, found)
	})
}

func requireEqualUsers(t *testing.T, expected, got *domain.User) {
	t.Helper()

	require.Equal(t, expected.ID(), got.ID())
	require.Equal(t, expected.Email(), got.Email())
	require.Equal(t, expected.Role(), got.Role())
	require.Equal(t, expected.Password(), got.Password())
	require.WithinDuration(t, expected.CreatedAt(), got.CreatedAt(), time.Second)
}

func seedUser(t *testing.T, repo chadapters.UsersClickhouseRepository) *domain.User {
	user, err := domain.NewUser(testhelpers.MustEmail(t), testhelpers.MustPassword(t), domain.UserRoleAdmin)
	require.NoError(t, err)

	err = repo.Save(ctx, user)
	require.NoError(t, err)

	return user
}
