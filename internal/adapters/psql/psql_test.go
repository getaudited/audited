package psql_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/firminochangani/audited/internal/common/postgres"
	"github.com/firminochangani/audited/misc/tools/wait/wait_for"
)

var (
	db  *sql.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	var err error
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	// Wait for postgres and rabbitmq and other default dependencies running in containers
	wait_for.Run()

	db, err = postgres.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	err = postgres.ApplyMigrations(db, "../../../misc/sql/migrations")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
