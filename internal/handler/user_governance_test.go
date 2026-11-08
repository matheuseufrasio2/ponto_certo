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

type mockUserGovManager struct {
	toggleUserStatusFn func(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error)
}

func (m *mockUserGovManager) ToggleUserStatus(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
	if m.toggleUserStatusFn != nil {
		return m.toggleUserStatusFn(ctx, claims, targetUserID, isActive)
	}
	return &domain.ToggleUserStatusResponse{
		UserID:   targetUserID,
		IsActive: isActive,
		Role:     domain.RoleBarber,
	}, nil
}

func generateGovTestToken(userID string, role domain.Role, secret string) string {
	claims := service.UserClaims{
		UserID: userID,
		Email:  "user@teste.com",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestUserGovernanceHandler_ToggleUserStatus(t *testing.T) {
	secret := "test-secret-key-32-chars-minimum!"
	adminToken := generateGovTestToken("admin-1", domain.RoleSuperAdmin, secret)
	validTargetID := "c0000000-0000-0000-0000-000000000099"

	mockMgr := &mockUserGovManager{}
	h := handler.NewUserGovernanceHandler(mockMgr)

	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(secret))
	r.Patch("/api/v1/users/{id}/toggle-status", h.ToggleUserStatus)

	t.Run("successfully toggles status to false and returns 200 OK", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp domain.ToggleUserStatusResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.UserID != validTargetID || resp.IsActive != false {
			t.Errorf("unexpected response: %+v", resp)
		}
	})

	t.Run("rejects request without Authorization header with 401", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("rejects invalid UUID in URL with 400 Bad Request", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/not-a-uuid/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("rejects malformed JSON body with 400 Bad Request", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{invalid-json}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("returns 403 when trying to deactivate super admin", func(t *testing.T) {
		mockMgr.toggleUserStatusFn = func(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
			return nil, domain.ErrCannotDeactivateSuperAdmin
		}

		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("returns 403 when role hierarchy is insufficient", func(t *testing.T) {
		mockMgr.toggleUserStatusFn = func(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
			return nil, domain.ErrInsufficientRoleHierarchy
		}

		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("returns 403 when trying to deactivate self", func(t *testing.T) {
		mockMgr.toggleUserStatusFn = func(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
			return nil, domain.ErrCannotDeactivateSelf
		}

		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("returns 404 when target user is not found", func(t *testing.T) {
		mockMgr.toggleUserStatusFn = func(ctx context.Context, claims *service.UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
			return nil, domain.ErrUserNotFound
		}

		body := bytes.NewReader([]byte(`{"is_active":false}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+validTargetID+"/toggle-status", body)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}
