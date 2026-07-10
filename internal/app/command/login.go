package command

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/getaudited/audited/internal/domain"
	"github.com/golang-jwt/jwt"
)

type LogIn struct {
	Email             domain.Email
	PlainTextPassword string
}

type LogInHandler struct {
	repo          domain.UserRepository
	privateJwtKey *ecdsa.PrivateKey
	jwtSecret     string
}

func NewLogInHandler(repo domain.UserRepository, privateJwtKey *ecdsa.PrivateKey, jwtSecret string) LogInHandler {
	return LogInHandler{
		repo:          repo,
		jwtSecret:     jwtSecret,
		privateJwtKey: privateJwtKey,
	}
}

func (c LogInHandler) Execute(ctx context.Context, cmd LogIn) (string, error) {
	user, err := c.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return "", err
	}

	ok := user.Password().IsEqual(cmd.PlainTextPassword)
	if !ok {
		return "", domain.ErrAuthenticationFailedCredentialsMismatch
	}

	now := time.Now()
	claims := jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		Subject:   user.ID().String(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
	}

	if c.privateJwtKey != nil {
		return c.signWithPrivateKey(claims)
	}

	return c.signWithSecretKey(claims)
}

func (c LogInHandler) signWithPrivateKey(claims jwt.StandardClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	signedToken, err := token.SignedString(c.privateJwtKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign the jwt: %w", err)
	}

	return signedToken, nil
}

func (c LogInHandler) signWithSecretKey(claims jwt.StandardClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(c.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("unable to sign the jwt: %w", err)
	}

	return signedToken, nil
}
