package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// mockUserRepo simulates database queries in memory for isolated unit testing.
type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func TestAuthService_Login(t *testing.T) {
	rawPassword := "senha123"
	hash, err := service.GeneratePasswordHash(rawPassword)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	testUser := &domain.User{
		ID:           "user-123",
		Name:         "Carlos Admin",
		Email:        "carlos@barbearia.com",
		PasswordHash: hash,
		Role:         domain.RoleOrgOwner,
		IsActive:     true,
	}

	t.Run("successful login generates valid jwt", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		output, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != nil {
			t.Fatalf("expected successful login, got error: %v", err)
		}

		if output.Token == "" || output.TokenType != "Bearer" {
			t.Errorf("expected bearer token, got: %+v", output)
		}

		if output.User.Email != testUser.Email {
			t.Errorf("expected email %s, got %s", testUser.Email, output.User.Email)
		}

		if output.ExpiresIn != 3600 {
			t.Errorf("expected expiresIn 3600, got %d", output.ExpiresIn)
		}
	})

	t.Run("login with wrong password returns unauthorized", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha_errada",
		})

		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("login with inactive user returns forbidden", func(t *testing.T) {
		inactiveUser := *testUser
		inactiveUser.IsActive = false
		repo := &mockUserRepo{user: &inactiveUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("login with empty credentials returns unauthorized", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "",
			Password: "",
		})

		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("login with inactive organization rejects non-owner with ErrTenantInactive", func(t *testing.T) {
		falseVal := false
		barberUser := *testUser
		barberUser.Role = domain.RoleBarber
		barberUser.OrganizationIsActive = &falseVal

		repo := &mockUserRepo{user: &barberUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "barbeiro@barbearia.com",
			Password: "senha123",
		})

		if err != domain.ErrTenantInactive {
			t.Errorf("expected ErrTenantInactive, got %v", err)
		}
	})

	t.Run("login with inactive organization allows ORG_OWNER with organization_is_active=false", func(t *testing.T) {
		falseVal := false
		subStatus := domain.SubscriptionStatusPastDue
		ownerUser := *testUser
		ownerUser.Role = domain.RoleOrgOwner
		ownerUser.OrganizationIsActive = &falseVal
		ownerUser.SubscriptionStatus = &subStatus

		repo := &mockUserRepo{user: &ownerUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		out, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != nil {
			t.Fatalf("expected successful login for ORG_OWNER, got: %v", err)
		}
		if out.User.OrganizationIsActive {
			t.Errorf("expected organization_is_active to be false")
		}
		if out.User.SubscriptionStatus != "PAST_DUE" {
			t.Errorf("expected subscription_status to be PAST_DUE, got %s", out.User.SubscriptionStatus)
		}
	})

	t.Run("login with active organization allows user and returns organization_is_active=true", func(t *testing.T) {
		trueVal := true
		subStatus := domain.SubscriptionStatusActive
		ownerUser := *testUser
		ownerUser.Role = domain.RoleOrgOwner
		ownerUser.OrganizationIsActive = &trueVal
		ownerUser.SubscriptionStatus = &subStatus

		repo := &mockUserRepo{user: &ownerUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		out, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != nil {
			t.Fatalf("expected successful login, got: %v", err)
		}
		if !out.User.OrganizationIsActive {
			t.Errorf("expected organization_is_active to be true")
		}
	})

	t.Run("login rejects unit staff when assigned unit is inactive", func(t *testing.T) {
		falseVal := false
		unitID := "unit-uuid-1"
		roles := []domain.Role{domain.RoleUnitManager, domain.RoleReceptionist, domain.RoleBarber}

		for _, role := range roles {
			staff := *testUser
			staff.Role = role
			staff.UnitID = &unitID
			staff.UnitIsActive = &falseVal

			repo := &mockUserRepo{user: &staff}
			authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

			_, err := authSvc.Login(context.Background(), service.LoginInput{
				Email:    staff.Email,
				Password: rawPassword,
			})
			if err != domain.ErrUnitInactive {
				t.Errorf("role %s: expected ErrUnitInactive, got %v", role, err)
			}
		}
	})

	t.Run("login allows unit staff when assigned unit is active", func(t *testing.T) {
		trueVal := true
		unitID := "unit-uuid-1"
		staff := *testUser
		staff.Role = domain.RoleBarber
		staff.UnitID = &unitID
		staff.UnitIsActive = &trueVal

		repo := &mockUserRepo{user: &staff}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		out, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    staff.Email,
			Password: rawPassword,
		})
		if err != nil {
			t.Fatalf("expected successful login, got: %v", err)
		}
		if out.User.UnitID == nil || *out.User.UnitID != unitID {
			t.Errorf("expected unit_id %s, got %v", unitID, out.User.UnitID)
		}
	})
}
