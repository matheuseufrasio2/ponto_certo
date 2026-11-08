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

type mockTenantAdminManager struct {
	getMetricsFn          func(ctx context.Context, claims *service.UserClaims) (*domain.SaaSMetrics, error)
	listTenantsFn         func(ctx context.Context, claims *service.UserClaims, search, status, plan string, page, perPage int) (*domain.PaginatedOrganizations, error)
	getTenantOverviewFn   func(ctx context.Context, claims *service.UserClaims, tenantID string) (*domain.TenantOverview, error)
	updateTenantPlanFn    func(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.UpdateTenantPlanInput) (*domain.Organization, error)
	createTenantFn        func(ctx context.Context, claims *service.UserClaims, input domain.CreateTenantInput) (*domain.Organization, error)
	toggleTenantStatusFn  func(ctx context.Context, claims *service.UserClaims, tenantID string, isActive bool) error
	getTenantUsersFn      func(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.User, error)
	getTenantUnitsFn      func(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.BarbershopUnit, error)
	resetTenantPasswordFn func(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.ResetTenantPasswordInput) (*domain.ResetPasswordResponse, error)
	impersonateUserFn     func(ctx context.Context, claims *service.UserClaims, targetUserID string) (*domain.ImpersonateResponse, error)
}

func (m *mockTenantAdminManager) GetSaaSMetrics(ctx context.Context, claims *service.UserClaims) (*domain.SaaSMetrics, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx, claims)
	}
	return &domain.SaaSMetrics{
		MRRInCents:              1500000,
		ARRInCents:              18000000,
		ActiveTenants:           10,
		PastDueTenantsCount:     1,
		TotalUnitsAcrossTenants: 25,
	}, nil
}

func (m *mockTenantAdminManager) ListTenants(ctx context.Context, claims *service.UserClaims, search, status, plan string, page, perPage int) (*domain.PaginatedOrganizations, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, claims, search, status, plan, page, perPage)
	}
	return &domain.PaginatedOrganizations{
		Items:   []domain.Organization{{ID: "org-1", Name: "Navalha de Ouro"}},
		Total:   1,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (m *mockTenantAdminManager) CreateTenant(ctx context.Context, claims *service.UserClaims, input domain.CreateTenantInput) (*domain.Organization, error) {
	if m.createTenantFn != nil {
		return m.createTenantFn(ctx, claims, input)
	}
	return &domain.Organization{
		ID:       "org-new",
		Name:     input.Name,
		IsActive: true,
	}, nil
}

func (m *mockTenantAdminManager) ToggleTenantStatus(ctx context.Context, claims *service.UserClaims, tenantID string, isActive bool) error {
	if m.toggleTenantStatusFn != nil {
		return m.toggleTenantStatusFn(ctx, claims, tenantID, isActive)
	}
	return nil
}

func (m *mockTenantAdminManager) GetTenantUsers(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.User, error) {
	if m.getTenantUsersFn != nil {
		return m.getTenantUsersFn(ctx, claims, tenantID)
	}
	return []domain.User{
		{ID: "user-1", Name: "Carlos Silva", Role: domain.RoleOrgOwner},
	}, nil
}

func (m *mockTenantAdminManager) GetTenantUnits(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.BarbershopUnit, error) {
	if m.getTenantUnitsFn != nil {
		return m.getTenantUnitsFn(ctx, claims, tenantID)
	}
	return []domain.BarbershopUnit{
		{ID: "unit-1", OrganizationID: tenantID, Name: "Matriz Centro", ChairCount: 4, IsActive: true},
	}, nil
}

func (m *mockTenantAdminManager) GetTenantOverview(ctx context.Context, claims *service.UserClaims, tenantID string) (*domain.TenantOverview, error) {
	if m.getTenantOverviewFn != nil {
		return m.getTenantOverviewFn(ctx, claims, tenantID)
	}
	return &domain.TenantOverview{
		Organization:           domain.Organization{ID: tenantID, Name: "Navalha de Ouro"},
		GMVLast30DaysInCents:   5000000,
		AppointmentsLast30Days: 400,
		TotalBarbersCount:      6,
		DaysOverdue:            0,
	}, nil
}

func (m *mockTenantAdminManager) UpdateTenantPlan(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.UpdateTenantPlanInput) (*domain.Organization, error) {
	if m.updateTenantPlanFn != nil {
		return m.updateTenantPlanFn(ctx, claims, tenantID, input)
	}
	return &domain.Organization{
		ID:                tenantID,
		Plan:              input.Plan,
		MonthlyFeeInCents: input.MonthlyFeeInCents,
		IsActive:          true,
	}, nil
}

func (m *mockTenantAdminManager) ResetTenantPassword(ctx context.Context, claims *service.UserClaims, tenantID string, input domain.ResetTenantPasswordInput) (*domain.ResetPasswordResponse, error) {
	if m.resetTenantPasswordFn != nil {
		return m.resetTenantPasswordFn(ctx, claims, tenantID, input)
	}
	return &domain.ResetPasswordResponse{
		UserID:            "owner-1",
		UserEmail:         "owner@navalha.com",
		TemporaryPassword: "GeneratedPassword123",
	}, nil
}

func (m *mockTenantAdminManager) ImpersonateUser(ctx context.Context, claims *service.UserClaims, targetUserID string) (*domain.ImpersonateResponse, error) {
	if m.impersonateUserFn != nil {
		return m.impersonateUserFn(ctx, claims, targetUserID)
	}
	return &domain.ImpersonateResponse{
		Token:          "impersonated-jwt-token",
		TargetUserID:   targetUserID,
		TargetUserName: "Carlos Silva",
		TargetRole:     string(domain.RoleOrgOwner),
		ExpiresIn:      1800,
	}, nil
}

func generateAdminToken(userID string, role domain.Role, secret string) string {
	claims := &service.UserClaims{
		UserID: userID,
		Email:  "superadmin@pontocerto.com.br",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func setupTenantAdminTestRouter(mgr *mockTenantAdminManager, secret string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(secret))

	h := handler.NewTenantAdminHandler(mgr)

	r.With(middleware.RequireRoles(domain.RoleSuperAdmin)).Route("/api/v1/admin", func(r chi.Router) {
		r.Get("/metrics", h.GetPlatformMetrics)
		r.Get("/tenants", h.ListTenants)
		r.Post("/tenants", h.CreateTenant)
		r.Patch("/tenants/{id}/toggle-status", h.ToggleTenantStatus)
		r.Get("/tenants/{id}/users", h.GetTenantUsers)
		r.Get("/tenants/{id}/units", h.GetTenantUnits)
		r.Get("/tenants/{id}/overview", h.GetTenantOverview)
		r.Patch("/tenants/{id}/plan", h.UpdateTenantPlan)
		r.Post("/tenants/{id}/reset-password", h.ResetTenantPassword)
		r.Post("/users/{id}/impersonate", h.ImpersonateUser)
	})

	return r
}

func TestTenantAdminHandler_Routes(t *testing.T) {
	secret := "test-superadmin-secret"
	adminToken := generateAdminToken("c0000000-0000-0000-0000-000000000001", domain.RoleSuperAdmin, secret)
	barberToken := generateAdminToken("c0000000-0000-0000-0000-000000000005", domain.RoleBarber, secret)

	mockMgr := &mockTenantAdminManager{}
	router := setupTenantAdminTestRouter(mockMgr, secret)

	t.Run("GET /api/v1/admin/metrics returns 200 for SUPER_ADMIN", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var metrics domain.SaaSMetrics
		if err := json.Unmarshal(rec.Body.Bytes(), &metrics); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if metrics.MRRInCents != 1500000 {
			t.Errorf("expected MRR 1500000, got %d", metrics.MRRInCents)
		}
	})

	t.Run("GET /api/v1/admin/metrics returns 403 for non-SUPER_ADMIN", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+barberToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/admin/tenants returns paginated tenants", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants?search=navalha&page=1&per_page=10", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var paginated domain.PaginatedOrganizations
		if err := json.Unmarshal(rec.Body.Bytes(), &paginated); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(paginated.Items) != 1 || paginated.Items[0].Name != "Navalha de Ouro" {
			t.Errorf("unexpected tenants data: %+v", paginated)
		}
	})

	t.Run("POST /api/v1/admin/tenants creates new tenant", func(t *testing.T) {
		body := domain.CreateTenantInput{
			Name:              "Barbearia Premium",
			TradeName:         "Barbearia Premium LTDA",
			Document:          "12345678000199",
			Plan:              "PRO",
			MonthlyFeeInCents: 24900,
			AdminName:         "Lucas Lima",
			AdminEmail:        "lucas@premium.com",
			AdminPassword:     "securepassword123",
		}
		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var org domain.Organization
		if err := json.Unmarshal(rec.Body.Bytes(), &org); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if org.Name != "Barbearia Premium" {
			t.Errorf("expected name 'Barbearia Premium', got %q", org.Name)
		}
	})

	t.Run("PATCH /api/v1/admin/tenants/{id}/toggle-status updates status", func(t *testing.T) {
		payload := map[string]bool{"is_active": false}
		data, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/toggle-status", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/v1/admin/tenants/{id}/users returns tenant users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var users []domain.User
		if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(users) != 1 || users[0].Name != "Carlos Silva" {
			t.Errorf("unexpected users list: %+v", users)
		}
	})

	t.Run("POST /api/v1/admin/users/{id}/impersonate returns impersonation token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/c0000000-0000-0000-0000-000000000001/impersonate", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp domain.ImpersonateResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Token != "impersonated-jwt-token" || resp.ExpiresIn != 1800 {
			t.Errorf("unexpected impersonate response: %+v", resp)
		}
	})

	t.Run("POST /api/v1/admin/users/{id}/impersonate self returns 403", func(t *testing.T) {
		mockMgr.impersonateUserFn = func(ctx context.Context, claims *service.UserClaims, targetUserID string) (*domain.ImpersonateResponse, error) {
			return nil, domain.ErrCannotImpersonateSelf
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/c0000000-0000-0000-0000-000000000001/impersonate", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/v1/admin/tenants/{id}/units returns 200 with unit list", func(t *testing.T) {
		mockMgr.getTenantUnitsFn = func(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.BarbershopUnit, error) {
			return []domain.BarbershopUnit{
				{ID: "unit-1", OrganizationID: tenantID, Name: "Matriz Centro", ChairCount: 5, IsActive: true},
				{ID: "unit-2", OrganizationID: tenantID, Name: "Filial Jardins", ChairCount: 3, IsActive: true},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/units", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var units []domain.BarbershopUnit
		if err := json.Unmarshal(rec.Body.Bytes(), &units); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(units) != 2 || units[0].Name != "Matriz Centro" {
			t.Errorf("unexpected units list: %+v", units)
		}
	})

	t.Run("GET /api/v1/admin/tenants/{id}/units forbidden for non-superadmin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/units", nil)
		req.Header.Set("Authorization", "Bearer "+barberToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/admin/tenants passes status and plan query parameters", func(t *testing.T) {
		var capturedStatus, capturedPlan string
		mockMgr.listTenantsFn = func(ctx context.Context, claims *service.UserClaims, search, status, plan string, page, perPage int) (*domain.PaginatedOrganizations, error) {
			capturedStatus = status
			capturedPlan = plan
			return &domain.PaginatedOrganizations{
				Items:   []domain.Organization{{ID: "org-1", Name: "Navalha de Ouro"}},
				Total:   1,
				Page:    page,
				PerPage: perPage,
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants?search=Navalha&status=ACTIVE&plan=PRO&page=1&per_page=15", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if capturedStatus != "ACTIVE" || capturedPlan != "PRO" {
			t.Errorf("expected status=ACTIVE, plan=PRO, got status=%s, plan=%s", capturedStatus, capturedPlan)
		}
	})

	t.Run("GET /api/v1/admin/tenants/{id}/overview returns 200 with metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/overview", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var overview domain.TenantOverview
		if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if overview.GMVLast30DaysInCents != 5000000 || overview.AppointmentsLast30Days != 400 {
			t.Errorf("unexpected overview response: %+v", overview)
		}
	})

	t.Run("PATCH /api/v1/admin/tenants/{id}/plan updates subscription tier", func(t *testing.T) {
		payload := domain.UpdateTenantPlanInput{
			Plan:              "ENTERPRISE",
			MonthlyFeeInCents: 59900,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/plan", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var org domain.Organization
		if err := json.Unmarshal(rec.Body.Bytes(), &org); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if org.Plan != "ENTERPRISE" || org.MonthlyFeeInCents != 59900 {
			t.Errorf("unexpected updated organization: %+v", org)
		}
	})

	t.Run("POST /api/v1/admin/tenants/{id}/reset-password resets owner password", func(t *testing.T) {
		payload := domain.ResetTenantPasswordInput{NewPassword: "NovaSenhaAdmin123"}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants/a0000000-0000-0000-0000-000000000001/reset-password", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp domain.ResetPasswordResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.UserEmail != "owner@navalha.com" || resp.TemporaryPassword != "GeneratedPassword123" {
			t.Errorf("unexpected reset response: %+v", resp)
		}
	})

	t.Run("Rejects non-UUID tenant id with 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenants/org-1/toggle-status", bytes.NewReader([]byte(`{"is_active":false}`)))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var errResp handler.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "invalid tenant id format, expected uuid" {
			t.Errorf("expected error %q, got %q", "invalid tenant id format, expected uuid", errResp.Error)
		}
	})

	t.Run("Rejects non-UUID user id for impersonation with 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/user-1/impersonate", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var errResp handler.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "invalid user id format, expected uuid" {
			t.Errorf("expected error %q, got %q", "invalid user id format, expected uuid", errResp.Error)
		}
	})
}
