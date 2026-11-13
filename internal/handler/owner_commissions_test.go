package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type stubOwnerCommissionsManager struct {
	resp *domain.UpdateCommissionsResponse
	err  error
}

func (s *stubOwnerCommissionsManager) UpdateCommissions(_ context.Context, _ *service.UserClaims, _ domain.UpdateCommissionsInput) (*domain.UpdateCommissionsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func TestOwnerCommissionsHandler_UpdateCommissions(t *testing.T) {
	orgID := "11111111-1111-1111-1111-111111111111"
	claims := &service.UserClaims{
		UserID:         "44444444-4444-4444-4444-444444444444",
		Role:           domain.RoleOrgOwner,
		OrganizationID: &orgID,
	}

	t.Run("returns 401 when claims are missing", func(t *testing.T) {
		h := handler.NewOwnerCommissionsHandler(&stubOwnerCommissionsManager{})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/owner/commissions", bytes.NewBufferString(`{}`))
		rec := httptest.NewRecorder()

		h.UpdateCommissions(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("returns 400 on malformed JSON", func(t *testing.T) {
		h := handler.NewOwnerCommissionsHandler(&stubOwnerCommissionsManager{})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/owner/commissions", bytes.NewBufferString(`{bad`))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		h.UpdateCommissions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns 403 when barber is not in org", func(t *testing.T) {
		h := handler.NewOwnerCommissionsHandler(&stubOwnerCommissionsManager{err: domain.ErrBarberNotInOrg})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/owner/commissions", bytes.NewBufferString(`{"scope":"BARBER","barber_id":"33333333-3333-3333-3333-333333333333","percentage":50}`))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		h.UpdateCommissions(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("returns 400 on invalid commission scope", func(t *testing.T) {
		h := handler.NewOwnerCommissionsHandler(&stubOwnerCommissionsManager{err: domain.ErrInvalidCommissionScope})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/owner/commissions", bytes.NewBufferString(`{"scope":"OTHER","percentage":50}`))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		h.UpdateCommissions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns 200 with summary payload on success", func(t *testing.T) {
		h := handler.NewOwnerCommissionsHandler(&stubOwnerCommissionsManager{
			resp: &domain.UpdateCommissionsResponse{
				Scope:           domain.CommissionScopeAllOrganization,
				Percentage:      55,
				BarbersAffected: 4,
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/owner/commissions", bytes.NewBufferString(`{"scope":"ALL_ORGANIZATION","percentage":55}`))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		h.UpdateCommissions(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var got domain.UpdateCommissionsResponse
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.Scope != domain.CommissionScopeAllOrganization || got.Percentage != 55 || got.BarbersAffected != 4 {
			t.Fatalf("unexpected response payload: %+v", got)
		}
	})
}
