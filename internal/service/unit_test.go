package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockUnitRepo struct {
	listByOrgFn         func(ctx context.Context, orgID string, status string) ([]domain.BarbershopUnit, error)
	existsByNameInOrgFn func(ctx context.Context, orgID string, name string, excludeUnitID string) (bool, error)
	createFn            func(ctx context.Context, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error)
	updateFn            func(ctx context.Context, unitID string, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error)
	toggleStatusFn      func(ctx context.Context, unitID string, orgID string, isActive bool) (*domain.BarbershopUnit, error)
}

func (m *mockUnitRepo) ListByOrg(ctx context.Context, orgID string, status string) ([]domain.BarbershopUnit, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, status)
	}
	return []domain.BarbershopUnit{}, nil
}

func (m *mockUnitRepo) ExistsByNameInOrg(ctx context.Context, orgID string, name string, excludeUnitID string) (bool, error) {
	if m.existsByNameInOrgFn != nil {
		return m.existsByNameInOrgFn(ctx, orgID, name, excludeUnitID)
	}
	return false, nil
}

func (m *mockUnitRepo) Create(ctx context.Context, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
	if m.createFn != nil {
		return m.createFn(ctx, orgID, name, phone, address, chairCount)
	}
	return &domain.BarbershopUnit{
		ID:             "33333333-3333-3333-3333-333333333333",
		OrganizationID: orgID,
		Name:           name,
		ChairCount:     chairCount,
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}, nil
}

func (m *mockUnitRepo) Update(ctx context.Context, unitID string, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, unitID, orgID, name, phone, address, chairCount)
	}
	return &domain.BarbershopUnit{
		ID:             unitID,
		OrganizationID: orgID,
		Name:           name,
		ChairCount:     chairCount,
		IsActive:       true,
	}, nil
}

func (m *mockUnitRepo) ToggleStatus(ctx context.Context, unitID string, orgID string, isActive bool) (*domain.BarbershopUnit, error) {
	if m.toggleStatusFn != nil {
		return m.toggleStatusFn(ctx, unitID, orgID, isActive)
	}
	return &domain.BarbershopUnit{
		ID:             unitID,
		OrganizationID: orgID,
		Name:           "Matriz",
		ChairCount:     4,
		IsActive:       isActive,
	}, nil
}

type mockUnitAuditRecorder struct {
	logs []domain.AuditLog
}

func (m *mockUnitAuditRecorder) Record(ctx context.Context, log domain.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func TestUnitService_ListUnits(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	ownerClaims := &service.UserClaims{
		UserID:         "22222222-2222-2222-2222-222222222222",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("rejects nil claims with ErrUnauthorized", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.ListUnits(context.Background(), nil, "")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects claims without organization id with ErrForbidden", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.ListUnits(context.Background(), &service.UserClaims{Role: domain.RoleSuperAdmin}, "")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("rejects invalid status filter with ErrValidationFailed", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.ListUnits(context.Background(), ownerClaims, "archived")
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed, got %v", err)
		}
	})

	t.Run("forwards normalized status filter and returns units", func(t *testing.T) {
		var receivedStatus string
		repo := &mockUnitRepo{
			listByOrgFn: func(ctx context.Context, id string, status string) ([]domain.BarbershopUnit, error) {
				receivedStatus = status
				return []domain.BarbershopUnit{
					{ID: "33333333-3333-3333-3333-333333333333", OrganizationID: id, Name: "Filial Centro", ChairCount: 4, IsActive: true},
				}, nil
			},
		}
		svc := service.NewUnitService(repo)
		units, err := svc.ListUnits(context.Background(), ownerClaims, " ACTIVE ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedStatus != "active" {
			t.Errorf("expected normalized status 'active', got %q", receivedStatus)
		}
		if len(units) != 1 || units[0].Name != "Filial Centro" {
			t.Errorf("unexpected units: %+v", units)
		}
	})
}

func TestUnitService_CreateUnit(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	actorID := "22222222-2222-2222-2222-222222222222"
	ownerClaims := &service.UserClaims{
		UserID:         actorID,
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("defaults chair_count to 4 when omitted and records audit log", func(t *testing.T) {
		var savedName string
		var savedChairs int
		repo := &mockUnitRepo{
			createFn: func(ctx context.Context, id string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
				savedName = name
				savedChairs = chairCount
				return &domain.BarbershopUnit{
					ID:             "33333333-3333-3333-3333-333333333333",
					OrganizationID: id,
					Name:           name,
					ChairCount:     chairCount,
					IsActive:       true,
				}, nil
			},
		}
		audit := &mockUnitAuditRecorder{}
		svc := service.NewUnitService(repo)
		svc.SetAuditRecorder(audit)

		ctx := domain.WithAuditMetadata(context.Background(), domain.AuditMetadata{
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		})
		unit, err := svc.CreateUnit(ctx, ownerClaims, domain.CreateUnitInput{
			Name: "  Filial Morumbi  ",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if savedName != "Filial Morumbi" || savedChairs != 4 || unit.ChairCount != 4 {
			t.Errorf("expected trimmed name and default 4 chairs, got name=%q chairs=%d", savedName, savedChairs)
		}
		if len(audit.logs) != 1 || audit.logs[0].Action != "UNIT_CREATED" || audit.logs[0].TargetID != unit.ID {
			t.Fatalf("expected UNIT_CREATED audit log, got %+v", audit.logs)
		}
	})

	t.Run("rejects empty name with ErrUnitNameRequired", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{Name: "   "})
		if !errors.Is(err, domain.ErrUnitNameRequired) {
			t.Fatalf("expected ErrUnitNameRequired, got %v", err)
		}
	})

	t.Run("rejects name shorter than 2 runes or longer than 120 runes", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})

		_, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{Name: "A"})
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed for short name, got %v", err)
		}

		longAccented := strings.Repeat("ã", 121)
		_, err = svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{Name: longAccented})
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed for long name, got %v", err)
		}

		exact120 := strings.Repeat("ã", 120)
		if _, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{Name: exact120}); err != nil {
			t.Fatalf("expected 120 accented runes to pass, got %v", err)
		}
	})

	t.Run("rejects phone over 20 runes or address over 255 runes", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		longPhone := strings.Repeat("9", 21)
		_, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{
			Name:  "Filial Teste",
			Phone: &longPhone,
		})
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed for long phone, got %v", err)
		}

		longAddr := strings.Repeat("x", 256)
		_, err = svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{
			Name:    "Filial Teste",
			Address: &longAddr,
		})
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed for long address, got %v", err)
		}
	})

	t.Run("rejects explicit zero or negative chair_count with ErrInvalidChairCount", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		zero := 0
		_, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{
			Name:       "Filial Jardins",
			ChairCount: &zero,
		})
		if !errors.Is(err, domain.ErrInvalidChairCount) {
			t.Fatalf("expected ErrInvalidChairCount for 0, got %v", err)
		}

		neg := -2
		_, err = svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{
			Name:       "Filial Jardins",
			ChairCount: &neg,
		})
		if !errors.Is(err, domain.ErrInvalidChairCount) {
			t.Fatalf("expected ErrInvalidChairCount for -2, got %v", err)
		}
	})

	t.Run("rejects duplicate name in same org with ErrUnitNameAlreadyExists", func(t *testing.T) {
		repo := &mockUnitRepo{
			existsByNameInOrgFn: func(ctx context.Context, id string, name string, excludeUnitID string) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewUnitService(repo)
		_, err := svc.CreateUnit(context.Background(), ownerClaims, domain.CreateUnitInput{
			Name: "Matriz Centro",
		})
		if !errors.Is(err, domain.ErrUnitNameAlreadyExists) {
			t.Fatalf("expected ErrUnitNameAlreadyExists, got %v", err)
		}
	})
}

func TestUnitService_UpdateUnit(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	unitID := "33333333-3333-3333-3333-333333333333"
	ownerClaims := &service.UserClaims{
		UserID:         "22222222-2222-2222-2222-222222222222",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("rejects invalid unit uuid", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.UpdateUnit(context.Background(), ownerClaims, "bad-uuid", domain.UpdateUnitInput{
			Name:       "Matriz",
			ChairCount: 5,
		})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("rejects zero chair_count in PUT without defaulting to 4", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.UpdateUnit(context.Background(), ownerClaims, unitID, domain.UpdateUnitInput{
			Name:       "Matriz",
			ChairCount: 0,
		})
		if !errors.Is(err, domain.ErrInvalidChairCount) {
			t.Fatalf("expected ErrInvalidChairCount, got %v", err)
		}
	})

	t.Run("excludes own unitID when checking duplicate name and records audit log", func(t *testing.T) {
		var excludedID string
		repo := &mockUnitRepo{
			existsByNameInOrgFn: func(ctx context.Context, id string, name string, excludeUnitID string) (bool, error) {
				excludedID = excludeUnitID
				return false, nil
			},
		}
		audit := &mockUnitAuditRecorder{}
		svc := service.NewUnitService(repo)
		svc.SetAuditRecorder(audit)

		phone := "11999998888"
		addr := "Av Paulista, 1000"
		unit, err := svc.UpdateUnit(context.Background(), ownerClaims, unitID, domain.UpdateUnitInput{
			Name:       "Matriz Paulista",
			Phone:      &phone,
			Address:    &addr,
			ChairCount: 6,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if excludedID != unitID {
			t.Errorf("expected excludeUnitID %s, got %s", unitID, excludedID)
		}
		if unit.ChairCount != 6 {
			t.Errorf("expected 6 chairs, got %d", unit.ChairCount)
		}
		if len(audit.logs) != 1 || audit.logs[0].Action != "UNIT_UPDATED" {
			t.Fatalf("expected UNIT_UPDATED audit log, got %+v", audit.logs)
		}
	})

	t.Run("propagates ErrUnitNotFound for cross-tenant unit", func(t *testing.T) {
		repo := &mockUnitRepo{
			updateFn: func(ctx context.Context, uID string, oID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
				return nil, domain.ErrUnitNotFound
			},
		}
		svc := service.NewUnitService(repo)
		_, err := svc.UpdateUnit(context.Background(), ownerClaims, unitID, domain.UpdateUnitInput{
			Name:       "Outra Org",
			ChairCount: 4,
		})
		if !errors.Is(err, domain.ErrUnitNotFound) {
			t.Fatalf("expected ErrUnitNotFound, got %v", err)
		}
	})
}

func TestUnitService_ToggleUnitStatus(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	unitID := "33333333-3333-3333-3333-333333333333"
	ownerClaims := &service.UserClaims{
		UserID:         "22222222-2222-2222-2222-222222222222",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("toggles unit status and records UNIT_STATUS_TOGGLED audit log", func(t *testing.T) {
		audit := &mockUnitAuditRecorder{}
		svc := service.NewUnitService(&mockUnitRepo{})
		svc.SetAuditRecorder(audit)

		unit, err := svc.ToggleUnitStatus(context.Background(), ownerClaims, unitID, domain.ToggleUnitStatusInput{
			IsActive: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if unit.IsActive {
			t.Errorf("expected IsActive false")
		}
		if len(audit.logs) != 1 || audit.logs[0].Action != "UNIT_STATUS_TOGGLED" {
			t.Fatalf("expected UNIT_STATUS_TOGGLED audit log, got %+v", audit.logs)
		}
		var details map[string]any
		if err := json.Unmarshal(audit.logs[0].Details, &details); err != nil {
			t.Fatalf("failed to unmarshal audit details: %v", err)
		}
		if activeVal, ok := details["is_active"].(bool); !ok || activeVal {
			t.Errorf("expected details.is_active=false, got %+v", details)
		}
	})

	t.Run("rejects invalid unit uuid", func(t *testing.T) {
		svc := service.NewUnitService(&mockUnitRepo{})
		_, err := svc.ToggleUnitStatus(context.Background(), ownerClaims, "not-a-uuid", domain.ToggleUnitStatusInput{IsActive: false})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("propagates ErrUnitNotFound when unit belongs to another org", func(t *testing.T) {
		repo := &mockUnitRepo{
			toggleStatusFn: func(ctx context.Context, uID string, oID string, isActive bool) (*domain.BarbershopUnit, error) {
				return nil, domain.ErrUnitNotFound
			},
		}
		svc := service.NewUnitService(repo)
		_, err := svc.ToggleUnitStatus(context.Background(), ownerClaims, unitID, domain.ToggleUnitStatusInput{IsActive: false})
		if !errors.Is(err, domain.ErrUnitNotFound) {
			t.Fatalf("expected ErrUnitNotFound, got %v", err)
		}
	})
}
