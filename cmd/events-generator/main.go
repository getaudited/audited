package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/common/cli"
	_ "github.com/lib/pq"

	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/domain"
)

const maxEvents = 1000

var (
	actorTypes  = []string{"user", "service", "admin", "system"}
	targetTypes = []string{"document", "record", "file", "account", "resource"}
)

func main() {
	ctx := context.Background()
	logger := logs.New(cmp.Or(os.Getenv("ADT_LOG_LEVEL"), "INFO"))

	application, closer, err := cli.NewApp(ctx, logger, nil, cli.Config{
		ActiveDatabase:        os.Getenv("ADT_ACTIVE_DATABASE"),
		DatabaseURL:           os.Getenv("ADT_DATABASE_URL"),
		ClickhouseDatabaseURL: os.Getenv("ADT_CLICKHOUSE_DATABASE_URL"),
	})
	if err != nil {
		log.Fatalf("build app: %v", err)
	}
	defer func() { _ = closer.Close() }()

	if err := generate(ctx, application); err != nil {
		log.Fatal(err)
	}
}

func generate(ctx context.Context, application *app.App) error {
	source, err := domain.NewSource(fmt.Sprintf("%s (generator)", gofakeit.Company()))
	if err != nil {
		return fmt.Errorf("build source: %w", err)
	}

	err = application.Commands.CreateSource.Execute(ctx, command.CreateSource{
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
		UpdatedAt:                    time.Now(),
	}
	err = application.Commands.CreateEventType.Execute(ctx, command.CreateEventType{
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

	err = application.Commands.CreateToken.Execute(ctx, command.CreateToken{
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

		err = application.Commands.CreateEvent.Execute(ctx, command.CreateEvent{
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
