package domain

import "time"

type Event struct {
	Id         string
	OccurredAt time.Time
	Version    int
	Actor      Actor
	Targets    []Target
	Context    Context
	Metadata   Metadata
}

type Metadata map[string]any

type Context struct {
	Location  string
	UserAgent string
}

type Actor struct {
	Id        string
	ActorType string
	Metadata  Metadata
}

type Target struct {
	Id         string
	TargetType string
	Metadata   Metadata
}
