package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose"
)

func Connect(ctx context.Context, uri string) (*sql.DB, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("could not open a connection to postgres: \nuri: %s\nerror: %w", uri, err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping postgres: \nuri: %s\nerror: %w", uri, err)
	}

	return db, nil
}

func ApplyMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get the current working directory: %w", err)
	}

	if err = goose.Up(db, fmt.Sprintf("%s/%s", workdir, dir)); err != nil {
		return fmt.Errorf("unable to apply migrations: %w", err)
	}

	return nil
}
