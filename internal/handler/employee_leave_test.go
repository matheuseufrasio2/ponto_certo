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
	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockEmployeeLeaveManager struct {
	createLeaveFn func(ctx context.Context, input domain.CreateEmployeeLeaveInput, claims *service.UserClaims) (*domain.EmployeeLeave, error)
	listLeavesFn  func(ctx context.Context, claims *service.UserClaims, unitID *string, status string) ([]domain.EmployeeLeave, error)
	cancelLeaveFn func(ctx context.Context, leaveID string, claims *service.UserClaims) error
}

func (m *mockEmployeeLeaveManager) CreateLeave(ctx context.Context, input domain.CreateEmployeeLeaveInput, claims *service.UserClaims) (*domain.EmployeeLeave, error) {
	if m.createLeaveFn != nil {
		return m.createLeaveFn(ctx, input, claims)
	}
	return &domain.EmployeeLeave{ID: "leave-1", UserID: input.UserID}, nil
}

func (m *mockEmployeeLeaveManager) ListLeaves(ctx context.Context, claims *service.UserClaims, unitID *string, status string) ([]domain.EmployeeLeave, error) {
	if m.listLeavesFn != nil {
		return m.listLeavesFn(ctx, claims, unitID, status)
	}
	return nil, nil
}

func (m *mockEmployeeLeaveManager) CancelLeave(ctx context.Context, leaveID string, claims *service.UserClaims) error {
	if m.cancelLeaveFn != nil {
		return m.cancelLeaveFn(ctx, leaveID, claims)
	}
	return nil
}

func TestEmployeeLeaveHandler_Routes(t *testing.T) {
	jwtSecret := "test-secret"
	claims := &service.UserClaims{
		UserID: "owner-1",
		Role:   domain.RoleOrgOwner,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	validLeaveUUID := "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33"

	t.Run("successfully creates employee leave with 201 Created", func(t *testing.T) {
		h := handler.NewEmployeeLeaveHandler(&mockEmployeeLeaveManager{})
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Post("/owner/leaves", h.CreateLeave)

		payload := domain.CreateEmployeeLeaveInput{
			UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			LeaveType: domain.LeaveTypeVacation,
			StartDate: "2026-12-01",
			EndDate:   "2026-12-15",
			Reason:    "Year end vacation",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/owner/leaves", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenString)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("returns 409 Conflict when leave conflict exists", func(t *testing.T) {
		mockMgr := &mockEmployeeLeaveManager{
			createLeaveFn: func(ctx context.Context, input domain.CreateEmployeeLeaveInput, claims *service.UserClaims) (*domain.EmployeeLeave, error) {
				return nil, domain.ErrEmployeeLeaveConflict
			},
		}
		h := handler.NewEmployeeLeaveHandler(mockMgr)
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Post("/owner/leaves", h.CreateLeave)

		payload := domain.CreateEmployeeLeaveInput{
			UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			LeaveType: domain.LeaveTypeVacation,
			StartDate: "2026-12-01",
			EndDate:   "2026-12-15",
			Reason:    "Year end vacation",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/owner/leaves", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("successfully cancels leave with 204 No Content", func(t *testing.T) {
		h := handler.NewEmployeeLeaveHandler(&mockEmployeeLeaveManager{})
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Delete("/owner/leaves/{id}", h.CancelLeave)

		req := httptest.NewRequest("DELETE", "/owner/leaves/"+validLeaveUUID, nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d: %s", w.Code, w.Body.String())
		}
	})
}
