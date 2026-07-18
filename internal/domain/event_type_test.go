package domain_test

import (
	"testing"
	"time"

	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewEventType(t *testing.T) {
	testCases := []struct {
		name           string
		action         string
		validateSchema bool
		targetTypes    []string
		schema         domain.Schema
		expectedErr    string
	}{
		{
			name:           "new_event_type",
			action:         "user.created",
			validateSchema: false,
			targetTypes:    []string{"user"},
			schema:         nil,
		},
		{
			name:           "error_missing_action",
			action:         "",
			validateSchema: false,
			targetTypes:    []string{"user"},
			schema:         nil,
			expectedErr:    "action cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eventType, err := domain.NewEventType(tc.action, tc.validateSchema, tc.targetTypes, tc.schema)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.action, eventType.Action())
			require.Equal(t, tc.targetTypes, eventType.TargetTypes())
			require.Equal(t, 1, eventType.Version())
			require.Equal(t, tc.validateSchema, eventType.ShouldValidateMetadataSchema())
			require.Equal(t, tc.schema, eventType.Schema())
			require.True(t, eventType.CreatedAt().Before(time.Now()))
		})
	}
}

func TestNewEvent_WithDuplicateTargetTypes(t *testing.T) {
	eventType, err := domain.NewEventType("user.created", false, []string{"user", "account", "user"}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"user", "account"}, eventType.TargetTypes())
}
