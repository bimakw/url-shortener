package repository

import (
	"context"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
)

type ClickRepository interface {
	Create(ctx context.Context, click *entity.Click) error
	GetByURLID(ctx context.Context, urlID string, limit, offset int) ([]*entity.Click, error)
	GetStatsByURLID(ctx context.Context, urlID string, from, to time.Time) (*entity.ClickStats, error)
	CountByURLID(ctx context.Context, urlID string) (int64, error)
	CountUniqueByURLID(ctx context.Context, urlID string) (int64, error)
}
