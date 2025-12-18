package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
	"github.com/bimakw/url-shortener/pkg/nanoid"
	"github.com/bimakw/url-shortener/pkg/preview"
)

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrURLExpired       = errors.New("url has expired")
	ErrURLInactive      = errors.New("url is inactive")
	ErrAliasExists      = errors.New("custom alias already exists")
	ErrInvalidURL       = errors.New("invalid url")
	ErrShortCodeExists  = errors.New("short code already exists")
)

type URLUseCase struct {
	urlRepo      repository.URLRepository
	urlCache     repository.URLCacheRepository
	clickRepo    repository.ClickRepository
	baseURL      string
	codeLength   int
}

type URLUseCaseConfig struct {
	URLRepo    repository.URLRepository
	URLCache   repository.URLCacheRepository
	ClickRepo  repository.ClickRepository
	BaseURL    string
	CodeLength int
}

func NewURLUseCase(cfg URLUseCaseConfig) *URLUseCase {
	if cfg.CodeLength <= 0 {
		cfg.CodeLength = 8
	}
	return &URLUseCase{
		urlRepo:    cfg.URLRepo,
		urlCache:   cfg.URLCache,
		clickRepo:  cfg.ClickRepo,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		codeLength: cfg.CodeLength,
	}
}

func (uc *URLUseCase) CreateShortURL(ctx context.Context, req entity.CreateURLRequest) (*entity.URLResponse, error) {
	// Validate URL
	if !isValidURL(req.OriginalURL) {
		return nil, ErrInvalidURL
	}

	// Generate or use custom alias
	shortCode := req.CustomAlias
	if shortCode == "" {
		var err error
		shortCode, err = uc.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		// Check if custom alias exists
		exists, err := uc.urlRepo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrAliasExists
		}
	}

	// Calculate expiry
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &exp
	}

	// Create URL entity
	url := &entity.URL{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		CustomAlias: req.CustomAlias,
		UserID:      req.UserID,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	// Save to database
	if err := uc.urlRepo.Create(ctx, url); err != nil {
		return nil, err
	}

	// Cache the URL
	if uc.urlCache != nil {
		_ = uc.urlCache.Set(ctx, url)
	}

	return uc.toResponse(url), nil
}

func (uc *URLUseCase) GetOriginalURL(ctx context.Context, shortCode string) (*entity.URL, error) {
	// Try cache first
	if uc.urlCache != nil {
		url, err := uc.urlCache.Get(ctx, shortCode)
		if err == nil && url != nil {
			if !url.CanRedirect() {
				if url.IsExpired() {
					return nil, ErrURLExpired
				}
				return nil, ErrURLInactive
			}
			return url, nil
		}
	}

	// Get from database
	url, err := uc.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, ErrURLNotFound
	}

	// Check if can redirect
	if !url.CanRedirect() {
		if url.IsExpired() {
			return nil, ErrURLExpired
		}
		return nil, ErrURLInactive
	}

	// Cache for next time
	if uc.urlCache != nil {
		_ = uc.urlCache.Set(ctx, url)
	}

	return url, nil
}

func (uc *URLUseCase) RecordClick(ctx context.Context, click *entity.Click) error {
	// Increment click count in cache
	if uc.urlCache != nil {
		_, _ = uc.urlCache.IncrementClickCount(ctx, click.ShortCode)
	}

	// Increment in database
	url, err := uc.urlRepo.GetByShortCode(ctx, click.ShortCode)
	if err != nil || url == nil {
		return err
	}

	click.URLID = url.ID

	// Record click asynchronously (fire and forget)
	go func() {
		ctx := context.Background()
		_ = uc.urlRepo.IncrementClickCount(ctx, url.ID)
		if uc.clickRepo != nil {
			_ = uc.clickRepo.Create(ctx, click)
		}
	}()

	return nil
}

func (uc *URLUseCase) GetURLByID(ctx context.Context, id string) (*entity.URLResponse, error) {
	url, err := uc.urlRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, ErrURLNotFound
	}

	return uc.toResponse(url), nil
}

func (uc *URLUseCase) GetURLByShortCode(ctx context.Context, shortCode string) (*entity.URLResponse, error) {
	url, err := uc.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, ErrURLNotFound
	}

	return uc.toResponse(url), nil
}

func (uc *URLUseCase) GetUserURLs(ctx context.Context, userID string, limit, offset int) ([]*entity.URLResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	urls, err := uc.urlRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*entity.URLResponse, len(urls))
	for i, url := range urls {
		responses[i] = uc.toResponse(url)
	}

	return responses, nil
}

func (uc *URLUseCase) DeleteURL(ctx context.Context, id, userID string) error {
	url, err := uc.urlRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if url == nil {
		return ErrURLNotFound
	}

	// Check ownership
	if url.UserID != "" && url.UserID != userID {
		return errors.New("unauthorized")
	}

	// Delete from cache
	if uc.urlCache != nil {
		_ = uc.urlCache.Delete(ctx, url.ShortCode)
	}

	return uc.urlRepo.Delete(ctx, id)
}

func (uc *URLUseCase) GetStats(ctx context.Context, shortCode string, from, to time.Time) (*entity.ClickStats, error) {
	url, err := uc.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, ErrURLNotFound
	}

	if uc.clickRepo == nil {
		return &entity.ClickStats{
			TotalClicks: url.ClickCount,
		}, nil
	}

	return uc.clickRepo.GetStatsByURLID(ctx, url.ID, from, to)
}

func (uc *URLUseCase) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code, err := nanoid.Generate(uc.codeLength)
		if err != nil {
			return "", err
		}

		exists, err := uc.urlRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}

		if !exists {
			return code, nil
		}
	}

	return "", ErrShortCodeExists
}

func (uc *URLUseCase) toResponse(url *entity.URL) *entity.URLResponse {
	shortURL := uc.baseURL + "/" + url.ShortCode
	if url.CustomAlias != "" {
		shortURL = uc.baseURL + "/" + url.CustomAlias
	}

	return &entity.URLResponse{
		ShortCode:   url.ShortCode,
		ShortURL:    shortURL,
		OriginalURL: url.OriginalURL,
		ExpiresAt:   url.ExpiresAt,
		CreatedAt:   url.CreatedAt,
		ClickCount:  url.ClickCount,
		QRCodeURL:   uc.baseURL + "/api/urls/" + url.ShortCode + "/qr",
	}
}

func isValidURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func (uc *URLUseCase) BulkCreateShortURLs(ctx context.Context, req entity.BulkCreateURLRequest) *entity.BulkCreateURLResponse {
	results := make([]entity.BulkURLResult, len(req.URLs))
	successful := 0

	for i, urlReq := range req.URLs {
		result := entity.BulkURLResult{
			OriginalURL: urlReq.OriginalURL,
		}

		response, err := uc.CreateShortURL(ctx, urlReq)
		if err != nil {
			result.Success = false
			switch err {
			case ErrAliasExists:
				result.Error = "custom alias already exists"
			case ErrInvalidURL:
				result.Error = "invalid URL format"
			default:
				result.Error = "failed to create short URL"
			}
		} else {
			result.Success = true
			result.Data = response
			successful++
		}

		results[i] = result
	}

	return &entity.BulkCreateURLResponse{
		Total:      len(req.URLs),
		Successful: successful,
		Failed:     len(req.URLs) - successful,
		Results:    results,
	}
}

func (uc *URLUseCase) GetLinkPreview(ctx context.Context, url string) (*entity.LinkPreview, error) {
	if !isValidURL(url) {
		return nil, ErrInvalidURL
	}

	p, err := preview.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	return &entity.LinkPreview{
		Title:       p.Title,
		Description: p.Description,
		Image:       p.Image,
		SiteName:    p.SiteName,
		URL:         p.URL,
	}, nil
}
