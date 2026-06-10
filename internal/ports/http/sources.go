package http

import (
	"net/http"

	"github.com/friendsofgo/errors"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) CreateSource(c echo.Context) error {
	var body CreateSourceJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	source, err := domain.NewSource(body.Name)
	if err != nil {
		return err
	}

	err = h.application.Commands.CreateSource.Execute(mapEchoCtxToCtx(c), command.CreateSource{
		Source: source,
	})
	if errors.Is(err, domain.ErrSourceWithProvidedNameExists) {
		return NewHandlerErrorWithStatus(err, "error-source-exists", http.StatusConflict)
	}
	if err != nil {
		return NewHandlerError(err, "error-creating-source")
	}

	return c.JSON(http.StatusCreated, Source{
		Id:        source.ID().String(),
		Name:      source.Name(),
		CreatedAt: source.CreatedAt(),
		UpdatedAt: source.UpdatedAt(),
	})
}

func (h handlers) GetSources(c echo.Context, params GetSourcesParams) error {
	result, err := h.application.Queries.AllSources.Execute(mapEchoCtxToCtx(c), query.AllSources{
		Name:       params.Name,
		Pagination: mapToQueryPaginationParams(params.Page, params.Limit),
	})
	if err != nil {
		return NewHandlerError(err, "error-retrieving-sources")
	}

	return c.JSON(http.StatusOK, SourceList{
		Data: mapToSources(result.Data),
		Pagination: Pagination{
			Total:       result.Total,
			PerPage:     result.PerPage,
			CurrentPage: result.CurrentPage,
			TotalPages:  result.TotalPages,
		},
	})
}

func (h handlers) GetSourceByID(c echo.Context, sourceId SourceId) error {
	source, err := h.application.Queries.SourceByID.Execute(mapEchoCtxToCtx(c), query.SourceByID{
		ID: sourceId,
	})
	if errors.Is(err, domain.ErrSourceNotFound) {
		return NewNotFoundError(err, "source-not-found")
	}
	if err != nil {
		return NewHandlerError(err, "error-retrieving-source")
	}

	return c.JSON(http.StatusOK, mapToSource(*source))
}
