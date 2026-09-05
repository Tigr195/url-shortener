package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	urlTTL    = 24 * time.Hour
	urlPrefix = "url:"
)

type URLCache struct {
	client *redis.Client
}

func NewURLCache(client *redis.Client) *URLCache {
	return &URLCache{client: client}
}

func (c *URLCache) Get(ctx context.Context, code string) (*model.URL, error) {
	key := urlPrefix + code

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("cache.Get: %w", err)
	}

	var url model.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, fmt.Errorf("cache.Get unmarshal: %w", err)
	}

	return &url, nil
}

func (c *URLCache) Set(ctx context.Context, url *model.URL) error {
	key := urlPrefix + url.ShortCode

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("cache.Set marshal: %w", err)
	}

	if err := c.client.Set(ctx, key, data, urlTTL).Err(); err != nil {
		return fmt.Errorf("cache.Set: %w", err)
	}

	return nil
}
