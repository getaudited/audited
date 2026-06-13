package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", os.Getenv("ADT_DATABASE_URL"))
	if err != nil {
		log.Fatalf("unable to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	adminUser, err := models.Users(models.UserWhere.Role.EQ(domain.UserRoleAdmin.String())).One(context.Background(), db)
	if err != nil {
		log.Fatalf("error querying user: %v", err)
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
		Subject:   adminUser.ID,
	})

	signed, err := token.SignedString(ecKey)
	if err != nil {
		log.Fatalf("failed to sign token: %v", err)
	}

	fmt.Println(signed)
}
