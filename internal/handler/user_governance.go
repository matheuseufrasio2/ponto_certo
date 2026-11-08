package handler

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// UserGovernanceManager defines business operations for managing user lifecycle and status.
type UserGovernanceManager interface {
	ToggleUserStatus(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error)
}

// UserGovernanceHandler handles HTTP requests for user governance and status transitions.
type UserGovernanceHandler struct {
	svc UserGovernanceManager
}

// NewUserGovernanceHandler creates a new UserGovernanceHandler.
func NewUserGovernanceHandler(svc UserGovernanceManager) *UserGovernanceHandler {
	return &UserGovernanceHandler{svc: svc}
}

// ToggleUserStatus godoc
// @Summary Activate or deactivate a user
// @Description Updates user active status based on strict hierarchical RBAC rules (Super Admin > Owner > Manager)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Target User UUID"
// @Param body body domain.ToggleUserStatusInput true "Toggle status payload"
// @Success 200 {object} domain.ToggleUserStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/toggle-status [patch]
func (h *UserGovernanceHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" || !domain.IsValidUUID(targetUserID) {
		WriteError(w, http.StatusBadRequest, "invalid user id format, expected uuid")
		return
	}

	var req domain.ToggleUserStatusInput
	if err := ReadJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.ToggleUserStatus(r.Context(), claims, targetUserID, req.IsActive)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, resp)
}
