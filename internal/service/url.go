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

type URLCache interface {
	Get(ctx context.Context, code string) (*model.URL, error)
	Set(ctx context.Context, url *model.URL) error
}

type URLService struct {
	repo    URLRepository
	cache   URLCache
	baseURL string
}

func NewURLService(repo URLRepository, cache URLCache, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		cache:   cache,
		baseURL: baseURL,
	}
}

func (s *URLService) Resolve(ctx context.Context, code string) (*model.URL, error) {
	url, err := s.cache.Get(ctx, code)
	if err == nil {
		go s.repo.IncrementClicks(context.Background(), code)
		return url, nil
	}

	url, err = s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("service.Resolve: %w", err)
	}

	go s.cache.Set(context.Background(), url)

	go s.repo.IncrementClicks(context.Background(), code)

	return url, nil
}

func (s *URLService) Shorten(ctx context.Context, originalURL string) (*model.ShortenResponse, error) {
	existing, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		go s.cache.Set(context.Background(), existing)
		return &model.ShortenResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existing.ShortCode),
			OriginalURL: existing.OriginalURL,
		}, nil
	}

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

	go s.cache.Set(context.Background(), url)

	return &model.ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, code),
		OriginalURL: originalURL,
	}, nil
}

func generateCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
