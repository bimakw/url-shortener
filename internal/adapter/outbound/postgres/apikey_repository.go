package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
	"github.com/google/uuid"
)

var _ repository.APIKeyRepository = (*APIKeyRepository)(nil)

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *entity.APIKey) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	key.CreatedAt = time.Now()
	key.IsActive = true

	scopesJSON, err := json.Marshal(key.Scopes)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO api_keys (id, key, name, user_id, scopes, rate_limit, expires_at, created_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		key.ID,
		key.Key,
		key.Name,
		key.UserID,
		scopesJSON,
		key.RateLimit,
		nullTime(key.ExpiresAt),
		key.CreatedAt,
		key.IsActive,
	)

	return err
}

func (r *APIKeyRepository) GetByKey(ctx context.Context, key string) (*entity.APIKey, error) {
	query := `
		SELECT id, key, name, user_id, scopes, rate_limit, expires_at, created_at, last_used, is_active
		FROM api_keys
		WHERE key = $1
	`

	apiKey := &entity.APIKey{}
	var scopesJSON []byte
	var expiresAt, lastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&apiKey.ID,
		&apiKey.Key,
		&apiKey.Name,
		&apiKey.UserID,
		&scopesJSON,
		&apiKey.RateLimit,
		&expiresAt,
		&apiKey.CreatedAt,
		&lastUsed,
		&apiKey.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(scopesJSON, &apiKey.Scopes); err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	if lastUsed.Valid {
		apiKey.LastUsed = &lastUsed.Time
	}

	return apiKey, nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id string) (*entity.APIKey, error) {
	query := `
		SELECT id, key, name, user_id, scopes, rate_limit, expires_at, created_at, last_used, is_active
		FROM api_keys
		WHERE id = $1
	`

	apiKey := &entity.APIKey{}
	var scopesJSON []byte
	var expiresAt, lastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.Key,
		&apiKey.Name,
		&apiKey.UserID,
		&scopesJSON,
		&apiKey.RateLimit,
		&expiresAt,
		&apiKey.CreatedAt,
		&lastUsed,
		&apiKey.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(scopesJSON, &apiKey.Scopes); err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	if lastUsed.Valid {
		apiKey.LastUsed = &lastUsed.Time
	}

	return apiKey, nil
}

func (r *APIKeyRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.APIKey, error) {
	query := `
		SELECT id, key, name, user_id, scopes, rate_limit, expires_at, created_at, last_used, is_active
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entity.APIKey
	for rows.Next() {
		apiKey := &entity.APIKey{}
		var scopesJSON []byte
		var expiresAt, lastUsed sql.NullTime

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Key,
			&apiKey.Name,
			&apiKey.UserID,
			&scopesJSON,
			&apiKey.RateLimit,
			&expiresAt,
			&apiKey.CreatedAt,
			&lastUsed,
			&apiKey.IsActive,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(scopesJSON, &apiKey.Scopes); err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			apiKey.ExpiresAt = &expiresAt.Time
		}
		if lastUsed.Valid {
			apiKey.LastUsed = &lastUsed.Time
		}

		keys = append(keys, apiKey)
	}

	return keys, rows.Err()
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET last_used = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *APIKeyRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
