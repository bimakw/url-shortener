package entity

import (
	"time"
)

type URL struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CustomAlias string     `json:"custom_alias,omitempty"`
	UserID      string     `json:"user_id,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClickCount  int64      `json:"click_count"`
	IsActive    bool       `json:"is_active"`
}

func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

func (u *URL) CanRedirect() bool {
	return u.IsActive && !u.IsExpired()
}

type CreateURLRequest struct {
	OriginalURL string  `json:"original_url" validate:"required,url"`
	CustomAlias string  `json:"custom_alias,omitempty" validate:"omitempty,min=3,max=20,alphanum"`
	ExpiresIn   *int    `json:"expires_in,omitempty"` // in hours
	UserID      string  `json:"-"`
}

type URLResponse struct {
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ClickCount  int64      `json:"click_count"`
	QRCodeURL   string     `json:"qr_code_url,omitempty"`
}
