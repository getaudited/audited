package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app"
	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

var _ ServerInterface = (*handlers)(nil)

type handlers struct {
	application *app.App
}

func (h handlers) DeleteToken(c echo.Context, sourceId SourceId, tokenId TokenId) error {
	err := h.application.Commands.DeleteToken.Execute(mapEchoCtxToCtx(c), command.DeleteToken{
		TokenID:  domain.ID(tokenId),
		SourceID: domain.ID(sourceId),
	})
	if err != nil {
		return NewHandlerError(err, "error-deleting-token")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h handlers) CreateToken(c echo.Context, sourceId SourceId) error {
	var body CreateSourceJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	token, err := domain.NewToken(domain.ID(sourceId), body.Name)
	if err != nil {
		return NewBadRequestError(err, "error-validating-data")
	}

	err = h.application.Commands.CreateToken.Execute(mapEchoCtxToCtx(c), command.CreateToken{
		Token: token,
	})
	if err != nil {
		return NewHandlerError(err, "error-creating-token")
	}

	return c.JSON(http.StatusCreated, Token{
		Name:      token.Name(),
		SourceId:  token.SourceID().String(),
		Id:        token.ID().String(),
		Value:     token.Value().String(),
		CreatedAt: token.CreatedAt(),
	})
}

func (h handlers) GetTokens(c echo.Context, sourceId SourceId) error {
	tokens, err := h.application.Queries.AllTokens.Execute(mapEchoCtxToCtx(c), query.AllTokens{
		SourceID: domain.ID(sourceId),
	})
	if err != nil {
		return NewHandlerError(err, "error-querying-tokens")
	}

	return c.JSON(http.StatusOK, TokenList{
		Data: mapToTokens(tokens),
	})
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
