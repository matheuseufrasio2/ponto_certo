package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockTenantStatusChecker struct {
	isActiveFn func(ctx context.Context, orgID string) (bool, error)
}

func (m *mockTenantStatusChecker) IsOrganizationActive(ctx context.Context, orgID string) (bool, error) {
	if m.isActiveFn != nil {
		return m.isActiveFn(ctx, orgID)
	}
	return true, nil
}

func generateTenantTestToken(secret string, role domain.Role, orgID *string) string {
	claims := service.UserClaims{
		UserID:         "user-test-1",
		Role:           role,
		OrganizationID: orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func generateImpersonatedTenantTestToken(secret string, role domain.Role, orgID *string) string {
	claims := service.UserClaims{
		UserID:         "user-test-1",
		Role:           role,
		OrganizationID: orgID,
		IsImpersonated: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestRequireActiveTenant(t *testing.T) {
	secret := "test-secret-for-tenant-active-status"
	validOrgID := "org-uuid-1234"

	t.Run("Super Admin bypasses tenant status check even if inactive", func(t *testing.T) {
		checker := &mockTenantStatusChecker{
			isActiveFn: func(ctx context.Context, orgID string) (bool, error) {
				return false, nil // Inactive org
			},
		}

		token := generateTenantTestToken(secret, domain.RoleSuperAdmin, nil)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(next))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !nextCalled {
			t.Errorf("expected next handler to be called")
		}
	})

	t.Run("Impersonation session by Super Admin bypasses tenant status check even if inactive", func(t *testing.T) {
		checker := &mockTenantStatusChecker{
			isActiveFn: func(ctx context.Context, orgID string) (bool, error) {
				return false, nil // Inactive org
			},
		}

		token := generateImpersonatedTenantTestToken(secret, domain.RoleOrgOwner, &validOrgID)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(next))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !nextCalled {
			t.Errorf("expected next handler to be called")
		}
	})

	t.Run("Rejects request without authenticated claims with 401", func(t *testing.T) {
		checker := &mockTenantStatusChecker{}
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.RequireActiveTenant(checker)(next)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Rejects non-superadmin caller missing OrganizationID with 403", func(t *testing.T) {
		checker := &mockTenantStatusChecker{}
		token := generateTenantTestToken(secret, domain.RoleOrgOwner, nil)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(next))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("Allows active tenant caller to pass through", func(t *testing.T) {
		checker := &mockTenantStatusChecker{
			isActiveFn: func(ctx context.Context, orgID string) (bool, error) {
				if orgID == validOrgID {
					return true, nil
				}
				return false, nil
			},
		}

		token := generateTenantTestToken(secret, domain.RoleOrgOwner, &validOrgID)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(next))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !nextCalled {
			t.Errorf("expected next handler to be called")
		}
	})

	t.Run("Blocks inactive tenant caller with 403 ErrTenantInactive", func(t *testing.T) {
		checker := &mockTenantStatusChecker{
			isActiveFn: func(ctx context.Context, orgID string) (bool, error) {
				return false, nil // Inactive organization
			},
		}

		token := generateTenantTestToken(secret, domain.RoleBarber, &validOrgID)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(next))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
		if nextCalled {
			t.Errorf("did not expect next handler to be called for inactive tenant")
		}
	})

	t.Run("Blocks caller with 403 when checker returns error", func(t *testing.T) {
		checker := &mockTenantStatusChecker{
			isActiveFn: func(ctx context.Context, orgID string) (bool, error) {
				return false, errors.New("database connection failed")
			},
		}

		token := generateTenantTestToken(secret, domain.RoleUnitManager, &validOrgID)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler := middleware.AuthMiddleware(secret)(middleware.RequireActiveTenant(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})
}
