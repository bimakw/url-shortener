package utm

import (
	"net/url"
	"strings"
)

type Params struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

func Build(originalURL string, params Params) (string, error) {
	u, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	q := u.Query()

	if params.Source != "" {
		q.Set("utm_source", params.Source)
	}
	if params.Medium != "" {
		q.Set("utm_medium", params.Medium)
	}
	if params.Campaign != "" {
		q.Set("utm_campaign", params.Campaign)
	}
	if params.Term != "" {
		q.Set("utm_term", params.Term)
	}
	if params.Content != "" {
		q.Set("utm_content", params.Content)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func HasUTM(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	q := u.Query()
	utmKeys := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"}

	for _, key := range utmKeys {
		if q.Get(key) != "" {
			return true
		}
	}
	return false
}

func Strip(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	utmKeys := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"}

	for _, key := range utmKeys {
		q.Del(key)
	}

	u.RawQuery = q.Encode()

	// Remove trailing ? if no query params remain
	result := u.String()
	result = strings.TrimSuffix(result, "?")

	return result, nil
}
