package http_test

import "testing"

// nolint
var privateJwtKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgx5srcR5QZFDtLc4L
diH+XXbsg4ULgtF3pyOq9UjysfShRANCAAT07VfaoQhKmlOJmgabWwMuBOp8iCgY
Q9xDVH7wvsPl7Wt46/IpgapKcSer3Z0XrWsrYbszIC6fCPJFhOY3T6Yf
-----END PRIVATE KEY-----
`

// nolint
var publicJwtKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE9O1X2qEISppTiZoGm1sDLgTqfIgo
GEPcQ1R+8L7D5e1reOvyKYGqSnEnq92dF61rK2G7MyAunwjyRYTmN0+mHw==
-----END PUBLIC KEY-----
`

func TestJWTMiddleware_Authenticate(t *testing.T) {
	// Implement me
}
