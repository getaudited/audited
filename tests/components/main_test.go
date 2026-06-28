package components_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/common/postgres"
	"github.com/getaudited/audited/misc/tools/waitfor"
	"github.com/getaudited/audited/tests/client"
	"github.com/stretchr/testify/require"
)

var (
	ctx context.Context
)

// nolint
func newApiClient(t *testing.T) *client.ClientWithResponses {
	cli, err := client.NewClientWithResponses("http://localhost:8080")
	require.NoError(t, err)
	return cli
}

func TestMain(m *testing.M) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	// Wait for the backend to be spun up
	waitFor := waitfor.NewWaitFor(logs.New("DEBUG"))

	waitFor.Do(func() error {
		db, err := postgres.Connect(ctx, strings.Replace(os.Getenv("ADT_DATABASE_URL"), "@postgres", "@localhost", 1))
		if err != nil {
			return err
		}

		return db.Ping()
	}, "postgres", time.Second*30)

	waitFor.Do(func() error {
		req, err := http.Get(fmt.Sprintf("http://localhost:%s/health", os.Getenv("ADT_HTTP_PORT")))
		if err != nil {
			return err
		}

		if req.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status code 200 instead got %d", req.StatusCode)
		}

		return nil
	}, "webhooks-service", time.Minute*1)
	waitFor.Wait()

	os.Exit(m.Run())
}
