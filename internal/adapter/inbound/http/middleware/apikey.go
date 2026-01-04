package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bimakw/url-shortener/internal/application/usecase"
)

type contextKey string

const (
	ContextKeyUserID     contextKey = "user_id"
	ContextKeyAPIKeyID   contextKey = "api_key_id"
	ContextKeyAPIScopes  contextKey = "api_key_scopes"
)

type APIKeyMiddleware struct {
	apiKeyUseCase *usecase.APIKeyUseCase
}

func NewAPIKeyMiddleware(uc *usecase.APIKeyUseCase) *APIKeyMiddleware {
	return &APIKeyMiddleware{apiKeyUseCase: uc}
}

func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key in header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No API key, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Extract key from "Bearer sk_xxx" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"success":false,"message":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		key := parts[1]
		if !strings.HasPrefix(key, "sk_") {
			http.Error(w, `{"success":false,"message":"Invalid API key format"}`, http.StatusUnauthorized)
			return
		}

		// Validate API key
		apiKey, err := m.apiKeyUseCase.ValidateAPIKey(r.Context(), key)
		if err != nil {
			switch err {
			case usecase.ErrAPIKeyNotFound:
				http.Error(w, `{"success":false,"message":"Invalid API key"}`, http.StatusUnauthorized)
			case usecase.ErrAPIKeyExpired:
				http.Error(w, `{"success":false,"message":"API key has expired"}`, http.StatusUnauthorized)
			case usecase.ErrAPIKeyInactive:
				http.Error(w, `{"success":false,"message":"API key is inactive"}`, http.StatusUnauthorized)
			default:
				http.Error(w, `{"success":false,"message":"Authentication failed"}`, http.StatusInternalServerError)
			}
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), ContextKeyUserID, apiKey.UserID)
		ctx = context.WithValue(ctx, ContextKeyAPIKeyID, apiKey.ID)
		ctx = context.WithValue(ctx, ContextKeyAPIScopes, apiKey.Scopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
