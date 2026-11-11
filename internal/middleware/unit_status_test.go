package middleware_test

import (
	"context"
	"encoding/json"
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

type mockUnitStatusChecker struct {
	isUnitActiveFn func(ctx context.Context, unitID string) (bool, error)
}

func (m *mockUnitStatusChecker) IsUnitActive(ctx context.Context, unitID string) (bool, error) {
	if m.isUnitActiveFn != nil {
		return m.isUnitActiveFn(ctx, unitID)
	}
	return true, nil
}

func generateUnitTestToken(secret string, role domain.Role, unitID *string, impersonated bool) string {
	orgID := "org-uuid-1234"
	claims := service.UserClaims{
		UserID:         "user-test-1",
		Role:           role,
		OrganizationID: &orgID,
		UnitID:         unitID,
		IsImpersonated: impersonated,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestRequireActiveUnit(t *testing.T) {
	secret := "test-secret-for-unit-active-status"
	validUnitID := "unit-uuid-1234"

	t.Run("Rejects request without authenticated claims with 401", func(t *testing.T) {
		checker := &mockUnitStatusChecker{}
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		h := middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Super Admin bypasses unit status check", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return false, nil
			},
		}
		token := generateUnitTestToken(secret, domain.RoleSuperAdmin, nil, false)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("expected 200 and next called, got code=%d nextCalled=%v", rec.Code, nextCalled)
		}
	})

	t.Run("ORG_OWNER bypasses unit status check", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return false, nil
			},
		}
		token := generateUnitTestToken(secret, domain.RoleOrgOwner, nil, false)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("expected 200 and next called, got code=%d nextCalled=%v", rec.Code, nextCalled)
		}
	})

	t.Run("Impersonated session bypasses unit status check", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return false, nil
			},
		}
		token := generateUnitTestToken(secret, domain.RoleBarber, &validUnitID, true)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("expected 200 and next called, got code=%d nextCalled=%v", rec.Code, nextCalled)
		}
	})

	t.Run("Rejects unit staff missing UnitID with 403", func(t *testing.T) {
		checker := &mockUnitStatusChecker{}
		token := generateUnitTestToken(secret, domain.RoleUnitManager, nil, false)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("Allows unit staff when unit is active", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return unitID == validUnitID, nil
			},
		}
		token := generateUnitTestToken(secret, domain.RoleBarber, &validUnitID, false)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		nextCalled := false
		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("expected 200 and next called, got code=%d nextCalled=%v", rec.Code, nextCalled)
		}
	})

	t.Run("Blocks unit staff when unit is inactive with 403 ErrUnitInactive", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return false, nil
			},
		}

		for _, role := range []domain.Role{domain.RoleUnitManager, domain.RoleReceptionist, domain.RoleBarber} {
			token := generateUnitTestToken(secret, role, &validUnitID, false)
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			nextCalled := false
			h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})))
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden || nextCalled {
				t.Errorf("role %s: expected 403 and next not called, got code=%d nextCalled=%v", role, rec.Code, nextCalled)
			}
			var body map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&body)
			if body["error"] != domain.ErrUnitInactive.Error() {
				t.Errorf("role %s: expected error %q, got %q", role, domain.ErrUnitInactive.Error(), body["error"])
			}
		}
	})

	t.Run("Blocks caller with 403 when checker returns error", func(t *testing.T) {
		checker := &mockUnitStatusChecker{
			isUnitActiveFn: func(ctx context.Context, unitID string) (bool, error) {
				return false, errors.New("db error")
			},
		}
		token := generateUnitTestToken(secret, domain.RoleReceptionist, &validUnitID, false)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		h := middleware.AuthMiddleware(secret)(middleware.RequireActiveUnit(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})
}
