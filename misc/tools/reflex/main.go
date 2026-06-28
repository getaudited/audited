package main

import (
	"context"
	"os"
	"time"

	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/common/postgres"
	"github.com/getaudited/audited/misc/tools/waitfor"
)

func main() {
	waitFor := waitfor.NewWaitFor(logs.New("DEBUG"))
	waitFor.Do(func() error {
		db, err := postgres.Connect(context.Background(), os.Getenv("ADT_DATABASE_URL"))
		if err != nil {
			return err
		}

		return db.Ping()
	}, "postgres", time.Second*30)
	waitFor.Wait()
}
