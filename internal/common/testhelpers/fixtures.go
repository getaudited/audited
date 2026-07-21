package testhelpers

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/internal/domain"
)

func FixtureActor() domain.Actor {
	return domain.Actor{
		ID:        domain.NewID().String(),
		ActorType: "user",
		Name:      new(fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())),
		Metadata: new(map[string]interface{}{
			"user_role": "admin",
		}),
	}
}

func FixtureTarget() domain.Target {
	return domain.Target{
		ID:         domain.NewID().String(),
		Name:       new(gofakeit.AppName()),
		TargetType: "account",
		Metadata: new(map[string]interface{}{
			"prop": "value",
		}),
	}
}

func FixtureContext() domain.Context {
	return domain.Context{
		Location:  gofakeit.IPv4Address(),
		UserAgent: new(gofakeit.UserAgent()),
	}
}

func FixtureMetadata() *domain.Metadata {
	return new(map[string]interface{}{
		"prop": "value",
	})
}
