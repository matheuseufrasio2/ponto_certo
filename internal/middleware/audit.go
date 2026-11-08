package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// AuditMetadataMiddleware extracts IP address and User-Agent header, injecting them into request context.
func AuditMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip != "" {
			parts := strings.Split(ip, ",")
			ip = strings.TrimSpace(parts[0])
		} else if rip := r.Header.Get("X-Real-IP"); rip != "" {
			ip = strings.TrimSpace(rip)
		} else {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				ip = host
			} else {
				ip = r.RemoteAddr
			}
		}

		ctx := domain.WithAuditMetadata(r.Context(), domain.AuditMetadata{
			IPAddress: ip,
			UserAgent: r.UserAgent(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
