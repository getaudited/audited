package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
)

// nolint
var privateJwtKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgx5srcR5QZFDtLc4L
diH+XXbsg4ULgtF3pyOq9UjysfShRANCAAT07VfaoQhKmlOJmgabWwMuBOp8iCgY
Q9xDVH7wvsPl7Wt46/IpgapKcSer3Z0XrWsrYbszIC6fCPJFhOY3T6Yf
-----END PRIVATE KEY-----
`

func main() {
	block, _ := pem.Decode([]byte(privateJwtKey))
	if block == nil {
		log.Fatal("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("failed to parse private key: %v", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		log.Fatal("private key is not ECDSA")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(ecKey)
	if err != nil {
		log.Fatalf("failed to sign token: %v", err)
	}

	fmt.Println(signed)
}
