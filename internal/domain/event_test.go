package domain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/domain"
	"github.com/stretchr/testify/require"
)

type eventTestCase struct {
	name        string
	sourceID    domain.ID
	version     int
	action      string
	actor       domain.Actor
	targets     []domain.Target
	context     domain.Context
	metadata    *domain.Metadata
	occurredAt  time.Time
	expectedErr string
}

func TestNewEvent(t *testing.T) {
	testCases := []eventTestCase{
		{
			name:     "new_event",
			sourceID: domain.NewID(),
			version:  1,
			action:   "user.created",
			actor: domain.Actor{
				ID:        domain.NewID().String(),
				ActorType: "user",
				Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
				Metadata: new(map[string]interface{}{
					"user_role": "admin",
				}),
			},
			targets: []domain.Target{
				{
					ID:         domain.NewID().String(),
					Name:       new(gofakeit.AppName()),
					TargetType: "account",
					Metadata: new(map[string]interface{}{
						"prop": "value",
					}),
				},
			},
			context: domain.Context{
				Location:  gofakeit.IPv4Address(),
				UserAgent: new(gofakeit.UserAgent()),
			},
			metadata: new(map[string]interface{}{
				"prop": "value",
			}),
			occurredAt:  time.Now(),
			expectedErr: "",
		},
		{
			name:     "error_empty_source_id",
			sourceID: "",
			version:  1,
			action:   "user.created",
			actor: domain.Actor{
				ID:        domain.NewID().String(),
				ActorType: "user",
				Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
				Metadata: new(map[string]interface{}{
					"user_role": "admin",
				}),
			},
			targets: []domain.Target{
				{
					ID:         domain.NewID().String(),
					Name:       new(gofakeit.AppName()),
					TargetType: "account",
					Metadata: new(map[string]interface{}{
						"prop": "value",
					}),
				},
			},
			context: domain.Context{
				Location:  gofakeit.IPv4Address(),
				UserAgent: new(gofakeit.UserAgent()),
			},
			metadata: new(map[string]interface{}{
				"prop": "value",
			}),
			occurredAt:  time.Now(),
			expectedErr: "sourceID cannot be empty",
		},
		{
			name:     "error_invalid_version",
			sourceID: domain.NewID(),
			version:  0,
			action:   "user.created",
			actor: domain.Actor{
				ID:        domain.NewID().String(),
				ActorType: "user",
				Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
				Metadata: new(map[string]interface{}{
					"user_role": "admin",
				}),
			},
			targets: []domain.Target{
				{
					ID:         domain.NewID().String(),
					Name:       new(gofakeit.AppName()),
					TargetType: "account",
					Metadata: new(map[string]interface{}{
						"prop": "value",
					}),
				},
			},
			context: domain.Context{
				Location:  gofakeit.IPv4Address(),
				UserAgent: new(gofakeit.UserAgent()),
			},
			metadata: new(map[string]interface{}{
				"prop": "value",
			}),
			occurredAt:  time.Now(),
			expectedErr: "version cannot be less than 1",
		},
		{
			name:     "error_empty_action",
			sourceID: domain.NewID(),
			version:  1,
			action:   "",
			actor: domain.Actor{
				ID:        domain.NewID().String(),
				ActorType: "user",
				Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
				Metadata: new(map[string]interface{}{
					"user_role": "admin",
				}),
			},
			targets: []domain.Target{
				{
					ID:         domain.NewID().String(),
					Name:       new(gofakeit.AppName()),
					TargetType: "account",
					Metadata: new(map[string]interface{}{
						"prop": "value",
					}),
				},
			},
			context: domain.Context{
				Location:  gofakeit.IPv4Address(),
				UserAgent: new(gofakeit.UserAgent()),
			},
			metadata: new(map[string]interface{}{
				"prop": "value",
			}),
			occurredAt:  time.Now(),
			expectedErr: "action cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			event, err := domain.NewEvent(
				tc.sourceID,
				tc.version,
				tc.action,
				tc.actor,
				tc.targets,
				tc.context,
				tc.metadata,
				tc.occurredAt,
			)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			require.False(t, event.ID().Empty())
			require.Equal(t, tc.action, event.Action())
			require.Equal(t, tc.sourceID, event.SourceID())
			require.Equal(t, tc.version, event.Version())
			require.Equal(t, tc.actor, event.Actor())
			require.Equal(t, tc.targets, event.Targets())
			require.Equal(t, tc.context, event.Context())
			require.Equal(t, tc.metadata, event.Metadata())
			require.Equal(t, tc.occurredAt, event.OccurredAt())
		})
	}
}

func TestMarshallToEvent(t *testing.T) {
	id := domain.NewID()
	sourceID := domain.NewID()
	action := "user.created"
	version := 1
	actor := fixtureActor()
	targets := []domain.Target{
		fixtureTarget(),
	}
	context := fixtureContext()
	metadata := fixtureMetadata()
	occurredAt := time.Now()

	event := domain.MarshallToEvent(
		id.String(),
		sourceID.String(),
		action,
		version,
		actor,
		targets,
		context,
		metadata,
		occurredAt,
	)

	require.False(t, event.ID().Empty())
	require.Equal(t, action, event.Action())
	require.Equal(t, sourceID, event.SourceID())
	require.Equal(t, version, event.Version())
	require.Equal(t, actor, event.Actor())
	require.Equal(t, targets, event.Targets())
	require.Equal(t, context, event.Context())
	require.Equal(t, metadata, event.Metadata())
	require.Equal(t, occurredAt, event.OccurredAt())
}

func fixtureActor() domain.Actor {
	return domain.Actor{
		ID:        domain.NewID().String(),
		ActorType: "user",
		Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
		Metadata: new(map[string]interface{}{
			"user_role": "admin",
		}),
	}
}

func fixtureTarget() domain.Target {
	return domain.Target{
		ID:         domain.NewID().String(),
		Name:       new(gofakeit.AppName()),
		TargetType: "account",
		Metadata: new(map[string]interface{}{
			"prop": "value",
		}),
	}
}

func fixtureContext() domain.Context {
	return domain.Context{
		Location:  gofakeit.IPv4Address(),
		UserAgent: new(gofakeit.UserAgent()),
	}
}

func fixtureMetadata() *domain.Metadata {
	return new(map[string]interface{}{
		"prop": "value",
	})
}
