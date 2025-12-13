package repository

import (
	"context"

	"github.com/bimakw/url-shortener/internal/domain/entity"
)

type URLRepository interface {
	Create(ctx context.Context, url *entity.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)
	GetByID(ctx context.Context, id string) (*entity.URL, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.URL, error)
	Update(ctx context.Context, url *entity.URL) error
	Delete(ctx context.Context, id string) error
	IncrementClickCount(ctx context.Context, id string) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
}

type URLCacheRepository interface {
	Get(ctx context.Context, shortCode string) (*entity.URL, error)
	Set(ctx context.Context, url *entity.URL) error
	Delete(ctx context.Context, shortCode string) error
	IncrementClickCount(ctx context.Context, shortCode string) (int64, error)
}
