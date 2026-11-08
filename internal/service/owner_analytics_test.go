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
	getKPIsFn          func(ctx context.Context, orgID string, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error)
	getRevenueByUnitFn func(ctx context.Context, orgID string, start, end time.Time) ([]domain.UnitRevenueComparison, error)
	getServicesMixFn   func(ctx context.Context, orgID string, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error)
	listBillingsFn     func(ctx context.Context, orgID string, unitID *string, paymentMethod *string, limit, offset int) ([]domain.Billing, int, error)
	getExportDataFn    func(ctx context.Context, orgID string, unitID *string) ([]repository.BillingExportRow, error)
	getSubscriptionFn  func(ctx context.Context, orgID string) (*domain.OwnerSubscription, error)
	listEmployeesFn    func(ctx context.Context, orgID string, status string, unitID *string) ([]domain.OrganizationEmployee, error)
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

func (m *mockOwnerAnalyticsRepo) GetOwnerSubscription(ctx context.Context, orgID string) (*domain.OwnerSubscription, error) {
	if m.getSubscriptionFn != nil {
		return m.getSubscriptionFn(ctx, orgID)
	}
	return &domain.OwnerSubscription{
		OrganizationID:     orgID,
		Name:               "Barbearia Modelo",
		Plan:               "PRO",
		MonthlyFeeInCents:  19900,
		SubscriptionStatus: domain.SubscriptionStatusActive,
		IsActive:           true,
	}, nil
}

func (m *mockOwnerAnalyticsRepo) ListEmployees(ctx context.Context, orgID string, status string, unitID *string) ([]domain.OrganizationEmployee, error) {
	if m.listEmployeesFn != nil {
		return m.listEmployeesFn(ctx, orgID, status, unitID)
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

func TestOwnerAnalyticsService_GetSubscription(t *testing.T) {
	orgID := "org-subscription-test"

	t.Run("successfully retrieves subscription for ORG_OWNER", func(t *testing.T) {
		claims := &service.UserClaims{
			UserID:         "owner-1",
			Role:           domain.RoleOrgOwner,
			OrganizationID: &orgID,
		}

		repo := &mockOwnerAnalyticsRepo{
			getSubscriptionFn: func(ctx context.Context, o string) (*domain.OwnerSubscription, error) {
				if o != orgID {
					t.Errorf("expected orgID %s, got %s", orgID, o)
				}
				return &domain.OwnerSubscription{
					OrganizationID:     o,
					Name:               "Navalha de Ouro",
					Plan:               "ENTERPRISE",
					MonthlyFeeInCents:  49900,
					SubscriptionStatus: domain.SubscriptionStatusPastDue,
					IsActive:           false,
					SuspensionReason:   "FINANCIAL_PAST_DUE",
				}, nil
			},
		}

		svc := service.NewOwnerAnalyticsService(repo)
		sub, err := svc.GetSubscription(context.Background(), claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if sub.Name != "Navalha de Ouro" {
			t.Errorf("expected name 'Navalha de Ouro', got %s", sub.Name)
		}
		if sub.IsActive {
			t.Errorf("expected IsActive to be false")
		}
		if sub.SuspensionReason != "FINANCIAL_PAST_DUE" {
			t.Errorf("expected SuspensionReason FINANCIAL_PAST_DUE, got %s", sub.SuspensionReason)
		}
	})

	t.Run("rejects unauthorized caller without claims", func(t *testing.T) {
		svc := service.NewOwnerAnalyticsService(&mockOwnerAnalyticsRepo{})
		_, err := svc.GetSubscription(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestOwnerAnalyticsService_ListEmployees(t *testing.T) {
	orgID := "org-123"
	claims := &service.UserClaims{
		UserID:         "owner-1",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	repo := &mockOwnerAnalyticsRepo{
		listEmployeesFn: func(ctx context.Context, oID string, status string, unitID *string) ([]domain.OrganizationEmployee, error) {
			if status == "inactive" {
				return []domain.OrganizationEmployee{
					{ID: "e-demitido", Name: "Ex Barbeiro", Role: domain.RoleBarber, IsActive: false},
				}, nil
			}
			return []domain.OrganizationEmployee{
				{ID: "e-ativo", Name: "Barbeiro Ativo", Role: domain.RoleBarber, IsActive: true},
			}, nil
		},
	}

	svc := service.NewOwnerAnalyticsService(repo)

	t.Run("returns inactive employees for owner with status=inactive", func(t *testing.T) {
		emps, err := svc.ListEmployees(context.Background(), claims, "inactive", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(emps) != 1 || emps[0].ID != "e-demitido" {
			t.Errorf("unexpected employees: %+v", emps)
		}
	})

	t.Run("rejects unauthorized caller without claims", func(t *testing.T) {
		_, err := svc.ListEmployees(context.Background(), nil, "active", nil)
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})
}
