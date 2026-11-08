package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// UnitOperationsManager defines required operations for the unit manager HTTP handler.
type UnitOperationsManager interface {
	GetManagerMetrics(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.ManagerDashboardMetrics, error)
	GetDailyCashClosing(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.CashClosingSummary, error)
	UpdateBarberStatus(ctx context.Context, unitID, barberID string, input domain.UpdateBarberStatusInput, claims *service.UserClaims) error
	ToggleService(ctx context.Context, unitID, serviceID string, input domain.ToggleServiceInput, claims *service.UserClaims) error
	ListBarbers(ctx context.Context, unitID string, status string, claims *service.UserClaims) ([]domain.Barber, error)
}

// UnitOperationsHandler manages unit manager endpoints.
type UnitOperationsHandler struct {
	svc UnitOperationsManager
}

// NewUnitOperationsHandler instantiates a UnitOperationsHandler.
func NewUnitOperationsHandler(svc UnitOperationsManager) *UnitOperationsHandler {
	return &UnitOperationsHandler{svc: svc}
}

func resolveUnitID(r *http.Request, claims *service.UserClaims) string {
	unitID := r.URL.Query().Get("unit_id")
	if unitID != "" {
		return unitID
	}
	if claims != nil && claims.UnitID != nil {
		return *claims.UnitID
	}
	return ""
}

func parseDateQuery(r *http.Request) time.Time {
	dateStr := r.URL.Query().Get("date")
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			return parsed
		}
	}
	return time.Now().UTC()
}

// GetMetrics handles GET /api/v1/manager/metrics.
func (h *UnitOperationsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := resolveUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	date := parseDateQuery(r)
	metrics, err := h.svc.GetManagerMetrics(r.Context(), unitID, date, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, metrics)
}

// GetCashClosing handles GET /api/v1/manager/cash-closing.
func (h *UnitOperationsHandler) GetCashClosing(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := resolveUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	date := parseDateQuery(r)
	closing, err := h.svc.GetDailyCashClosing(r.Context(), unitID, date, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, closing)
}

// ListBarbers handles GET /api/v1/manager/barbers.
// @Summary List barbers of a branch
// @Description Returns active or inactive barbers assigned to a specific branch unit
// @Tags manager
// @Produce json
// @Security BearerAuth
// @Param unit_id query string false "Unit UUID (defaults to manager's assigned unit)"
// @Param status query string false "Filter barbers by status: active, inactive, all"
// @Success 200 {array} domain.Barber
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /manager/barbers [get]
func (h *UnitOperationsHandler) ListBarbers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := resolveUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	status := r.URL.Query().Get("status")
	barbers, err := h.svc.ListBarbers(r.Context(), unitID, status, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, barbers)
}

// UpdateBarberStatus handles PATCH /api/v1/manager/barbers/{barber_id}/status.
func (h *UnitOperationsHandler) UpdateBarberStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	barberID := chi.URLParam(r, "barber_id")
	if barberID == "" {
		WriteError(w, http.StatusBadRequest, "barber_id path parameter is required")
		return
	}

	unitID := resolveUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var input domain.UpdateBarberStatusInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateBarberStatus(r.Context(), unitID, barberID, input, claims); err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]any{
		"id":         barberID,
		"status":     input.Status,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// ToggleService handles PATCH /api/v1/manager/services/{service_id}/toggle.
func (h *UnitOperationsHandler) ToggleService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "service_id")
	if serviceID == "" {
		WriteError(w, http.StatusBadRequest, "service_id path parameter is required")
		return
	}

	unitID := resolveUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var input domain.ToggleServiceInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.ToggleService(r.Context(), unitID, serviceID, input, claims); err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]any{
		"id":         serviceID,
		"is_active":  input.IsActive,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}
