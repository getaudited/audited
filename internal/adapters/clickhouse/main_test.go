package clickhouse_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/common/config"
	"github.com/getaudited/audited/internal/common/logs"
)

var (
	db                  clickhouse.Conn
	dbError             mockDBError
	errMockedClickhouse = errors.New("clickhouse mock error")
	ctx                 context.Context
)

func TestMain(m *testing.M) {
	var err error
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	chConfig := clickhouseconn.Config{
		Hosts:    cfg.ClickhouseHosts,
		Database: cfg.ClickhouseDbName,
		Username: cfg.ClickhouseUsername,
		Password: cfg.ClickhousePassword,
	}

	db, err = clickhouseconn.NewConnection(ctx, chConfig)
	if err != nil {
		panic(err)
	}

	err = clickhouseconn.ApplyMigrations(ctx, chConfig, "../../../misc/clickhouse", logs.New("DEBUG"))
	if err != nil {
		panic(err)
	}

	dbError = mockDBError{}

	os.Exit(m.Run())
}

type mockDBError struct{}

func (d mockDBError) Contributors() []string {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) ServerVersion() (*driver.ServerVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) Select(_ context.Context, dest any, query string, args ...any) error {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) PrepareBatch(_ context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) AsyncInsert(_ context.Context, query string, wait bool, args ...any) error {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) Ping(_ context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) Stats() driver.Stats {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d mockDBError) Query(_ context.Context, query string, args ...any) (driver.Rows, error) {
	return mockRowsError{}, nil
}

func (d mockDBError) QueryRow(_ context.Context, query string, args ...any) driver.Row {
	return mockRowError{}
}

func (d mockDBError) Exec(ctx context.Context, query string, args ...any) error {
	return errMockedClickhouse
}

type mockRowError struct{}

func (m mockRowError) Err() error {
	return errMockedClickhouse
}

func (m mockRowError) Scan(dest ...any) error {
	return errMockedClickhouse
}

func (m mockRowError) ScanStruct(dest any) error {
	return errMockedClickhouse
}

type mockRowsError struct{}

func (m mockRowsError) Next() bool {
	return true
}

func (m mockRowsError) Scan(dest ...any) error {
	return errMockedClickhouse
}

func (m mockRowsError) ScanStruct(dest any) error {
	return errMockedClickhouse
}

func (m mockRowsError) ColumnTypes() []driver.ColumnType {
	//TODO implement me
	panic("implement me")
}

func (m mockRowsError) Totals(dest ...any) error {
	//TODO implement me
	panic("implement me")
}

func (m mockRowsError) Columns() []string {
	//TODO implement me
	panic("implement me")
}

func (m mockRowsError) Close() error {
	return errMockedClickhouse
}

func (m mockRowsError) Err() error {
	return errMockedClickhouse
}

func (m mockRowsError) HasData() bool {
	//TODO implement me
	panic("implement me")
}
