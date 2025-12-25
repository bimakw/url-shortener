package entity

import (
	"testing"
	"time"
)

func TestURL_IsExpired(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{"nil expiry", nil, false},
		{"expired", &pastTime, true},
		{"not expired", &futureTime, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URL{ExpiresAt: tt.expiresAt}
			if got := u.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURL_CanRedirect(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		isActive  bool
		expiresAt *time.Time
		want      bool
	}{
		{"active and not expired", true, nil, true},
		{"active with future expiry", true, &futureTime, true},
		{"active but expired", true, &pastTime, false},
		{"inactive and not expired", false, nil, false},
		{"inactive and expired", false, &pastTime, false},
		{"inactive with future expiry", false, &futureTime, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URL{
				IsActive:  tt.isActive,
				ExpiresAt: tt.expiresAt,
			}
			if got := u.CanRedirect(); got != tt.want {
				t.Errorf("CanRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLStruct(t *testing.T) {
	now := time.Now()
	expiry := now.Add(24 * time.Hour)

	url := URL{
		ID:          "123",
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CustomAlias: "my-link",
		UserID:      "user123",
		ExpiresAt:   &expiry,
		CreatedAt:   now,
		UpdatedAt:   now,
		ClickCount:  42,
		IsActive:    true,
	}

	if url.ID != "123" {
		t.Errorf("ID = %q, want 123", url.ID)
	}
	if url.ShortCode != "abc123" {
		t.Errorf("ShortCode = %q, want abc123", url.ShortCode)
	}
	if url.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL = %q, want https://example.com", url.OriginalURL)
	}
	if url.ClickCount != 42 {
		t.Errorf("ClickCount = %d, want 42", url.ClickCount)
	}
	if !url.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestURLDefaultValues(t *testing.T) {
	url := URL{}

	if url.ID != "" {
		t.Errorf("default ID should be empty, got %q", url.ID)
	}
	if url.ClickCount != 0 {
		t.Errorf("default ClickCount should be 0, got %d", url.ClickCount)
	}
	if url.IsActive {
		t.Error("default IsActive should be false")
	}
	if url.ExpiresAt != nil {
		t.Error("default ExpiresAt should be nil")
	}
}

func TestCreateURLRequest(t *testing.T) {
	hours := 24
	req := CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomAlias: "custom",
		ExpiresIn:   &hours,
		Password:    "secret",
		UserID:      "user123",
	}

	if req.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL = %q, want https://example.com", req.OriginalURL)
	}
	if req.CustomAlias != "custom" {
		t.Errorf("CustomAlias = %q, want custom", req.CustomAlias)
	}
	if *req.ExpiresIn != 24 {
		t.Errorf("ExpiresIn = %d, want 24", *req.ExpiresIn)
	}
}

func TestURLResponse(t *testing.T) {
	now := time.Now()
	resp := URLResponse{
		ShortCode:         "abc123",
		ShortURL:          "https://short.ly/abc123",
		OriginalURL:       "https://example.com",
		ExpiresAt:         nil,
		CreatedAt:         now,
		ClickCount:        10,
		QRCodeURL:         "https://qr.service/abc123",
		PasswordProtected: true,
	}

	if resp.ShortCode != "abc123" {
		t.Errorf("ShortCode = %q, want abc123", resp.ShortCode)
	}
	if resp.ClickCount != 10 {
		t.Errorf("ClickCount = %d, want 10", resp.ClickCount)
	}
	if !resp.PasswordProtected {
		t.Error("PasswordProtected should be true")
	}
}

func TestBulkCreateURLRequest(t *testing.T) {
	req := BulkCreateURLRequest{
		URLs: []CreateURLRequest{
			{OriginalURL: "https://example1.com"},
			{OriginalURL: "https://example2.com"},
		},
	}

	if len(req.URLs) != 2 {
		t.Errorf("len(URLs) = %d, want 2", len(req.URLs))
	}
}

func TestBulkURLResult(t *testing.T) {
	// Success case
	successResult := BulkURLResult{
		OriginalURL: "https://example.com",
		Success:     true,
		Data:        &URLResponse{ShortCode: "abc"},
		Error:       "",
	}

	if !successResult.Success {
		t.Error("Success should be true")
	}
	if successResult.Data == nil {
		t.Error("Data should not be nil for success")
	}

	// Failure case
	failResult := BulkURLResult{
		OriginalURL: "https://invalid",
		Success:     false,
		Data:        nil,
		Error:       "invalid URL",
	}

	if failResult.Success {
		t.Error("Success should be false")
	}
	if failResult.Error == "" {
		t.Error("Error should not be empty for failure")
	}
}

func TestLinkPreview(t *testing.T) {
	preview := LinkPreview{
		Title:       "Example Site",
		Description: "An example website",
		Image:       "https://example.com/image.png",
		SiteName:    "Example",
		URL:         "https://example.com",
	}

	if preview.Title != "Example Site" {
		t.Errorf("Title = %q, want Example Site", preview.Title)
	}
}

func TestUTMParams(t *testing.T) {
	utm := UTMParams{
		Source:   "google",
		Medium:   "cpc",
		Campaign: "summer_sale",
		Term:     "shoes",
		Content:  "banner",
	}

	if utm.Source != "google" {
		t.Errorf("Source = %q, want google", utm.Source)
	}
	if utm.Medium != "cpc" {
		t.Errorf("Medium = %q, want cpc", utm.Medium)
	}
}

func TestUTMBuildRequest(t *testing.T) {
	req := UTMBuildRequest{
		URL: "https://example.com",
		UTM: UTMParams{
			Source: "google",
			Medium: "cpc",
		},
	}

	if req.URL != "https://example.com" {
		t.Errorf("URL = %q, want https://example.com", req.URL)
	}
	if req.UTM.Source != "google" {
		t.Errorf("UTM.Source = %q, want google", req.UTM.Source)
	}
}

func TestVerifyPasswordRequest(t *testing.T) {
	req := VerifyPasswordRequest{
		Password: "secret123",
	}

	if req.Password != "secret123" {
		t.Errorf("Password = %q, want secret123", req.Password)
	}
}
