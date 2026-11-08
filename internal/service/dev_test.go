package service_test

import (
	"context"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockDevRepo struct {
	listSuperAdminsFn func(ctx context.Context) ([]domain.DevSuperAdmin, error)
	listTenantsFn     func(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error)
	listTenantUsersFn func(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error)
	listUsersByRoleFn func(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error)
}

func (m *mockDevRepo) ListSuperAdmins(ctx context.Context) ([]domain.DevSuperAdmin, error) {
	if m.listSuperAdminsFn != nil {
		return m.listSuperAdminsFn(ctx)
	}
	return nil, nil
}

func (m *mockDevRepo) ListTenants(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, search, scenario, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockDevRepo) ListTenantUsers(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error) {
	if m.listTenantUsersFn != nil {
		return m.listTenantUsersFn(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockDevRepo) ListUsersByRole(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error) {
	if m.listUsersByRoleFn != nil {
		return m.listUsersByRoleFn(ctx, role, limit, offset)
	}
	return nil, 0, nil
}

func TestDevService_GetSuperAdmins(t *testing.T) {
	repo := &mockDevRepo{
		listSuperAdminsFn: func(ctx context.Context) ([]domain.DevSuperAdmin, error) {
			return []domain.DevSuperAdmin{
				{
					ID:    "c0000000-0000-0000-0000-000000000001",
					Name:  "Carlos Super Admin",
					Email: "superadmin@pontocerto.com.br",
					Role:  "SUPER_ADMIN",
				},
			}, nil
		},
	}

	svc := service.NewDevService(repo)
	admins, err := svc.GetSuperAdmins(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if len(admins) != 1 {
		t.Fatalf("expected 1 admin, got: %d", len(admins))
	}
	if admins[0].Email != "superadmin@pontocerto.com.br" {
		t.Errorf("expected email superadmin@pontocerto.com.br, got: %s", admins[0].Email)
	}
}

func TestDevService_GetTenants(t *testing.T) {
	repo := &mockDevRepo{
		listTenantsFn: func(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error) {
			if limit != 10 || offset != 0 {
				t.Errorf("expected limit=10, offset=0, got limit=%d, offset=%d", limit, offset)
			}
			return []domain.DevTenant{
				{
					ID:        "a0000000-0000-0000-0000-000000000001",
					TradeName: "Don Corleone",
					Plan:      "PRO",
					Status:    "ACTIVE",
					Owner: &domain.DevTenantOwner{
						Name:  "Vito",
						Email: "owner.corleone@pontocerto.com.br",
					},
				},
			}, 25, nil
		},
	}

	svc := service.NewDevService(repo)
	res, err := svc.GetTenants(context.Background(), "", "", 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if res.Total != 25 {
		t.Errorf("expected total 25, got: %d", res.Total)
	}
	if !res.HasMore {
		t.Errorf("expected HasMore to be true for page 1 of 25 with limit 10")
	}
	if len(res.Tenants) != 1 {
		t.Errorf("expected 1 tenant in page, got: %d", len(res.Tenants))
	}
}

func TestDevService_GetTenantUsers(t *testing.T) {
	t.Run("invalid tenant UUID returns ErrNotFound", func(t *testing.T) {
		svc := service.NewDevService(&mockDevRepo{})
		_, err := svc.GetTenantUsers(context.Background(), "invalid-uuid")
		if err == nil {
			t.Fatal("expected error for invalid UUID, got nil")
		}
	})

	t.Run("valid tenant UUID returns users", func(t *testing.T) {
		tenantID := "a0000000-0000-0000-0000-000000000001"
		repo := &mockDevRepo{
			listTenantUsersFn: func(ctx context.Context, id string) ([]domain.DevTenantUser, error) {
				return []domain.DevTenantUser{
					{
						ID:       "c0000000-0000-0000-0000-000000000002",
						Name:     "Vito",
						Email:    "owner.corleone@pontocerto.com.br",
						Role:     "ORG_OWNER",
						UnitName: "Matriz",
					},
				}, nil
			},
		}

		svc := service.NewDevService(repo)
		users, err := svc.GetTenantUsers(context.Background(), tenantID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(users) != 1 {
			t.Fatalf("expected 1 user, got: %d", len(users))
		}
		if users[0].UnitName != "Matriz" {
			t.Errorf("expected unit_name 'Matriz', got: %s", users[0].UnitName)
		}
	})
}

func TestDevService_GetUsersByRole(t *testing.T) {
	repo := &mockDevRepo{
		listUsersByRoleFn: func(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error) {
			if role != "BARBER" {
				t.Errorf("expected role 'BARBER', got: %s", role)
			}
			return []domain.DevUser{
				{
					ID:         "c0000000-0000-0000-0000-000000000005",
					Name:       "Marcos",
					Email:      "marcos.barbeiro@pontocerto.com.br",
					Role:       "BARBER",
					TenantName: "Barbearias Ponto Certo",
				},
			}, 1, nil
		},
	}

	svc := service.NewDevService(repo)
	res, err := svc.GetUsersByRole(context.Background(), "BARBER", 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if res.Total != 1 {
		t.Errorf("expected total 1, got: %d", res.Total)
	}
	if res.HasMore {
		t.Errorf("expected HasMore to be false when total is 1 and limit is 10")
	}
	if len(res.Users) != 1 {
		t.Errorf("expected 1 user, got: %d", len(res.Users))
	}
}
