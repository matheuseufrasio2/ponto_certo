package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
)

func TestRequestLogger_DevelopmentAndProduction(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		statusCode int
		authSecret string
	}{
		{
			name:       "dev logger with 200 OK unauthenticated",
			env:        "development",
			statusCode: http.StatusOK,
		},
		{
			name:       "dev logger with 404 Not Found",
			env:        "development",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "dev logger with 500 Internal Server Error",
			env:        "development",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "prod logger with authenticated user",
			env:        "production",
			statusCode: http.StatusOK,
			authSecret: "test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMw := middleware.RequestLogger(tt.env)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			})

			var finalHandler http.Handler = handler
			if tt.authSecret != "" {
				finalHandler = middleware.AuthMiddleware(tt.authSecret)(handler)
			}
			wrapped := loggerMw(finalHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
			if tt.authSecret != "" {
				token := generateTestToken(tt.authSecret, domain.RoleOrgOwner, time.Now().Add(time.Hour))
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, rec.Code)
			}
		})
	}
}
