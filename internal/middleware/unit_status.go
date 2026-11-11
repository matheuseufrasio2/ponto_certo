package middleware

import (
	"context"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// UnitStatusChecker verifies whether a barbershop branch is currently active in real time.
type UnitStatusChecker interface {
	IsUnitActive(ctx context.Context, unitID string) (bool, error)
}

// RequireActiveUnit ensures unit-bound staff (manager, receptionist, barber) belong to an active branch.
// Super Admin, Org Owner, and Impersonation sessions bypass branch status restrictions.
func RequireActiveUnit(checker UnitStatusChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserFromContext(r.Context())
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
				return
			}

			if claims.Role == domain.RoleSuperAdmin || claims.Role == domain.RoleOrgOwner || claims.IsImpersonated {
				next.ServeHTTP(w, r)
				return
			}

			if claims.UnitID == nil || *claims.UnitID == "" {
				writeJSONError(w, http.StatusForbidden, domain.ErrUnitInactive.Error())
				return
			}

			isActive, err := checker.IsUnitActive(r.Context(), *claims.UnitID)
			if err != nil || !isActive {
				writeJSONError(w, http.StatusForbidden, domain.ErrUnitInactive.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
