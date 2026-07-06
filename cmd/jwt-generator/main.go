package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/common/config"
	"github.com/getaudited/audited/internal/domain"
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

	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	conn, err := clickhouseconn.NewConnection(ctx, clickhouseconn.Config{
		Version:  "development",
		Hosts:    cfg.ClickhouseHosts,
		Database: cfg.ClickhouseDbName,
		Username: cfg.ClickhouseUsername,
		Password: cfg.ClickhousePassword,
	})
	if err != nil {
		log.Fatalf("unable to open db: %v", err)
	}
	defer func() { _ = conn.Close() }()

	var adminUserID string
	row := conn.QueryRow(
		ctx,
		`SELECT id FROM users WHERE role = ? LIMIT 1`,
		domain.UserRoleAdmin.String(),
	)
	if err := row.Scan(&adminUserID); err != nil {
		log.Fatalf("error querying user: %v", err)
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
		Subject:   adminUserID,
	})

	signed, err := token.SignedString(ecKey)
	if err != nil {
		log.Fatalf("failed to sign token: %v", err)
	}

	fmt.Println(signed)
}
