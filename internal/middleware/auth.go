package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type contextKey struct{}

var userContextKey = contextKey{}

func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.HandleError(w, domain.ErrUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				handler.HandleError(w, domain.ErrUnauthorized)
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
				handler.HandleError(w, domain.ErrUnauthorized)
				return
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
				handler.HandleError(w, domain.ErrUnauthorized)
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

			handler.HandleError(w, domain.ErrForbidden)
		})
	}
}

func RequireUnitAccess(targetUnitID string, claims *service.UserClaims) bool {
	if claims == nil {
		return false
	}

	if claims.Role == domain.RoleSuperAdmin {
		return true
	}

	if claims.Role == domain.RoleOrgOwner {
		return true
	}

	if claims.UnitID != nil && *claims.UnitID == targetUnitID {
		return true
	}

	return false
}
