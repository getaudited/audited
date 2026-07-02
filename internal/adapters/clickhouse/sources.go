package clickhouse

import (
	"context"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/getaudited/audited/internal/domain"
)

type SourcesClickhouseRepository struct {
	db clickhousedb.Conn
}

func NewSourcesClickhouseRepository(db clickhousedb.Conn) SourcesClickhouseRepository {
	return SourcesClickhouseRepository{
		db: db,
	}
}

func (r SourcesClickhouseRepository) Save(ctx context.Context, s *domain.Source) error {
	err := r.db.Exec(
		ctx,
		`INSERT INTO sources (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		s.ID().String(),
		s.Name(),
		s.CreatedAt(),
		s.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("error saving source '%s' due to: %w", s.Name(), err)
	}

	return nil
}
