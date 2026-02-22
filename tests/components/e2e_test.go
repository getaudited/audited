package components_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/tachyonhqdev/webhooks/tests/client"
	"github.com/tachyonhqdev/webhooks/tests/emulators"
)

func TestE2E(t *testing.T) {
	apiClient := newApiClient(t)

	const maxEndpoints = 1

	// Create application
	applicationBody := client.CreateApplicationJSONRequestBody{
		Name: gofakeit.AppName(),
	}
	applicationResp01, err := apiClient.CreateApplicationWithResponse(ctx, applicationBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, applicationResp01.StatusCode())
	require.Equal(t, applicationBody.Name, applicationResp01.JSON201.Name)

	// Create event types
	schema := map[string]interface{}{
		"name": "Schema",
	}
	eventTypeBody := client.CreateEventTypeJSONRequestBody{
		Description: gofakeit.SentenceSimple(),
		Name:        gofakeit.AppName(),
		Schema:      &schema,
	}

	eventTypeResp01, err := apiClient.CreateEventTypeWithResponse(ctx, eventTypeBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, eventTypeResp01.StatusCode())
	require.Equal(t, eventTypeBody.Name, eventTypeResp01.JSON201.Name)
	require.NotEmpty(t, eventTypeResp01.JSON201.Id)

	// Created endpoints
	endpointPaths := make([]string, maxEndpoints)
	sentHeaders := map[string]string{
		"Content-Type": "application/json",
	}
	for i := 0; i < maxEndpoints; i++ {
		endpointPaths[i] = fmt.Sprintf("/webhooks/%s", ulid.Make().String())
		endpointBody := client.CreateEndpointJSONRequestBody{
			ApplicationId: applicationResp01.JSON201.Id,
			Description:   toPtr(gofakeit.SentenceSimple()),
			EventTypeIds:  []string{eventTypeResp01.JSON201.Id},
			Headers:       sentHeaders,
			Secret:        &emulators.TestEndpointSecret,
			Url:           getEndpointUrl() + endpointPaths[i],
		}
		endpoint01, err := apiClient.CreateEndpointWithResponse(ctx, endpointBody)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, endpoint01.StatusCode())
		require.NotEmpty(t, endpoint01.JSON201.Id)
		require.NotEmpty(t, endpoint01.JSON201.Headers)
		require.Equal(t, endpointBody.Url, endpoint01.JSON201.Url)
	}

	// Create message
	idempotencyKey := gofakeit.UUID()
	messageBody := client.CreateMessageJSONRequestBody{
		ApplicationId: applicationResp01.JSON201.Id,
		EventType:     eventTypeResp01.JSON201.Name,
		Payload: map[string]interface{}{
			"email":      gofakeit.Email(),
			"created_at": time.Now().String(),
		},
	}
	// expectedMessagePayload, err := json.Marshal(messageBody.Payload)
	// require.NoError(t, err)

	message01, err := apiClient.CreateMessageWithResponse(
		ctx,
		&client.CreateMessageParams{IdempotencyKey: &idempotencyKey},
		messageBody,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, message01.StatusCode())
	require.Equal(t, messageBody.Payload, message01.JSON201.Payload)
	require.Equal(t, messageBody.EventType, message01.JSON201.EventType)
	require.Equal(t, messageBody.ApplicationId, message01.JSON201.ApplicationId)

	// Check for calls to the webhook endpoint
	require.Eventually(t, func() bool {
		resp := getCallsToEndpointUrl(t)

		for _, path := range endpointPaths {
			call, exists := resp.Calls[path]
			if !exists {
				return false
			}

			// check sentHeaders
			for headerName, headerValue := range sentHeaders {
				foundHeaderValue, headerExists := call.Headers[headerName]
				require.True(t, headerExists)
				require.Equal(t, headerValue, foundHeaderValue)
			}
		}

		return true
	}, time.Second*5, time.Millisecond*250)

	// Check for message attempts
}

func toPtr[V any](v V) *V {
	return &v
}

func getEndpointUrl() string {
	return os.Getenv("MOCK_ENDPOINT_URL")
}

func getCallsToEndpointUrl(t *testing.T) emulators.GetWebhookCallsResponse {
	resp, err := http.Get("http://localhost:9090/webhooks/calls")
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var respBody emulators.GetWebhookCallsResponse
	require.NoError(t, json.Unmarshal(body, &respBody))

	return respBody
}
