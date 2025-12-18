package entity

import (
	"time"
)

type APIKey struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	UserID    string     `json:"user_id"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	IsActive  bool       `json:"is_active"`
}

type CreateAPIKeyRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	Scopes    []string `json:"scopes,omitempty"`
	RateLimit int      `json:"rate_limit,omitempty"`
	ExpiresIn *int     `json:"expires_in,omitempty"` // in days
}

type APIKeyResponse struct {
	ID        string     `json:"id"`
	Key       string     `json:"key,omitempty"` // Only shown on creation
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	IsActive  bool       `json:"is_active"`
}

var DefaultScopes = []string{"urls:read", "urls:write", "stats:read"}
