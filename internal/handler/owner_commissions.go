package handler

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// OwnerCommissionsManager defines the service contract for owner commission management.
type OwnerCommissionsManager interface {
	UpdateCommissions(ctx context.Context, claims *service.UserClaims, input domain.UpdateCommissionsInput) (*domain.UpdateCommissionsResponse, error)
}

// OwnerCommissionsHandler handles HTTP requests for owner 3-level commission rules.
type OwnerCommissionsHandler struct {
	svc OwnerCommissionsManager
}

// NewOwnerCommissionsHandler creates an OwnerCommissionsHandler.
func NewOwnerCommissionsHandler(svc OwnerCommissionsManager) *OwnerCommissionsHandler {
	return &OwnerCommissionsHandler{svc: svc}
}

// UpdateCommissions godoc
// @Summary      Update barber commissions
// @Description  Updates commission percentage across the organization, a single unit, or a specific barber.
// @Tags         owner
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.UpdateCommissionsInput  true  "Commission update scope and percentage"
// @Success      200      {object}  domain.UpdateCommissionsResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /owner/commissions [put]
func (h *OwnerCommissionsHandler) UpdateCommissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	var input domain.UpdateCommissionsInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	resp, err := h.svc.UpdateCommissions(r.Context(), claims, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, resp)
}
