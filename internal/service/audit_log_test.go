package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockAuditLogRepo struct {
	listFn func(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error)
}

func (m *mockAuditLogRepo) List(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error) {
	if m.listFn != nil {
		return m.listFn(ctx, orgID, limit, offset)
	}
	return nil, nil
}

func TestAuditLogService_ListOrgAuditLogs(t *testing.T) {
	ctx := context.Background()
	org1 := "org-1"
	superAdmin := &service.UserClaims{UserID: "sa-1", Role: domain.RoleSuperAdmin}
	owner := &service.UserClaims{UserID: "owner-1", Role: domain.RoleOrgOwner, OrganizationID: &org1}
	manager := &service.UserClaims{UserID: "mgr-1", Role: domain.RoleUnitManager}

	t.Run("rejects caller without claims", func(t *testing.T) {
		svc := service.NewAuditLogService(&mockAuditLogRepo{})
		_, err := svc.ListOrgAuditLogs(ctx, nil, 10, 0)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects manager from listing org audit logs", func(t *testing.T) {
		svc := service.NewAuditLogService(&mockAuditLogRepo{})
		_, err := svc.ListOrgAuditLogs(ctx, manager, 10, 0)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("owner successfully queries scoped audit logs", func(t *testing.T) {
		repo := &mockAuditLogRepo{
			listFn: func(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error) {
				if orgID == nil || *orgID != org1 {
					t.Errorf("expected orgID %s, got %v", org1, orgID)
				}
				return []domain.AuditLog{
					{ID: "log-1", Action: "USER_DEACTIVATED", CreatedAt: time.Now()},
				}, nil
			},
		}
		svc := service.NewAuditLogService(repo)
		logs, err := svc.ListOrgAuditLogs(ctx, owner, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
	})

	t.Run("super admin queries global audit logs", func(t *testing.T) {
		repo := &mockAuditLogRepo{
			listFn: func(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error) {
				if orgID != nil {
					t.Errorf("expected nil orgID for global, got %v", orgID)
				}
				return []domain.AuditLog{
					{ID: "log-global", Action: "IMPERSONATE_USER", CreatedAt: time.Now()},
				}, nil
			},
		}
		svc := service.NewAuditLogService(repo)
		logs, err := svc.ListGlobalAuditLogs(ctx, superAdmin, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
	})
}
