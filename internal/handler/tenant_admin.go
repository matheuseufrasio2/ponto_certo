package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// TenantAdminManager specifies business logic methods for SaaS platform management.
type TenantAdminManager interface {
	GetSaaSMetrics(ctx context.Context, claims *service.UserClaims) (*domain.SaaSMetrics, error)
	ListTenants(ctx context.Context, claims *service.UserClaims, search, status, plan string, page, perPage int) (*domain.PaginatedOrganizations, error)
	GetTenantOverview(ctx context.Context, claims *service.UserClaims, tenantID string) (*domain.TenantOverview, error)
	UpdateTenantPlan(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.UpdateTenantPlanInput) (*domain.Organization, error)
	CreateTenant(ctx context.Context, claims *service.UserClaims, input domain.CreateTenantInput) (*domain.Organization, error)
	ToggleTenantStatus(ctx context.Context, claims *service.UserClaims, tenantID string, isActive bool) error
	GetTenantUsers(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.User, error)
	GetTenantUnits(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.BarbershopUnit, error)
	ResetTenantPassword(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.ResetTenantPasswordInput) (*domain.ResetPasswordResponse, error)
	ImpersonateUser(ctx context.Context, claims *service.UserClaims, targetUserID string) (*domain.ImpersonateResponse, error)
}

// TenantAdminHandler handles HTTP requests for Super Admin platform governance.
type TenantAdminHandler struct {
	svc TenantAdminManager
}

// NewTenantAdminHandler creates a new TenantAdminHandler instance.
func NewTenantAdminHandler(svc TenantAdminManager) *TenantAdminHandler {
	return &TenantAdminHandler{svc: svc}
}

// ToggleStatusRequest represents the payload for enabling/disabling a tenant.
type ToggleStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// GetPlatformMetrics godoc
// @Summary Get SaaS platform KPIs
// @Description Returns platform-wide MRR, ARR, active tenants, churn rate, and total units
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SaaSMetrics
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/metrics [get]
func (h *TenantAdminHandler) GetPlatformMetrics(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	metrics, err := h.svc.GetSaaSMetrics(r.Context(), claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, metrics)
}

// ListTenants godoc
// @Summary List and search organizations
// @Description Returns paginated organizations matching search term, status, and plan filters
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name, trade name or document"
// @Param status query string false "Filter by subscription status (ACTIVE, PAST_DUE, TRIAL, CANCELLED) or BLOCKED"
// @Param plan query string false "Filter by plan (BASIC, PRO, ENTERPRISE)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Items per page (default 15)"
// @Success 200 {object} domain.PaginatedOrganizations
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants [get]
func (h *TenantAdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	plan := r.URL.Query().Get("plan")
	page := 1
	perPage := 15

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if val, err := strconv.Atoi(pp); err == nil && val > 0 {
			perPage = val
		}
	}

	result, err := h.svc.ListTenants(r.Context(), claims, search, status, plan, page, perPage)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, result)
}

// CreateTenant godoc
// @Summary Onboard a new barbershop franchise tenant
// @Description Creates organization, headquarters branch and owner user in a transaction
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateTenantInput true "Tenant creation payload"
// @Success 201 {object} domain.Organization
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants [post]
func (h *TenantAdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	var input domain.CreateTenantInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	org, err := h.svc.CreateTenant(r.Context(), claims, input)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteCreated(w, org)
}

// ToggleTenantStatus godoc
// @Summary Activate or suspend an organization
// @Description Updates organization active status and revokes or restores access
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body ToggleStatusRequest true "Toggle status payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/toggle-status [patch]
func (h *TenantAdminHandler) ToggleTenantStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	var req ToggleStatusRequest
	if err := ReadJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.ToggleTenantStatus(r.Context(), claims, tenantID, req.IsActive); err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"tenant_id": tenantID,
		"is_active": req.IsActive,
	})
}

// GetTenantUsers godoc
// @Summary List all users in a tenant organization
// @Description Returns users of a tenant for diagnostic impersonation
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {array} domain.User
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/users [get]
func (h *TenantAdminHandler) GetTenantUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	users, err := h.svc.GetTenantUsers(r.Context(), claims, tenantID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, users)
}

// GetTenantUnits godoc
// @Summary List all physical branches/units in a tenant organization
// @Description Returns branches/units belonging to an organization for Super Admin inspection
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {array} domain.BarbershopUnit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/units [get]
func (h *TenantAdminHandler) GetTenantUnits(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	units, err := h.svc.GetTenantUnits(r.Context(), claims, tenantID)
	if err != nil {
		HandleError(w, err)
		return
	}

	if units == nil {
		units = []domain.BarbershopUnit{}
	}

	_ = WriteJSON(w, http.StatusOK, units)
}

// GetTenantOverview godoc
// @Summary Get aggregated engagement and financial metrics for a tenant
// @Description Returns 30-day GMV, 30-day appointment volume, total barbers, and overdue status
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.TenantOverview
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/overview [get]
func (h *TenantAdminHandler) GetTenantOverview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	overview, err := h.svc.GetTenantOverview(r.Context(), claims, tenantID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, overview)
}

// UpdateTenantPlan godoc
// @Summary Upgrade or downgrade a tenant's subscription plan and fee
// @Description Updates organization plan tier (BASIC, PRO, ENTERPRISE) and monthly fee
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.UpdateTenantPlanInput true "Plan update payload"
// @Success 200 {object} domain.Organization
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/plan [patch]
func (h *TenantAdminHandler) UpdateTenantPlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	var input domain.UpdateTenantPlanInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	org, err := h.svc.UpdateTenantPlan(r.Context(), claims, tenantID, input)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, org)
}

// ResetTenantPassword godoc
// @Summary Administratively reset franchise owner's password
// @Description Sets a specified password or generates a temporary credential for tenant owner
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.ResetTenantPasswordInput false "Optional custom password payload"
// @Success 200 {object} domain.ResetPasswordResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/tenants/{id}/reset-password [post]
func (h *TenantAdminHandler) ResetTenantPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		WriteError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	var input domain.ResetTenantPasswordInput
	if r.Body != nil && r.ContentLength > 0 {
		_ = ReadJSON(w, r, &input)
	}

	resp, err := h.svc.ResetTenantPassword(r.Context(), claims, tenantID, input)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, resp)
}

// ImpersonateUser godoc
// @Summary Impersonate a user for diagnostic support
// @Description Issues a temporary 30-minute JWT token with target user identity
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Target User UUID"
// @Success 200 {object} domain.ImpersonateResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/{id}/impersonate [post]
func (h *TenantAdminHandler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		WriteError(w, http.StatusBadRequest, "target user id is required")
		return
	}

	resp, err := h.svc.ImpersonateUser(r.Context(), claims, targetUserID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, resp)
}
