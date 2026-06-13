package http

import (
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
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
		return NewHandlerErrorWithStatus(err, "authentication-failed", http.StatusUnauthorized)
	}
	if errors.Is(err, domain.ErrAuthenticationFailedCredentialsMismatch) {
		return NewHandlerErrorWithStatus(err, "authentication-failed", http.StatusUnauthorized)
	}
	if err != nil {
		return NewHandlerError(err, "error-authenticating")
	}

	return c.JSON(http.StatusOK, LogIn{
		Jwt: signedToken,
	})
}

func (h handlers) GetUserProfile(c echo.Context) error {
	userID, err := mapRetrieveUserIdFromCtx(c)
	if err != nil {
		return NewHandlerError(err, "error-retrieving-user-id")
	}

	user, err := h.application.Queries.UserProfile.Execute(mapEchoCtxToCtx(c), query.UserProfile{
		UserID: domain.ID(userID),
	})
	if err != nil {
		return NewHandlerError(err, "unable-retrieve-user-details")
	}

	return c.JSON(http.StatusOK, User{
		Id:    user.ID().String(),
		Email: user.Email().String(),
		Role:  user.Role().String(),
	})
}
