package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyExpired  = errors.New("api key has expired")
	ErrAPIKeyInactive = errors.New("api key is inactive")
)

type APIKeyUseCase struct {
	apiKeyRepo repository.APIKeyRepository
}

func NewAPIKeyUseCase(repo repository.APIKeyRepository) *APIKeyUseCase {
	return &APIKeyUseCase{apiKeyRepo: repo}
}

func (uc *APIKeyUseCase) CreateAPIKey(ctx context.Context, userID string, req entity.CreateAPIKeyRequest) (*entity.APIKeyResponse, error) {
	// Generate secure API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	key := "sk_" + hex.EncodeToString(keyBytes)

	// Set default scopes if not provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = entity.DefaultScopes
	}

	// Set default rate limit
	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 100
	}

	// Calculate expiry
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &exp
	}

	apiKey := &entity.APIKey{
		Key:       key,
		Name:      req.Name,
		UserID:    userID,
		Scopes:    scopes,
		RateLimit: rateLimit,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := uc.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return &entity.APIKeyResponse{
		ID:        apiKey.ID,
		Key:       key, // Only show key on creation
		Name:      apiKey.Name,
		Scopes:    apiKey.Scopes,
		RateLimit: apiKey.RateLimit,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
		IsActive:  apiKey.IsActive,
	}, nil
}

func (uc *APIKeyUseCase) ValidateAPIKey(ctx context.Context, key string) (*entity.APIKey, error) {
	apiKey, err := uc.apiKeyRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}

	if !apiKey.IsActive {
		return nil, ErrAPIKeyInactive
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	// Update last used asynchronously
	go func() {
		_ = uc.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
	}()

	return apiKey, nil
}

func (uc *APIKeyUseCase) GetUserAPIKeys(ctx context.Context, userID string) ([]*entity.APIKeyResponse, error) {
	keys, err := uc.apiKeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*entity.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = &entity.APIKeyResponse{
			ID:        key.ID,
			Name:      key.Name,
			Scopes:    key.Scopes,
			RateLimit: key.RateLimit,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			LastUsed:  key.LastUsed,
			IsActive:  key.IsActive,
		}
	}

	return responses, nil
}

func (uc *APIKeyUseCase) RevokeAPIKey(ctx context.Context, id, userID string) error {
	apiKey, err := uc.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return ErrAPIKeyNotFound
	}

	// Check ownership
	if apiKey.UserID != userID {
		return errors.New("unauthorized")
	}

	return uc.apiKeyRepo.Deactivate(ctx, id)
}

func (uc *APIKeyUseCase) DeleteAPIKey(ctx context.Context, id, userID string) error {
	apiKey, err := uc.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return ErrAPIKeyNotFound
	}

	// Check ownership
	if apiKey.UserID != userID {
		return errors.New("unauthorized")
	}

	return uc.apiKeyRepo.Delete(ctx, id)
}
