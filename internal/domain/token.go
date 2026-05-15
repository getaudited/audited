package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTokenNotFound = errors.New("token not found")

const tokenMaxUnmaskedCharacters = 8

type TokenValue string

func (t TokenValue) String() string {
	return string(t)
}

func NewTokenValue() TokenValue {
	return TokenValue(fmt.Sprintf("tkn_%s", NewID().String()))
}

type Token struct {
	id        ID
	sourceID  ID
	name      string
	value     TokenValue
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
		value:     NewTokenValue(),
		createdAt: time.Now(),
	}, nil
}

func (t *Token) ID() ID {
	return t.id
}

func (t *Token) SourceID() ID {
	return t.sourceID
}

func (t *Token) Value() TokenValue {
	return t.value
}

func (t *Token) MaskedValue() string {
	result := ""

	for i := 0; i < len(t.value)-tokenMaxUnmaskedCharacters; i++ {
		result += "*"
	}

	return result + t.value.String()[len(t.value)-tokenMaxUnmaskedCharacters:]
}

func (t *Token) Name() string {
	return t.name
}

func (t *Token) CreatedAt() time.Time {
	return t.createdAt
}

func MarshallToToken(id, sourceID, value, name string, createdAt time.Time) *Token {
	return &Token{
		id:        ID(id),
		sourceID:  ID(sourceID),
		name:      name,
		value:     TokenValue(value),
		createdAt: createdAt,
	}
}

type TokenRepository interface {
	Save(ctx context.Context, token *Token) error
	Delete(ctx context.Context, id, sourceID ID) error
	QueryAll(ctx context.Context, sourceID ID) ([]*Token, error)
}
