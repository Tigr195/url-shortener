package handler

import (
	"context"
	"encoding/json"

	"log/slog"
	"net/http"

	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	Shorten(ctx context.Context, originalURL string) (*model.ShortenResponse, error)
	Resolve(ctx context.Context, code string) (*model.URL, error)
}

type URLHandler struct {
	service URLService
	log     *slog.Logger
}

func NewURLHandler(service URLService, log *slog.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		log:     log,
	}
}

// @Summary      Shorten URL
// @Description  Creates a short URL from a long one
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request body model.ShortenRequest true "Original URL"
// @Success      201  {object}  model.ShortenResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/shorten [post]
func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		h.writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	resp, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		h.log.Error("failed to shorten url", "error", err)
		h.writeError(w, http.StatusInternalServerError, "failed to shorten url")
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// @Summary      Redirect to original URL
// @Description  Redirects to the original URL by short code
// @Tags         urls
// @Param        code path string true "Short code"
// @Success      301
// @Failure      404  {object}  map[string]string
// @Router       /{code} [get]
func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	url, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		h.log.Error("failed to resolve url", "code", code, "error", err)
		h.writeError(w, http.StatusNotFound, "url not found")
		return
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *URLHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
