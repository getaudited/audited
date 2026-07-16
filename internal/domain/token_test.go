package domain_test

import (
	"testing"
	"time"

	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	testCases := []struct {
		name        string
		sourceID    domain.ID
		tokenName   string
		expectedErr string
	}{
		{
			name:      "create_token",
			sourceID:  domain.NewID(),
			tokenName: "svc-users-token",
		},
		{
			name:        "error_missing_source_id",
			sourceID:    "",
			tokenName:   "svc-users-token",
			expectedErr: "sourceID cannot be empty",
		},
		{
			name:        "error_missing_name",
			sourceID:    domain.NewID(),
			tokenName:   "",
			expectedErr: "name cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := domain.NewToken(tc.sourceID, tc.tokenName)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.Nil(t, token)
				return
			}

			require.Equal(t, tc.tokenName, token.Name())
			require.Equal(t, tc.sourceID, token.SourceID())
			require.True(t, token.CreatedAt().Before(time.Now()))
			require.NotEmpty(t, token.Value())
			require.NotEqual(t, token.Value(), token.MaskedValue())
			require.Contains(t, token.MaskedValue(), "****")
		})
	}
}

func TestMarshallToToken(t *testing.T) {
	id := domain.NewID().String()
	sourceID := domain.NewID().String()
	tokenValue := "**********TOKEN"
	name := "svc-users-token"
	createdAt := time.Now()

	token := domain.MarshallToToken(id, sourceID, tokenValue, name, createdAt)
	require.NotNil(t, token)
	require.Equal(t, id, token.ID().String())
	require.Equal(t, sourceID, token.SourceID().String())
	require.Equal(t, tokenValue, token.Value().String())
	require.Equal(t, name, token.Name())
	require.WithinDuration(t, createdAt, token.CreatedAt(), time.Second)
}
