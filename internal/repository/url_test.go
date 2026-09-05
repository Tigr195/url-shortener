package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/Tigr195/url-shortener/internal/repository"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err)

	// создаём таблицу
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id          BIGSERIAL PRIMARY KEY,
			short_code  VARCHAR(10)  NOT NULL UNIQUE,
			original_url TEXT        NOT NULL,
			created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
			expires_at  TIMESTAMP,
			clicks      BIGINT       NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

func TestURLRepository_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewURLRepository(db)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now(),
	}

	err := repo.Save(context.Background(), url)

	assert.NoError(t, err)
	assert.NotZero(t, url.ID)
}

func TestURLRepository_GetByShortCode(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewURLRepository(db)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now(),
	}
	err := repo.Save(context.Background(), url)
	require.NoError(t, err)

	found, err := repo.GetByShortCode(context.Background(), "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "https://google.com", found.OriginalURL)
	assert.Equal(t, "abc123", found.ShortCode)
}

func TestURLRepository_GetByShortCode_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewURLRepository(db)

	_, err := repo.GetByShortCode(context.Background(), "notexist")

	assert.Error(t, err)
}

func TestURLRepository_GetByOriginalURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewURLRepository(db)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now(),
	}
	err := repo.Save(context.Background(), url)
	require.NoError(t, err)

	found, err := repo.GetByOriginalURL(context.Background(), "https://google.com")

	assert.NoError(t, err)
	assert.Equal(t, "abc123", found.ShortCode)
}

func TestURLRepository_IncrementClicks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewURLRepository(db)

	url := &model.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://google.com",
		CreatedAt:   time.Now(),
	}
	err := repo.Save(context.Background(), url)
	require.NoError(t, err)

	err = repo.IncrementClicks(context.Background(), "abc123")
	assert.NoError(t, err)

	found, err := repo.GetByShortCode(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), found.Clicks)
}
