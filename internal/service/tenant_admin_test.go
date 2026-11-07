package service_test

import (
	"context"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockTenantAdminRepo struct {
	getMetricsFn    func(ctx context.Context) (*domain.SaaSMetrics, error)
	listTenantsFn   func(ctx context.Context, search string, limit, offset int) ([]domain.Organization, int, error)
	createTenantFn  func(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error)
	toggleStatusFn  func(ctx context.Context, tenantID string, isActive bool) error
	getUsersFn      func(ctx context.Context, tenantID string) ([]domain.User, error)
	getUserByIDFn   func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockTenantAdminRepo) GetSaaSPlatformKPIs(ctx context.Context) (*domain.SaaSMetrics, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx)
	}
	return &domain.SaaSMetrics{}, nil
}

func (m *mockTenantAdminRepo) ListTenants(ctx context.Context, search string, limit, offset int) ([]domain.Organization, int, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, search, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockTenantAdminRepo) CreateTenantWithAdmin(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error) {
	if m.createTenantFn != nil {
		return m.createTenantFn(ctx, input, passwordHash)
	}
	return &domain.Organization{ID: "new-org-1"}, nil
}

func (m *mockTenantAdminRepo) ToggleTenantActiveStatus(ctx context.Context, tenantID string, isActive bool) error {
	if m.toggleStatusFn != nil {
		return m.toggleStatusFn(ctx, tenantID, isActive)
	}
	return nil
}

func (m *mockTenantAdminRepo) GetTenantUsers(ctx context.Context, tenantID string) ([]domain.User, error) {
	if m.getUsersFn != nil {
		return m.getUsersFn(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockTenantAdminRepo) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, userID)
	}
	return nil, domain.ErrUserNotFound
}

func TestTenantAdminService_Impersonation(t *testing.T) {
	secret := "test-super-secret-key-32-chars-long!"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
		Email:  "superadmin@pontocerto.com.br",
	}

	targetBarber := &domain.User{
		ID:       "barber-1",
		Name:     "Marcos Barbeiro",
		Email:    "marcos@pontocerto.com.br",
		Role:     domain.RoleBarber,
		IsActive: true,
	}

	targetSuperAdmin := &domain.User{
		ID:       "super-admin-2",
		Name:     "Outro Admin",
		Email:    "outro@pontocerto.com.br",
		Role:     domain.RoleSuperAdmin,
		IsActive: true,
	}

	targetInactive := &domain.User{
		ID:       "inactive-1",
		Name:     "Usuario Inativo",
		Role:     domain.RoleReceptionist,
		IsActive: false,
	}

	repo := &mockTenantAdminRepo{
		getUserByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			switch id {
			case "barber-1":
				return targetBarber, nil
			case "super-admin-2":
				return targetSuperAdmin, nil
			case "inactive-1":
				return targetInactive, nil
			case "super-admin-1":
				return &domain.User{ID: "super-admin-1", Role: domain.RoleSuperAdmin}, nil
			default:
				return nil, domain.ErrUserNotFound
			}
		},
	}

	svc := service.NewTenantAdminService(repo, secret)
	ctx := context.Background()

	t.Run("successfully generates impersonation token for barber", func(t *testing.T) {
		resp, err := svc.ImpersonateUser(ctx, adminClaims, "barber-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TargetUserID != "barber-1" {
			t.Errorf("expected target_user_id barber-1, got %s", resp.TargetUserID)
		}
		if resp.TargetRole != "BARBER" {
			t.Errorf("expected target_role BARBER, got %s", resp.TargetRole)
		}
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("rejects impersonating self", func(t *testing.T) {
		_, err := svc.ImpersonateUser(ctx, adminClaims, "super-admin-1")
		if err != domain.ErrCannotImpersonateSelf {
			t.Errorf("expected ErrCannotImpersonateSelf, got %v", err)
		}
	})

	t.Run("rejects impersonating another super admin", func(t *testing.T) {
		_, err := svc.ImpersonateUser(ctx, adminClaims, "super-admin-2")
		if err != domain.ErrCannotImpersonateSuperAdmin {
			t.Errorf("expected ErrCannotImpersonateSuperAdmin, got %v", err)
		}
	})

	t.Run("rejects impersonating inactive user", func(t *testing.T) {
		_, err := svc.ImpersonateUser(ctx, adminClaims, "inactive-1")
		if err == nil {
			t.Fatal("expected error for inactive user, got nil")
		}
	})

	t.Run("rejects unauthorized caller without super admin role", func(t *testing.T) {
		ownerClaims := &service.UserClaims{
			UserID: "owner-1",
			Role:   domain.RoleOrgOwner,
		}
		_, err := svc.ImpersonateUser(ctx, ownerClaims, "barber-1")
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}
