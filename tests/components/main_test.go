package components_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/getaudited/audited/tests/client"
	"github.com/stretchr/testify/require"
)

var ctx context.Context

func newApiClient(t *testing.T, token string) *client.ClientWithResponses {
	cli, err := client.NewClientWithResponses(
		"http://localhost:8080",
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			return nil
		}),
	)
	require.NoError(t, err)
	return cli
}

func TestMain(m *testing.M) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	os.Exit(m.Run())
}
