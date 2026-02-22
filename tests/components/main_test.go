package components_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
	"github.com/tachyonhqdev/webhooks/misc/tools/wait/wait_for"
	"github.com/tachyonhqdev/webhooks/tests/client"
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

	// Wait for postgres and rabbitmq and other default dependencies running in containers
	wait_for.Run()

	// Wait for the backend to be spun up
	waitFor := wait_for.NewWaitFor(logs.New())
	waitFor.Do(func() error {
		req, err := http.Get(fmt.Sprintf("http://localhost:%s/health", os.Getenv("HTTP_PORT")))
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
