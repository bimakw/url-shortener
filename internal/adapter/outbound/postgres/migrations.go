package postgres

import (
	"context"
	"database/sql"
)

func RunMigrations(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(36) PRIMARY KEY,
			short_code VARCHAR(20) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			custom_alias VARCHAR(20) UNIQUE,
			user_id VARCHAR(36),
			expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			click_count BIGINT NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_custom_alias ON urls(custom_alias)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at)`,
		`CREATE TABLE IF NOT EXISTS clicks (
			id VARCHAR(36) PRIMARY KEY,
			url_id VARCHAR(36) NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
			short_code VARCHAR(20) NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			referrer TEXT,
			country VARCHAR(100),
			city VARCHAR(100),
			device VARCHAR(50),
			browser VARCHAR(50),
			os VARCHAR(50),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_clicks_url_id ON clicks(url_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clicks_created_at ON clicks(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_clicks_ip_address ON clicks(ip_address)`,
	}

	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return err
		}
	}

	return nil
}
