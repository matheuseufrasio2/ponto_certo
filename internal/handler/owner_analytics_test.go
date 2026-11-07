package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockOwnerAnalyticsManager struct {
	getMetricsFn      func(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error)
	getRevenueByUnitFn func(ctx context.Context, claims *service.UserClaims, start, end time.Time) ([]domain.UnitRevenueComparison, error)
	getServicesMixFn  func(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error)
	listBillingsFn    func(ctx context.Context, claims *service.UserClaims, unitID *string, paymentMethod *string, page, perPage int) (*domain.PaginatedBillings, error)
	exportCSVFn       func(ctx context.Context, claims *service.UserClaims, unitID *string, w io.Writer) error
}

func (m *mockOwnerAnalyticsManager) GetExecutiveKPIs(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx, claims, unitID, start, end)
	}
	return &domain.OwnerExecutiveKPIs{}, nil
}

func (m *mockOwnerAnalyticsManager) GetRevenueComparisonByUnit(ctx context.Context, claims *service.UserClaims, start, end time.Time) ([]domain.UnitRevenueComparison, error) {
	if m.getRevenueByUnitFn != nil {
		return m.getRevenueByUnitFn(ctx, claims, start, end)
	}
	return nil, nil
}

func (m *mockOwnerAnalyticsManager) GetServicesMix(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error) {
	if m.getServicesMixFn != nil {
		return m.getServicesMixFn(ctx, claims, unitID, start, end)
	}
	return nil, nil
}

func (m *mockOwnerAnalyticsManager) ListConsolidatedBillings(ctx context.Context, claims *service.UserClaims, unitID *string, paymentMethod *string, page, perPage int) (*domain.PaginatedBillings, error) {
	if m.listBillingsFn != nil {
		return m.listBillingsFn(ctx, claims, unitID, paymentMethod, page, perPage)
	}
	return nil, nil
}

func (m *mockOwnerAnalyticsManager) ExportBillingsCSV(ctx context.Context, claims *service.UserClaims, unitID *string, w io.Writer) error {
	if m.exportCSVFn != nil {
		return m.exportCSVFn(ctx, claims, unitID, w)
	}
	_, err := w.Write([]byte("ID,Data\n1,2026-11-06\n"))
	return err
}

func generateOwnerTestToken(userID, orgID string, role domain.Role, secret string) string {
	claims := &service.UserClaims{
		UserID:         userID,
		Email:          "owner@pontocerto.com.br",
		Role:           role,
		OrganizationID: &orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestOwnerAnalyticsHandler_Routes(t *testing.T) {
	secret := "test-secret-key-32-chars-long!"
	orgID := "org-1"
	token := generateOwnerTestToken("user-1", orgID, domain.RoleOrgOwner, secret)

	mockMgr := &mockOwnerAnalyticsManager{
		getMetricsFn: func(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
			return &domain.OwnerExecutiveKPIs{
				GrossRevenueInCents: 1000000,
			}, nil
		},
		getRevenueByUnitFn: func(ctx context.Context, claims *service.UserClaims, start, end time.Time) ([]domain.UnitRevenueComparison, error) {
			return []domain.UnitRevenueComparison{
				{UnitID: "u-1", UnitName: "Matriz", RevenueInCents: 600000},
			}, nil
		},
		getServicesMixFn: func(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error) {
			return []domain.ServiceMixAnalytics{
				{ServiceID: "s-1", ServiceName: "Corte", Count: 50, RevenueInCents: 300000},
			}, nil
		},
		listBillingsFn: func(ctx context.Context, claims *service.UserClaims, unitID *string, paymentMethod *string, page, perPage int) (*domain.PaginatedBillings, error) {
			return &domain.PaginatedBillings{
				Items:   []domain.Billing{{ID: "b-1"}},
				Total:   1,
				Page:    1,
				PerPage: 20,
			}, nil
		},
	}

	h := handler.NewOwnerAnalyticsHandler(mockMgr)
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(secret))
		r.Use(middleware.RequireRoles(domain.RoleOrgOwner, domain.RoleSuperAdmin))

		r.Route("/owner", func(r chi.Router) {
			r.Get("/metrics", h.GetMetrics)
			r.Get("/revenue-by-unit", h.GetRevenueByUnit)
			r.Get("/services-mix", h.GetServicesMix)
			r.Get("/billings", h.ListBillings)
			r.Get("/export/csv", h.ExportCSV)
		})
	})

	t.Run("GET /owner/metrics returns 200 with JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/metrics?unit_id=ALL", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var kpis domain.OwnerExecutiveKPIs
		if err := json.NewDecoder(w.Body).Decode(&kpis); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if kpis.GrossRevenueInCents != 1000000 {
			t.Errorf("expected gross revenue 1000000, got %d", kpis.GrossRevenueInCents)
		}
	})

	t.Run("GET /owner/revenue-by-unit returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/revenue-by-unit", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GET /owner/export/csv returns 200 with text/csv content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/export/csv", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if ct != "text/csv; charset=utf-8" {
			t.Errorf("expected text/csv; charset=utf-8, got %s", ct)
		}
	})

	t.Run("rejects unauthorized request without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/metrics", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}
