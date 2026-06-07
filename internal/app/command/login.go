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
}

func NewLogInHandler(repo domain.UserRepository, privateJwtKey *ecdsa.PrivateKey) LogInHandler {
	return LogInHandler{
		repo:          repo,
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
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(c.privateJwtKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign the jwt: %w", err)
	}

	return signedToken, nil
}
