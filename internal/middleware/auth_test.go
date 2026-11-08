package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

func generateTestToken(secret string, role domain.Role, expiresAt time.Time) string {
	claims := service.UserClaims{
		UserID: "user-1",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "my-secret"
	token := generateTestToken(secret, domain.RoleBarber, time.Now().Add(1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.UserFromContext(r.Context())
		if !ok || claims == nil || claims.Role != domain.RoleBarber {
			t.Errorf("failed to retrieve valid claims from context")
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.AuthMiddleware(secret)(nextHandler)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_MissingOrInvalidToken(t *testing.T) {
	secret := "my-secret"

	t.Run("missing header returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		mw := middleware.AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("tampered secret returns 401", func(t *testing.T) {
		token := generateTestToken("wrong-secret", domain.RoleBarber, time.Now().Add(1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		mw := middleware.AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		mw.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestPreventImpersonatedMutations(t *testing.T) {
	secret := "my-secret"

	t.Run("blocks mutation when session is impersonated", func(t *testing.T) {
		claims := service.UserClaims{
			UserID:         "user-1",
			Role:           domain.RoleOrgOwner,
			IsImpersonated: true,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := token.SignedString([]byte(secret))

		req := httptest.NewRequest(http.MethodPost, "/sensitive", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		rec := httptest.NewRecorder()

		handler := middleware.AuthMiddleware(secret)(
			middleware.PreventImpersonatedMutations(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("allows normal non-impersonated session", func(t *testing.T) {
		token := generateTestToken(secret, domain.RoleOrgOwner, time.Now().Add(1*time.Hour))

		req := httptest.NewRequest(http.MethodPost, "/sensitive", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler := middleware.AuthMiddleware(secret)(
			middleware.PreventImpersonatedMutations(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestRequireRoles(t *testing.T) {
	secret := "my-secret"

	t.Run("super admin can access anything", func(t *testing.T) {
		token := generateTestToken(secret, domain.RoleSuperAdmin, time.Now().Add(1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler := middleware.AuthMiddleware(secret)(
			middleware.RequireRoles(domain.RoleOrgOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected super admin to bypass, got %d", rec.Code)
		}
	})

	t.Run("barber cannot access manager route", func(t *testing.T) {
		token := generateTestToken(secret, domain.RoleBarber, time.Now().Add(1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/manager", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler := middleware.AuthMiddleware(secret)(
			middleware.RequireRoles(domain.RoleUnitManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", rec.Code)
		}
	})
}

func TestRequireUnitAccess(t *testing.T) {
	unitMatriz := "unit-matriz"
	unitFilial := "unit-filial"

	t.Run("super admin has access to any unit", func(t *testing.T) {
		claims := &service.UserClaims{Role: domain.RoleSuperAdmin}
		if !middleware.RequireUnitAccess(unitMatriz, claims) {
			t.Errorf("expected super admin to have unit access")
		}
	})

	t.Run("org owner has access to any unit", func(t *testing.T) {
		claims := &service.UserClaims{Role: domain.RoleOrgOwner}
		if !middleware.RequireUnitAccess(unitMatriz, claims) {
			t.Errorf("expected org owner to have unit access")
		}
	})

	t.Run("manager can only access own unit", func(t *testing.T) {
		claims := &service.UserClaims{
			Role:   domain.RoleUnitManager,
			UnitID: &unitMatriz,
		}

		if !middleware.RequireUnitAccess(unitMatriz, claims) {
			t.Errorf("expected manager to access own unit")
		}

		if middleware.RequireUnitAccess(unitFilial, claims) {
			t.Errorf("expected manager not to access other unit")
		}
	})
}
