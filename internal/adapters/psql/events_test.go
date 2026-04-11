package psql_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"

	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/adapters/psql"
	"github.com/firminochangani/audited/internal/domain"
)

func TestEventsPsqlRepository_Save(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// GIVEN
	source := fixtureSource(t)
	storeSource(t, source)

	event := fixtureEvent(t, source.ID())

	// WHEN
	err := repo.Save(context.Background(), event)
	require.NoError(t, err)

	// THEN
	storedEvent := queryEventByID(t, event.ID())
	require.NotNil(t, storedEvent)

	require.Equal(t, event.ID().String(), storedEvent.ID)
	require.Equal(t, event.Version(), storedEvent.Version)
	require.WithinDuration(t, event.OccurredAt(), storedEvent.OccurredAt, time.Millisecond)

	require.Equal(t, event.Actor().Id, storedEvent.ActorID)
	require.Equal(t, event.Actor().ActorType, storedEvent.ActorType)
	require.Equal(t, event.Actor().Name, storedEvent.ActorName.Ptr())
	requireEqualMetadata(t, event.Actor().Metadata, storedEvent.ActorMetadata)

	require.Equal(t, event.Context().Location, storedEvent.ContextLocation)
	require.Equal(t, event.Context().UserAgent, storedEvent.ContextUserAgent.Ptr())

	requireEqualMetadata(t, event.Metadata(), storedEvent.Metadata)

	require.Len(t, storedEvent.R.EventTargets, len(event.Targets()))

	for _, target := range event.Targets() {
		storedTarget := findStoredTarget(storedEvent.R.EventTargets, target.Id)
		require.NotNilf(t, storedTarget, "target %s not found in stored event", target.Id)
		require.Equal(t, target.Id, storedTarget.ID)
		require.Equal(t, event.ID().String(), storedTarget.EventID)
		require.Equal(t, target.TargetType, storedTarget.Type)
		require.Equal(t, target.Name, storedTarget.Name.Ptr())
		requireEqualMetadata(t, target.Metadata, storedTarget.Metadata)
	}
}

func TestEventsPsqlRepository_SaveErrSourceNotFoundWhileSavingEvent(t *testing.T) {
	repo := psql.NewEventsPsqlRepository(db)

	// GIVEN
	nonExistentSourceID := domain.NewID()

	event := fixtureEvent(t, nonExistentSourceID)

	// WHEN
	err := repo.Save(context.Background(), event)

	// THEN
	require.ErrorIs(t, err, domain.ErrSourceNotFoundWhileSavingEvent)
}

func requireEqualMetadata(t *testing.T, expected *domain.Metadata, actual null.JSON) {
	t.Helper()

	if expected == nil {
		require.False(t, actual.Valid)
		return
	}

	require.True(t, actual.Valid)

	var actualMap domain.Metadata
	require.NoError(t, json.Unmarshal(actual.JSON, &actualMap))
	require.Equal(t, *expected, actualMap)
}

func findStoredTarget(targets []*models.EventTarget, id string) *models.EventTarget {
	for _, t := range targets {
		if t.ID == id {
			return t
		}
	}

	return nil
}

func storeSource(t *testing.T, source *domain.Source) {
	repo := psql.NewSourcesPsqlRepository(db)
	require.NoError(t, repo.Save(ctx, source))
}

func queryEventByID(t *testing.T, eventID domain.ID) *models.Event {
	row, err := models.Events(
		models.EventWhere.ID.EQ(eventID.String()),
		qm.Load(models.EventRels.EventTargets),
	).One(ctx, db)
	require.NoError(t, err)

	return row
}

func fixtureEvent(t *testing.T, sourceID domain.ID) domain.Event {
	event, err := domain.NewEvent(
		sourceID,
		1,
		domain.Actor{
			Id:        ulid.Make().String(),
			ActorType: "user",
			Name:      new(gofakeit.Name()),
			Metadata: &domain.Metadata{
				"role": "admin",
			},
		},
		[]domain.Target{
			{
				Id:         ulid.Make().String(),
				TargetType: "user",
				Name:       new(gofakeit.Name()),
				Metadata: &domain.Metadata{
					"role": "admin",
				},
			},
			{
				Id:         ulid.Make().String(),
				TargetType: "team",
				Name:       new(gofakeit.Name()),
				Metadata: &domain.Metadata{
					"owner_id": ulid.Make().String(),
				},
			},
		},
		domain.Context{
			Location:  gofakeit.IPv4Address(),
			UserAgent: new(gofakeit.UserAgent()),
		},
		&domain.Metadata{
			"user_id": ulid.Make().String(),
		},
		time.Now(),
	)
	require.NoError(t, err)

	return event
}
