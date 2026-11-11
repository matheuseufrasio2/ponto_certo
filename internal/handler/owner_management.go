package handler

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
)

// OrganizationProfileManager defines operations for managing organization corporate profile.
type OrganizationProfileManager interface {
	GetProfile(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error)
	UpdateProfile(ctx context.Context, orgID string, input domain.UpdateOrganizationInput) (*domain.OrganizationProfileResponse, error)
}

// OwnerManagementHandler handles HTTP requests for organization owner administration.
type OwnerManagementHandler struct {
	orgService OrganizationProfileManager
}

// NewOwnerManagementHandler creates a new OwnerManagementHandler instance.
func NewOwnerManagementHandler(orgService OrganizationProfileManager) *OwnerManagementHandler {
	return &OwnerManagementHandler{orgService: orgService}
}

// GetOrganization godoc
// @Summary Obter perfil da organização do Dono
// @Description Retorna os dados cadastrais da organização do Dono autenticado (Razão social, Nome fantasia, CNPJ/CPF, Telefone, Logo).
// @Tags Owner
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.OrganizationProfileResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/organization [get]
func (h *OwnerManagementHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	if claims.OrganizationID == nil || *claims.OrganizationID == "" {
		HandleError(w, domain.ErrForbidden, r)
		return
	}

	profile, err := h.orgService.GetProfile(r.Context(), *claims.OrganizationID)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, profile)
}

// UpdateOrganization godoc
// @Summary Atualizar dados cadastrais da organização
// @Description Permite ao Dono atualizar razão social, nome fantasia, documento (CNPJ/CPF), telefone e logotipo da empresa.
// @Tags Owner
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.UpdateOrganizationInput true "Dados cadastrais da organização"
// @Success 200 {object} domain.OrganizationProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/organization [put]
func (h *OwnerManagementHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	if claims.OrganizationID == nil || *claims.OrganizationID == "" {
		HandleError(w, domain.ErrForbidden, r)
		return
	}

	var input domain.UpdateOrganizationInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error(), domain.ErrCodeValidationFailed)
		return
	}

	profile, err := h.orgService.UpdateProfile(r.Context(), *claims.OrganizationID, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, profile)
}
