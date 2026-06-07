package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrSourceNotFoundWhileSavingEvent = errors.New("source not found while saving event")
)

type Event struct {
	id         ID
	sourceID   ID
	version    int
	action     string
	actor      Actor
	targets    []Target
	context    Context
	metadata   *Metadata
	occurredAt time.Time
}

func NewEvent(
	sourceID ID,
	version int,
	action string,
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

	if strings.TrimSpace(action) == "" {
		return Event{}, errors.New("action cannot be empty")
	}

	return Event{
		id:         NewID(),
		sourceID:   sourceID,
		action:     action,
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

func (e *Event) Action() string {
	return e.action
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

func MarshallToEvent(
	id, sourceID, action string,
	version int,
	actor Actor,
	targets []Target,
	ctx Context,
	metadata *Metadata,
	occurredAt time.Time,
) Event {
	return Event{
		id:         ID(id),
		sourceID:   ID(sourceID),
		action:     action,
		version:    version,
		actor:      actor,
		targets:    targets,
		context:    ctx,
		metadata:   metadata,
		occurredAt: occurredAt,
	}
}

type Metadata = map[string]interface{}

type Context struct {
	Location  string
	UserAgent *string
}

type Actor struct {
	ID        string
	ActorType string
	Name      *string
	Metadata  *Metadata
}

type Target struct {
	ID         string
	Name       *string
	TargetType string
	Metadata   *Metadata
}
