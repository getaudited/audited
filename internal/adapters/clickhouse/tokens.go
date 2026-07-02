package clickhouse

import (
	"context"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/getaudited/audited/internal/domain"
)

type TokenChRepository struct {
	db clickhousedb.Conn
}

func NewTokenChRepository(db clickhousedb.Conn) *TokenChRepository {
	return &TokenChRepository{
		db: db,
	}
}

func (r *TokenChRepository) Save(ctx context.Context, t *domain.Token) error {
	err := r.db.Exec(
		ctx,
		`INSERT INTO tokens (id, name, value, source_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		t.ID().String(),
		t.Name(),
		t.Value().String(),
		t.SourceID().String(),
		t.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("error saving token '%s' due to: %w", t.Name(), err)
	}

	return nil
}

func (r *TokenChRepository) Delete(ctx context.Context, id, sourceID domain.ID) error {
	err := r.db.Exec(
		ctx,
		`DELETE FROM tokens WHERE id = ? AND source_id = ?`,
		id.String(),
		sourceID.String(),
	)
	if err != nil {
		return fmt.Errorf("error deleting token with id '%s' due to: %w", id, err)
	}

	return nil
}

func (r *TokenChRepository) QueryAll(ctx context.Context, sourceID domain.ID) ([]*domain.Token, error) {
	//TODO implement me
	panic("implement me")
}
