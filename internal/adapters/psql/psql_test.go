package psql_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/common/postgres"
	"github.com/getaudited/audited/internal/domain"
	"github.com/getaudited/audited/misc/tools/waitfor"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

var (
	db  *sql.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	var err error
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	// Wait for postgres and other dependencies running in containers
	waitFor := waitfor.NewWaitFor(logs.New("DEBUG"))

	waitFor.Do(func() error {
		db, err = postgres.Connect(ctx, strings.Replace(os.Getenv("ADT_DATABASE_URL"), "@postgres", "@localhost", 1))
		if err != nil {
			return err
		}

		return db.Ping()
	}, "postgres", time.Second*30)
	waitFor.Wait()

	err = postgres.ApplyMigrations(db, "../../../misc/sql/migrations")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
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

//
// Test utils
//

func findStoredTarget(targets []*models.EventTarget, id string) *models.EventTarget {
	for _, t := range targets {
		if t.ID == id {
			return t
		}
	}

	return nil
}

func storeEvent(t *testing.T, event domain.Event, token domain.TokenValue) {
	t.Helper()
	repo := psql.NewEventsPsqlRepository(db)
	require.NoError(t, repo.Save(ctx, event, token))
}

func storeSource(t *testing.T, source *domain.Source) {
	repo := psql.NewSourcesPsqlRepository(db)
	require.NoError(t, repo.Save(ctx, source))
}

func storeToken(t *testing.T, token *domain.Token) {
	repo := psql.NewTokensPsqlRepository(db)
	require.NoError(t, repo.Save(ctx, token))
}

func queryEventByID(t *testing.T, eventID domain.ID) *models.Event {
	row, err := models.Events(
		models.EventWhere.ID.EQ(eventID.String()),
		qm.Load(models.EventRels.EventTargets),
	).One(ctx, db)
	require.NoError(t, err)

	return row
}

func fixtureEventWith(
	t *testing.T,
	eventTypeAction string,
	sourceID domain.ID,
	actor domain.Actor,
	targets []domain.Target,
	occurredAt time.Time,
) domain.Event {
	t.Helper()
	e, err := domain.NewEvent(
		sourceID,
		1,
		eventTypeAction,
		actor,
		targets,
		domain.Context{Location: gofakeit.IPv4Address()},
		nil,
		occurredAt,
	)
	require.NoError(t, err)
	return e
}

func fixtureEvent(t *testing.T, sourceID domain.ID, eventTypeAction string) domain.Event {
	event, err := domain.NewEvent(
		sourceID,
		1,
		eventTypeAction,
		domain.Actor{
			ID:        ulid.Make().String(),
			ActorType: "user",
			Name:      new(gofakeit.Name()),
			Metadata: &domain.Metadata{
				"role": "admin",
			},
		},
		[]domain.Target{
			{
				ID:         ulid.Make().String(),
				TargetType: "user",
				Name:       new(gofakeit.Name()),
				Metadata: &domain.Metadata{
					"role": "admin",
				},
			},
			{
				ID:         ulid.Make().String(),
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
