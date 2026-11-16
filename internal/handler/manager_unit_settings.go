package handler

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// ManagerUnitSettingsService defines the contract for unit settings operations for managers.
type ManagerUnitSettingsService interface {
	GetUnitSettings(ctx context.Context, unitID string, claims *service.UserClaims) (*domain.UnitSettingsResponse, error)
	UpdateUnitSettings(ctx context.Context, unitID string, input domain.UpdateUnitSettingsInput, claims *service.UserClaims) (*domain.UnitSettingsResponse, error)
}

// ManagerUnitSettingsHandler handles HTTP endpoints for branch unit operational settings.
type ManagerUnitSettingsHandler struct {
	svc ManagerUnitSettingsService
}

// NewManagerUnitSettingsHandler instantiates a ManagerUnitSettingsHandler.
func NewManagerUnitSettingsHandler(svc ManagerUnitSettingsService) *ManagerUnitSettingsHandler {
	return &ManagerUnitSettingsHandler{svc: svc}
}

// GetUnitSettings godoc
// @Summary      Get operational settings and operating hours for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Produce      json
// @Param        unit_id   query     string  false  "Unit UUID (defaults to manager's unit)"
// @Success      200       {object}  domain.UnitSettingsResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      404       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/unit-settings [get]
func (h *ManagerUnitSettingsHandler) GetUnitSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	settings, err := h.svc.GetUnitSettings(r.Context(), unitID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, settings)
}

// UpdateUnitSettings godoc
// @Summary      Update operational settings and operating hours for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        unit_id   query     string                         false  "Unit UUID (defaults to manager's unit)"
// @Param        request   body      domain.UpdateUnitSettingsInput true   "Unit settings payload"
// @Success      200       {object}  domain.UnitSettingsResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      404       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/unit-settings [put]
func (h *ManagerUnitSettingsHandler) UpdateUnitSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var input domain.UpdateUnitSettingsInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	settings, err := h.svc.UpdateUnitSettings(r.Context(), unitID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, settings)
}
