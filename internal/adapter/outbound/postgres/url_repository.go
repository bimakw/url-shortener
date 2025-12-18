package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
	"github.com/google/uuid"
)

var _ repository.URLRepository = (*URLRepository)(nil)

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, url *entity.URL) error {
	if url.ID == "" {
		url.ID = uuid.New().String()
	}
	url.CreatedAt = time.Now()
	url.UpdatedAt = time.Now()
	url.IsActive = true

	query := `
		INSERT INTO urls (id, short_code, original_url, custom_alias, user_id, expires_at, created_at, updated_at, click_count, is_active, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.ShortCode,
		url.OriginalURL,
		nullString(url.CustomAlias),
		nullString(url.UserID),
		nullTime(url.ExpiresAt),
		url.CreatedAt,
		url.UpdatedAt,
		url.ClickCount,
		url.IsActive,
		nullString(url.PasswordHash),
	)

	return err
}

func (r *URLRepository) GetByShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, expires_at, created_at, updated_at, click_count, is_active, password_hash
		FROM urls
		WHERE short_code = $1 OR custom_alias = $1
	`

	url := &entity.URL{}
	var customAlias, userID, passwordHash sql.NullString
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&customAlias,
		&userID,
		&expiresAt,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
		&url.IsActive,
		&passwordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	url.CustomAlias = customAlias.String
	url.UserID = userID.String
	url.PasswordHash = passwordHash.String
	if expiresAt.Valid {
		url.ExpiresAt = &expiresAt.Time
	}

	return url, nil
}

func (r *URLRepository) GetByID(ctx context.Context, id string) (*entity.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, expires_at, created_at, updated_at, click_count, is_active, password_hash
		FROM urls
		WHERE id = $1
	`

	url := &entity.URL{}
	var customAlias, userID, passwordHash sql.NullString
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&customAlias,
		&userID,
		&expiresAt,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
		&url.IsActive,
		&passwordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	url.CustomAlias = customAlias.String
	url.UserID = userID.String
	url.PasswordHash = passwordHash.String
	if expiresAt.Valid {
		url.ExpiresAt = &expiresAt.Time
	}

	return url, nil
}

func (r *URLRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, user_id, expires_at, created_at, updated_at, click_count, is_active, password_hash
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*entity.URL
	for rows.Next() {
		url := &entity.URL{}
		var customAlias, uid, passwordHash sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&url.ID,
			&url.ShortCode,
			&url.OriginalURL,
			&customAlias,
			&uid,
			&expiresAt,
			&url.CreatedAt,
			&url.UpdatedAt,
			&url.ClickCount,
			&url.IsActive,
			&passwordHash,
		)
		if err != nil {
			return nil, err
		}

		url.CustomAlias = customAlias.String
		url.UserID = uid.String
		url.PasswordHash = passwordHash.String
		if expiresAt.Valid {
			url.ExpiresAt = &expiresAt.Time
		}

		urls = append(urls, url)
	}

	return urls, rows.Err()
}

func (r *URLRepository) Update(ctx context.Context, url *entity.URL) error {
	url.UpdatedAt = time.Now()

	query := `
		UPDATE urls
		SET original_url = $2, custom_alias = $3, expires_at = $4, updated_at = $5, is_active = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.OriginalURL,
		nullString(url.CustomAlias),
		nullTime(url.ExpiresAt),
		url.UpdatedAt,
		url.IsActive,
	)

	return err
}

func (r *URLRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM urls WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *URLRepository) IncrementClickCount(ctx context.Context, id string) error {
	query := `UPDATE urls SET click_count = click_count + 1, updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

func (r *URLRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1 OR custom_alias = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	return exists, err
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
