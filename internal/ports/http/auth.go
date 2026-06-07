package http

import (
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) LogIn(c echo.Context) error {
	var body LogInJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	email, err := domain.NewEmail(body.Email)
	if err != nil {
		return NewBadRequestError(err, "invalid-email")
	}

	signedToken, err := h.application.Commands.LogIn.Execute(mapEchoCtxToCtx(c), command.LogIn{
		Email:             email,
		PlainTextPassword: body.Password,
	})
	if errors.Is(err, domain.ErrUserNotFound) {
		return NewHandlerErrorWithStatus(err, "authentication-failed-user-not-found", http.StatusUnauthorized)
	}
	if errors.Is(err, domain.ErrAuthenticationFailedCredentialsMismatch) {
		return NewHandlerErrorWithStatus(err, "authentication-failed-credentials-mismatch", http.StatusUnauthorized)
	}
	if err != nil {
		return NewHandlerError(err, "error-authenticating")
	}

	return c.JSON(http.StatusOK, LogIn{
		Jwt: signedToken,
	})
}
