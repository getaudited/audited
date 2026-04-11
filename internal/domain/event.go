package domain

import (
	"errors"
	"time"
)

var (
	ErrSourceNotFoundWhileSavingEvent = errors.New("source not found while saving event")
)

type Event struct {
	id         ID
	sourceID   ID
	version    int
	actor      Actor
	targets    []Target
	context    Context
	metadata   *Metadata
	occurredAt time.Time
}

func NewEvent(
	sourceID ID,
	version int,
	actor Actor,
	targets []Target,
	context Context,
	metadata *Metadata,
	occurredAt time.Time,
) (Event, error) {
	if sourceID.Empty() {
		return Event{}, errors.New("sourceID cannot be empty")
	}

	if version == 0 {
		return Event{}, errors.New("version cannot be less than 1")
	}

	// TODO: add more validations

	return Event{
		id:         NewID(),
		sourceID:   sourceID,
		version:    version,
		actor:      actor,
		targets:    targets,
		context:    context,
		metadata:   metadata,
		occurredAt: occurredAt,
	}, nil
}

func (e *Event) ID() ID {
	return e.id
}

func (e *Event) SourceID() ID {
	return e.sourceID
}

func (e *Event) Version() int {
	return e.version
}

func (e *Event) Actor() Actor {
	return e.actor
}

func (e *Event) Targets() []Target {
	return e.targets
}

func (e *Event) Context() Context {
	return e.context
}

func (e *Event) Metadata() *Metadata {
	return e.metadata
}

func (e *Event) OccurredAt() time.Time {
	return e.occurredAt
}

type Metadata = map[string]interface{}

type Context struct {
	Location  string
	UserAgent *string
}

type Actor struct {
	Id        string
	ActorType string
	Name      *string
	Metadata  *Metadata
}

type Target struct {
	Id         string
	Name       *string
	TargetType string
	Metadata   *Metadata
}
