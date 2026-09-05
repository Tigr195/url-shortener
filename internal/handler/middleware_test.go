package handler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tigr195/url-shortener/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	middleware := handler.LoggerMiddleware(log)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggerMiddleware_WritesStatus(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	middleware := handler.LoggerMiddleware(log)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
