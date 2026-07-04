package main

import (
	"context"
	"os"
	"time"

	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/misc/tools/waitfor"
)

func main() {
	waitFor := waitfor.NewWaitFor(logs.New("DEBUG"))
	waitFor.Do(func() error {
		ctx := context.Background()
		db, err := clickhouseconn.NewConnection(ctx, os.Getenv("ADT_DATABASE_URL"))
		if err != nil {
			return err
		}

		return db.Ping(ctx)
	}, "clickhouse", time.Second*30)
	waitFor.Wait()
}
