package psql

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/firminochangani/audited/internal/domain"
)

type EventsPsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventsPsqlRepository(db boil.ContextExecutor) EventsPsqlRepository {
	return EventsPsqlRepository{
		db: db,
	}
}

func (a EventsPsqlRepository) Save(ctx context.Context, e domain.Event) error {
	row, err := mapDomainEventToModelEvent(e)
	if err != nil {
		return err
	}

	err = row.Insert(ctx, a.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("error saving event: %v", err)
	}

	targetRows, err := mapDomainEventTargetsToModelEventTargets(e.Id, e.Targets)
	if err != nil {
		return err
	}

	err = row.AddEventTargets(ctx, a.db, true, targetRows...)
	if err != nil {
		return fmt.Errorf("error saving event_targets: %v", err)
	}

	return nil
}
