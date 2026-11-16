package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type stubManagerTeamService struct {
	members      []domain.TeamMemberItem
	created      *domain.TeamMemberItem
	workingHours []domain.BarberWorkingHour
	errToReturn  error
}

func (s *stubManagerTeamService) ListTeamMembers(ctx context.Context, unitID string, claims *service.UserClaims) ([]domain.TeamMemberItem, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.members, nil
}

func (s *stubManagerTeamService) CreateTeamMember(ctx context.Context, unitID string, input domain.CreateTeamMemberInput, claims *service.UserClaims) (*domain.TeamMemberItem, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.created, nil
}

func (s *stubManagerTeamService) GetWorkingHours(ctx context.Context, unitID, identifier string, claims *service.UserClaims) ([]domain.BarberWorkingHour, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.workingHours, nil
}

func (s *stubManagerTeamService) UpdateWorkingHours(ctx context.Context, unitID, identifier string, input domain.UpdateWorkingHoursInput, claims *service.UserClaims) ([]domain.BarberWorkingHour, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.workingHours, nil
}

func TestManagerTeamHandler_ListTeamMembers(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("200 OK with members", func(t *testing.T) {
		barberID := "barber-1"
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			members: []domain.TeamMemberItem{
				{
					ID:        "user-1",
					BarberID:  &barberID,
					Name:      "Carlos Barber",
					Email:     "carlos@barber.com",
					Role:      domain.RoleBarber,
					IsActive:  true,
					CreatedAt: time.Now().UTC(),
				},
			},
		})
		r := chi.NewRouter()
		r.Get("/team", h.ListTeamMembers)

		req := httptest.NewRequest(http.MethodGet, "/team", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp []domain.TeamMemberItem
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode err: %v", err)
		}
		if len(resp) != 1 || resp[0].Name != "Carlos Barber" {
			t.Fatalf("unexpected members payload: %+v", resp)
		}
	})

	t.Run("401 Unauthorized when missing claims", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{})
		r := chi.NewRouter()
		r.Get("/team", h.ListTeamMembers)

		req := httptest.NewRequest(http.MethodGet, "/team", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("400 Bad Request when unit_id missing", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{})
		r := chi.NewRouter()
		r.Get("/team", h.ListTeamMembers)

		noUnitClaims := &service.UserClaims{UserID: "admin-1", Role: domain.RoleSuperAdmin}
		req := httptest.NewRequest(http.MethodGet, "/team", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), noUnitClaims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("403 Forbidden when unauthorized", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			errToReturn: domain.ErrForbidden,
		})
		r := chi.NewRouter()
		r.Get("/team", h.ListTeamMembers)

		req := httptest.NewRequest(http.MethodGet, "/team", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})
}

func TestManagerTeamHandler_CreateTeamMember(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("201 Created", func(t *testing.T) {
		barberID := "barber-2"
		expected := &domain.TeamMemberItem{
			ID:        "user-2",
			BarberID:  &barberID,
			Name:      "Novo Barbeiro",
			Email:     "novo@barber.com",
			Role:      domain.RoleBarber,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
		}
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			created: expected,
		})
		r := chi.NewRouter()
		r.Post("/team", h.CreateTeamMember)

		body := map[string]any{
			"name":     "Novo Barbeiro",
			"email":    "novo@barber.com",
			"password": "Password123!",
			"role":     "BARBER",
		}
		buf, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("400 Bad Request on invalid JSON", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{})
		r := chi.NewRouter()
		r.Post("/team", h.CreateTeamMember)

		req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("409 Conflict when email already registered", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			errToReturn: domain.ErrManagerEmailTaken,
		})
		r := chi.NewRouter()
		r.Post("/team", h.CreateTeamMember)

		body := map[string]any{
			"name":     "Barbeiro Duplicado",
			"email":    "duplicado@barber.com",
			"password": "Password123!",
			"role":     "BARBER",
		}
		buf, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})
}

func TestManagerTeamHandler_WorkingHours(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	barberID := "barber-uuid-1234"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	dummyHours := []domain.BarberWorkingHour{
		{
			ID:        "bwh-1",
			BarberID:  barberID,
			DayOfWeek: 1,
			StartTime: "09:00",
			EndTime:   "18:00",
			IsOff:     false,
		},
	}

	t.Run("GET /team/{id}/working-hours 200 OK", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			workingHours: dummyHours,
		})
		r := chi.NewRouter()
		r.Get("/team/{id}/working-hours", h.GetWorkingHours)

		req := httptest.NewRequest(http.MethodGet, "/team/"+barberID+"/working-hours", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /team/{id}/working-hours 404 Not Found", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			errToReturn: domain.ErrNotFound,
		})
		r := chi.NewRouter()
		r.Get("/team/{id}/working-hours", h.GetWorkingHours)

		req := httptest.NewRequest(http.MethodGet, "/team/unknown-id/working-hours", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("PUT /team/{id}/working-hours 200 OK", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{
			workingHours: dummyHours,
		})
		r := chi.NewRouter()
		r.Put("/team/{id}/working-hours", h.UpdateWorkingHours)

		body := map[string]any{
			"schedule": []map[string]any{
				{"day_of_week": 1, "start_time": "09:00", "end_time": "18:00", "is_off": false},
			},
		}
		buf, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/team/"+barberID+"/working-hours", bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT /team/{id}/working-hours 400 Bad Request on malformed body", func(t *testing.T) {
		h := handler.NewManagerTeamHandler(&stubManagerTeamService{})
		r := chi.NewRouter()
		r.Put("/team/{id}/working-hours", h.UpdateWorkingHours)

		req := httptest.NewRequest(http.MethodPut, "/team/"+barberID+"/working-hours", bytes.NewReader([]byte("{malformed")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}
