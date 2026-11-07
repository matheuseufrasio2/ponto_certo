package service_test

import (
	"context"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockTenantAdminRepo struct {
	getMetricsFn         func(ctx context.Context) (*domain.SaaSMetrics, error)
	listTenantsFn        func(ctx context.Context, search, status, plan string, limit, offset int) ([]domain.Organization, int, error)
	getTenantOverviewFn  func(ctx context.Context, tenantID string) (*domain.TenantOverview, error)
	updateTenantPlanFn   func(ctx context.Context, tenantID, plan string, feeInCents int64) (*domain.Organization, error)
	createTenantFn       func(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error)
	toggleStatusFn       func(ctx context.Context, tenantID string, isActive bool) error
	getUsersFn           func(ctx context.Context, tenantID string) ([]domain.User, error)
	getUnitsFn           func(ctx context.Context, tenantID string) ([]domain.BarbershopUnit, error)
	getPrimaryOwnerFn    func(ctx context.Context, tenantID string) (*domain.User, error)
	updateUserPasswordFn func(ctx context.Context, userID, passwordHash string) error
	getUserByIDFn        func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockTenantAdminRepo) GetSaaSPlatformKPIs(ctx context.Context) (*domain.SaaSMetrics, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx)
	}
	return &domain.SaaSMetrics{}, nil
}

func (m *mockTenantAdminRepo) ListTenants(ctx context.Context, search, status, plan string, limit, offset int) ([]domain.Organization, int, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, search, status, plan, limit, offset)
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

func (m *mockTenantAdminRepo) GetTenantUnits(ctx context.Context, tenantID string) ([]domain.BarbershopUnit, error) {
	if m.getUnitsFn != nil {
		return m.getUnitsFn(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockTenantAdminRepo) GetTenantOverview(ctx context.Context, tenantID string) (*domain.TenantOverview, error) {
	if m.getTenantOverviewFn != nil {
		return m.getTenantOverviewFn(ctx, tenantID)
	}
	return &domain.TenantOverview{
		Organization: domain.Organization{ID: tenantID, Name: "Test Org"},
	}, nil
}

func (m *mockTenantAdminRepo) UpdateTenantPlan(ctx context.Context, tenantID, plan string, feeInCents int64) (*domain.Organization, error) {
	if m.updateTenantPlanFn != nil {
		return m.updateTenantPlanFn(ctx, tenantID, plan, feeInCents)
	}
	return &domain.Organization{
		ID:                tenantID,
		Plan:              plan,
		MonthlyFeeInCents: feeInCents,
	}, nil
}

func (m *mockTenantAdminRepo) GetPrimaryOwnerByTenantID(ctx context.Context, tenantID string) (*domain.User, error) {
	if m.getPrimaryOwnerFn != nil {
		return m.getPrimaryOwnerFn(ctx, tenantID)
	}
	return &domain.User{
		ID:             "owner-1",
		OrganizationID: &tenantID,
		Name:           "Dono Barbearia",
		Email:          "dono@barbearia.com",
		Role:           domain.RoleOrgOwner,
		IsActive:       true,
	}, nil
}

func (m *mockTenantAdminRepo) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	if m.updateUserPasswordFn != nil {
		return m.updateUserPasswordFn(ctx, userID, passwordHash)
	}
	return nil
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

func TestTenantAdminService_GetTenantUnits(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
	}

	expectedUnits := []domain.BarbershopUnit{
		{
			ID:             "unit-1",
			OrganizationID: "org-1",
			Name:           "Matriz Centro",
			ChairCount:     5,
			IsActive:       true,
		},
		{
			ID:             "unit-2",
			OrganizationID: "org-1",
			Name:           "Filial Jardins",
			ChairCount:     3,
			IsActive:       true,
		},
	}

	repo := &mockTenantAdminRepo{
		getUnitsFn: func(ctx context.Context, tenantID string) ([]domain.BarbershopUnit, error) {
			if tenantID == "org-1" {
				return expectedUnits, nil
			}
			return []domain.BarbershopUnit{}, nil
		},
	}

	svc := service.NewTenantAdminService(repo, secret)

	t.Run("successfully retrieves units for super admin", func(t *testing.T) {
		units, err := svc.GetTenantUnits(ctx, adminClaims, "org-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(units) != 2 {
			t.Fatalf("expected 2 units, got %d", len(units))
		}
		if units[0].Name != "Matriz Centro" {
			t.Errorf("expected unit name 'Matriz Centro', got %s", units[0].Name)
		}
	})

	t.Run("rejects request without claims", func(t *testing.T) {
		_, err := svc.GetTenantUnits(ctx, nil, "org-1")
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects request from non-superadmin", func(t *testing.T) {
		managerClaims := &service.UserClaims{
			UserID: "manager-1",
			Role:   domain.RoleUnitManager,
		}
		_, err := svc.GetTenantUnits(ctx, managerClaims, "org-1")
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("rejects empty tenant id", func(t *testing.T) {
		_, err := svc.GetTenantUnits(ctx, adminClaims, "")
		if err == nil {
			t.Fatal("expected error for empty tenant id, got nil")
		}
	})
}

func TestTenantAdminService_ListTenants_Filters(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
	}

	var capturedSearch, capturedStatus, capturedPlan string
	repo := &mockTenantAdminRepo{
		listTenantsFn: func(ctx context.Context, search, status, plan string, limit, offset int) ([]domain.Organization, int, error) {
			capturedSearch = search
			capturedStatus = status
			capturedPlan = plan
			return []domain.Organization{
				{ID: "org-1", Name: "Barbearia Alfa", Plan: "PRO", SubscriptionStatus: domain.SubscriptionStatusActive},
			}, 1, nil
		},
	}

	svc := service.NewTenantAdminService(repo, secret)

	res, err := svc.ListTenants(ctx, adminClaims, "Alfa", "ACTIVE", "PRO", 1, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Total != 1 {
		t.Errorf("expected total 1, got %d", res.Total)
	}
	if capturedSearch != "Alfa" || capturedStatus != "ACTIVE" || capturedPlan != "PRO" {
		t.Errorf("filters not passed correctly to repo: search=%s status=%s plan=%s", capturedSearch, capturedStatus, capturedPlan)
	}
}

func TestTenantAdminService_GetTenantOverview(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
	}

	repo := &mockTenantAdminRepo{
		getTenantOverviewFn: func(ctx context.Context, tenantID string) (*domain.TenantOverview, error) {
			if tenantID == "org-1" {
				return &domain.TenantOverview{
					Organization:           domain.Organization{ID: "org-1", Name: "Navalha de Ouro"},
					GMVLast30DaysInCents:   4500000,
					AppointmentsLast30Days: 320,
					TotalBarbersCount:      8,
					DaysOverdue:            0,
				}, nil
			}
			return nil, domain.ErrOrganizationNotFound
		},
	}

	svc := service.NewTenantAdminService(repo, secret)

	t.Run("successfully retrieves tenant overview for super admin", func(t *testing.T) {
		overview, err := svc.GetTenantOverview(ctx, adminClaims, "org-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if overview.GMVLast30DaysInCents != 4500000 || overview.AppointmentsLast30Days != 320 {
			t.Errorf("unexpected overview metrics: %+v", overview)
		}
	})

	t.Run("rejects non-superadmin caller", func(t *testing.T) {
		managerClaims := &service.UserClaims{UserID: "m-1", Role: domain.RoleUnitManager}
		_, err := svc.GetTenantOverview(ctx, managerClaims, "org-1")
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestTenantAdminService_UpdateTenantPlan(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
	}

	repo := &mockTenantAdminRepo{
		updateTenantPlanFn: func(ctx context.Context, tenantID, plan string, feeInCents int64) (*domain.Organization, error) {
			return &domain.Organization{
				ID:                tenantID,
				Plan:              plan,
				MonthlyFeeInCents: feeInCents,
			}, nil
		},
	}

	svc := service.NewTenantAdminService(repo, secret)

	t.Run("successfully updates plan to ENTERPRISE", func(t *testing.T) {
		res, err := svc.UpdateTenantPlan(ctx, adminClaims, "org-1", domain.UpdateTenantPlanInput{
			Plan:              "ENTERPRISE",
			MonthlyFeeInCents: 49900,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Plan != "ENTERPRISE" || res.MonthlyFeeInCents != 49900 {
			t.Errorf("unexpected updated organization: %+v", res)
		}
	})

	t.Run("rejects invalid plan name", func(t *testing.T) {
		_, err := svc.UpdateTenantPlan(ctx, adminClaims, "org-1", domain.UpdateTenantPlanInput{
			Plan:              "PLATINUM",
			MonthlyFeeInCents: 49900,
		})
		if err == nil {
			t.Fatal("expected error for invalid plan name, got nil")
		}
	})
}

func TestTenantAdminService_ResetTenantPassword(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	adminClaims := &service.UserClaims{
		UserID: "super-admin-1",
		Role:   domain.RoleSuperAdmin,
	}

	var updatedPasswordHash string
	repo := &mockTenantAdminRepo{
		getPrimaryOwnerFn: func(ctx context.Context, tenantID string) (*domain.User, error) {
			return &domain.User{
				ID:             "owner-123",
				OrganizationID: &tenantID,
				Name:           "Dono Barba",
				Email:          "dono@barba.com",
				Role:           domain.RoleOrgOwner,
			}, nil
		},
		updateUserPasswordFn: func(ctx context.Context, userID, passwordHash string) error {
			updatedPasswordHash = passwordHash
			return nil
		},
	}

	svc := service.NewTenantAdminService(repo, secret)

	t.Run("successfully resets password with custom string", func(t *testing.T) {
		res, err := svc.ResetTenantPassword(ctx, adminClaims, "org-1", domain.ResetTenantPasswordInput{
			NewPassword: "NovaSenhaForte123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.UserEmail != "dono@barba.com" || res.TemporaryPassword != "NovaSenhaForte123" {
			t.Errorf("unexpected reset response: %+v", res)
		}
		if updatedPasswordHash == "" {
			t.Error("expected password hash to be stored, was empty")
		}
	})

	t.Run("successfully generates temporary password if empty", func(t *testing.T) {
		res, err := svc.ResetTenantPassword(ctx, adminClaims, "org-1", domain.ResetTenantPasswordInput{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.TemporaryPassword) < 8 {
			t.Errorf("expected generated password >= 8 chars, got: %s", res.TemporaryPassword)
		}
	})
}
