package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tigr195/url-shortener/internal/cache"
	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: connStr[len("redis://"):],
	})

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestURLCache_SetAndGet(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	c := cache.NewURLCache(client)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now(),
	}

	err := c.Set(context.Background(), url)
	assert.NoError(t, err)

	found, err := c.Get(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.Equal(t, "https://google.com", found.OriginalURL)
	assert.Equal(t, "abc123", found.ShortCode)
}

func TestURLCache_Get_NotFound(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	c := cache.NewURLCache(client)

	_, err := c.Get(context.Background(), "notexist")
	assert.Error(t, err)
}
