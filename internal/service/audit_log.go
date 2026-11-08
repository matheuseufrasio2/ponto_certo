package service

import (
	"context"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// AuditLogReaderRepository defines read access to immutable compliance records.
type AuditLogReaderRepository interface {
	List(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error)
}

// AuditLogService provides authorized queries into system audit logs.
type AuditLogService struct {
	repo AuditLogReaderRepository
}

// NewAuditLogService instantiates a new audit log service.
func NewAuditLogService(repo AuditLogReaderRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// ListOrgAuditLogs retrieves audit logs scoped to the caller's organization (for owners and super admins).
func (s *AuditLogService) ListOrgAuditLogs(ctx context.Context, claims *UserClaims, limit, offset int) ([]domain.AuditLog, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if claims.Role != domain.RoleSuperAdmin && claims.Role != domain.RoleOrgOwner {
		return nil, domain.ErrForbidden
	}

	var orgID *string
	if claims.Role == domain.RoleOrgOwner {
		if claims.OrganizationID == nil || *claims.OrganizationID == "" {
			return nil, domain.ErrForbidden
		}
		orgID = claims.OrganizationID
	}

	return s.repo.List(ctx, orgID, limit, offset)
}

// ListGlobalAuditLogs retrieves platform-wide audit logs (exclusive to super admins).
func (s *AuditLogService) ListGlobalAuditLogs(ctx context.Context, claims *UserClaims, limit, offset int) ([]domain.AuditLog, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, nil, limit, offset)
}
