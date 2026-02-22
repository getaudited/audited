package emulators

var (
	TestEndpointSecret = "test-endpoint-secret"
	//nolint:gosec
	TestEndpointSecretBase64 = "dGVzdC1lbmRwb2ludC1zZWNyZXQ="
)

type WebhookCall struct {
	Headers map[string]string `json:"headers"`
	Body    string
}

type GetWebhookCallsResponse struct {
	Calls map[string]WebhookCall `json:"calls"`
}
