package domain_test

import (
	"testing"

	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	id := domain.NewID()
	require.NotEmpty(t, id.String())
	require.False(t, id.Empty())
}
