package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/lib/pq"

	"github.com/firminochangani/audited/internal/adapters/psql"
	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/common/postgres"
	"github.com/firminochangani/audited/internal/domain"
)

const maxEvents = 500

var (
	locations   = []string{"US", "EU", "APAC", "BR", "UK"}
	actorTypes  = []string{"user", "service", "admin", "system"}
	targetTypes = []string{"document", "record", "file", "account", "resource"}
)

func main() {
	count := new(500)
	// flag.Int("count", maxEvents, "number of events to insert (1–500)")
	// flag.Parse()

	if *count < 1 {
		*count = 1
	}
	if *count > maxEvents {
		*count = maxEvents
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()

	db, err := postgres.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	sourcesRepo := psql.NewSourcesPsqlRepository(db)
	eventsRepo := psql.NewEventsPsqlRepository(db)

	createSource := command.NewCreateSourceHandler(sourcesRepo)
	createEvent := command.NewCreateEventHandler(eventsRepo)

	source, err := domain.NewSource(fmt.Sprintf("%s (generator)", gofakeit.Company()))
	if err != nil {
		log.Fatalf("build source: %v", err)
	}

	if err := createSource.Execute(ctx, command.CreateSource{Source: source}); err != nil {
		log.Fatalf("save source: %v", err)
	}
	fmt.Printf("source created  id=%-26s  name=%s\n", source.ID(), source.Name())

	since := time.Now().AddDate(0, -6, 0)

	inserted := 0
	for i := range *count {
		actorName := gofakeit.Name()
		targetName := gofakeit.ProductName()
		ua := gofakeit.UserAgent()

		event, err := domain.NewEvent(
			source.ID(),
			i+1,
			domain.Actor{
				ID:        gofakeit.UUID(),
				ActorType: pick(actorTypes),
				Name:      &actorName,
			},
			[]domain.Target{
				{
					ID:         gofakeit.UUID(),
					TargetType: pick(targetTypes),
					Name:       &targetName,
				},
			},
			domain.Context{
				Location:  pick(locations),
				UserAgent: &ua,
			},
			&domain.Metadata{
				"ip":      gofakeit.IPv4Address(),
				"session": gofakeit.UUID(),
			},
			gofakeit.DateRange(since, time.Now()),
		)
		if err != nil {
			log.Fatalf("build event %d: %v", i+1, err)
		}

		if err := createEvent.Execute(ctx, command.CreateEvent{Event: event}); err != nil {
			log.Printf("insert event %d: %v", i+1, err)
			continue
		}
		inserted++
	}

	fmt.Printf("done  inserted=%d  total=%d  source=%s\n", inserted, *count, source.ID())
}

func pick(s []string) string {
	return s[gofakeit.Number(0, len(s)-1)]
}
