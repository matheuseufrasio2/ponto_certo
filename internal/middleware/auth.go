package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type contextKey struct{}

var userContextKey = contextKey{}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			tokenString := parts[1]
			claims := &service.UserClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			if holder, ok := r.Context().Value(userHolderKey).(*userHolder); ok && holder != nil {
				holder.email = claims.Email
				holder.role = string(claims.Role)
			}

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (*service.UserClaims, bool) {
	claims, ok := ctx.Value(userContextKey).(*service.UserClaims)
	return claims, ok
}

func RequireRoles(allowedRoles ...domain.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserFromContext(r.Context())
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			if claims.Role == domain.RoleSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			for _, allowed := range allowedRoles {
				if claims.Role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeJSONError(w, http.StatusForbidden, domain.ErrForbidden.Error())
		})
	}
}

func RequireUnitAccess(targetUnitID string, claims *service.UserClaims) bool {
	return claims.HasUnitAccess(targetUnitID)
}

// PreventImpersonatedMutations restricts high-risk actions when a session is impersonated by Super Admin.
func PreventImpersonatedMutations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := UserFromContext(r.Context())
		if ok && claims != nil && claims.IsImpersonated {
			writeJSONError(w, http.StatusForbidden, domain.ErrImpersonationActionBlocked.Error())
			return
		}
		next.ServeHTTP(w, r)
	})
}
