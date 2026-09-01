package repository

import (
	"context"
	"fmt"

	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Save(ctx context.Context, url *model.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at)
		VALUES (:short_code, :original_url, :created_at)
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, url)
	if err != nil {
		return fmt.Errorf("repository.Save: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&url.ID); err != nil {
			return fmt.Errorf("repository.Save scan id: %w", err)
		}
	}

	return nil
}

func (r *URLRepository) GetByShortCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL

	query := `SELECT * FROM urls WHERE short_code = $1`

	if err := r.db.GetContext(ctx, &url, query, code); err != nil {
		return nil, fmt.Errorf("repository.GetByShortCode: %w", err)
	}

	return &url, nil
}

func (r *URLRepository) IncrementClicks(ctx context.Context, code string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`

	if _, err := r.db.ExecContext(ctx, query, code); err != nil {
		return fmt.Errorf("repository.IncrementClicks: %w", err)
	}

	return nil
}

func (r *URLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	var url model.URL

	query := `SELECT * FROM urls WHERE original_url = $1 LIMIT 1`

	if err := r.db.GetContext(ctx, &url, query, originalURL); err != nil {
		return nil, fmt.Errorf("repository.GetByOriginalURL: %w", err)
	}

	return &url, nil
}
