package preview

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type LinkPreview struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	URL         string `json:"url"`
}

var (
	titleRegex       = regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	ogTitleRegex     = regexp.MustCompile(`<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	ogTitleRegex2    = regexp.MustCompile(`<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:title["']`)
	ogDescRegex      = regexp.MustCompile(`<meta[^>]+property=["']og:description["'][^>]+content=["']([^"']+)["']`)
	ogDescRegex2     = regexp.MustCompile(`<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:description["']`)
	ogImageRegex     = regexp.MustCompile(`<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	ogImageRegex2    = regexp.MustCompile(`<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`)
	ogSiteNameRegex  = regexp.MustCompile(`<meta[^>]+property=["']og:site_name["'][^>]+content=["']([^"']+)["']`)
	ogSiteNameRegex2 = regexp.MustCompile(`<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:site_name["']`)
	descRegex        = regexp.MustCompile(`<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	descRegex2       = regexp.MustCompile(`<meta[^>]+content=["']([^"']+)["'][^>]+name=["']description["']`)
)

func Fetch(ctx context.Context, url string) (*LinkPreview, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; URLShortener/1.0; +https://github.com/bimakw/url-shortener)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read first 50KB only
	limitedReader := io.LimitReader(resp.Body, 50*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	html := string(body)
	preview := &LinkPreview{URL: url}

	// Extract title
	if match := ogTitleRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.Title = strings.TrimSpace(match[1])
	} else if match := ogTitleRegex2.FindStringSubmatch(html); len(match) > 1 {
		preview.Title = strings.TrimSpace(match[1])
	} else if match := titleRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.Title = strings.TrimSpace(match[1])
	}

	// Extract description
	if match := ogDescRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.Description = strings.TrimSpace(match[1])
	} else if match := ogDescRegex2.FindStringSubmatch(html); len(match) > 1 {
		preview.Description = strings.TrimSpace(match[1])
	} else if match := descRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.Description = strings.TrimSpace(match[1])
	} else if match := descRegex2.FindStringSubmatch(html); len(match) > 1 {
		preview.Description = strings.TrimSpace(match[1])
	}

	// Extract image
	if match := ogImageRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.Image = strings.TrimSpace(match[1])
	} else if match := ogImageRegex2.FindStringSubmatch(html); len(match) > 1 {
		preview.Image = strings.TrimSpace(match[1])
	}

	// Extract site name
	if match := ogSiteNameRegex.FindStringSubmatch(html); len(match) > 1 {
		preview.SiteName = strings.TrimSpace(match[1])
	} else if match := ogSiteNameRegex2.FindStringSubmatch(html); len(match) > 1 {
		preview.SiteName = strings.TrimSpace(match[1])
	}

	return preview, nil
}
