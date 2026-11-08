package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockAuditLogManager struct {
	listOrgFn    func(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error)
	listGlobalFn func(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error)
}

func (m *mockAuditLogManager) ListOrgAuditLogs(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error) {
	if m.listOrgFn != nil {
		return m.listOrgFn(ctx, claims, limit, offset)
	}
	return nil, nil
}

func (m *mockAuditLogManager) ListGlobalAuditLogs(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error) {
	if m.listGlobalFn != nil {
		return m.listGlobalFn(ctx, claims, limit, offset)
	}
	return nil, nil
}

func TestAuditLogHandler_ListOrgAuditLogs(t *testing.T) {
	jwtSecret := "test-secret"
	claims := &service.UserClaims{
		UserID: "owner-1",
		Role:   domain.RoleOrgOwner,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	t.Run("successfully lists org audit logs", func(t *testing.T) {
		mockMgr := &mockAuditLogManager{
			listOrgFn: func(ctx context.Context, claims *service.UserClaims, limit, offset int) ([]domain.AuditLog, error) {
				return []domain.AuditLog{
					{ID: "log-1", Action: "USER_DEACTIVATED", CreatedAt: time.Now()},
				}, nil
			},
		}
		h := handler.NewAuditLogHandler(mockMgr)

		req := httptest.NewRequest("GET", "/api/v1/owner/audit-logs?limit=10&offset=0", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		router := middleware.AuthMiddleware(jwtSecret)(http.HandlerFunc(h.ListOrgAuditLogs))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects request without authorization header", func(t *testing.T) {
		h := handler.NewAuditLogHandler(&mockAuditLogManager{})

		req := httptest.NewRequest("GET", "/api/v1/owner/audit-logs", nil)
		w := httptest.NewRecorder()

		router := middleware.AuthMiddleware(jwtSecret)(http.HandlerFunc(h.ListOrgAuditLogs))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})
}
