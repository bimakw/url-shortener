package entity

import (
	"time"
)

type Click struct {
	ID        string    `json:"id"`
	URLID     string    `json:"url_id"`
	ShortCode string    `json:"short_code"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer,omitempty"`
	Country   string    `json:"country,omitempty"`
	City      string    `json:"city,omitempty"`
	Device    string    `json:"device,omitempty"`
	Browser   string    `json:"browser,omitempty"`
	OS        string    `json:"os,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ClickStats struct {
	TotalClicks   int64            `json:"total_clicks"`
	UniqueClicks  int64            `json:"unique_clicks"`
	ClicksByDate  map[string]int64 `json:"clicks_by_date"`
	TopReferrers  []ReferrerStat   `json:"top_referrers"`
	TopCountries  []CountryStat    `json:"top_countries"`
	TopBrowsers   []BrowserStat    `json:"top_browsers"`
	TopDevices    []DeviceStat     `json:"top_devices"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

type BrowserStat struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

type DeviceStat struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}
