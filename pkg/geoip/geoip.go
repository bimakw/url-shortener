package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type GeoInfo struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	Region      string `json:"regionName"`
	ISP         string `json:"isp"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Lookup(ctx context.Context, ip string) (*GeoInfo, error) {
	// Skip private/local IPs
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return &GeoInfo{
			Country: "Local",
			City:    "Local",
		}, nil
	}

	// Use ip-api.com (free, no API key needed, 45 req/min)
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,regionName,city,lat,lon,isp", ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		GeoInfo
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status == "fail" {
		return &GeoInfo{
			Country: "Unknown",
			City:    "Unknown",
		}, nil
	}

	return &result.GeoInfo, nil
}
