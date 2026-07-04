package components_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/getaudited/audited/tests/client"
	"github.com/stretchr/testify/require"
)

var ctx context.Context

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

	os.Exit(m.Run())
}
