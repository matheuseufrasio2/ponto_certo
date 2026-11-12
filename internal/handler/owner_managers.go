package handler

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// OwnerManagersManager defines operations for managing unit managers under an owner's organization.
type OwnerManagersManager interface {
	ListManagers(ctx context.Context, claims *service.UserClaims) ([]domain.ManagerListItem, error)
	CreateManager(ctx context.Context, claims *service.UserClaims, input domain.CreateManagerInput) (*domain.ManagerListItem, error)
}

// OwnerManagersHandler handles HTTP endpoints for unit manager management.
type OwnerManagersHandler struct {
	svc OwnerManagersManager
}

// NewOwnerManagersHandler instantiates an OwnerManagersHandler.
func NewOwnerManagersHandler(svc OwnerManagersManager) *OwnerManagersHandler {
	return &OwnerManagersHandler{svc: svc}
}

// ListManagers godoc
// @Summary List unit managers of the organization
// @Description Returns all UNIT_MANAGER users under the authenticated owner's organization, ordered by unit then name.
// @Tags Owner
// @Security BearerAuth
// @Produce json
// @Success 200 {array} domain.ManagerListItem
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/managers [get]
func (h *OwnerManagersHandler) ListManagers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	managers, err := h.svc.ListManagers(r.Context(), claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, managers)
}

// CreateManager godoc
// @Summary Create a unit manager and link to a branch
// @Description Creates a new UNIT_MANAGER user and assigns them to the specified barbershop unit within the owner's organization.
// @Tags Owner
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateManagerInput true "Manager creation payload"
// @Success 201 {object} domain.ManagerListItem
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/managers [post]
func (h *OwnerManagersHandler) CreateManager(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	var input domain.CreateManagerInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error(), domain.ErrCodeValidationFailed)
		return
	}

	manager, err := h.svc.CreateManager(r.Context(), claims, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusCreated, manager)
}
