package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

type mockUnitOpsManager struct {
	getMetricsFn         func(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.ManagerDashboardMetrics, error)
	getCashClosingFn     func(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.CashClosingSummary, error)
	updateBarberStatusFn func(ctx context.Context, unitID, barberID string, input domain.UpdateBarberStatusInput, claims *service.UserClaims) error
	toggleServiceFn      func(ctx context.Context, unitID, serviceID string, input domain.ToggleServiceInput, claims *service.UserClaims) error
	listBarbersFn        func(ctx context.Context, unitID string, status string, claims *service.UserClaims) ([]domain.Barber, error)
}

func (m *mockUnitOpsManager) GetManagerMetrics(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.ManagerDashboardMetrics, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx, unitID, date, claims)
	}
	return &domain.ManagerDashboardMetrics{
		TodayRevenueInCents:       145000,
		ChairOccupancyRatePercent: 75,
		TodayAppointmentsCount:    18,
		ActiveBarbersCount:        4,
		WaitingClientsCount:       3,
	}, nil
}

func (m *mockUnitOpsManager) GetDailyCashClosing(ctx context.Context, unitID string, date time.Time, claims *service.UserClaims) (*domain.CashClosingSummary, error) {
	if m.getCashClosingFn != nil {
		return m.getCashClosingFn(ctx, unitID, date, claims)
	}
	return &domain.CashClosingSummary{
		Date:               date,
		UnitID:             unitID,
		TotalAmountInCents: 145000,
		TotalTransactions:  6,
		ByPaymentMethod: map[domain.PaymentMethod]int64{
			domain.PaymentMethodPix:        85000,
			domain.PaymentMethodCreditCard: 40000,
			domain.PaymentMethodCash:       20000,
		},
	}, nil
}

func (m *mockUnitOpsManager) UpdateBarberStatus(ctx context.Context, unitID, barberID string, input domain.UpdateBarberStatusInput, claims *service.UserClaims) error {
	if m.updateBarberStatusFn != nil {
		return m.updateBarberStatusFn(ctx, unitID, barberID, input, claims)
	}
	return nil
}

func (m *mockUnitOpsManager) ToggleService(ctx context.Context, unitID, serviceID string, input domain.ToggleServiceInput, claims *service.UserClaims) error {
	if m.toggleServiceFn != nil {
		return m.toggleServiceFn(ctx, unitID, serviceID, input, claims)
	}
	return nil
}

func (m *mockUnitOpsManager) ListBarbers(ctx context.Context, unitID string, status string, claims *service.UserClaims) ([]domain.Barber, error) {
	if m.listBarbersFn != nil {
		return m.listBarbersFn(ctx, unitID, status, claims)
	}
	return []domain.Barber{
		{ID: "b-1", UnitID: unitID, Name: "Barber 1", Status: domain.BarberStatusAvailable},
	}, nil
}

func generateTestToken(userID string, role domain.Role, unitID string, secret string) string {
	now := time.Now().UTC()
	claims := service.UserClaims{
		UserID: userID,
		Email:  "test@pontocerto.com.br",
		Role:   role,
		UnitID: &unitID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := t.SignedString([]byte(secret))
	return tokenString
}

func setupManagerTestRouter(mockSvc *mockUnitOpsManager, jwtSecret string) *chi.Mux {
	h := handler.NewUnitOperationsHandler(mockSvc)

	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(jwtSecret))
	r.Route("/api/v1/manager", func(sub chi.Router) {
		sub.Get("/metrics", h.GetMetrics)
		sub.Get("/cash-closing", h.GetCashClosing)
		sub.Get("/barbers", h.ListBarbers)
		sub.Patch("/barbers/{barber_id}/status", h.UpdateBarberStatus)
		sub.Patch("/services/{service_id}/toggle", h.ToggleService)
	})
	return r
}

func TestUnitOperationsHandler_GetMetrics(t *testing.T) {
	jwtSecret := "test-secret-key-for-manager-unit-tests-32"
	unitID := "unit-1"
	token := generateTestToken("user-mgr", domain.RoleUnitManager, unitID, jwtSecret)

	mockSvc := &mockUnitOpsManager{}
	r := setupManagerTestRouter(mockSvc, jwtSecret)

	t.Run("returns 200 OK with manager metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/manager/metrics?unit_id=unit-1&date=2026-11-06", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp domain.ManagerDashboardMetrics
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if resp.TodayRevenueInCents != 145000 {
			t.Errorf("expected 145000, got %d", resp.TodayRevenueInCents)
		}
		if resp.ChairOccupancyRatePercent != 75 {
			t.Errorf("expected 75, got %d", resp.ChairOccupancyRatePercent)
		}
	})

	t.Run("returns 401 Unauthorized when token is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/manager/metrics?unit_id=unit-1", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})
}

func TestUnitOperationsHandler_GetCashClosing(t *testing.T) {
	jwtSecret := "test-secret-key-for-manager-unit-tests-32"
	unitID := "unit-1"
	token := generateTestToken("user-mgr", domain.RoleUnitManager, unitID, jwtSecret)

	mockSvc := &mockUnitOpsManager{}
	r := setupManagerTestRouter(mockSvc, jwtSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/manager/cash-closing?unit_id=unit-1&date=2026-11-06", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.CashClosingSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode cash closing: %v", err)
	}

	if resp.TotalAmountInCents != 145000 {
		t.Errorf("expected 145000 total, got %d", resp.TotalAmountInCents)
	}
	if resp.ByPaymentMethod[domain.PaymentMethodPix] != 85000 {
		t.Errorf("expected 85000 PIX, got %d", resp.ByPaymentMethod[domain.PaymentMethodPix])
	}
}

func TestUnitOperationsHandler_UpdateBarberStatus(t *testing.T) {
	jwtSecret := "test-secret-key-for-manager-unit-tests-32"
	unitID := "unit-1"
	token := generateTestToken("user-mgr", domain.RoleUnitManager, unitID, jwtSecret)

	mockSvc := &mockUnitOpsManager{}
	r := setupManagerTestRouter(mockSvc, jwtSecret)

	payload := domain.UpdateBarberStatusInput{Status: domain.BarberStatusLunch}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/manager/barbers/barber-1/status?unit_id=unit-1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUnitOperationsHandler_ToggleService(t *testing.T) {
	jwtSecret := "test-secret-key-for-manager-unit-tests-32"
	unitID := "unit-1"
	token := generateTestToken("user-mgr", domain.RoleUnitManager, unitID, jwtSecret)

	mockSvc := &mockUnitOpsManager{}
	r := setupManagerTestRouter(mockSvc, jwtSecret)

	payload := domain.ToggleServiceInput{IsActive: false}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/manager/services/svc-1/toggle?unit_id=unit-1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}
