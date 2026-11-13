package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type stubOwnerCommissionsRepo struct {
	unitBelongs   bool
	barberBelongs bool
	affected      int
	err           error
	lastCall      string
}

func (s *stubOwnerCommissionsRepo) UnitBelongsToOrg(_ context.Context, _, _ string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.unitBelongs, nil
}

func (s *stubOwnerCommissionsRepo) BarberBelongsToOrg(_ context.Context, _, _ string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.barberBelongs, nil
}

func (s *stubOwnerCommissionsRepo) UpdateAllOrganization(_ context.Context, _ string, _ int) (int, error) {
	s.lastCall = "ALL_ORGANIZATION"
	return s.affected, s.err
}

func (s *stubOwnerCommissionsRepo) UpdateByUnit(_ context.Context, _, _ string, _ int) (int, error) {
	s.lastCall = "UNIT"
	return s.affected, s.err
}

func (s *stubOwnerCommissionsRepo) UpdateByBarber(_ context.Context, _, _ string, _ int) (int, error) {
	s.lastCall = "BARBER"
	return s.affected, s.err
}

type stubCommissionAuditRecorder struct {
	logs []domain.AuditLog
}

func (a *stubCommissionAuditRecorder) Record(_ context.Context, log domain.AuditLog) error {
	a.logs = append(a.logs, log)
	return nil
}

func TestOwnerCommissionsService_UpdateCommissions(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	unitID := "22222222-2222-2222-2222-222222222222"
	barberID := "33333333-3333-3333-3333-333333333333"
	ownerClaims := &service.UserClaims{
		UserID:         "44444444-4444-4444-4444-444444444444",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("rejects nil claims", func(t *testing.T) {
		svc := service.NewOwnerCommissionsService(&stubOwnerCommissionsRepo{})
		_, err := svc.UpdateCommissions(context.Background(), nil, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeAllOrganization,
			Percentage: 50,
		})
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects manager role", func(t *testing.T) {
		svc := service.NewOwnerCommissionsService(&stubOwnerCommissionsRepo{})
		managerClaims := &service.UserClaims{
			UserID:         "44444444-4444-4444-4444-444444444444",
			Role:           domain.RoleUnitManager,
			OrganizationID: &orgID,
		}
		_, err := svc.UpdateCommissions(context.Background(), managerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeAllOrganization,
			Percentage: 50,
		})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("updates all organization and records audit log", func(t *testing.T) {
		repo := &stubOwnerCommissionsRepo{affected: 8}
		audit := &stubCommissionAuditRecorder{}
		svc := service.NewOwnerCommissionsService(repo)
		svc.SetAuditRecorder(audit)

		resp, err := svc.UpdateCommissions(context.Background(), ownerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeAllOrganization,
			Percentage: 45,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.BarbersAffected != 8 || resp.Percentage != 45 || resp.Scope != domain.CommissionScopeAllOrganization {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if len(audit.logs) != 1 || audit.logs[0].Action != "COMMISSIONS_UPDATED" {
			t.Fatalf("expected COMMISSIONS_UPDATED audit log, got %+v", audit.logs)
		}
	})

	t.Run("rejects unit outside organization", func(t *testing.T) {
		repo := &stubOwnerCommissionsRepo{unitBelongs: false}
		svc := service.NewOwnerCommissionsService(repo)

		_, err := svc.UpdateCommissions(context.Background(), ownerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeUnit,
			UnitID:     unitID,
			Percentage: 40,
		})
		if !errors.Is(err, domain.ErrUnitNotFound) {
			t.Fatalf("expected ErrUnitNotFound, got %v", err)
		}
	})

	t.Run("updates unit commissions when unit belongs to org", func(t *testing.T) {
		repo := &stubOwnerCommissionsRepo{unitBelongs: true, affected: 3}
		svc := service.NewOwnerCommissionsService(repo)

		resp, err := svc.UpdateCommissions(context.Background(), ownerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeUnit,
			UnitID:     unitID,
			Percentage: 55,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.lastCall != "UNIT" || resp.BarbersAffected != 3 {
			t.Fatalf("unexpected repo call %q or resp %+v", repo.lastCall, resp)
		}
	})

	t.Run("rejects barber outside organization", func(t *testing.T) {
		repo := &stubOwnerCommissionsRepo{barberBelongs: false}
		svc := service.NewOwnerCommissionsService(repo)

		_, err := svc.UpdateCommissions(context.Background(), ownerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeBarber,
			BarberID:   barberID,
			Percentage: 60,
		})
		if !errors.Is(err, domain.ErrBarberNotInOrg) {
			t.Fatalf("expected ErrBarberNotInOrg, got %v", err)
		}
	})

	t.Run("updates barber commission when barber belongs to org", func(t *testing.T) {
		repo := &stubOwnerCommissionsRepo{barberBelongs: true, affected: 1}
		svc := service.NewOwnerCommissionsService(repo)

		resp, err := svc.UpdateCommissions(context.Background(), ownerClaims, domain.UpdateCommissionsInput{
			Scope:      domain.CommissionScopeBarber,
			BarberID:   barberID,
			Percentage: 60,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.lastCall != "BARBER" || resp.BarbersAffected != 1 {
			t.Fatalf("unexpected repo call %q or resp %+v", repo.lastCall, resp)
		}
	})
}
