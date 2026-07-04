package cli

import (
	"context"
	"crypto/ecdsa"

	"github.com/friendsofgo/errors"

	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/common/logs"
)

// Config holds the settings required to build an app.App backed by the
// configured database.
type Config struct {
	ActiveDatabase        string
	DatabaseURL           string
	ClickhouseDatabaseURL string
}

// Closer releases the resources held by the database connection backing an
// app.App (e.g. the *sql.DB or the clickhouse connection).
type Closer interface {
	Close() error
}

// NewApp builds an app.App wired to the database selected in config.ActiveDatabase.
func NewApp(
	ctx context.Context,
	logger *logs.Logger,
	jwtPrivateKey *ecdsa.PrivateKey,
	config Config,
) (*app.App, Closer, error) {
	switch config.ActiveDatabase {
	case "postgres":
		return newPostgresApp(ctx, logger, jwtPrivateKey, config)
	case "clickhouse":
		return newClickhouseApp(ctx, jwtPrivateKey, config)
	default:
		return nil, nil, errors.New("no database has been set")
	}
}
