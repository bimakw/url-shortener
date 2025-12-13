package http

import (
	"net/http"
	"strconv"

	"github.com/bimakw/url-shortener/internal/application/usecase"
	"github.com/skip2/go-qrcode"
)

type QRHandler struct {
	urlUseCase *usecase.URLUseCase
	baseURL    string
}

func NewQRHandler(urlUseCase *usecase.URLUseCase, baseURL string) *QRHandler {
	return &QRHandler{
		urlUseCase: urlUseCase,
		baseURL:    baseURL,
	}
}

func (h *QRHandler) GenerateQR(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		Error(w, http.StatusBadRequest, "Short code is required")
		return
	}

	// Verify URL exists
	_, err := h.urlUseCase.GetURLByShortCode(r.Context(), shortCode)
	if err != nil {
		if err == usecase.ErrURLNotFound {
			Error(w, http.StatusNotFound, "URL not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to verify URL")
		return
	}

	// Get size from query param (default 256)
	size := 256
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s >= 64 && s <= 1024 {
			size = s
		}
	}

	// Generate QR code
	shortURL := h.baseURL + "/" + shortCode
	png, err := qrcode.Encode(shortURL, qrcode.Medium, size)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}
