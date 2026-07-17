package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/common/testhelpers"
	"github.com/getaudited/audited/internal/domain"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		name        string
		email       domain.Email
		password    domain.Password
		role        domain.UserRole
		expectedErr string
	}{
		{
			name:     "new_user",
			email:    testhelpers.MustEmail(t),
			password: testhelpers.MustPassword(t),
			role:     domain.UserRoleAdmin,
		},
		{
			name:        "error_missing_email",
			email:       domain.Email{},
			password:    testhelpers.MustPassword(t),
			role:        domain.UserRoleAdmin,
			expectedErr: "email cannot be empty",
		},
		{
			name:        "error_missing_password",
			email:       testhelpers.MustEmail(t),
			password:    domain.Password{},
			role:        domain.UserRoleAdmin,
			expectedErr: "password cannot be empty",
		},
		{
			name:        "error_missing_role",
			email:       testhelpers.MustEmail(t),
			password:    testhelpers.MustPassword(t),
			role:        "",
			expectedErr: "user role must be 'admin' or 'member'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user, err := domain.NewUser(tc.email, tc.password, tc.role)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.Nil(t, user)
				return
			}

			require.False(t, user.ID().Empty())
			require.Equal(t, tc.email, user.Email())
			require.Equal(t, tc.password, user.Password())
			require.Equal(t, tc.role, user.Role())
			require.True(t, user.CreatedAt().Before(time.Now()))
		})
	}
}

func TestMarshallToUser(t *testing.T) {
	id := domain.NewID()
	email := testhelpers.MustEmail(t)
	password := testhelpers.MustPassword(t)
	role := domain.UserRoleAdmin
	createdAt := time.Now()

	user := domain.MarshallToUser(
		id.String(),
		email.String(),
		password.String(),
		role.String(),
		createdAt,
	)
	require.NotNil(t, user)
	require.Equal(t, id, user.ID())
	require.Equal(t, email, user.Email())
	require.Equal(t, password, user.Password())
	require.Equal(t, role, user.Role())
	require.WithinDuration(t, createdAt, user.CreatedAt(), time.Second)
}

func TestNewEmail(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		expectedErr string
	}{
		{
			name:  "new_email",
			email: "john.doe@example.com",
		},
		{
			name:        "error_invalid_email",
			email:       "john.doe example.com",
			expectedErr: "email is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			email, err := domain.NewEmail(tc.email)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.True(t, email.Empty())
				return
			}

			require.Equal(t, tc.email, email.String())
		})
	}
}

func TestNewPassword(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		expectedErr string
	}{
		{
			name:     "new_password",
			password: "super-secure-password",
		},
		{
			name:        "error_invalid_password",
			password:    "",
			expectedErr: "password cannot be empty",
		},
		{
			name:        "error_invalid_password",
			password:    "",
			expectedErr: "password cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			password, err := domain.NewPassword(tc.password)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.True(t, password.Empty())
				return
			}

			require.True(t, password.IsEqual(tc.password))
		})
	}
}
