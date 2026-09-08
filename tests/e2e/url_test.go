package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Tigr195/url-shortener/internal/cache"
	"github.com/Tigr195/url-shortener/internal/handler"
	"github.com/Tigr195/url-shortener/internal/logger"
	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/Tigr195/url-shortener/internal/repository"
	"github.com/Tigr195/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testApp struct {
	server  *http.Server
	baseURL string
}

func setupApp(t *testing.T) (*testApp, func()) {
	ctx := context.Background()

	// поднимаем postgres
	pgContainer, err := tcpostgres.Run(ctx,
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

	pgConn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", pgConn)
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

	// поднимаем redis
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	redisConn, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisConn[len("redis://"):],
	})

	// собираем приложение
	log := logger.New()
	urlRepo := repository.NewURLRepository(db)
	urlCache := cache.NewURLCache(redisClient)
	urlService := service.NewURLService(urlRepo, urlCache, "http://localhost:8765")
	urlHandler := handler.NewURLHandler(urlService, log)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Post("/api/shorten", urlHandler.Shorten)
	r.Get("/{code}", urlHandler.Redirect)

	server := &http.Server{
		Addr:    ":8765",
		Handler: r,
	}

	go server.ListenAndServe()
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		server.Shutdown(ctx)
		db.Close()
		redisClient.Close()
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
	}

	return &testApp{
		server:  server,
		baseURL: "http://localhost:8765",
	}, cleanup
}

func TestE2E_Shorten(t *testing.T) {
	app, cleanup := setupApp(t)
	defer cleanup()

	body, _ := json.Marshal(model.ShortenRequest{URL: "https://google.com"})
	resp, err := http.Post(
		app.baseURL+"/api/shorten",
		"application/json",
		bytes.NewBuffer(body),
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result model.ShortenResponse
	json.NewDecoder(resp.Body).Decode(&result)

	assert.NotEmpty(t, result.ShortURL)
	assert.Equal(t, "https://google.com", result.OriginalURL)
}

func TestE2E_ShortenAndRedirect(t *testing.T) {
	app, cleanup := setupApp(t)
	defer cleanup()

	// создаём короткую ссылку
	body, _ := json.Marshal(model.ShortenRequest{URL: "https://google.com"})
	resp, err := http.Post(
		app.baseURL+"/api/shorten",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)

	var result model.ShortenResponse
	json.NewDecoder(resp.Body).Decode(&result)

	// извлекаем код из short_url
	code := result.ShortURL[len(app.baseURL)+1:]

	// делаем редирект запрос без следования редиректу
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	redirectResp, err := client.Get(fmt.Sprintf("%s/%s", app.baseURL, code))
	require.NoError(t, err)

	assert.Equal(t, http.StatusMovedPermanently, redirectResp.StatusCode)
	assert.Equal(t, "https://google.com", redirectResp.Header.Get("Location"))
}

func TestE2E_SameURLReturnsSameCode(t *testing.T) {
	app, cleanup := setupApp(t)
	defer cleanup()

	body, _ := json.Marshal(model.ShortenRequest{URL: "https://google.com"})

	resp1, _ := http.Post(app.baseURL+"/api/shorten", "application/json", bytes.NewBuffer(body))
	var result1 model.ShortenResponse
	json.NewDecoder(resp1.Body).Decode(&result1)

	body, _ = json.Marshal(model.ShortenRequest{URL: "https://google.com"})
	resp2, _ := http.Post(app.baseURL+"/api/shorten", "application/json", bytes.NewBuffer(body))
	var result2 model.ShortenResponse
	json.NewDecoder(resp2.Body).Decode(&result2)

	assert.Equal(t, result1.ShortURL, result2.ShortURL)
}

func TestE2E_NotFound(t *testing.T) {
	app, cleanup := setupApp(t)
	defer cleanup()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(app.baseURL + "/notexist")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func init() {
	os.Setenv("LOGS_DIR", os.TempDir())
}
