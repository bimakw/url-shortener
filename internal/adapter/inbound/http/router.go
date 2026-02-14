package http

import (
	"log/slog"
	"net/http"

	"github.com/bimakw/url-shortener/internal/adapter/inbound/http/middleware"
)

type RouterConfig struct {
	URLHandler    *URLHandler
	QRHandler     *QRHandler
	APIKeyHandler *APIKeyHandler
	Logger        *slog.Logger
	RateLimit     int
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		Success(w, http.StatusOK, "OK", nil)
	})

	// API routes
	mux.HandleFunc("POST /api/urls", cfg.URLHandler.CreateShortURL)
	mux.HandleFunc("POST /api/urls/bulk", cfg.URLHandler.BulkCreateShortURLs)
	mux.HandleFunc("GET /api/urls/{code}", cfg.URLHandler.GetURLInfo)
	mux.HandleFunc("GET /api/urls/{code}/stats", cfg.URLHandler.GetStats)
	mux.HandleFunc("DELETE /api/urls/{id}", cfg.URLHandler.DeleteURL)
	mux.HandleFunc("GET /api/urls", cfg.URLHandler.GetUserURLs)

	// QR Code
	mux.HandleFunc("GET /api/urls/{code}/qr", cfg.QRHandler.GenerateQR)

	// Link Preview
	mux.HandleFunc("GET /api/preview", cfg.URLHandler.GetLinkPreview)

	// Password Protected URLs
	mux.HandleFunc("GET /api/urls/{code}/protected", cfg.URLHandler.CheckPasswordProtected)
	mux.HandleFunc("POST /api/urls/{code}/verify", cfg.URLHandler.VerifyPassword)

	mux.HandleFunc("POST /api/utm/build", cfg.URLHandler.BuildUTMUrl)
	mux.HandleFunc("GET /api/utm/strip", cfg.URLHandler.StripUTM)

	// API Key Management
	if cfg.APIKeyHandler != nil {
		mux.HandleFunc("POST /api/keys", cfg.APIKeyHandler.CreateAPIKey)
		mux.HandleFunc("GET /api/keys", cfg.APIKeyHandler.GetAPIKeys)
		mux.HandleFunc("POST /api/keys/{id}/revoke", cfg.APIKeyHandler.RevokeAPIKey)
		mux.HandleFunc("DELETE /api/keys/{id}", cfg.APIKeyHandler.DeleteAPIKey)
	}

	// Redirect (must be last as it's a catch-all)
	mux.HandleFunc("GET /{code}", cfg.URLHandler.Redirect)

	var handler http.Handler = mux

	// CORS
	corsConfig := middleware.DefaultCORSConfig()
	handler = middleware.CORS(corsConfig)(handler)

	// Logging
	if cfg.Logger != nil {
		handler = middleware.Logging(cfg.Logger)(handler)
		handler = middleware.Recovery(cfg.Logger)(handler)
	}

	// Rate limiting
	if cfg.RateLimit > 0 {
		rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, cfg.RateLimit*2)
		handler = rateLimiter.Limit(handler)
	}

	return handler
}
