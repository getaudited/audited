package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	clickhouseadapter "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/domain"
)

const maxEvents = 1000

var (
	actorTypes  = []string{"user", "service", "admin", "system"}
	targetTypes = []string{"document", "record", "file", "account", "resource"}
)

// generator holds the command handlers the events generator needs, wired
// directly against the ClickHouse adapters without depending on app.App.
type generator struct {
	createSource    command.CreateSourceHandler
	createEventType command.CreateEventTypeHandler
	createToken     command.CreateTokenHandler
	createEvent     command.CreateEventHandler
}

func main() {
	ctx := context.Background()

	conn, err := clickhouseconn.NewConnection(ctx, os.Getenv("ADT_DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect to clickhouse: %v", err)
	}
	defer func() { _ = conn.Close() }()

	g := &generator{
		createSource:    command.NewCreateSourceHandler(clickhouseadapter.NewSourcesClickhouseRepository(conn)),
		createEventType: command.NewCreateEventTypeHandler(clickhouseadapter.NewEventTypesClickhouseRepository(conn)),
		createToken:     command.NewCreateTokenHandler(clickhouseadapter.NewTokenChRepository(conn)),
		createEvent:     command.NewCreateEventHandler(clickhouseadapter.NewEventsClickhouseRepository(conn)),
	}

	if err := g.generate(ctx); err != nil {
		log.Fatal(err)
	}
}

func (g *generator) generate(ctx context.Context) error {
	source, err := domain.NewSource(fmt.Sprintf("%s (generator)", gofakeit.Company()))
	if err != nil {
		return fmt.Errorf("build source: %w", err)
	}

	err = g.createSource.Execute(ctx, command.CreateSource{
		Source: source,
	})
	if err != nil {
		return fmt.Errorf("save source: %w", err)
	}
	fmt.Printf("source created  id=%-26s  name=%s\n", source.ID(), source.Name())

	eventType := domain.EventType{
		Action:                       "user.created",
		ShouldValidateMetadataSchema: false,
		LastVersion:                  domain.NewEventTypeVersion([]string{"user"}, nil),
		CreatedAt:                    time.Now(),
	}
	err = g.createEventType.Execute(ctx, command.CreateEventType{
		EventType: eventType,
	})
	if err != nil && !errors.Is(err, domain.ErrEventTypeExists) {
		return fmt.Errorf("error creating event type: %w", err)
	}

	token, err := domain.NewToken(
		source.ID(),
		fmt.Sprintf("Event Generator Token %s", time.Now()),
	)
	if err != nil {
		return fmt.Errorf("error creating token: %w", err)
	}

	err = g.createToken.Execute(ctx, command.CreateToken{
		Token: token,
	})
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	since := time.Now().AddDate(0, -6, 0)

	inserted := 0

	targetName := gofakeit.ProductName()
	target := domain.Target{
		ID:         gofakeit.UUID(),
		TargetType: pick(targetTypes),
		Name:       new(targetName),
	}

	for i := range maxEvents {
		actorName := gofakeit.Name()
		ua := gofakeit.UserAgent()

		event, err := domain.NewEvent(
			source.ID(),
			i+1,
			eventType.Action,
			domain.Actor{
				ID:        gofakeit.UUID(),
				ActorType: pick(actorTypes),
				Name:      new(actorName),
			},
			[]domain.Target{
				target,
			},
			domain.Context{
				Location:  gofakeit.IPv4Address(),
				UserAgent: new(ua),
			},
			&domain.Metadata{
				"ip":      gofakeit.IPv4Address(),
				"session": gofakeit.UUID(),
			},
			gofakeit.DateRange(since, time.Now()),
		)
		if err != nil {
			return fmt.Errorf("build event %d: %w", i+1, err)
		}

		err = g.createEvent.Execute(ctx, command.CreateEvent{
			Event: event,
			Token: token.Value(),
		})
		if err != nil {
			log.Printf("insert event %d: %v", i+1, err)
			continue
		}
		inserted++
	}

	fmt.Printf("done  inserted=%d  total=%d  source=%s\n", inserted, maxEvents, source.ID())

	return nil
}

func pick(s []string) string {
	return s[gofakeit.Number(0, len(s)-1)]
}
