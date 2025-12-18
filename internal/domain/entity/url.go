package entity

import (
	"time"
)

type URL struct {
	ID           string     `json:"id"`
	ShortCode    string     `json:"short_code"`
	OriginalURL  string     `json:"original_url"`
	CustomAlias  string     `json:"custom_alias,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClickCount   int64      `json:"click_count"`
	IsActive     bool       `json:"is_active"`
	PasswordHash string     `json:"-"`
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
	Password    string  `json:"password,omitempty" validate:"omitempty,min=4,max=50"`
	UserID      string  `json:"-"`
}

type URLResponse struct {
	ShortCode         string     `json:"short_code"`
	ShortURL          string     `json:"short_url"`
	OriginalURL       string     `json:"original_url"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	ClickCount        int64      `json:"click_count"`
	QRCodeURL         string     `json:"qr_code_url,omitempty"`
	PasswordProtected bool       `json:"password_protected"`
}

type VerifyPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type BulkCreateURLRequest struct {
	URLs []CreateURLRequest `json:"urls" validate:"required,min=1,max=100,dive"`
}

type BulkURLResult struct {
	OriginalURL string       `json:"original_url"`
	Success     bool         `json:"success"`
	Data        *URLResponse `json:"data,omitempty"`
	Error       string       `json:"error,omitempty"`
}

type BulkCreateURLResponse struct {
	Total      int              `json:"total"`
	Successful int              `json:"successful"`
	Failed     int              `json:"failed"`
	Results    []BulkURLResult  `json:"results"`
}

type LinkPreview struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	URL         string `json:"url"`
}

type UTMParams struct {
	Source   string `json:"utm_source,omitempty" validate:"omitempty,max=100"`
	Medium   string `json:"utm_medium,omitempty" validate:"omitempty,max=100"`
	Campaign string `json:"utm_campaign,omitempty" validate:"omitempty,max=100"`
	Term     string `json:"utm_term,omitempty" validate:"omitempty,max=100"`
	Content  string `json:"utm_content,omitempty" validate:"omitempty,max=100"`
}

type UTMBuildRequest struct {
	URL string    `json:"url" validate:"required,url"`
	UTM UTMParams `json:"utm" validate:"required"`
}

type UTMBuildResponse struct {
	OriginalURL string `json:"original_url"`
	UTMUrl      string `json:"utm_url"`
}
