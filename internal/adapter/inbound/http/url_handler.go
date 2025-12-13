package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bimakw/url-shortener/internal/application/usecase"
	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/go-playground/validator/v10"
)

type URLHandler struct {
	urlUseCase *usecase.URLUseCase
	validate   *validator.Validate
}

func NewURLHandler(urlUseCase *usecase.URLUseCase) *URLHandler {
	return &URLHandler{
		urlUseCase: urlUseCase,
		validate:   validator.New(),
	}
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req entity.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from context if authenticated
	if userID, ok := r.Context().Value("user_id").(string); ok {
		req.UserID = userID
	}

	response, err := h.urlUseCase.CreateShortURL(r.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrAliasExists:
			Error(w, http.StatusConflict, "Custom alias already exists")
		case usecase.ErrInvalidURL:
			Error(w, http.StatusBadRequest, "Invalid URL format")
		default:
			Error(w, http.StatusInternalServerError, "Failed to create short URL")
		}
		return
	}

	Success(w, http.StatusCreated, "Short URL created", response)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		Error(w, http.StatusBadRequest, "Short code is required")
		return
	}

	url, err := h.urlUseCase.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		switch err {
		case usecase.ErrURLNotFound:
			Error(w, http.StatusNotFound, "URL not found")
		case usecase.ErrURLExpired:
			Error(w, http.StatusGone, "URL has expired")
		case usecase.ErrURLInactive:
			Error(w, http.StatusGone, "URL is no longer active")
		default:
			Error(w, http.StatusInternalServerError, "Failed to redirect")
		}
		return
	}

	// Record click
	click := &entity.Click{
		ShortCode: shortCode,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Referrer:  r.Referer(),
	}

	// Parse user agent for device/browser info
	parseUserAgent(click)

	_ = h.urlUseCase.RecordClick(r.Context(), click)

	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) GetURLInfo(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		Error(w, http.StatusBadRequest, "Short code is required")
		return
	}

	response, err := h.urlUseCase.GetURLByShortCode(r.Context(), shortCode)
	if err != nil {
		if err == usecase.ErrURLNotFound {
			Error(w, http.StatusNotFound, "URL not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to get URL info")
		return
	}

	Success(w, http.StatusOK, "URL info retrieved", response)
}

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		Error(w, http.StatusBadRequest, "Short code is required")
		return
	}

	// Parse date range from query params
	from := time.Now().AddDate(0, 0, -30) // Default: last 30 days
	to := time.Now()

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsed
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsed.Add(24*time.Hour - time.Second) // End of day
		}
	}

	stats, err := h.urlUseCase.GetStats(r.Context(), shortCode, from, to)
	if err != nil {
		if err == usecase.ErrURLNotFound {
			Error(w, http.StatusNotFound, "URL not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	Success(w, http.StatusOK, "Stats retrieved", stats)
}

func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	urls, err := h.urlUseCase.GetUserURLs(r.Context(), userID, limit, offset)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get URLs")
		return
	}

	Success(w, http.StatusOK, "URLs retrieved", urls)
}

func (h *URLHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		Error(w, http.StatusBadRequest, "URL ID is required")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	err := h.urlUseCase.DeleteURL(r.Context(), id, userID)
	if err != nil {
		if err == usecase.ErrURLNotFound {
			Error(w, http.StatusNotFound, "URL not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to delete URL")
		return
	}

	Success(w, http.StatusOK, "URL deleted", nil)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

func parseUserAgent(click *entity.Click) {
	ua := strings.ToLower(click.UserAgent)

	// Simple device detection
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		click.Device = "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		click.Device = "Tablet"
	} else {
		click.Device = "Desktop"
	}

	// Simple browser detection
	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		click.Browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		click.Browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		click.Browser = "Safari"
	case strings.Contains(ua, "edg"):
		click.Browser = "Edge"
	default:
		click.Browser = "Other"
	}

	// Simple OS detection
	switch {
	case strings.Contains(ua, "windows"):
		click.OS = "Windows"
	case strings.Contains(ua, "mac"):
		click.OS = "macOS"
	case strings.Contains(ua, "linux"):
		click.OS = "Linux"
	case strings.Contains(ua, "android"):
		click.OS = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		click.OS = "iOS"
	default:
		click.OS = "Other"
	}
}
