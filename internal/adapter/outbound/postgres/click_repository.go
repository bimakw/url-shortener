package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/bimakw/url-shortener/internal/domain/repository"
	"github.com/google/uuid"
)

var _ repository.ClickRepository = (*ClickRepository)(nil)

type ClickRepository struct {
	db *sql.DB
}

func NewClickRepository(db *sql.DB) *ClickRepository {
	return &ClickRepository{db: db}
}

func (r *ClickRepository) Create(ctx context.Context, click *entity.Click) error {
	if click.ID == "" {
		click.ID = uuid.New().String()
	}
	click.CreatedAt = time.Now()

	query := `
		INSERT INTO clicks (id, url_id, short_code, ip_address, user_agent, referrer, country, city, device, browser, os, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		click.ID,
		click.URLID,
		click.ShortCode,
		click.IPAddress,
		click.UserAgent,
		click.Referrer,
		click.Country,
		click.City,
		click.Device,
		click.Browser,
		click.OS,
		click.CreatedAt,
	)

	return err
}

func (r *ClickRepository) GetByURLID(ctx context.Context, urlID string, limit, offset int) ([]*entity.Click, error) {
	query := `
		SELECT id, url_id, short_code, ip_address, user_agent, referrer, country, city, device, browser, os, created_at
		FROM clicks
		WHERE url_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []*entity.Click
	for rows.Next() {
		click := &entity.Click{}
		err := rows.Scan(
			&click.ID,
			&click.URLID,
			&click.ShortCode,
			&click.IPAddress,
			&click.UserAgent,
			&click.Referrer,
			&click.Country,
			&click.City,
			&click.Device,
			&click.Browser,
			&click.OS,
			&click.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}

	return clicks, rows.Err()
}

func (r *ClickRepository) GetStatsByURLID(ctx context.Context, urlID string, from, to time.Time) (*entity.ClickStats, error) {
	stats := &entity.ClickStats{
		ClicksByDate: make(map[string]int64),
	}

	// Total clicks
	totalQuery := `SELECT COUNT(*) FROM clicks WHERE url_id = $1 AND created_at BETWEEN $2 AND $3`
	err := r.db.QueryRowContext(ctx, totalQuery, urlID, from, to).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, err
	}

	// Unique clicks (by IP)
	uniqueQuery := `SELECT COUNT(DISTINCT ip_address) FROM clicks WHERE url_id = $1 AND created_at BETWEEN $2 AND $3`
	err = r.db.QueryRowContext(ctx, uniqueQuery, urlID, from, to).Scan(&stats.UniqueClicks)
	if err != nil {
		return nil, err
	}

	// Clicks by date
	dateQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY date
	`
	dateRows, err := r.db.QueryContext(ctx, dateQuery, urlID, from, to)
	if err != nil {
		return nil, err
	}
	defer dateRows.Close()

	for dateRows.Next() {
		var date time.Time
		var count int64
		if err := dateRows.Scan(&date, &count); err != nil {
			return nil, err
		}
		stats.ClicksByDate[date.Format("2006-01-02")] = count
	}

	// Top referrers
	stats.TopReferrers, _ = r.getTopReferrers(ctx, urlID, from, to, 5)

	// Top countries
	stats.TopCountries, _ = r.getTopCountries(ctx, urlID, from, to, 5)

	// Top browsers
	stats.TopBrowsers, _ = r.getTopBrowsers(ctx, urlID, from, to, 5)

	// Top devices
	stats.TopDevices, _ = r.getTopDevices(ctx, urlID, from, to, 5)

	return stats, nil
}

func (r *ClickRepository) getTopReferrers(ctx context.Context, urlID string, from, to time.Time, limit int) ([]entity.ReferrerStat, error) {
	query := `
		SELECT COALESCE(NULLIF(referrer, ''), 'Direct') as referrer, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.ReferrerStat
	for rows.Next() {
		var stat entity.ReferrerStat
		if err := rows.Scan(&stat.Referrer, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (r *ClickRepository) getTopCountries(ctx context.Context, urlID string, from, to time.Time, limit int) ([]entity.CountryStat, error) {
	query := `
		SELECT COALESCE(NULLIF(country, ''), 'Unknown') as country, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY country
		ORDER BY count DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.CountryStat
	for rows.Next() {
		var stat entity.CountryStat
		if err := rows.Scan(&stat.Country, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (r *ClickRepository) getTopBrowsers(ctx context.Context, urlID string, from, to time.Time, limit int) ([]entity.BrowserStat, error) {
	query := `
		SELECT COALESCE(NULLIF(browser, ''), 'Unknown') as browser, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY browser
		ORDER BY count DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.BrowserStat
	for rows.Next() {
		var stat entity.BrowserStat
		if err := rows.Scan(&stat.Browser, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (r *ClickRepository) getTopDevices(ctx context.Context, urlID string, from, to time.Time, limit int) ([]entity.DeviceStat, error) {
	query := `
		SELECT COALESCE(NULLIF(device, ''), 'Unknown') as device, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY device
		ORDER BY count DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.DeviceStat
	for rows.Next() {
		var stat entity.DeviceStat
		if err := rows.Scan(&stat.Device, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (r *ClickRepository) CountByURLID(ctx context.Context, urlID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM clicks WHERE url_id = $1`
	err := r.db.QueryRowContext(ctx, query, urlID).Scan(&count)
	return count, err
}

func (r *ClickRepository) CountUniqueByURLID(ctx context.Context, urlID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(DISTINCT ip_address) FROM clicks WHERE url_id = $1`
	err := r.db.QueryRowContext(ctx, query, urlID).Scan(&count)
	return count, err
}
