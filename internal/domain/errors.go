package domain

import "errors"

var (
	ErrEventTypeNotFound                       = errors.New("event_type not found")
	ErrEventTypeActionNotFound                 = errors.New("event type action not found")
	ErrSourceNotFound                          = errors.New("source not found")
	ErrSourceWithProvidedNameExists            = errors.New("the source with the name provided already exists")
	ErrEventTypeExists                         = errors.New("the action specified for the event type already exists")
	ErrVersionOneOfEventTypeCannotBeRolledBack = errors.New("version 1 of event type cannot be rolledback")
	ErrEventNotFound                           = errors.New("event not found")
)
