package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockOwnerManagersRepo struct {
	unitBelongsFn   func(ctx context.Context, unitID, orgID string) (bool, error)
	emailExistsFn   func(ctx context.Context, email string) (bool, error)
	listManagersFn  func(ctx context.Context, orgID string) ([]domain.ManagerListItem, error)
	createManagerFn func(ctx context.Context, orgID, unitID string, input domain.CreateManagerInput, passwordHash string, audit *domain.AuditLog) (*domain.ManagerListItem, error)
}

func (m *mockOwnerManagersRepo) UnitBelongsToOrg(ctx context.Context, unitID, orgID string) (bool, error) {
	if m.unitBelongsFn != nil {
		return m.unitBelongsFn(ctx, unitID, orgID)
	}
	return true, nil
}

func (m *mockOwnerManagersRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.emailExistsFn != nil {
		return m.emailExistsFn(ctx, email)
	}
	return false, nil
}

func (m *mockOwnerManagersRepo) ListManagers(ctx context.Context, orgID string) ([]domain.ManagerListItem, error) {
	if m.listManagersFn != nil {
		return m.listManagersFn(ctx, orgID)
	}
	return []domain.ManagerListItem{}, nil
}

func (m *mockOwnerManagersRepo) CreateManager(ctx context.Context, orgID, unitID string, input domain.CreateManagerInput, passwordHash string, audit *domain.AuditLog) (*domain.ManagerListItem, error) {
	if m.createManagerFn != nil {
		return m.createManagerFn(ctx, orgID, unitID, input, passwordHash, audit)
	}
	return &domain.ManagerListItem{
		ID:        "99999999-9999-9999-9999-999999999999",
		Name:      input.Name,
		Email:     input.Email,
		Phone:     input.Phone,
		UnitID:    unitID,
		UnitName:  "Unidade Centro",
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func TestOwnerManagersService(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	unitID := "22222222-2222-2222-2222-222222222222"
	userID := "33333333-3333-3333-3333-333333333333"

	ownerClaims := &service.UserClaims{
		UserID:         userID,
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("ListManagers rejects nil claims with ErrUnauthorized", func(t *testing.T) {
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{})
		_, err := svc.ListManagers(context.Background(), nil)
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("ListManagers rejects non-owner roles with ErrForbidden", func(t *testing.T) {
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{})
		_, err := svc.ListManagers(context.Background(), &service.UserClaims{
			UserID:         userID,
			Role:           domain.RoleUnitManager,
			OrganizationID: &orgID,
		})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("ListManagers forwards org managers", func(t *testing.T) {
		expected := []domain.ManagerListItem{
			{ID: "mgr-1", Name: "João", Email: "joao@barbearia.com", UnitID: unitID, UnitName: "Centro", IsActive: true},
		}
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{
			listManagersFn: func(ctx context.Context, id string) ([]domain.ManagerListItem, error) {
				if id != orgID {
					t.Fatalf("expected orgID %s, got %s", orgID, id)
				}
				return expected, nil
			},
		})
		got, err := svc.ListManagers(context.Background(), ownerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].Name != "João" {
			t.Fatalf("unexpected result: %+v", got)
		}
	})

	t.Run("CreateManager rejects invalid payload with ValidationError", func(t *testing.T) {
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{})
		_, err := svc.CreateManager(context.Background(), ownerClaims, domain.CreateManagerInput{
			Name:     "J",
			Email:    "not-an-email",
			Phone:    "123",
			Password: "123",
			UnitID:   "invalid",
		})
		var valErr *domain.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *domain.ValidationError, got %v", err)
		}
	})

	t.Run("CreateManager returns ErrUnitNotInOrg when unit belongs to another tenant", func(t *testing.T) {
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{
			unitBelongsFn: func(ctx context.Context, uID, oID string) (bool, error) {
				return false, nil
			},
		})
		_, err := svc.CreateManager(context.Background(), ownerClaims, domain.CreateManagerInput{
			Name:     "João Silva",
			Email:    "joao@barbearia.com.br",
			Phone:    "11987654321",
			Password: "StrongPassword123",
			UnitID:   unitID,
		})
		if !errors.Is(err, domain.ErrUnitNotInOrg) {
			t.Fatalf("expected ErrUnitNotInOrg, got %v", err)
		}
	})

	t.Run("CreateManager returns ErrManagerEmailTaken on duplicate email", func(t *testing.T) {
		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{
			emailExistsFn: func(ctx context.Context, email string) (bool, error) {
				if email != "joao@barbearia.com.br" {
					t.Fatalf("expected lowercased normalized email, got %q", email)
				}
				return true, nil
			},
		})
		_, err := svc.CreateManager(context.Background(), ownerClaims, domain.CreateManagerInput{
			Name:     "João Silva",
			Email:    "  JOAO@BARBEARIA.COM.BR ",
			Phone:    "11987654321",
			Password: "StrongPassword123",
			UnitID:   unitID,
		})
		if !errors.Is(err, domain.ErrManagerEmailTaken) {
			t.Fatalf("expected ErrManagerEmailTaken, got %v", err)
		}
	})

	t.Run("CreateManager hashes password with bcrypt and passes MANAGER_CREATED audit log", func(t *testing.T) {
		var capturedHash string
		var capturedAudit *domain.AuditLog

		svc := service.NewOwnerManagersService(&mockOwnerManagersRepo{
			createManagerFn: func(ctx context.Context, oID, uID string, in domain.CreateManagerInput, hash string, audit *domain.AuditLog) (*domain.ManagerListItem, error) {
				capturedHash = hash
				capturedAudit = audit
				return &domain.ManagerListItem{
					ID:       "44444444-4444-4444-4444-444444444444",
					Name:     in.Name,
					Email:    "joao@barbearia.com.br",
					Phone:    in.Phone,
					UnitID:   uID,
					UnitName: "Centro",
					IsActive: true,
				}, nil
			},
		})

		rawPassword := "ManagerPass123!"
		created, err := svc.CreateManager(context.Background(), ownerClaims, domain.CreateManagerInput{
			Name:     "João Silva",
			Email:    "joao@barbearia.com.br",
			Phone:    "(11) 98765-4321",
			Password: rawPassword,
			UnitID:   unitID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.UnitID != unitID {
			t.Fatalf("expected unitID %s, got %s", unitID, created.UnitID)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte(rawPassword)); err != nil {
			t.Fatalf("expected valid bcrypt hash matching password, got %v", err)
		}
		if capturedAudit == nil || capturedAudit.Action != "MANAGER_CREATED" {
			t.Fatalf("expected MANAGER_CREATED audit log, got %+v", capturedAudit)
		}
		var details map[string]string
		if err := json.Unmarshal(capturedAudit.Details, &details); err != nil {
			t.Fatalf("failed to decode audit details: %v", err)
		}
		if details["manager_email"] != "joao@barbearia.com.br" {
			t.Errorf("expected manager_email in audit details, got %+v", details)
		}
	})
}
