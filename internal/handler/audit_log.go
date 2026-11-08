package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// AuditLogManager defines query operations for audit logs.
type AuditLogManager interface {
	ListOrgAuditLogs(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error)
	ListGlobalAuditLogs(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error)
}

// AuditLogHandler handles HTTP requests for inspecting compliance audit logs.
type AuditLogHandler struct {
	svc AuditLogManager
}

// NewAuditLogHandler creates a new AuditLogHandler.
func NewAuditLogHandler(svc AuditLogManager) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}
	return limit, offset
}

// ListOrgAuditLogs godoc
// @Summary List organization audit logs
// @Description Returns audit trail of actions within the caller's organization (Owner / Super Admin)
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of records to return (default: 50)"
// @Param offset query int false "Pagination offset (default: 0)"
// @Success 200 {array} domain.AuditLog
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/audit-logs [get]
func (h *AuditLogHandler) ListOrgAuditLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	limit, offset := parsePagination(r)
	logs, err := h.svc.ListOrgAuditLogs(r.Context(), claims, limit, offset)
	if err != nil {
		HandleError(w, err)
		return
	}

	if logs == nil {
		logs = []domain.AuditLog{}
	}
	_ = WriteJSON(w, http.StatusOK, logs)
}

// ListGlobalAuditLogs godoc
// @Summary List global SaaS audit logs
// @Description Returns global platform-wide audit trail across all tenants (Super Admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of records to return (default: 50)"
// @Param offset query int false "Pagination offset (default: 0)"
// @Success 200 {array} domain.AuditLog
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/audit-logs [get]
func (h *AuditLogHandler) ListGlobalAuditLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	limit, offset := parsePagination(r)
	logs, err := h.svc.ListGlobalAuditLogs(r.Context(), claims, limit, offset)
	if err != nil {
		HandleError(w, err)
		return
	}

	if logs == nil {
		logs = []domain.AuditLog{}
	}
	_ = WriteJSON(w, http.StatusOK, logs)
}
