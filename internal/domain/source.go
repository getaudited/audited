package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Source struct {
	id        ID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func NewSource(name string) (*Source, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name cannot be empty")
	}

	return &Source{
		id:        NewID(),
		name:      name,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

func (s Source) ID() ID {
	return s.id
}

func (s Source) Name() string {
	return s.name
}

func (s Source) CreatedAt() time.Time {
	return s.createdAt
}

func (s Source) UpdatedAt() time.Time {
	return s.updatedAt
}

func MarshallToSource(id, name string, createdAt, updatedAt time.Time) Source {
	return Source{
		id:        ID(id),
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

type SourceRepository interface {
	Save(ctx context.Context, s *Source) error
}
