package testhelpers

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/domain"
)

func MustEmail(t *testing.T) domain.Email {
	email, err := domain.NewEmail(fmt.Sprintf("%s.%s", domain.NewID().String(), gofakeit.Email()))
	require.NoError(t, err)
	return email
}

func MustPassword(t *testing.T) domain.Password {
	password, err := domain.NewPassword(gofakeit.Password(
		true,
		true,
		true,
		true,
		false,
		12,
	))
	require.NoError(t, err)
	return password
}
