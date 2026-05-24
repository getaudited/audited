package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type SourceByID struct {
	ID string
}

type SourceByIDHandler struct {
	finder sourceByIDFinder
}

func NewSourceByIDHandler(finder sourceByIDFinder) SourceByIDHandler {
	return SourceByIDHandler{finder: finder}
}

func (h SourceByIDHandler) Execute(ctx context.Context, q SourceByID) (*domain.Source, error) {
	return h.finder.FindByID(ctx, q.ID)
}
