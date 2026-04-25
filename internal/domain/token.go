package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Token struct {
	id        ID
	sourceID  ID
	name      string
	createdAt time.Time
}

func NewToken(sourceID ID, name string) (*Token, error) {
	if sourceID.Empty() {
		return nil, errors.New("sourceID cannot be empty")
	}

	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name cannot be empty")
	}

	return &Token{
		id:        NewID(),
		sourceID:  sourceID,
		name:      name,
		createdAt: time.Now(),
	}, nil
}

func (t *Token) ID() ID {
	return t.id
}

func (t *Token) SourceID() ID {
	return t.sourceID
}

func (t *Token) Name() string {
	return t.name
}

func (t *Token) CreatedAt() time.Time {
	return t.createdAt
}

func MarshallToToken(id, sourceID, name string, createdAt time.Time) *Token {
	return &Token{
		id:        ID(id),
		sourceID:  ID(sourceID),
		name:      name,
		createdAt: createdAt,
	}
}

type TokenRepository interface {
	Save(ctx context.Context, token Token) error
}
