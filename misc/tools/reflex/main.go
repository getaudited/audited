package main

import (
	"context"
	"time"

	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/common/config"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/misc/tools/waitfor"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	waitFor := waitfor.NewWaitFor(logs.New("DEBUG"))
	waitFor.Do(func() error {
		ctx := context.Background()
		db, err := clickhouseconn.NewConnection(ctx, clickhouseconn.Config{
			Version:  "development",
			Hosts:    cfg.ClickhouseHosts,
			Database: cfg.ClickhouseDbName,
			Username: cfg.ClickhouseUsername,
			Password: cfg.ClickhousePassword,
		})
		if err != nil {
			return err
		}

		return db.Ping(ctx)
	}, "clickhouse", time.Second*30)
	waitFor.Wait()
}
