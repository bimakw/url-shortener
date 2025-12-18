package http

import (
	"encoding/json"
	"net/http"

	"github.com/bimakw/url-shortener/internal/application/usecase"
	"github.com/bimakw/url-shortener/internal/domain/entity"
	"github.com/go-playground/validator/v10"
)

type APIKeyHandler struct {
	apiKeyUseCase *usecase.APIKeyUseCase
	validate      *validator.Validate
}

func NewAPIKeyHandler(uc *usecase.APIKeyUseCase) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyUseCase: uc,
		validate:      validator.New(),
	}
}

func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req entity.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.apiKeyUseCase.CreateAPIKey(r.Context(), userID, req)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	Success(w, http.StatusCreated, "API key created. Store it safely, it won't be shown again.", response)
}

func (h *APIKeyHandler) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	keys, err := h.apiKeyUseCase.GetUserAPIKeys(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get API keys")
		return
	}

	Success(w, http.StatusOK, "API keys retrieved", keys)
}

func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		Error(w, http.StatusBadRequest, "API key ID is required")
		return
	}

	err := h.apiKeyUseCase.RevokeAPIKey(r.Context(), id, userID)
	if err != nil {
		if err == usecase.ErrAPIKeyNotFound {
			Error(w, http.StatusNotFound, "API key not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to revoke API key")
		return
	}

	Success(w, http.StatusOK, "API key revoked", nil)
}

func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		Error(w, http.StatusBadRequest, "API key ID is required")
		return
	}

	err := h.apiKeyUseCase.DeleteAPIKey(r.Context(), id, userID)
	if err != nil {
		if err == usecase.ErrAPIKeyNotFound {
			Error(w, http.StatusNotFound, "API key not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to delete API key")
		return
	}

	Success(w, http.StatusOK, "API key deleted", nil)
}
