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

// OwnerUnitsManager defines operations for managing barbershop branches under an owner's organization.
type OwnerUnitsManager interface {
	ListUnits(ctx context.Context, claims *service.UserClaims, status string) ([]domain.BarbershopUnit, error)
	CreateUnit(ctx context.Context, claims *service.UserClaims, input domain.CreateUnitInput) (*domain.BarbershopUnit, error)
	UpdateUnit(ctx context.Context, claims *service.UserClaims, unitID string, input domain.UpdateUnitInput) (*domain.BarbershopUnit, error)
	ToggleUnitStatus(ctx context.Context, claims *service.UserClaims, unitID string, input domain.ToggleUnitStatusInput) (*domain.BarbershopUnit, error)
}

// OwnerUnitsHandler handles HTTP endpoints for branch/unit management.
type OwnerUnitsHandler struct {
	unitService OwnerUnitsManager
}

// NewOwnerUnitsHandler instantiates an OwnerUnitsHandler.
func NewOwnerUnitsHandler(unitService OwnerUnitsManager) *OwnerUnitsHandler {
	return &OwnerUnitsHandler{unitService: unitService}
}

func requireOwnerClaims(w http.ResponseWriter, r *http.Request) (*service.UserClaims, bool) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return nil, false
	}
	if claims.OrganizationID == nil || *claims.OrganizationID == "" {
		HandleError(w, domain.ErrForbidden, r)
		return nil, false
	}
	return claims, true
}

// ListUnits godoc
// @Summary Listar filiais/unidades da organização
// @Description Retorna as filiais da organização do Dono autenticado ordenadas por nome, com filtro opcional de status (active, inactive, all).
// @Tags Owner
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filtrar por status: active, inactive, all"
// @Success 200 {array} domain.BarbershopUnit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/units [get]
func (h *OwnerUnitsHandler) ListUnits(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireOwnerClaims(w, r)
	if !ok {
		return
	}

	status := r.URL.Query().Get("status")
	units, err := h.unitService.ListUnits(r.Context(), claims, status)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if units == nil {
		units = []domain.BarbershopUnit{}
	}

	_ = WriteJSON(w, http.StatusOK, units)
}

// CreateUnit godoc
// @Summary Cadastrar nova filial/unidade
// @Description Cria uma nova filial para a organização do Dono autenticado (chair_count padrão é 4 se omitido).
// @Tags Owner
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateUnitInput true "Dados da nova filial"
// @Success 201 {object} domain.BarbershopUnit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/units [post]
func (h *OwnerUnitsHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireOwnerClaims(w, r)
	if !ok {
		return
	}

	var input domain.CreateUnitInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	unit, err := h.unitService.CreateUnit(r.Context(), claims, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteCreated(w, unit)
}

// UpdateUnit godoc
// @Summary Atualizar dados de uma filial/unidade
// @Description Atualiza nome, telefone, endereço e quantidade de cadeiras de uma filial pertencente à organização do Dono.
// @Tags Owner
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "UUID da filial"
// @Param request body domain.UpdateUnitInput true "Dados atualizados da filial"
// @Success 200 {object} domain.BarbershopUnit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/units/{id} [put]
func (h *OwnerUnitsHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireOwnerClaims(w, r)
	if !ok {
		return
	}

	unitID := chi.URLParam(r, "id")
	if !IsValidUUID(unitID) {
		HandleError(w, domain.ErrInvalidUUID, r)
		return
	}

	var input domain.UpdateUnitInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	unit, err := h.unitService.UpdateUnit(r.Context(), claims, unitID, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, unit)
}

// ToggleUnitStatus godoc
// @Summary Ativar ou pausar uma filial/unidade
// @Description Alterna o status ativo (is_active) de uma filial. Unidades pausadas bloqueiam novos agendamentos e o login da equipe local.
// @Tags Owner
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "UUID da filial"
// @Param request body domain.ToggleUnitStatusInput true "Novo status de ativação"
// @Success 200 {object} domain.BarbershopUnit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/units/{id}/status [patch]
func (h *OwnerUnitsHandler) ToggleUnitStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireOwnerClaims(w, r)
	if !ok {
		return
	}

	unitID := chi.URLParam(r, "id")
	if !IsValidUUID(unitID) {
		HandleError(w, domain.ErrInvalidUUID, r)
		return
	}

	var input domain.ToggleUnitStatusInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	unit, err := h.unitService.ToggleUnitStatus(r.Context(), claims, unitID, input)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, unit)
}
