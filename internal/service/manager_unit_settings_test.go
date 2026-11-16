package service

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

type mockManagerUnitSettingsRepo struct {
	settings  *domain.UnitSettingsResponse
	updateErr error
}

func (m *mockManagerUnitSettingsRepo) GetUnitSettings(ctx context.Context, unitID string) (*domain.UnitSettingsResponse, error) {
	if m.settings != nil {
		return m.settings, nil
	}
	return &domain.UnitSettingsResponse{
		UnitID:     unitID,
		UnitName:   "Unidade Teste",
		ChairCount: 4,
	}, nil
}

func (m *mockManagerUnitSettingsRepo) UpdateUnitSettings(ctx context.Context, unitID string, input domain.UpdateUnitSettingsInput, audit *domain.AuditLog) (*domain.UnitSettingsResponse, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	phone := ""
	if input.Phone != nil {
		phone = *input.Phone
	}
	address := ""
	if input.Address != nil {
		address = *input.Address
	}
	return &domain.UnitSettingsResponse{
		UnitID:     unitID,
		UnitName:   "Unidade Teste",
		Phone:      &phone,
		Address:    &address,
		ChairCount: 4,
	}, nil
}

func (m *mockManagerUnitSettingsRepo) GetUnitOperatingHours(ctx context.Context, unitID string) ([]domain.UnitOperatingHour, error) {
	return nil, nil
}

func TestManagerUnitSettingsService(t *testing.T) {
	unitID := "unit-1111"
	otherUnit := "unit-2222"
	claims := &UserClaims{Role: domain.RoleUnitManager, UnitID: &unitID}

	t.Run("authorization guards", func(t *testing.T) {
		repo := &mockManagerUnitSettingsRepo{}
		svc := NewManagerUnitSettingsService(repo)

		// nil claims
		_, err := svc.GetUnitSettings(context.Background(), unitID, nil)
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}

		// wrong unit
		wrongClaims := &UserClaims{Role: domain.RoleUnitManager, UnitID: &otherUnit}
		_, err = svc.GetUnitSettings(context.Background(), unitID, wrongClaims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		// barber role
		barberClaims := &UserClaims{Role: domain.RoleBarber, UnitID: &unitID}
		_, err = svc.GetUnitSettings(context.Background(), unitID, barberClaims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("GetUnitSettings succeeds for manager", func(t *testing.T) {
		repo := &mockManagerUnitSettingsRepo{}
		svc := NewManagerUnitSettingsService(repo)

		res, err := svc.GetUnitSettings(context.Background(), unitID, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.UnitID != unitID {
			t.Fatalf("unexpected unit id: %s", res.UnitID)
		}
	})

	t.Run("UpdateUnitSettings validates and updates", func(t *testing.T) {
		repo := &mockManagerUnitSettingsRepo{}
		svc := NewManagerUnitSettingsService(repo)

		phone := "11988887777"
		input := domain.UpdateUnitSettingsInput{
			Phone: &phone,
		}

		res, err := svc.UpdateUnitSettings(context.Background(), unitID, input, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Phone == nil || *res.Phone != phone {
			t.Fatalf("expected phone %s, got %+v", phone, res.Phone)
		}
	})
}
