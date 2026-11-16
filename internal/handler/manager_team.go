package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// ManagerTeamService defines the contract for team operations for managers.
type ManagerTeamService interface {
	ListTeamMembers(ctx context.Context, unitID string, claims *service.UserClaims) ([]domain.TeamMemberItem, error)
	CreateTeamMember(ctx context.Context, unitID string, input domain.CreateTeamMemberInput, claims *service.UserClaims) (*domain.TeamMemberItem, error)
	GetWorkingHours(ctx context.Context, unitID, identifier string, claims *service.UserClaims) ([]domain.BarberWorkingHour, error)
	UpdateWorkingHours(ctx context.Context, unitID, identifier string, input domain.UpdateWorkingHoursInput, claims *service.UserClaims) ([]domain.BarberWorkingHour, error)
}

// ManagerTeamHandler handles HTTP endpoints for manager team members and working hours.
type ManagerTeamHandler struct {
	svc ManagerTeamService
}

// NewManagerTeamHandler instantiates a ManagerTeamHandler.
func NewManagerTeamHandler(svc ManagerTeamService) *ManagerTeamHandler {
	return &ManagerTeamHandler{svc: svc}
}

// ListTeamMembers godoc
// @Summary      List team members for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Produce      json
// @Param        unit_id   query     string  false  "Unit UUID (defaults to manager's unit)"
// @Success      200       {array}   domain.TeamMemberItem
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/team [get]
func (h *ManagerTeamHandler) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.svc.ListTeamMembers(r.Context(), unitID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if members == nil {
		members = []domain.TeamMemberItem{}
	}
	_ = WriteJSON(w, http.StatusOK, members)
}

// CreateTeamMember godoc
// @Summary      Create a new team member for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        unit_id   query     string                      false  "Unit UUID"
// @Param        request   body      domain.CreateTeamMemberInput  true   "Team member creation payload"
// @Success      201       {object}  domain.TeamMemberItem
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      409       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/team [post]
func (h *ManagerTeamHandler) CreateTeamMember(w http.ResponseWriter, r *http.Request) {
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

	var input domain.CreateTeamMemberInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	member, err := h.svc.CreateTeamMember(r.Context(), unitID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteCreated(w, member)
}

// GetWorkingHours godoc
// @Summary      Get working hours for team member
// @Tags         manager
// @Security     BearerAuth
// @Produce      json
// @Param        id        path      string  true   "Barber UUID or User UUID"
// @Param        unit_id   query     string  false  "Unit UUID"
// @Success      200       {array}   domain.BarberWorkingHour
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      404       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/team/{id}/working-hours [get]
func (h *ManagerTeamHandler) GetWorkingHours(w http.ResponseWriter, r *http.Request) {
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

	identifier := chi.URLParam(r, "id")
	if identifier == "" {
		WriteError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	hours, err := h.svc.GetWorkingHours(r.Context(), unitID, identifier, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if hours == nil {
		hours = []domain.BarberWorkingHour{}
	}
	_ = WriteJSON(w, http.StatusOK, hours)
}

// UpdateWorkingHours godoc
// @Summary      Update working hours for team member
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id        path      string                         true   "Barber UUID or User UUID"
// @Param        unit_id   query     string                         false  "Unit UUID"
// @Param        request   body      domain.UpdateWorkingHoursInput true   "Working hours payload"
// @Success      200       {array}   domain.BarberWorkingHour
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      404       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/team/{id}/working-hours [put]
func (h *ManagerTeamHandler) UpdateWorkingHours(w http.ResponseWriter, r *http.Request) {
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

	identifier := chi.URLParam(r, "id")
	if identifier == "" {
		WriteError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var input domain.UpdateWorkingHoursInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	hours, err := h.svc.UpdateWorkingHours(r.Context(), unitID, identifier, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if hours == nil {
		hours = []domain.BarberWorkingHour{}
	}
	_ = WriteJSON(w, http.StatusOK, hours)
}
