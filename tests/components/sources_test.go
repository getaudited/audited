package components_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getaudited/audited/tests/client"
	"github.com/stretchr/testify/require"
)

func TestSources(t *testing.T) {
	token := "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODU4NjgzODQsImlhdCI6MTc4NTc4MTk4NCwic3ViIjoiMDFLWDhDWTNCWVFDUTJRN1c4RzE2V1daM0cifQ.xeeyK32qqeUyy7q5BhOGHnKWzr8sRI4CFwkIc1Itbpc-Tk2lWHo31uuhkVkl4RzcVu9leHKIg5N3Kk_AIF_-ZQ"
	apiClient := newApiClient(t, token)

	resp, err := apiClient.CreateSource(ctx, client.CreateSourceJSONRequestBody{
		Name: fmt.Sprintf("Integration-Tests-%s-%d", gofakeit.AppName(), time.Now().Unix()),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}
