package service_test

import (
	"context"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockUserGovRepo struct {
	getByIDFn            func(ctx context.Context, id string) (*domain.User, error)
	toggleActiveStatusFn func(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error
}

func (m *mockUserGovRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserGovRepo) ToggleActiveStatus(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error {
	if m.toggleActiveStatusFn != nil {
		return m.toggleActiveStatusFn(ctx, userID, isActive, isBarber, ownerOrgID, auditLog)
	}
	return nil
}

func TestUserGovernanceService_ToggleUserStatus(t *testing.T) {
	ctx := context.Background()
	org1 := "org-1"
	org2 := "org-2"
	unit1 := "unit-1"
	unit2 := "unit-2"

	superAdminClaims := &service.UserClaims{UserID: "sa-1", Role: domain.RoleSuperAdmin}
	owner1Claims := &service.UserClaims{UserID: "owner-1", Role: domain.RoleOrgOwner, OrganizationID: &org1}
	mgr1Claims := &service.UserClaims{UserID: "mgr-1", Role: domain.RoleUnitManager, OrganizationID: &org1, UnitID: &unit1}
	barberClaims := &service.UserClaims{UserID: "barber-1", Role: domain.RoleBarber, OrganizationID: &org1, UnitID: &unit1}

	t.Run("rejects caller without claims", func(t *testing.T) {
		svc := service.NewUserGovernanceService(&mockUserGovRepo{})
		_, err := svc.ToggleUserStatus(ctx, nil, "any-id", false)
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects empty target user id", func(t *testing.T) {
		svc := service.NewUserGovernanceService(&mockUserGovRepo{})
		_, err := svc.ToggleUserStatus(ctx, superAdminClaims, "", false)
		if err == nil {
			t.Fatal("expected error for empty target user id, got nil")
		}
	})

	t.Run("rejects deactivating self", func(t *testing.T) {
		svc := service.NewUserGovernanceService(&mockUserGovRepo{})
		_, err := svc.ToggleUserStatus(ctx, owner1Claims, "owner-1", false)
		if err != domain.ErrCannotDeactivateSelf {
			t.Errorf("expected ErrCannotDeactivateSelf, got %v", err)
		}
	})

	t.Run("rejects deactivating Super Admin", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "sa-target", Role: domain.RoleSuperAdmin}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		_, err := svc.ToggleUserStatus(ctx, superAdminClaims, "sa-target", false)
		if err != domain.ErrCannotDeactivateSuperAdmin {
			t.Errorf("expected ErrCannotDeactivateSuperAdmin, got %v", err)
		}
	})

	t.Run("SuperAdmin successfully deactivates ORG_OWNER and cascades to organization", func(t *testing.T) {
		var receivedOrgID *string
		var receivedActive bool
		var receivedAudit *domain.AuditLog
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "owner-tgt", Role: domain.RoleOrgOwner, OrganizationID: &org1}, nil
			},
			toggleActiveStatusFn: func(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error {
				receivedOrgID = ownerOrgID
				receivedActive = isActive
				receivedAudit = auditLog
				return nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		resp, err := svc.ToggleUserStatus(ctx, superAdminClaims, "owner-tgt", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.IsActive {
			t.Errorf("expected is_active to be false")
		}
		if receivedOrgID == nil || *receivedOrgID != org1 {
			t.Errorf("expected organization cascading for %s, got %v", org1, receivedOrgID)
		}
		if receivedActive != false {
			t.Errorf("expected active false, got %v", receivedActive)
		}
		if resp.OrganizationIsActive == nil || *resp.OrganizationIsActive != false {
			t.Errorf("expected organization_is_active to be false in response")
		}
		if receivedAudit == nil || receivedAudit.Action != "USER_DEACTIVATED" {
			t.Errorf("expected audit log with action USER_DEACTIVATED, got %+v", receivedAudit)
		}
	})

	t.Run("ORG_OWNER cannot deactivate another ORG_OWNER", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "owner-other", Role: domain.RoleOrgOwner, OrganizationID: &org1}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		_, err := svc.ToggleUserStatus(ctx, owner1Claims, "owner-other", false)
		if err != domain.ErrInsufficientRoleHierarchy {
			t.Errorf("expected ErrInsufficientRoleHierarchy, got %v", err)
		}
	})

	t.Run("ORG_OWNER successfully deactivates UNIT_MANAGER of same organization", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "mgr-tgt", Role: domain.RoleUnitManager, OrganizationID: &org1, UnitID: &unit1}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		resp, err := svc.ToggleUserStatus(ctx, owner1Claims, "mgr-tgt", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.IsActive {
			t.Errorf("expected is_active to be false")
		}
	})

	t.Run("ORG_OWNER rejected when deactivating UNIT_MANAGER of different organization", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "mgr-other-org", Role: domain.RoleUnitManager, OrganizationID: &org2, UnitID: &unit2}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		_, err := svc.ToggleUserStatus(ctx, owner1Claims, "mgr-other-org", false)
		if err != domain.ErrInsufficientRoleHierarchy {
			t.Errorf("expected ErrInsufficientRoleHierarchy, got %v", err)
		}
	})

	t.Run("UNIT_MANAGER successfully deactivates BARBER of same unit", func(t *testing.T) {
		var isBarberToggled bool
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "barber-tgt", Role: domain.RoleBarber, OrganizationID: &org1, UnitID: &unit1}, nil
			},
			toggleActiveStatusFn: func(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error {
				isBarberToggled = isBarber
				return nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		resp, err := svc.ToggleUserStatus(ctx, mgr1Claims, "barber-tgt", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isBarberToggled {
			t.Errorf("expected isBarber to be true")
		}
		if resp.Role != domain.RoleBarber {
			t.Errorf("expected role BARBER, got %s", resp.Role)
		}
	})

	t.Run("UNIT_MANAGER rejected when deactivating BARBER of different unit", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "barber-other-unit", Role: domain.RoleBarber, OrganizationID: &org1, UnitID: &unit2}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		_, err := svc.ToggleUserStatus(ctx, mgr1Claims, "barber-other-unit", false)
		if err != domain.ErrInsufficientRoleHierarchy {
			t.Errorf("expected ErrInsufficientRoleHierarchy, got %v", err)
		}
	})

	t.Run("BARBER rejected from deactivating anyone", func(t *testing.T) {
		repo := &mockUserGovRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{ID: "another-barber", Role: domain.RoleBarber, OrganizationID: &org1, UnitID: &unit1}, nil
			},
		}
		svc := service.NewUserGovernanceService(repo)
		_, err := svc.ToggleUserStatus(ctx, barberClaims, "another-barber", false)
		if err != domain.ErrInsufficientRoleHierarchy {
			t.Errorf("expected ErrInsufficientRoleHierarchy, got %v", err)
		}
	})
}
