package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockDevRepoForHandler struct {
	listSuperAdminsFn func(ctx context.Context) ([]domain.DevSuperAdmin, error)
	listTenantsFn     func(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error)
	listTenantUsersFn func(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error)
	listUsersByRoleFn func(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error)
}

func (m *mockDevRepoForHandler) ListSuperAdmins(ctx context.Context) ([]domain.DevSuperAdmin, error) {
	if m.listSuperAdminsFn != nil {
		return m.listSuperAdminsFn(ctx)
	}
	return []domain.DevSuperAdmin{}, nil
}

func (m *mockDevRepoForHandler) ListTenants(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error) {
	if m.listTenantsFn != nil {
		return m.listTenantsFn(ctx, search, scenario, limit, offset)
	}
	return []domain.DevTenant{}, 0, nil
}

func (m *mockDevRepoForHandler) ListTenantUsers(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error) {
	if m.listTenantUsersFn != nil {
		return m.listTenantUsersFn(ctx, tenantID)
	}
	return []domain.DevTenantUser{}, nil
}

func (m *mockDevRepoForHandler) ListUsersByRole(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error) {
	if m.listUsersByRoleFn != nil {
		return m.listUsersByRoleFn(ctx, role, limit, offset)
	}
	return []domain.DevUser{}, 0, nil
}

func setupDevTestRouter(repo *mockDevRepoForHandler) *chi.Mux {
	svc := service.NewDevService(repo)
	h := handler.NewDevHandler(svc)

	r := chi.NewRouter()
	r.Route("/dev", func(r chi.Router) {
		r.Get("/super-admins", h.ListSuperAdmins)
		r.Get("/tenants", h.ListTenants)
		r.Get("/tenants/{id}/users", h.ListTenantUsers)
		r.Get("/users", h.ListUsers)
	})

	return r
}

func TestDevHandler_ListSuperAdmins(t *testing.T) {
	repo := &mockDevRepoForHandler{
		listSuperAdminsFn: func(ctx context.Context) ([]domain.DevSuperAdmin, error) {
			return []domain.DevSuperAdmin{
				{
					ID:    "c0000000-0000-0000-0000-000000000001",
					Name:  "Carlos Super Admin",
					Email: "superadmin@pontocerto.com.br",
					Role:  "SUPER_ADMIN",
				},
			}, nil
		},
	}

	r := setupDevTestRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/dev/super-admins", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rec.Code)
	}

	var admins []domain.DevSuperAdmin
	if err := json.NewDecoder(rec.Body).Decode(&admins); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(admins) != 1 || admins[0].Email != "superadmin@pontocerto.com.br" {
		t.Errorf("unexpected admins payload: %+v", admins)
	}
}

func TestDevHandler_ListTenants(t *testing.T) {
	repo := &mockDevRepoForHandler{
		listTenantsFn: func(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error) {
			return []domain.DevTenant{
				{
					ID:        "a0000000-0000-0000-0000-000000000001",
					TradeName: "Barbearia Don Corleone ME",
					Plan:      "PRO",
					Status:    "ACTIVE",
				},
			}, 1, nil
		},
	}

	r := setupDevTestRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/dev/tenants?page=1&limit=10&scenario=past_due", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rec.Code)
	}

	var res domain.DevTenantsResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(res.Tenants) != 1 || res.Total != 1 {
		t.Errorf("unexpected tenants response: %+v", res)
	}
}

func TestDevHandler_ListTenantUsers(t *testing.T) {
	tenantID := "a0000000-0000-0000-0000-000000000001"
	repo := &mockDevRepoForHandler{
		listTenantUsersFn: func(ctx context.Context, id string) ([]domain.DevTenantUser, error) {
			return []domain.DevTenantUser{
				{
					ID:       "c0000000-0000-0000-0000-000000000002",
					Name:     "Vito",
					Email:    "owner.corleone@pontocerto.com.br",
					Role:     "ORG_OWNER",
					UnitName: "Matriz",
				},
			}, nil
		},
	}

	r := setupDevTestRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/dev/tenants/"+tenantID+"/users", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rec.Code)
	}

	var users []domain.DevTenantUser
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(users) != 1 || users[0].Email != "owner.corleone@pontocerto.com.br" {
		t.Errorf("unexpected tenant users: %+v", users)
	}
}

func TestDevHandler_ListUsers(t *testing.T) {
	repo := &mockDevRepoForHandler{
		listUsersByRoleFn: func(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error) {
			return []domain.DevUser{
				{
					ID:         "c0000000-0000-0000-0000-000000000005",
					Name:       "Marcos",
					Email:      "marcos.barbeiro@pontocerto.com.br",
					Role:       "BARBER",
					TenantName: "Barbearias Ponto Certo",
				},
			}, 1, nil
		},
	}

	r := setupDevTestRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/dev/users?role=BARBER&page=1&limit=10", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rec.Code)
	}

	var res domain.DevUsersResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(res.Users) != 1 || res.Users[0].Role != "BARBER" {
		t.Errorf("unexpected users response: %+v", res)
	}
}
