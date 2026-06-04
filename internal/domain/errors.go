package domain

import "errors"

var (
	ErrEventTypeNotFound            = errors.New("event_type not found")
	ErrSourceNotFound               = errors.New("source not found")
	ErrSourceWithProvidedNameExists = errors.New("the source with the name provided already exists")
	ErrEventTypeExists              = errors.New("the action specified for the event type already exists")
)
