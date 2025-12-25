package utm

import (
	"testing"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		params    Params
		wantErr   bool
		checkFunc func(result string) bool
	}{
		{
			name: "full UTM params",
			url:  "https://example.com",
			params: Params{
				Source:   "google",
				Medium:   "cpc",
				Campaign: "summer_sale",
				Term:     "shoes",
				Content:  "banner",
			},
			checkFunc: func(result string) bool {
				return contains(result, "utm_source=google") &&
					contains(result, "utm_medium=cpc") &&
					contains(result, "utm_campaign=summer_sale") &&
					contains(result, "utm_term=shoes") &&
					contains(result, "utm_content=banner")
			},
		},
		{
			name: "partial UTM params",
			url:  "https://example.com",
			params: Params{
				Source: "newsletter",
				Medium: "email",
			},
			checkFunc: func(result string) bool {
				return contains(result, "utm_source=newsletter") &&
					contains(result, "utm_medium=email") &&
					!contains(result, "utm_campaign") &&
					!contains(result, "utm_term") &&
					!contains(result, "utm_content")
			},
		},
		{
			name:   "empty params",
			url:    "https://example.com",
			params: Params{},
			checkFunc: func(result string) bool {
				return result == "https://example.com"
			},
		},
		{
			name: "URL with existing query params",
			url:  "https://example.com?page=1&sort=asc",
			params: Params{
				Source: "facebook",
			},
			checkFunc: func(result string) bool {
				return contains(result, "page=1") &&
					contains(result, "sort=asc") &&
					contains(result, "utm_source=facebook")
			},
		},
		{
			name: "URL with existing UTM params gets overwritten",
			url:  "https://example.com?utm_source=old",
			params: Params{
				Source: "new",
			},
			checkFunc: func(result string) bool {
				return contains(result, "utm_source=new") &&
					!contains(result, "utm_source=old")
			},
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			params:  Params{Source: "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Build(tt.url, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("Build() = %q, check failed", result)
			}
		})
	}
}

func TestHasUTM(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"no UTM", "https://example.com", false},
		{"no UTM with other params", "https://example.com?page=1", false},
		{"with utm_source", "https://example.com?utm_source=google", true},
		{"with utm_medium", "https://example.com?utm_medium=cpc", true},
		{"with utm_campaign", "https://example.com?utm_campaign=sale", true},
		{"with utm_term", "https://example.com?utm_term=shoes", true},
		{"with utm_content", "https://example.com?utm_content=banner", true},
		{"with multiple UTM", "https://example.com?utm_source=google&utm_medium=cpc", true},
		{"mixed params", "https://example.com?page=1&utm_source=google", true},
		{"invalid URL", "://invalid", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasUTM(tt.url); got != tt.want {
				t.Errorf("HasUTM(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "strip all UTM params",
			url:  "https://example.com?utm_source=google&utm_medium=cpc&utm_campaign=sale",
			want: "https://example.com",
		},
		{
			name: "preserve non-UTM params",
			url:  "https://example.com?page=1&utm_source=google&sort=asc",
			want: "https://example.com?page=1&sort=asc",
		},
		{
			name: "no UTM params",
			url:  "https://example.com?page=1",
			want: "https://example.com?page=1",
		},
		{
			name: "empty query string after strip",
			url:  "https://example.com?utm_source=google",
			want: "https://example.com",
		},
		{
			name: "strip utm_term and utm_content",
			url:  "https://example.com?utm_term=test&utm_content=banner&id=123",
			want: "https://example.com?id=123",
		},
		{
			name: "no query params",
			url:  "https://example.com/path",
			want: "https://example.com/path",
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Strip(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Strip(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestParamsStruct(t *testing.T) {
	params := Params{
		Source:   "google",
		Medium:   "cpc",
		Campaign: "summer",
		Term:     "shoes",
		Content:  "banner",
	}

	if params.Source != "google" {
		t.Errorf("Source = %q, want google", params.Source)
	}
	if params.Medium != "cpc" {
		t.Errorf("Medium = %q, want cpc", params.Medium)
	}
	if params.Campaign != "summer" {
		t.Errorf("Campaign = %q, want summer", params.Campaign)
	}
	if params.Term != "shoes" {
		t.Errorf("Term = %q, want shoes", params.Term)
	}
	if params.Content != "banner" {
		t.Errorf("Content = %q, want banner", params.Content)
	}
}

// helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
