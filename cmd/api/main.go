package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	handler "github.com/bimakw/url-shortener/internal/adapter/inbound/http"
	"github.com/bimakw/url-shortener/internal/adapter/outbound/postgres"
	redisRepo "github.com/bimakw/url-shortener/internal/adapter/outbound/redis"
	"github.com/bimakw/url-shortener/internal/application/usecase"
	"github.com/bimakw/url-shortener/internal/infrastructure"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg := infrastructure.LoadConfig()

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}

	// Run migrations
	if err := postgres.RunMigrations(ctx, db); err != nil {
		logger.Error("failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("database connected and migrations applied")

	// Initialize repositories
	urlRepo := postgres.NewURLRepository(db)
	clickRepo := postgres.NewClickRepository(db)

	// Connect to Redis (optional)
	var urlCache *redisRepo.URLCacheRepository
	redisClient, err := redisRepo.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		logger.Warn("redis not available, running without cache", slog.Any("error", err))
	} else {
		urlCache = redisRepo.NewURLCacheRepository(redisClient, cfg.App.CacheTTL)
		logger.Info("redis connected")
	}

	// Initialize use cases
	urlUseCase := usecase.NewURLUseCase(usecase.URLUseCaseConfig{
		URLRepo:    urlRepo,
		URLCache:   urlCache,
		ClickRepo:  clickRepo,
		BaseURL:    cfg.App.BaseURL,
		CodeLength: cfg.App.ShortCodeLength,
	})

	// Initialize handlers
	urlHandler := handler.NewURLHandler(urlUseCase)
	qrHandler := handler.NewQRHandler(urlUseCase, cfg.App.BaseURL)

	// Setup router
	router := handler.NewRouter(handler.RouterConfig{
		URLHandler: urlHandler,
		QRHandler:  qrHandler,
		Logger:     logger,
		RateLimit:  cfg.App.RateLimit,
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", slog.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("server exited")
}
