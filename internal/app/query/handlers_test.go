package query_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

func TestAllEventTypes_Execute(t *testing.T) {
	handler := query.NewAllEventTypesHandler(&mockEventTypeFinder{})
	result, err := handler.Execute(context.Background(), query.AllEventTypes{})

	require.NoError(t, err)
	require.NotNil(t, result.Data)
}

func TestAllEvents_Execute(t *testing.T) {
	handler := query.NewAllEventsHandler(&mockAllEventsFinder{})
	result, err := handler.Execute(context.Background(), query.AllEvents{})

	require.NoError(t, err)
	require.True(t, result.HasMore)
	require.NotNil(t, result.Data)
}

func TestAllSources_Execute(t *testing.T) {
	handler := query.NewAllSourcesHandler(&mockSourcesFinder{})
	result, err := handler.Execute(context.Background(), query.AllSources{})

	require.NoError(t, err)
	require.NotNil(t, result.Data)
}

func TestAllTokens_Execute(t *testing.T) {
	handler := query.NewAllTokensHandler(&mockTokensRepo{})
	result, err := handler.Execute(context.Background(), query.AllTokens{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestEventByID_Execute(t *testing.T) {
	handler := query.NewEventByIDHandler(&mockEventsFinder{})
	result, err := handler.Execute(context.Background(), query.EventByID{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestEventTypeByAction_Execute(t *testing.T) {
	handler := query.NewEventTypeByActionHandler(&mockEventTypeFinder{})
	result, err := handler.Execute(context.Background(), query.EventTypeByAction{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestEventTypeVersions_Execute(t *testing.T) {
	handler := query.NewEventTypeVersionsHandler(&mockEventTypeFinder{})
	result, err := handler.Execute(context.Background(), query.EventTypeVersions{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestSourceByID_Execute(t *testing.T) {
	handler := query.NewSourceByIDHandler(&mockSourcesFinder{})
	result, err := handler.Execute(context.Background(), query.SourceByID{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestUserProfile_Execute(t *testing.T) {
	handler := query.NewUserProfileHandler(&mockUserRepo{})
	result, err := handler.Execute(context.Background(), query.UserProfile{})

	require.NoError(t, err)
	require.NotNil(t, result)
}

type mockAllEventsFinder struct{}

func (m mockAllEventsFinder) QueryAll(ctx context.Context, params query.AllEvents) (query.CursorPaginationResult[domain.Event], error) {
	return query.CursorPaginationResult[domain.Event]{
		HasMore: true,
		Data:    []domain.Event{},
	}, nil
}

type mockEventTypeFinder struct{}

func (m mockEventTypeFinder) FindByAction(ctx context.Context, action string) (query.EventType, error) {
	return query.EventType{}, nil
}

func (m mockEventTypeFinder) AllVersionsByAction(ctx context.Context, action string) ([]query.EventType, error) {
	return []query.EventType{}, nil
}

func (m mockEventTypeFinder) QueryAll(ctx context.Context, params query.AllEventTypes) (query.Pagination[query.EventType], error) {
	return query.Pagination[query.EventType]{
		Data: []query.EventType{},
	}, nil
}

type mockSourcesFinder struct{}

func (m mockSourcesFinder) FindByID(ctx context.Context, id string) (*domain.Source, error) {
	return &domain.Source{}, nil
}

func (m mockSourcesFinder) QueryAll(ctx context.Context, params query.AllSources) (query.Pagination[domain.Source], error) {
	return query.Pagination[domain.Source]{
		Data: []domain.Source{},
	}, nil
}

type mockTokensRepo struct{}

func (m mockTokensRepo) Save(ctx context.Context, token *domain.Token) error {
	return nil
}

func (m mockTokensRepo) Delete(ctx context.Context, id, sourceID domain.ID) error {
	return nil
}

func (m mockTokensRepo) QueryAll(ctx context.Context, sourceID domain.ID) ([]*domain.Token, error) {
	return []*domain.Token{}, nil
}

type mockEventsFinder struct{}

func (m mockEventsFinder) Save(ctx context.Context, evt domain.Event, token domain.TokenValue) error {
	return nil
}

func (m mockEventsFinder) FindByID(ctx context.Context, id domain.ID) (domain.Event, error) {
	return domain.Event{}, nil
}

func (m mockEventsFinder) QueryAll(ctx context.Context, params query.AllEvents) (query.CursorPaginationResult[domain.Event], error) {
	return query.CursorPaginationResult[domain.Event]{
		Data: []domain.Event{},
	}, nil
}

type mockUserRepo struct{}

func (m mockUserRepo) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	return &domain.User{}, nil
}

func (m mockUserRepo) FindByID(ctx context.Context, id domain.ID) (*domain.User, error) {
	return &domain.User{}, nil
}

func (m mockUserRepo) Save(ctx context.Context, user *domain.User) error {
	return nil
}
