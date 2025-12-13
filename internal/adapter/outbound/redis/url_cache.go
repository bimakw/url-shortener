package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
	"github.com/redis/go-redis/v9"
)

var _ repository.URLCacheRepository = (*URLCacheRepository)(nil)

const (
	urlKeyPrefix   = "url:"
	clickKeyPrefix = "click:"
	defaultTTL     = 1 * time.Hour
)

type URLCacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewURLCacheRepository(client *redis.Client, ttl time.Duration) *URLCacheRepository {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &URLCacheRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *URLCacheRepository) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
	key := urlKeyPrefix + shortCode

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var url entity.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLCacheRepository) Set(ctx context.Context, url *entity.URL) error {
	key := urlKeyPrefix + url.ShortCode

	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	ttl := r.ttl
	// If URL has expiry, use that instead
	if url.ExpiresAt != nil {
		remaining := time.Until(*url.ExpiresAt)
		if remaining > 0 && remaining < ttl {
			ttl = remaining
		}
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *URLCacheRepository) Delete(ctx context.Context, shortCode string) error {
	key := urlKeyPrefix + shortCode
	return r.client.Del(ctx, key).Err()
}

func (r *URLCacheRepository) IncrementClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := clickKeyPrefix + shortCode
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set TTL if this is the first increment
	if count == 1 {
		r.client.Expire(ctx, key, 24*time.Hour)
	}

	return count, nil
}

func (r *URLCacheRepository) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := clickKeyPrefix + shortCode
	count, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func NewRedisClient(host, port, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
