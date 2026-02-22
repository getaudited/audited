package endpointcaller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

const (
	webhookSecHeaderName    = "x-whk-secret"
	requestTimeoutInSeconds = 10
	maxSuccessCode          = 299
	maxResponseSizeInBytes  = 1024 * 1024 // 1024 kb
)

type CallPayload struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

type Service struct {
	logger *logs.Logger
}

func NewService(logger *logs.Logger) Service {
	return Service{
		logger: logger,
	}
}

func (s Service) Call(
	ctx context.Context,
	endpoint *domain.Endpoint,
	msg *domain.Message,
) (*domain.MessageAttempt, error) {
	client := &http.Client{
		Timeout: time.Second * requestTimeoutInSeconds,
	}

	payload, err := marshallToCallPayload(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling endpoint payload: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.Url().String(),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %v", err)
	}

	// Map headers into the request
	for key, value := range endpoint.Headers() {
		req.Header.Add(key, value)
	}

	signature, err := generateWebhookSignature(msg, endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("error signing message: %v", err)
	}

	// Standard webhooks headers
	req.Header[standardwebhooks.HeaderWebhookID] = []string{msg.ID().String()}
	req.Header[standardwebhooks.HeaderWebhookTimestamp] = []string{fmt.Sprintf("%d", msg.CreatedAt().Unix())}
	req.Header[standardwebhooks.HeaderWebhookSignature] = []string{signature}

	headersSent, err := json.Marshal(req.Header)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall headers sent: %v", err)
	}

	ma, err := domain.NewMessageAttempt(
		msg.ID(),
		headersSent,
		endpoint.Url(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating message attempt: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("error calling endpoint", "endpoint_url", endpoint.Url(), "error", err)
		// return the message attempt instead of an error
		return ma, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > maxSuccessCode {
		ma.MarkAttemptedAsFailed(resp.StatusCode)
	} else {
		ma.MarkAttemptedAsSucceeded(resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSizeInBytes))
	if err != nil {
		s.logger.Error("error reading response", "endpoint_url", endpoint.Url())
	}

	ma.AttachResponseDetails(string(body))

	return ma, nil
}

func generateWebhookSignature(msg *domain.Message, endpoint *domain.Endpoint, payload []byte) (string, error) {
	var endpointSecret []byte
	if endpoint.Secret() != nil {
		endpointSecret = []byte(*endpoint.Secret())
	}

	wh, err := standardwebhooks.NewWebhook(base64.StdEncoding.EncodeToString(endpointSecret))
	if err != nil {
		return "", fmt.Errorf("error creating standard webhook: %v", err)
	}

	signature, err := wh.Sign(msg.ID().String(), msg.CreatedAt(), payload)
	if err != nil {
		return "", fmt.Errorf("error signing message: %v", err)
	}

	return signature, nil
}

func marshallToCallPayload(msg *domain.Message) ([]byte, error) {
	var data interface{}
	err := json.Unmarshal(msg.Payload(), &data)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshall messsage payload")
	}

	return json.Marshal(CallPayload{
		Data:      data,
		Type:      msg.EventType(),
		Timestamp: msg.CreatedAt().Format(time.RFC3339),
	})
}
