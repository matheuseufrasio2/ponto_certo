package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// CatalogHandler manages HTTP handlers for catalog endpoints.
type CatalogHandler struct {
	catalogService *service.CatalogService
}

// NewCatalogHandler creates a new CatalogHandler instance.
func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

// GetUnitCatalog handles GET /api/v1/units/{unit_id}/catalog.
// @Summary Obter catálogo da unidade
// @Description Retorna a lista de barbeiros e serviços ativos disponíveis para atendimento na filial especificada.
// @Tags Catalog
// @Security BearerAuth
// @Produce json
// @Param unit_id path string true "ID da unidade"
// @Success 200 {object} service.CatalogResponse "Catálogo com lista de barbeiros e serviços"
// @Failure 401 {object} ErrorResponse "Não autenticado"
// @Failure 403 {object} ErrorResponse "Acesso não permitido à unidade informada"
// @Failure 404 {object} ErrorResponse "Unidade não encontrada"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /units/{unit_id}/catalog [get]
func (h *CatalogHandler) GetUnitCatalog(w http.ResponseWriter, r *http.Request) {
	unitID := chi.URLParam(r, "unit_id")
	if unitID == "" {
		unitID = r.URL.Query().Get("unit_id")
	}

	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	if !middleware.RequireUnitAccess(unitID, claims) {
		HandleError(w, domain.ErrForbidden)
		return
	}

	catalog, err := h.catalogService.GetUnitCatalog(r.Context(), unitID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, catalog)
}
