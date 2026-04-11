package domain

import "time"

type Event struct {
	Id         string
	TenantID   string
	Version    int
	Actor      Actor
	Targets    []Target
	Context    Context
	Metadata   *Metadata
	OccurredAt time.Time
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
