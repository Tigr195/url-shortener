package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tigr195/url-shortener/internal/handler"
	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockURLService struct {
	mock.Mock
}

func (m *mockURLService) Shorten(ctx context.Context, url string) (*model.ShortenResponse, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ShortenResponse), args.Error(1)
}

func (m *mockURLService) Resolve(ctx context.Context, code string) (*model.URL, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestShorten_Success(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	svc.On("Shorten", mock.Anything, "https://google.com").
		Return(&model.ShortenResponse{
			ShortURL:    "http://localhost:8080/abc123",
			OriginalURL: "https://google.com",
		}, nil)

	body, _ := json.Marshal(model.ShortenRequest{URL: "https://google.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp model.ShortenResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)
	svc.AssertExpectations(t)
}

func TestShorten_EmptyBody(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	body, _ := json.Marshal(model.ShortenRequest{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_InvalidBody(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_ServiceError(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	svc.On("Shorten", mock.Anything, "https://google.com").
		Return(nil, errors.New("service error"))

	body, _ := json.Marshal(model.ShortenRequest{URL: "https://google.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Shorten(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirect_Success(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	svc.On("Resolve", mock.Anything, "abc123").
		Return(&model.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
		}, nil)

	r := chi.NewRouter()
	r.Get("/{code}", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
	svc := new(mockURLService)
	h := handler.NewURLHandler(svc, testLogger())

	svc.On("Resolve", mock.Anything, "notexist").
		Return(nil, errors.New("not found"))

	r := chi.NewRouter()
	r.Get("/{code}", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/notexist", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
