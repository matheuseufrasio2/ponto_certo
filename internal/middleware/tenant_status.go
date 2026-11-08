package middleware

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// TenantStatusChecker defines the capability to verify if a tenant organization is currently active in real time.
type TenantStatusChecker interface {
	IsOrganizationActive(ctx context.Context, orgID string) (bool, error)
}

// RequireActiveTenant ensures that incoming requests belong to an active organization.
// Super Admin is exempt.
// Inactive organizations result in 403 Forbidden with ErrTenantInactive.
func RequireActiveTenant(checker TenantStatusChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserFromContext(r.Context())
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			// Super Admin or Impersonation session bypasses tenant status restrictions (for diagnostics and support)
			if claims.Role == domain.RoleSuperAdmin || claims.IsImpersonated {
				next.ServeHTTP(w, r)
				return
			}

			if claims.OrganizationID == nil || *claims.OrganizationID == "" {
				writeJSONError(w, http.StatusForbidden, domain.ErrTenantInactive.Error())
				return
			}

			isActive, err := checker.IsOrganizationActive(r.Context(), *claims.OrganizationID)
			if err != nil || !isActive {
				writeJSONError(w, http.StatusForbidden, domain.ErrTenantInactive.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
