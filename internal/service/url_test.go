package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/Tigr195/url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// мок репозитория
type mockURLRepository struct {
	mock.Mock
}

func (m *mockURLRepository) Save(ctx context.Context, url *model.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *mockURLRepository) GetByShortCode(ctx context.Context, code string) (*model.URL, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *mockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	args := m.Called(ctx, originalURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *mockURLRepository) IncrementClicks(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func TestShorten_NewURL(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	// URL ещё не существует в БД
	repo.On("GetByOriginalURL", mock.Anything, "https://google.com").
		Return(nil, errors.New("not found"))

	repo.On("Save", mock.Anything, mock.AnythingOfType("*model.URL")).
		Return(nil)

	resp, err := svc.Shorten(context.Background(), "https://google.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ShortURL)
	assert.Equal(t, "https://google.com", resp.OriginalURL)
	repo.AssertExpectations(t)
}

func TestShorten_ExistingURL(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	existing := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
	}

	// URL уже есть в БД
	repo.On("GetByOriginalURL", mock.Anything, "https://google.com").
		Return(existing, nil)

	resp, err := svc.Shorten(context.Background(), "https://google.com")

	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)
	assert.Equal(t, "https://google.com", resp.OriginalURL)
	repo.AssertExpectations(t)
}

func TestShorten_EmptyURL(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	repo.On("GetByOriginalURL", mock.Anything, "").
		Return(nil, errors.New("not found"))

	repo.On("Save", mock.Anything, mock.AnythingOfType("*model.URL")).
		Return(errors.New("db error"))

	_, err := svc.Shorten(context.Background(), "")

	assert.Error(t, err)
}

func TestResolve_Found(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	expected := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
	}

	repo.On("GetByShortCode", mock.Anything, "abc123").
		Return(expected, nil)

	repo.On("IncrementClicks", mock.Anything, "abc123").
		Return(nil)

	url, err := svc.Resolve(context.Background(), "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "https://google.com", url.OriginalURL)
}

func TestResolve_NotFound(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	repo.On("GetByShortCode", mock.Anything, "notexist").
		Return(nil, errors.New("not found"))

	_, err := svc.Resolve(context.Background(), "notexist")

	assert.Error(t, err)
}

func TestShorten_SaveError(t *testing.T) {
	repo := new(mockURLRepository)
	svc := service.NewURLService(repo, "http://localhost:8080")

	repo.On("GetByOriginalURL", mock.Anything, "https://google.com").
		Return(nil, errors.New("not found"))

	repo.On("Save", mock.Anything, mock.AnythingOfType("*model.URL")).
		Return(errors.New("db error"))

	_, err := svc.Shorten(context.Background(), "https://google.com")

	assert.Error(t, err)
	repo.AssertExpectations(t)
}
