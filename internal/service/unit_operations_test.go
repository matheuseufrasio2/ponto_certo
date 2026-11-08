package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockUnitOpsRepo struct {
	getCashClosingFn     func(ctx context.Context, unitID string, date time.Time) (*domain.CashClosingSummary, error)
	getMetricsFn         func(ctx context.Context, unitID string, date time.Time) (*domain.ManagerDashboardMetrics, error)
	updateBarberStatusFn func(ctx context.Context, barberID string, status domain.BarberStatus, changedByUserID *string) error
	toggleServiceFn      func(ctx context.Context, unitID, serviceID string, isActive bool) error
	listBarbersFn        func(ctx context.Context, unitID string, status string) ([]domain.Barber, error)
	countFutureAptsFn    func(ctx context.Context, barberID string, fromTime time.Time) (int, error)
}

func (m *mockUnitOpsRepo) GetDailyCashClosing(ctx context.Context, unitID string, date time.Time) (*domain.CashClosingSummary, error) {
	if m.getCashClosingFn != nil {
		return m.getCashClosingFn(ctx, unitID, date)
	}
	return &domain.CashClosingSummary{
		UnitID:             unitID,
		Date:               date,
		TotalAmountInCents: 50000,
		TotalTransactions:  2,
		ByPaymentMethod: map[domain.PaymentMethod]int64{
			domain.PaymentMethodPix: 50000,
		},
	}, nil
}

func (m *mockUnitOpsRepo) GetManagerLiveMetrics(ctx context.Context, unitID string, date time.Time) (*domain.ManagerDashboardMetrics, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx, unitID, date)
	}
	return &domain.ManagerDashboardMetrics{
		TodayRevenueInCents:       50000,
		ChairOccupancyRatePercent: 50,
		TodayAppointmentsCount:    5,
		ActiveBarbersCount:        3,
		WaitingClientsCount:       2,
	}, nil
}

func (m *mockUnitOpsRepo) UpdateBarberStatus(ctx context.Context, barberID string, status domain.BarberStatus, changedByUserID *string) error {
	if m.updateBarberStatusFn != nil {
		return m.updateBarberStatusFn(ctx, barberID, status, changedByUserID)
	}
	return nil
}

func (m *mockUnitOpsRepo) ToggleServiceUnitAvailability(ctx context.Context, unitID, serviceID string, isActive bool) error {
	if m.toggleServiceFn != nil {
		return m.toggleServiceFn(ctx, unitID, serviceID, isActive)
	}
	return nil
}

func (m *mockUnitOpsRepo) ListBarbersByUnit(ctx context.Context, unitID string, status string) ([]domain.Barber, error) {
	if m.listBarbersFn != nil {
		return m.listBarbersFn(ctx, unitID, status)
	}
	return []domain.Barber{
		{ID: "barber-1", UnitID: unitID, Name: "Barber One", Status: domain.BarberStatusAvailable, IsActive: true},
	}, nil
}

func (m *mockUnitOpsRepo) CountFutureAppointmentsForBarber(ctx context.Context, barberID string, fromTime time.Time) (int, error) {
	if m.countFutureAptsFn != nil {
		return m.countFutureAptsFn(ctx, barberID, fromTime)
	}
	return 0, nil
}

func TestUnitOperationsService_GetMetrics_RequireUnitAccess(t *testing.T) {
	repo := &mockUnitOpsRepo{}
	svc := service.NewUnitOperationsService(repo)
	ctx := context.Background()

	unitA := "unit-a"
	unitB := "unit-b"

	managerClaims := &service.UserClaims{
		UserID: "user-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitA,
	}

	t.Run("allows manager to access their own unit", func(t *testing.T) {
		metrics, err := svc.GetManagerMetrics(ctx, unitA, time.Now(), managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if metrics.TodayRevenueInCents != 50000 {
			t.Errorf("expected 50000, got %d", metrics.TodayRevenueInCents)
		}
	})

	t.Run("forbids manager from accessing another unit", func(t *testing.T) {
		_, err := svc.GetManagerMetrics(ctx, unitB, time.Now(), managerClaims)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got: %v", err)
		}
	})
}

func TestUnitOperationsService_UpdateBarberStatus(t *testing.T) {
	repo := &mockUnitOpsRepo{}
	svc := service.NewUnitOperationsService(repo)
	ctx := context.Background()

	unitID := "unit-a"
	managerClaims := &service.UserClaims{
		UserID: "user-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("successfully sets status to LUNCH", func(t *testing.T) {
		err := svc.UpdateBarberStatus(ctx, unitID, "barber-1", domain.UpdateBarberStatusInput{
			Status: domain.BarberStatusLunch,
		}, managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		err := svc.UpdateBarberStatus(ctx, unitID, "barber-1", domain.UpdateBarberStatusInput{
			Status: domain.BarberStatus("SLEEPING"),
		}, managerClaims)
		if err != domain.ErrInvalidBarberStatus {
			t.Fatalf("expected ErrInvalidBarberStatus, got: %v", err)
		}
	})

	t.Run("rejects OFF_DUTY when barber has future appointments", func(t *testing.T) {
		conflictRepo := &mockUnitOpsRepo{
			countFutureAptsFn: func(ctx context.Context, barberID string, fromTime time.Time) (int, error) {
				return 3, nil
			},
		}
		conflictSvc := service.NewUnitOperationsService(conflictRepo)

		err := conflictSvc.UpdateBarberStatus(ctx, unitID, "barber-1", domain.UpdateBarberStatusInput{
			Status: domain.BarberStatusOffDuty,
		}, managerClaims)
		if err != domain.ErrBarberHasFutureAppointments {
			t.Fatalf("expected ErrBarberHasFutureAppointments, got: %v", err)
		}
	})

	t.Run("rejects VACATION when barber has future appointments", func(t *testing.T) {
		conflictRepo := &mockUnitOpsRepo{
			countFutureAptsFn: func(ctx context.Context, barberID string, fromTime time.Time) (int, error) {
				return 1, nil
			},
		}
		conflictSvc := service.NewUnitOperationsService(conflictRepo)

		err := conflictSvc.UpdateBarberStatus(ctx, unitID, "barber-1", domain.UpdateBarberStatusInput{
			Status: domain.BarberStatusVacation,
		}, managerClaims)
		if err != domain.ErrBarberHasFutureAppointments {
			t.Fatalf("expected ErrBarberHasFutureAppointments, got: %v", err)
		}
	})
}

func TestUnitOperationsService_ToggleService(t *testing.T) {
	repo := &mockUnitOpsRepo{}
	svc := service.NewUnitOperationsService(repo)
	ctx := context.Background()

	unitID := "unit-a"
	claims := &service.UserClaims{
		UserID: "user-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("successfully toggles service", func(t *testing.T) {
		err := svc.ToggleService(ctx, unitID, "svc-1", domain.ToggleServiceInput{IsActive: false}, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUnitOperationsService_ListBarbers(t *testing.T) {
	ctx := context.Background()
	unitID := "unit-a"
	claims := &service.UserClaims{
		UserID: "mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	repo := &mockUnitOpsRepo{
		listBarbersFn: func(ctx context.Context, uID string, status string) ([]domain.Barber, error) {
			if status == "inactive" {
				return []domain.Barber{
					{ID: "b-demitido", UnitID: uID, Name: "Ex Barbeiro", IsActive: false},
				}, nil
			}
			return []domain.Barber{
				{ID: "b-ativo", UnitID: uID, Name: "Barbeiro Ativo", IsActive: true},
			}, nil
		},
	}
	svc := service.NewUnitOperationsService(repo)

	t.Run("successfully filters inactive barbers", func(t *testing.T) {
		barbers, err := svc.ListBarbers(ctx, unitID, "inactive", claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(barbers) != 1 || barbers[0].ID != "b-demitido" {
			t.Errorf("unexpected barbers: %+v", barbers)
		}
	})

	t.Run("rejects unauthorized access to another unit", func(t *testing.T) {
		_, err := svc.ListBarbers(ctx, "unit-other", "active", claims)
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}
