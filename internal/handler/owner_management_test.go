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

type mockOrgProfileManager struct {
	getProfileFn    func(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error)
	updateProfileFn func(ctx context.Context, orgID string, input domain.UpdateOrganizationInput) (*domain.OrganizationProfileResponse, error)
}

func (m *mockOrgProfileManager) GetProfile(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error) {
	if m.getProfileFn != nil {
		return m.getProfileFn(ctx, orgID)
	}
	return nil, nil
}

func (m *mockOrgProfileManager) UpdateProfile(ctx context.Context, orgID string, input domain.UpdateOrganizationInput) (*domain.OrganizationProfileResponse, error) {
	if m.updateProfileFn != nil {
		return m.updateProfileFn(ctx, orgID, input)
	}
	return nil, nil
}

func TestOwnerManagementHandler(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!"
	orgID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	userID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, service.UserClaims{
		UserID:         userID,
		OrganizationID: &orgID,
		Role:           domain.RoleOrgOwner,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	mockMgr := &mockOrgProfileManager{
		getProfileFn: func(ctx context.Context, id string) (*domain.OrganizationProfileResponse, error) {
			return &domain.OrganizationProfileResponse{
				ID:                 id,
				Name:               "Barbearia Viking",
				TradeName:          "Viking Hair",
				Document:           "12345678000199",
				Phone:              "11988887777",
				LogoURL:            "https://cdn.example.com/logo.png",
				Plan:               "PRO",
				SubscriptionStatus: domain.SubscriptionStatusActive,
				IsActive:           true,
			}, nil
		},
		updateProfileFn: func(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.OrganizationProfileResponse, error) {
			if input.Name == "Invalido" {
				return nil, domain.ErrValidationFailed
			}
			return &domain.OrganizationProfileResponse{
				ID:        id,
				Name:      input.Name,
				TradeName: *input.TradeName,
				Document:  *input.Document,
				Phone:     *input.Phone,
				LogoURL:   *input.LogoURL,
				IsActive:  true,
			}, nil
		},
	}

	h := handler.NewOwnerManagementHandler(mockMgr)
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(secret))
		r.Use(middleware.RequireRoles(domain.RoleOrgOwner, domain.RoleSuperAdmin))

		r.Route("/owner", func(r chi.Router) {
			r.Get("/organization", h.GetOrganization)
			r.Put("/organization", h.UpdateOrganization)
		})
	})

	t.Run("GET /owner/organization returns 200 with profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/organization", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp domain.OrganizationProfileResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode json: %v", err)
		}
		if resp.Name != "Barbearia Viking" {
			t.Errorf("expected name 'Barbearia Viking', got %s", resp.Name)
		}
	})

	t.Run("PUT /owner/organization updates profile with 200", func(t *testing.T) {
		tradeName := "Viking New"
		doc := "12345678000199"
		phone := "11988887777"
		logo := "https://cdn.example.com/new.png"

		payload, _ := json.Marshal(domain.UpdateOrganizationInput{
			Name:      "Barbearia Viking Atualizada",
			TradeName: &tradeName,
			Document:  &doc,
			Phone:     &phone,
			LogoURL:   &logo,
		})

		req := httptest.NewRequest(http.MethodPut, "/owner/organization", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp domain.OrganizationProfileResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Name != "Barbearia Viking Atualizada" {
			t.Errorf("expected updated name, got %s", resp.Name)
		}
	})

	t.Run("GET /owner/organization returns 401 without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/organization", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("PUT /owner/organization returns 400 when validation fails", func(t *testing.T) {
		payload, _ := json.Marshal(domain.UpdateOrganizationInput{
			Name: "Invalido",
		})

		req := httptest.NewRequest(http.MethodPut, "/owner/organization", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}
