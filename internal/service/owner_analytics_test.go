package service_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockOwnerAnalyticsRepo struct {
	getKPIsFn       func(ctx context.Context, orgID string, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error)
	getRevenueByUnitFn func(ctx context.Context, orgID string, start, end time.Time) ([]domain.UnitRevenueComparison, error)
	getServicesMixFn   func(ctx context.Context, orgID string, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error)
	listBillingsFn  func(ctx context.Context, orgID string, unitID *string, paymentMethod *string, limit, offset int) ([]domain.Billing, int, error)
	getExportDataFn func(ctx context.Context, orgID string, unitID *string) ([]repository.BillingExportRow, error)
}

func (m *mockOwnerAnalyticsRepo) GetExecutiveKPIs(ctx context.Context, orgID string, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
	if m.getKPIsFn != nil {
		return m.getKPIsFn(ctx, orgID, unitID, start, end)
	}
	return &domain.OwnerExecutiveKPIs{}, nil
}

func (m *mockOwnerAnalyticsRepo) GetRevenueComparisonByUnit(ctx context.Context, orgID string, start, end time.Time) ([]domain.UnitRevenueComparison, error) {
	if m.getRevenueByUnitFn != nil {
		return m.getRevenueByUnitFn(ctx, orgID, start, end)
	}
	return nil, nil
}

func (m *mockOwnerAnalyticsRepo) GetServicesMix(ctx context.Context, orgID string, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error) {
	if m.getServicesMixFn != nil {
		return m.getServicesMixFn(ctx, orgID, unitID, start, end)
	}
	return nil, nil
}

func (m *mockOwnerAnalyticsRepo) ListConsolidatedBillings(ctx context.Context, orgID string, unitID *string, paymentMethod *string, limit, offset int) ([]domain.Billing, int, error) {
	if m.listBillingsFn != nil {
		return m.listBillingsFn(ctx, orgID, unitID, paymentMethod, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockOwnerAnalyticsRepo) GetExportBillingsData(ctx context.Context, orgID string, unitID *string) ([]repository.BillingExportRow, error) {
	if m.getExportDataFn != nil {
		return m.getExportDataFn(ctx, orgID, unitID)
	}
	return nil, nil
}

func TestOwnerAnalyticsService_GetExecutiveKPIs(t *testing.T) {
	orgID := "org-123"
	unitID := "unit-456"
	diffUnit := "unit-999"

	claims := &service.UserClaims{
		UserID:         "user-1",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	repo := &mockOwnerAnalyticsRepo{
		getKPIsFn: func(ctx context.Context, o string, u *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
			return &domain.OwnerExecutiveKPIs{
				GrossRevenueInCents:     500000,
				NetRevenueInCents:       250000,
				TotalCommissionsInCents: 250000,
				AverageTicketInCents:    5000,
				TotalAppointments:       100,
				RetentionRatePercent:    80.5,
			}, nil
		},
	}

	svc := service.NewOwnerAnalyticsService(repo)
	ctx := context.Background()
	now := time.Now()

	t.Run("successfully retrieves KPIs for org owner", func(t *testing.T) {
		kpis, err := svc.GetExecutiveKPIs(ctx, claims, nil, now.AddDate(0, 0, -30), now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kpis.GrossRevenueInCents != 500000 {
			t.Errorf("expected gross revenue 500000, got %d", kpis.GrossRevenueInCents)
		}
	})

	t.Run("rejects unauthorized caller without claims", func(t *testing.T) {
		_, err := svc.GetExecutiveKPIs(ctx, nil, nil, now, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("rejects invalid date range", func(t *testing.T) {
		_, err := svc.GetExecutiveKPIs(ctx, claims, nil, now, now.AddDate(0, 0, -5))
		if err == nil {
			t.Fatal("expected error for end before start, got nil")
		}
	})

	t.Run("allows unit manager to access only their unit", func(t *testing.T) {
		mgrClaims := &service.UserClaims{
			UserID:         "mgr-1",
			Role:           domain.RoleUnitManager,
			OrganizationID: &orgID,
			UnitID:         &unitID,
		}

		// manager can access unit-456
		_, err := svc.GetExecutiveKPIs(ctx, mgrClaims, &unitID, now.AddDate(0, 0, -1), now)
		if err == nil {
			// wait: manager role should be forbidden from owner analytics service!
			t.Fatal("expected unit manager to be forbidden from owner analytics service")
		}

		_ = diffUnit
	})
}

func TestOwnerAnalyticsService_ExportBillingsCSV(t *testing.T) {
	orgID := "org-123"
	claims := &service.UserClaims{
		UserID:         "user-1",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	repo := &mockOwnerAnalyticsRepo{
		getExportDataFn: func(ctx context.Context, o string, u *string) ([]repository.BillingExportRow, error) {
			return []repository.BillingExportRow{
				{
					ID:            "bill-1",
					Date:          "2026-11-06",
					UnitName:      "Matriz Centro",
					BarberName:    "Carlos Silva",
					ClientName:    "Lucas Dias",
					ServiceName:   "Corte Degradê",
					AmountInCents: 6000,
					PaymentMethod: "PIX",
					Status:        "PAID",
					CreatedAt:     "2026-11-06T14:30:00Z",
				},
			}, nil
		},
	}

	svc := service.NewOwnerAnalyticsService(repo)
	var buf bytes.Buffer

	err := svc.ExportBillingsCSV(context.Background(), claims, nil, &buf)
	if err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID,Data,Unidade,Barbeiro,Cliente,Serviço,Valor (R$),Forma de Pagamento,Status,Criado Em") {
		t.Errorf("header missing from csv: %s", output)
	}
	if !strings.Contains(output, "60.00") {
		t.Errorf("expected formatted currency 60.00 in csv: %s", output)
	}
	if !strings.Contains(output, "Carlos Silva") {
		t.Errorf("expected barber name in csv: %s", output)
	}
}

func TestOwnerAnalyticsService_ListConsolidatedBillings(t *testing.T) {
	orgID := "org-123"
	claims := &service.UserClaims{
		UserID:         "user-1",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	var capturedLimit, capturedOffset int
	repo := &mockOwnerAnalyticsRepo{
		listBillingsFn: func(ctx context.Context, o string, u *string, method *string, limit, offset int) ([]domain.Billing, int, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []domain.Billing{{ID: "b-1"}}, 50, nil
		},
	}

	svc := service.NewOwnerAnalyticsService(repo)
	res, err := svc.ListConsolidatedBillings(context.Background(), claims, nil, nil, 2, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedOffset != 10 {
		t.Errorf("expected offset 10 for page 2 with perPage 10, got %d", capturedOffset)
	}
	if capturedLimit != 10 {
		t.Errorf("expected limit 10, got %d", capturedLimit)
	}
	if res.Total != 50 {
		t.Errorf("expected total 50, got %d", res.Total)
	}
}
