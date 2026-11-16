package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type stubManagerUnitSettingsService struct {
	settings    *domain.UnitSettingsResponse
	errToReturn error
}

func (s *stubManagerUnitSettingsService) GetUnitSettings(ctx context.Context, unitID string, claims *service.UserClaims) (*domain.UnitSettingsResponse, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.settings, nil
}

func (s *stubManagerUnitSettingsService) UpdateUnitSettings(ctx context.Context, unitID string, input domain.UpdateUnitSettingsInput, claims *service.UserClaims) (*domain.UnitSettingsResponse, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.settings, nil
}

func TestManagerUnitSettingsHandler(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	phone := "11999998888"
	address := "Rua das Flores, 123"
	dummySettings := &domain.UnitSettingsResponse{
		UnitID:         unitID,
		UnitName:       "Unidade Central",
		Phone:          &phone,
		Address:        &address,
		ChairCount:     5,
		OperatingHours: []domain.UnitOperatingHour{
			{
				ID:        "uoh-1",
				UnitID:    unitID,
				DayOfWeek: 1,
				OpenTime:  "09:00",
				CloseTime: "19:00",
				IsClosed:  false,
			},
		},
	}

	t.Run("GET /unit-settings 200 OK", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{
			settings: dummySettings,
		})
		r := chi.NewRouter()
		r.Get("/unit-settings", h.GetUnitSettings)

		req := httptest.NewRequest(http.MethodGet, "/unit-settings", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp domain.UnitSettingsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if resp.UnitName != "Unidade Central" || len(resp.OperatingHours) != 1 {
			t.Fatalf("unexpected settings response: %+v", resp)
		}
	})

	t.Run("GET /unit-settings 401 Unauthorized when missing claims", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{})
		r := chi.NewRouter()
		r.Get("/unit-settings", h.GetUnitSettings)

		req := httptest.NewRequest(http.MethodGet, "/unit-settings", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("GET /unit-settings 400 Bad Request when unit_id missing", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{})
		r := chi.NewRouter()
		r.Get("/unit-settings", h.GetUnitSettings)

		noUnitClaims := &service.UserClaims{UserID: "admin-1", Role: domain.RoleSuperAdmin}
		req := httptest.NewRequest(http.MethodGet, "/unit-settings", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), noUnitClaims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("PUT /unit-settings 200 OK", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{
			settings: dummySettings,
		})
		r := chi.NewRouter()
		r.Put("/unit-settings", h.UpdateUnitSettings)

		body := map[string]any{
			"phone":   "11988887777",
			"address": "Av Paulista, 1000",
		}
		buf, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/unit-settings", bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT /unit-settings 400 Bad Request on invalid JSON", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{})
		r := chi.NewRouter()
		r.Put("/unit-settings", h.UpdateUnitSettings)

		req := httptest.NewRequest(http.MethodPut, "/unit-settings", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("PUT /unit-settings 404 Not Found", func(t *testing.T) {
		h := handler.NewManagerUnitSettingsHandler(&stubManagerUnitSettingsService{
			errToReturn: domain.ErrUnitNotFound,
		})
		r := chi.NewRouter()
		r.Put("/unit-settings", h.UpdateUnitSettings)

		body := map[string]any{
			"phone": "11988887777",
		}
		buf, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/unit-settings", bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}
