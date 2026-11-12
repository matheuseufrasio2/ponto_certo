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

type mockOwnerManagersService struct {
	listManagersFn  func(ctx context.Context, claims *service.UserClaims) ([]domain.ManagerListItem, error)
	createManagerFn func(ctx context.Context, claims *service.UserClaims, input domain.CreateManagerInput) (*domain.ManagerListItem, error)
}

func (m *mockOwnerManagersService) ListManagers(ctx context.Context, claims *service.UserClaims) ([]domain.ManagerListItem, error) {
	if m.listManagersFn != nil {
		return m.listManagersFn(ctx, claims)
	}
	return []domain.ManagerListItem{}, nil
}

func (m *mockOwnerManagersService) CreateManager(ctx context.Context, claims *service.UserClaims, input domain.CreateManagerInput) (*domain.ManagerListItem, error) {
	if m.createManagerFn != nil {
		return m.createManagerFn(ctx, claims, input)
	}
	return &domain.ManagerListItem{}, nil
}

func signOwnerManagerToken(t *testing.T, secret string, role domain.Role, orgID *string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, service.UserClaims{
		UserID:         "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return s
}

func TestOwnerManagersHandler(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!"
	orgID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	unitID := "cccccccc-cccc-cccc-cccc-cccccccccccc"

	ownerToken := signOwnerManagerToken(t, secret, domain.RoleOrgOwner, &orgID)
	managerToken := signOwnerManagerToken(t, secret, domain.RoleUnitManager, &orgID)

	mockSvc := &mockOwnerManagersService{
		listManagersFn: func(ctx context.Context, claims *service.UserClaims) ([]domain.ManagerListItem, error) {
			return []domain.ManagerListItem{
				{
					ID:        "dddddddd-dddd-dddd-dddd-dddddddddddd",
					Name:      "João Gerente",
					Email:     "joao@pontocerto.com.br",
					Phone:     "(11) 98765-4321",
					UnitID:    unitID,
					UnitName:  "Unidade Centro",
					IsActive:  true,
					CreatedAt: time.Now().UTC(),
				},
			}, nil
		},
		createManagerFn: func(ctx context.Context, claims *service.UserClaims, input domain.CreateManagerInput) (*domain.ManagerListItem, error) {
			if err := input.Validate(); err != nil {
				return nil, err
			}
			if input.Email == "taken@pontocerto.com.br" {
				return nil, domain.ErrManagerEmailTaken
			}
			if input.UnitID == "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee" {
				return nil, domain.ErrUnitNotInOrg
			}
			return &domain.ManagerListItem{
				ID:        "ffffffff-ffff-ffff-ffff-ffffffffffff",
				Name:      input.Name,
				Email:     input.Email,
				Phone:     input.Phone,
				UnitID:    input.UnitID,
				UnitName:  "Unidade Centro",
				IsActive:  true,
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}

	h := handler.NewOwnerManagersHandler(mockSvc)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(secret))
		r.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleOrgOwner))
		r.Route("/owner", func(r chi.Router) {
			r.Get("/managers", h.ListManagers)
			r.Post("/managers", h.CreateManager)
		})
	})

	t.Run("GET /owner/managers returns 401 without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/managers", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("GET /owner/managers returns 403 for UNIT_MANAGER", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/managers", nil)
		req.Header.Set("Authorization", "Bearer "+managerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("GET /owner/managers returns 200 with managers list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner/managers", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var managers []domain.ManagerListItem
		if err := json.NewDecoder(w.Body).Decode(&managers); err != nil {
			t.Fatalf("failed to decode json: %v", err)
		}
		if len(managers) != 1 || managers[0].UnitName != "Unidade Centro" {
			t.Fatalf("unexpected managers payload: %+v", managers)
		}
	})

	t.Run("POST /owner/managers returns 201 on valid creation", func(t *testing.T) {
		body, _ := json.Marshal(domain.CreateManagerInput{
			Name:     "João Centro",
			Email:    "joao@pontocerto.com.br",
			Phone:    "(11) 98765-4321",
			Password: "Password123",
			UnitID:   unitID,
		})
		req := httptest.NewRequest(http.MethodPost, "/owner/managers", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		var created domain.ManagerListItem
		if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode json: %v", err)
		}
		if created.Email != "joao@pontocerto.com.br" || !created.IsActive {
			t.Fatalf("unexpected created manager: %+v", created)
		}
	})

	t.Run("POST /owner/managers returns 400 on invalid email", func(t *testing.T) {
		body, _ := json.Marshal(domain.CreateManagerInput{
			Name:     "João Centro",
			Email:    "invalid-email",
			Phone:    "11987654321",
			Password: "Password123",
			UnitID:   unitID,
		})
		req := httptest.NewRequest(http.MethodPost, "/owner/managers", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
		var errRes handler.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errRes); err != nil {
			t.Fatalf("failed to decode err response: %v", err)
		}
		if errRes.Code != domain.ErrCodeValidationFailed || errRes.Fields["email"] == "" {
			t.Fatalf("expected VALIDATION_FAILED with email field error, got %+v", errRes)
		}
	})

	t.Run("POST /owner/managers returns 409 on duplicate email", func(t *testing.T) {
		body, _ := json.Marshal(domain.CreateManagerInput{
			Name:     "João Centro",
			Email:    "taken@pontocerto.com.br",
			Phone:    "11987654321",
			Password: "Password123",
			UnitID:   unitID,
		})
		req := httptest.NewRequest(http.MethodPost, "/owner/managers", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
		var errRes handler.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errRes); err != nil {
			t.Fatalf("failed to decode err response: %v", err)
		}
		if errRes.Code != domain.ErrCodeManagerEmailTaken {
			t.Fatalf("expected MANAGER_EMAIL_TAKEN, got %q", errRes.Code)
		}
	})

	t.Run("POST /owner/managers returns 403 when unit belongs to another organization", func(t *testing.T) {
		body, _ := json.Marshal(domain.CreateManagerInput{
			Name:     "João Centro",
			Email:    "novo@pontocerto.com.br",
			Phone:    "11987654321",
			Password: "Password123",
			UnitID:   "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
		})
		req := httptest.NewRequest(http.MethodPost, "/owner/managers", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
		var errRes handler.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errRes); err != nil {
			t.Fatalf("failed to decode err response: %v", err)
		}
		if errRes.Code != domain.ErrCodeUnitNotInOrg {
			t.Fatalf("expected UNIT_NOT_IN_ORG, got %q", errRes.Code)
		}
	})
}
