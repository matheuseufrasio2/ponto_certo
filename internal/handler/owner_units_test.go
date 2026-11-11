package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockOwnerUnitsManager struct {
	listUnitsFn        func(ctx context.Context, claims *service.UserClaims, status string) ([]domain.BarbershopUnit, error)
	createUnitFn       func(ctx context.Context, claims *service.UserClaims, input domain.CreateUnitInput) (*domain.BarbershopUnit, error)
	updateUnitFn       func(ctx context.Context, claims *service.UserClaims, unitID string, input domain.UpdateUnitInput) (*domain.BarbershopUnit, error)
	toggleUnitStatusFn func(ctx context.Context, claims *service.UserClaims, unitID string, input domain.ToggleUnitStatusInput) (*domain.BarbershopUnit, error)
}

func (m *mockOwnerUnitsManager) ListUnits(ctx context.Context, claims *service.UserClaims, status string) ([]domain.BarbershopUnit, error) {
	if m.listUnitsFn != nil {
		return m.listUnitsFn(ctx, claims, status)
	}
	return []domain.BarbershopUnit{}, nil
}

func (m *mockOwnerUnitsManager) CreateUnit(ctx context.Context, claims *service.UserClaims, input domain.CreateUnitInput) (*domain.BarbershopUnit, error) {
	if m.createUnitFn != nil {
		return m.createUnitFn(ctx, claims, input)
	}
	return &domain.BarbershopUnit{}, nil
}

func (m *mockOwnerUnitsManager) UpdateUnit(ctx context.Context, claims *service.UserClaims, unitID string, input domain.UpdateUnitInput) (*domain.BarbershopUnit, error) {
	if m.updateUnitFn != nil {
		return m.updateUnitFn(ctx, claims, unitID, input)
	}
	return &domain.BarbershopUnit{}, nil
}

func (m *mockOwnerUnitsManager) ToggleUnitStatus(ctx context.Context, claims *service.UserClaims, unitID string, input domain.ToggleUnitStatusInput) (*domain.BarbershopUnit, error) {
	if m.toggleUnitStatusFn != nil {
		return m.toggleUnitStatusFn(ctx, claims, unitID, input)
	}
	return &domain.BarbershopUnit{}, nil
}

func signOwnerUnitTestToken(t *testing.T, secret string, role domain.Role, orgID *string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, service.UserClaims{
		UserID:         "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenStr
}

func TestOwnerUnitsHandler(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!"
	orgID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	unitID := "cccccccc-cccc-cccc-cccc-cccccccccccc"

	ownerToken := signOwnerUnitTestToken(t, secret, domain.RoleOrgOwner, &orgID)
	managerToken := signOwnerUnitTestToken(t, secret, domain.RoleUnitManager, &orgID)
	superAdminNoOrgToken := signOwnerUnitTestToken(t, secret, domain.RoleSuperAdmin, nil)

	mockMgr := &mockOwnerUnitsManager{
		listUnitsFn: func(ctx context.Context, claims *service.UserClaims, status string) ([]domain.BarbershopUnit, error) {
			if status == "invalid" {
				return nil, domain.ErrValidationFailed
			}
			if status == "active" {
				return []domain.BarbershopUnit{
					{ID: unitID, OrganizationID: orgID, Name: "Matriz Centro", ChairCount: 4, IsActive: true},
				}, nil
			}
			return nil, nil
		},
		createUnitFn: func(ctx context.Context, claims *service.UserClaims, input domain.CreateUnitInput) (*domain.BarbershopUnit, error) {
			if input.Name == "" {
				return nil, domain.ErrUnitNameRequired
			}
			if input.Name == "Duplicada" {
				return nil, domain.ErrUnitNameAlreadyExists
			}
			if input.ChairCount != nil && *input.ChairCount <= 0 {
				return nil, domain.ErrInvalidChairCount
			}
			chairs := 4
			if input.ChairCount != nil {
				chairs = *input.ChairCount
			}
			return &domain.BarbershopUnit{
				ID:             unitID,
				OrganizationID: *claims.OrganizationID,
				Name:           input.Name,
				ChairCount:     chairs,
				IsActive:       true,
			}, nil
		},
		updateUnitFn: func(ctx context.Context, claims *service.UserClaims, id string, input domain.UpdateUnitInput) (*domain.BarbershopUnit, error) {
			if id == "dddddddd-dddd-dddd-dddd-dddddddddddd" {
				return nil, domain.ErrUnitNotFound
			}
			return &domain.BarbershopUnit{
				ID:             id,
				OrganizationID: *claims.OrganizationID,
				Name:           input.Name,
				ChairCount:     input.ChairCount,
				IsActive:       true,
			}, nil
		},
		toggleUnitStatusFn: func(ctx context.Context, claims *service.UserClaims, id string, input domain.ToggleUnitStatusInput) (*domain.BarbershopUnit, error) {
			if id == "dddddddd-dddd-dddd-dddd-dddddddddddd" {
				return nil, domain.ErrUnitNotFound
			}
			return &domain.BarbershopUnit{
				ID:             id,
				OrganizationID: *claims.OrganizationID,
				Name:           "Matriz Centro",
				ChairCount:     4,
				IsActive:       input.IsActive,
			}, nil
		},
	}

	h := handler.NewOwnerUnitsHandler(mockMgr)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(secret))
		r.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleOrgOwner))
		r.Route("/owner", func(r chi.Router) {
			r.Get("/units", h.ListUnits)
			r.Post("/units", h.CreateUnit)
			r.Put("/units/{id}", h.UpdateUnit)
			r.Patch("/units/{id}/status", h.ToggleUnitStatus)
		})
	})

	t.Run("GET /owner/units returns 200 empty array when nil and filters active", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/units", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if w.Body.String() != "[]\n" {
			t.Errorf("expected empty json array [], got %q", w.Body.String())
		}

		reqActive := httptest.NewRequest(http.MethodGet, "/owner/units?status=active", nil)
		reqActive.Header.Set("Authorization", "Bearer "+ownerToken)
		wActive := httptest.NewRecorder()
		r.ServeHTTP(wActive, reqActive)

		if wActive.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wActive.Code)
		}
		var units []domain.BarbershopUnit
		if err := json.NewDecoder(wActive.Body).Decode(&units); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(units) != 1 || units[0].Name != "Matriz Centro" {
			t.Errorf("unexpected units: %+v", units)
		}
	})

	t.Run("GET /owner/units?status=invalid returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/units?status=invalid", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("POST /owner/units creates branch with 201", func(t *testing.T) {
		chairs := 6
		payload, _ := json.Marshal(domain.CreateUnitInput{
			Name:       "Filial Shopping Morumbi",
			ChairCount: &chairs,
		})
		req := httptest.NewRequest(http.MethodPost, "/owner/units", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var unit domain.BarbershopUnit
		_ = json.NewDecoder(w.Body).Decode(&unit)
		if unit.Name != "Filial Shopping Morumbi" || unit.ChairCount != 6 {
			t.Errorf("unexpected created unit: %+v", unit)
		}
	})

	t.Run("POST /owner/units rejects unknown fields, empty name, and invalid chair_count with 400", func(t *testing.T) {
		reqUnknown := httptest.NewRequest(http.MethodPost, "/owner/units", bytes.NewBufferString(`{"name":"Filial","extra":1}`))
		reqUnknown.Header.Set("Authorization", "Bearer "+ownerToken)
		wUnknown := httptest.NewRecorder()
		r.ServeHTTP(wUnknown, reqUnknown)
		if wUnknown.Code != http.StatusBadRequest {
			t.Errorf("unknown field: expected 400, got %d", wUnknown.Code)
		}

		reqZero := httptest.NewRequest(http.MethodPost, "/owner/units", bytes.NewBufferString(`{"name":"Filial","chair_count":0}`))
		reqZero.Header.Set("Authorization", "Bearer "+ownerToken)
		reqZero.Header.Set("Accept-Language", "en-US")
		wZero := httptest.NewRecorder()
		r.ServeHTTP(wZero, reqZero)
		if wZero.Code != http.StatusBadRequest {
			t.Errorf("chair_count=0: expected 400, got %d", wZero.Code)
		}
		var errResp handler.ErrorResponse
		_ = json.NewDecoder(wZero.Body).Decode(&errResp)
		if errResp.Code != domain.ErrCodeInvalidChairCount || errResp.Message != "Chair count must be greater than zero." {
			t.Errorf("unexpected localized error response: %+v", errResp)
		}
	})

	t.Run("POST /owner/units returns 409 on duplicate name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/owner/units", bytes.NewBufferString(`{"name":"Duplicada"}`))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", w.Code)
		}
		var errResp handler.ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&errResp)
		if errResp.Code != domain.ErrCodeUnitNameAlreadyExists {
			t.Errorf("expected UNIT_NAME_ALREADY_EXISTS, got %q", errResp.Code)
		}
	})

	t.Run("PUT /owner/units/{id} updates unit, rejects bad uuid, and returns 404 cross-tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/owner/units/"+unitID, bytes.NewBufferString(`{"name":"Matriz Nova","chair_count":8}`))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		reqBadID := httptest.NewRequest(http.MethodPut, "/owner/units/invalid-id", bytes.NewBufferString(`{"name":"Matriz","chair_count":4}`))
		reqBadID.Header.Set("Authorization", "Bearer "+ownerToken)
		wBadID := httptest.NewRecorder()
		r.ServeHTTP(wBadID, reqBadID)
		if wBadID.Code != http.StatusBadRequest {
			t.Errorf("expected 400 on invalid uuid, got %d", wBadID.Code)
		}

		reqOtherTenant := httptest.NewRequest(http.MethodPut, "/owner/units/dddddddd-dddd-dddd-dddd-dddddddddddd", bytes.NewBufferString(`{"name":"Matriz","chair_count":4}`))
		reqOtherTenant.Header.Set("Authorization", "Bearer "+ownerToken)
		wOtherTenant := httptest.NewRecorder()
		r.ServeHTTP(wOtherTenant, reqOtherTenant)
		if wOtherTenant.Code != http.StatusNotFound {
			t.Errorf("expected 404 on cross-tenant unit, got %d", wOtherTenant.Code)
		}
	})

	t.Run("PATCH /owner/units/{id}/status toggles active status and handles 404 cross-tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/owner/units/"+unitID+"/status", bytes.NewBufferString(`{"is_active":false}`))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var unit domain.BarbershopUnit
		_ = json.NewDecoder(w.Body).Decode(&unit)
		if unit.IsActive {
			t.Errorf("expected IsActive=false")
		}

		reqOtherTenant := httptest.NewRequest(http.MethodPatch, "/owner/units/dddddddd-dddd-dddd-dddd-dddddddddddd/status", bytes.NewBufferString(`{"is_active":false}`))
		reqOtherTenant.Header.Set("Authorization", "Bearer "+ownerToken)
		wOtherTenant := httptest.NewRecorder()
		r.ServeHTTP(wOtherTenant, reqOtherTenant)
		if wOtherTenant.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", wOtherTenant.Code)
		}
	})

	t.Run("enforces 401 unauthenticated and 403 unauthorized role or missing org", func(t *testing.T) {
		reqNoAuth := httptest.NewRequest(http.MethodGet, "/owner/units", nil)
		wNoAuth := httptest.NewRecorder()
		r.ServeHTTP(wNoAuth, reqNoAuth)
		if wNoAuth.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", wNoAuth.Code)
		}

		reqMgr := httptest.NewRequest(http.MethodGet, "/owner/units", nil)
		reqMgr.Header.Set("Authorization", "Bearer "+managerToken)
		wMgr := httptest.NewRecorder()
		r.ServeHTTP(wMgr, reqMgr)
		if wMgr.Code != http.StatusForbidden {
			t.Errorf("expected 403 for manager, got %d", wMgr.Code)
		}

		reqSANoOrg := httptest.NewRequest(http.MethodGet, "/owner/units", nil)
		reqSANoOrg.Header.Set("Authorization", "Bearer "+superAdminNoOrgToken)
		wSANoOrg := httptest.NewRecorder()
		r.ServeHTTP(wSANoOrg, reqSANoOrg)
		if wSANoOrg.Code != http.StatusForbidden {
			t.Errorf("expected 403 for super admin without org context, got %d", wSANoOrg.Code)
		}
	})
}
