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
	getMetricsFn         func(ctx context.Context, claims *service.UserClaims) (*domain.SaaSMetrics, error)
	listTenantsFn        func(ctx context.Context, claims *service.UserClaims, search string, page, perPage int) (*domain.PaginatedOrganizations, error)
	createTenantFn       func(ctx context.Context, claims *service.UserClaims, input domain.CreateTenantInput) (*domain.Organization, error)
	toggleTenantStatusFn func(ctx context.Context, claims *service.UserClaims, tenantID string, isActive bool) error
	getTenantUsersFn     func(ctx context.Context, claims *service.UserClaims, tenantID string) ([]domain.User, error)
	impersonateUserFn    func(ctx context.Context, claims *service.UserClaims, targetUserID string) (*domain.ImpersonateResponse, error)
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

func (m *mockTenantAdminManager) ListTenants(ctx context.Context, claims *service.UserClaims, search string, page, perPage int) (*domain.PaginatedOrganizations, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, claims, search, page, perPage)
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
		r.Post("/users/{id}/impersonate", h.ImpersonateUser)
	})

	return r
}

func TestTenantAdminHandler_Routes(t *testing.T) {
	secret := "test-superadmin-secret"
	adminToken := generateAdminToken("admin-1", domain.RoleSuperAdmin, secret)
	barberToken := generateAdminToken("barber-1", domain.RoleBarber, secret)

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

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenants/org-123/toggle-status", bytes.NewReader(data))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/v1/admin/tenants/{id}/users returns tenant users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/org-123/users", nil)
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
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/user-123/impersonate", nil)
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

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/admin-1/impersonate", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
