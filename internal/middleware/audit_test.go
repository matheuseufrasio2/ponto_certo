package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestAuditMetadataMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		headerKey     string
		headerVal     string
		userAgent     string
		expectedIP    string
		expectedAgent string
	}{
		{
			name:          "extracts from RemoteAddr host:port",
			remoteAddr:    "192.168.1.100:54321",
			userAgent:     "Mozilla/5.0",
			expectedIP:    "192.168.1.100",
			expectedAgent: "Mozilla/5.0",
		},
		{
			name:          "prefers X-Forwarded-For over RemoteAddr",
			remoteAddr:    "127.0.0.1:8080",
			headerKey:     "X-Forwarded-For",
			headerVal:     "203.0.113.195, 70.41.3.18",
			userAgent:     "PostmanRuntime/7.28.4",
			expectedIP:    "203.0.113.195",
			expectedAgent: "PostmanRuntime/7.28.4",
		},
		{
			name:          "prefers X-Real-IP if X-Forwarded-For is missing",
			remoteAddr:    "127.0.0.1:8080",
			headerKey:     "X-Real-IP",
			headerVal:     "198.51.100.17",
			userAgent:     "curl/7.68.0",
			expectedIP:    "198.51.100.17",
			expectedAgent: "curl/7.68.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}

			var capturedMeta domain.AuditMetadata
			handler := AuditMetadataMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMeta = domain.GetAuditMetadata(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if capturedMeta.IPAddress != tt.expectedIP {
				t.Errorf("expected IP %q, got %q", tt.expectedIP, capturedMeta.IPAddress)
			}
			if capturedMeta.UserAgent != tt.expectedAgent {
				t.Errorf("expected User-Agent %q, got %q", tt.expectedAgent, capturedMeta.UserAgent)
			}
		})
	}
}

func TestWithAuditMetadata(t *testing.T) {
	ctx := context.Background()
	meta := domain.AuditMetadata{IPAddress: "10.0.0.1", UserAgent: "TestClient"}
	ctxWithMeta := domain.WithAuditMetadata(ctx, meta)

	got := domain.GetAuditMetadata(ctxWithMeta)
	if got.IPAddress != "10.0.0.1" || got.UserAgent != "TestClient" {
		t.Errorf("unexpected metadata: %+v", got)
	}

	nilGot := domain.GetAuditMetadata(nil)
	if nilGot.IPAddress != "" || nilGot.UserAgent != "" {
		t.Errorf("expected empty metadata for nil ctx, got %+v", nilGot)
	}
}
