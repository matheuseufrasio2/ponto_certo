package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// DevHandler handles developer testing discovery requests.
type DevHandler struct {
	devService *service.DevService
}

// NewDevHandler creates a new DevHandler.
func NewDevHandler(devService *service.DevService) *DevHandler {
	return &DevHandler{devService: devService}
}

// ListSuperAdmins returns all platform super admins for dev authentication.
// @Summary Listar super admins (Dev Only)
// @Description Retorna todos os usuários SUPER_ADMIN cadastrados para alimentar a aba de Super Admin no Dev Auth Modal.
// @Tags Dev
// @Produce json
// @Success 200 {array} domain.DevSuperAdmin "Lista de super admins"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /dev/super-admins [get]
func (h *DevHandler) ListSuperAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.devService.GetSuperAdmins(r.Context())
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, admins)
}

// ListTenants returns paginated tenants filtered by query or scenario.
// @Summary Listar tenants paginados (Dev Only)
// @Description Retorna barbearias/tenants para alimentar o select de empresas e cenários no Dev Auth Modal.
// @Tags Dev
// @Produce json
// @Param q query string false "Termo de busca textual"
// @Param scenario query string false "Cenário de teste (ex: past_due)"
// @Param page query int false "Página (padrão 1)"
// @Param limit query int false "Limite por página (padrão 10)"
// @Success 200 {object} domain.DevTenantsResponse "Tenants paginados"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /dev/tenants [get]
func (h *DevHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	scenario := r.URL.Query().Get("scenario")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	res, err := h.devService.GetTenants(r.Context(), q, scenario, page, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, res)
}

// ListTenantUsers returns all users associated with a specific tenant.
// @Summary Listar usuários do tenant (Dev Only)
// @Description Retorna todos os colaboradores vinculados a uma empresa selecionada.
// @Tags Dev
// @Produce json
// @Param id path string true "ID do Tenant"
// @Success 200 {array} domain.DevTenantUser "Usuários do tenant"
// @Failure 404 {object} ErrorResponse "Tenant não encontrado"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /dev/tenants/{id}/users [get]
func (h *DevHandler) ListTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" || !IsValidUUID(tenantID) {
		WriteError(w, http.StatusBadRequest, "invalid tenant id format, expected uuid")
		return
	}

	users, err := h.devService.GetTenantUsers(r.Context(), tenantID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, users)
}

// ListUsers returns paginated users filtered by role.
// @Summary Listar usuários por cargo (Dev Only)
// @Description Retorna usuários filtrados por cargo com scroll infinito / paginação para o Dev Auth Modal.
// @Tags Dev
// @Produce json
// @Param role query string false "Cargo do usuário (ORG_OWNER, UNIT_MANAGER, RECEPTIONIST, BARBER)"
// @Param page query int false "Página (padrão 1)"
// @Param limit query int false "Limite por página (padrão 10)"
// @Success 200 {object} domain.DevUsersResponse "Usuários paginados"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /dev/users [get]
func (h *DevHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	res, err := h.devService.GetUsersByRole(r.Context(), role, page, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, res)
}
