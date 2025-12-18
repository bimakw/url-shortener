package repository

import (
	"context"

	"github.com/bimakw/url-shortener/internal/domain/entity"
)

type APIKeyRepository interface {
	Create(ctx context.Context, key *entity.APIKey) error
	GetByKey(ctx context.Context, key string) (*entity.APIKey, error)
	GetByID(ctx context.Context, id string) (*entity.APIKey, error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.APIKey, error)
	UpdateLastUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
}
