package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Tigr195/url-shortener/internal/model"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	GetByShortCode(ctx context.Context, code string) (*model.URL, error)
	IncrementClicks(ctx context.Context, code string) error
	GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error)
}

type URLService struct {
	repo    URLRepository
	baseURL string
}

func NewURLService(repo URLRepository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *URLService) Shorten(ctx context.Context, originalURL string) (*model.ShortenResponse, error) {
	// проверяем есть ли уже такой URL
	existing, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		// уже есть — возвращаем существующий
		return &model.ShortenResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existing.ShortCode),
			OriginalURL: existing.OriginalURL,
		}, nil
	}

	// нет — создаём новый
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("service.Shorten generate code: %w", err)
	}

	url := &model.URL{
		ShortCode:   code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Save(ctx, url); err != nil {
		return nil, fmt.Errorf("service.Shorten save: %w", err)
	}

	return &model.ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, code),
		OriginalURL: originalURL,
	}, nil
}

func (s *URLService) Resolve(ctx context.Context, code string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("service.Resolve: %w", err)
	}

	go s.repo.IncrementClicks(context.Background(), code)

	return url, nil
}

func generateCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
