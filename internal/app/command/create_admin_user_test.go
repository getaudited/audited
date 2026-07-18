package command_test

import (
	"context"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/common/testhelpers"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestCreateAdminUser(t *testing.T) {
	email := testhelpers.MustEmail(t)
	password := testhelpers.MustPassword(t)
	existingUser, newUserErr := domain.NewUser(email, password, domain.UserRoleAdmin)
	require.NoError(t, newUserErr)

	testCases := []struct {
		name        string
		userRepo    domain.UserRepository
		email       string
		password    string
		expectedErr string
	}{
		{
			name: "new_admin_user",
			userRepo: &mockUserRepository{
				mu:    &sync.RWMutex{},
				users: make(map[string]*domain.User),
			},
			email:       gofakeit.Email(),
			password:    testhelpers.MustPassword(t).String(),
			expectedErr: "",
		},
		{
			name:        "error_invalid_email",
			userRepo:    nil,
			email:       "invalid-email",
			password:    testhelpers.MustPassword(t).String(),
			expectedErr: "email is invalid",
		},
		{
			name:        "error_invalid_password",
			userRepo:    nil,
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password cannot be empty",
		},
		{
			name: "error_admin_user_exists",
			userRepo: &mockUserRepository{
				mu: &sync.RWMutex{},
				users: map[string]*domain.User{
					existingUser.ID().String(): existingUser,
				},
			},
			email:       email.String(),
			password:    "some-password",
			expectedErr: domain.ErrUserExists.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := command.NewCreateAdminUserHandler(tc.userRepo)
			err := handler.Execute(context.Background(), command.CreateAdminUser{
				Email:    tc.email,
				Password: tc.password,
			})

			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

type mockUserRepository struct {
	users map[string]*domain.User
	mu    *sync.RWMutex
}

func (m *mockUserRepository) FindByEmail(_ context.Context, email domain.Email) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email().String() == email.String() {
			return user, nil
		}
	}

	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByID(_ context.Context, id domain.ID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id.String()]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (m *mockUserRepository) Save(_ context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[user.ID().String()] = user

	return nil
}
