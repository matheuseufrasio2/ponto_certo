package handler

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// EmployeeLeaveManager defines operations for managing employee vacations and leaves.
type EmployeeLeaveManager interface {
	CreateLeave(ctx context.Context, input domain.CreateEmployeeLeaveInput, claims *service.UserClaims) (*domain.EmployeeLeave, error)
	ListLeaves(ctx context.Context, claims *service.UserClaims, unitID *string, status string) ([]domain.EmployeeLeave, error)
	CancelLeave(ctx context.Context, leaveID string, claims *service.UserClaims) error
}

// EmployeeLeaveHandler handles HTTP requests for employee vacations and leaves.
type EmployeeLeaveHandler struct {
	svc EmployeeLeaveManager
}

// NewEmployeeLeaveHandler instantiates EmployeeLeaveHandler.
func NewEmployeeLeaveHandler(svc EmployeeLeaveManager) *EmployeeLeaveHandler {
	return &EmployeeLeaveHandler{svc: svc}
}

// CreateLeave godoc
// @Summary Schedule vacation or leave for an employee
// @Description Creates a vacation or absence period for any staff member (barber, receptionist, manager, owner)
// @Tags owner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateEmployeeLeaveInput true "Leave details"
// @Success 201 {object} domain.EmployeeLeave
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/leaves [post]
func (h *EmployeeLeaveHandler) CreateLeave(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	var req domain.CreateEmployeeLeaveInput
	if err := ReadJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	leave, err := h.svc.CreateLeave(r.Context(), req, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteCreated(w, leave)
}

// ListLeaves godoc
// @Summary List employee vacations and leaves
// @Description Retrieves all employee leaves in the organization, with optional status and unit filters
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param unit_id query string false "Filter by unit UUID"
// @Param status query string false "Filter status: all, active, upcoming, past (default: all)"
// @Success 200 {array} domain.EmployeeLeave
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/leaves [get]
func (h *EmployeeLeaveHandler) ListLeaves(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	var unitID *string
	if u := r.URL.Query().Get("unit_id"); u != "" && domain.IsValidUUID(u) {
		unitID = &u
	}

	leaves, err := h.svc.ListLeaves(r.Context(), claims, unitID, status)
	if err != nil {
		HandleError(w, err)
		return
	}

	if leaves == nil {
		leaves = []domain.EmployeeLeave{}
	}
	_ = WriteJSON(w, http.StatusOK, leaves)
}

// CancelLeave godoc
// @Summary Cancel an employee leave
// @Description Cancels a vacation or absence period, restoring barber availability if applicable
// @Tags owner
// @Security BearerAuth
// @Param id path string true "Employee Leave UUID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/leaves/{id} [delete]
func (h *EmployeeLeaveHandler) CancelLeave(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	leaveID := chi.URLParam(r, "id")
	if leaveID == "" || !domain.IsValidUUID(leaveID) {
		WriteError(w, http.StatusBadRequest, "invalid leave id format, expected uuid")
		return
	}

	if err := h.svc.CancelLeave(r.Context(), leaveID, claims); err != nil {
		HandleError(w, err)
		return
	}

	WriteNoContent(w)
}
