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

type mockScheduleBlockManager struct {
	createBlockFn func(ctx context.Context, barberID string, input domain.CreateScheduleBlockInput, claims *service.UserClaims) (*domain.BarberScheduleBlock, error)
	listBlocksFn  func(ctx context.Context, barberID string, from, to time.Time, claims *service.UserClaims) ([]domain.BarberScheduleBlock, error)
	deleteBlockFn func(ctx context.Context, blockID string, claims *service.UserClaims) error
}

func (m *mockScheduleBlockManager) CreateBlock(ctx context.Context, barberID string, input domain.CreateScheduleBlockInput, claims *service.UserClaims) (*domain.BarberScheduleBlock, error) {
	if m.createBlockFn != nil {
		return m.createBlockFn(ctx, barberID, input, claims)
	}
	return &domain.BarberScheduleBlock{ID: "b-1", BarberID: barberID}, nil
}

func (m *mockScheduleBlockManager) ListBlocks(ctx context.Context, barberID string, from, to time.Time, claims *service.UserClaims) ([]domain.BarberScheduleBlock, error) {
	if m.listBlocksFn != nil {
		return m.listBlocksFn(ctx, barberID, from, to, claims)
	}
	return nil, nil
}

func (m *mockScheduleBlockManager) DeleteBlock(ctx context.Context, blockID string, claims *service.UserClaims) error {
	if m.deleteBlockFn != nil {
		return m.deleteBlockFn(ctx, blockID, claims)
	}
	return nil
}

func TestScheduleBlockHandler_Routes(t *testing.T) {
	jwtSecret := "test-secret"
	claims := &service.UserClaims{
		UserID: "mgr-1",
		Role:   domain.RoleUnitManager,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	validUUID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	validBlockUUID := "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"

	t.Run("successfully creates schedule block with 201 Created", func(t *testing.T) {
		h := handler.NewScheduleBlockHandler(&mockScheduleBlockManager{})
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Post("/manager/barbers/{id}/blocks", h.CreateBlock)

		payload := domain.CreateScheduleBlockInput{
			StartTime: time.Now().Add(time.Hour),
			EndTime:   time.Now().Add(2 * time.Hour),
			Reason:    "Doctor appointment",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/manager/barbers/"+validUUID+"/blocks", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenString)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("returns 409 Conflict when appointment conflict occurs", func(t *testing.T) {
		mockMgr := &mockScheduleBlockManager{
			createBlockFn: func(ctx context.Context, barberID string, input domain.CreateScheduleBlockInput, claims *service.UserClaims) (*domain.BarberScheduleBlock, error) {
				return nil, domain.ErrScheduleBlockConflict
			},
		}
		h := handler.NewScheduleBlockHandler(mockMgr)
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Post("/manager/barbers/{id}/blocks", h.CreateBlock)

		payload := domain.CreateScheduleBlockInput{
			StartTime: time.Now().Add(time.Hour),
			EndTime:   time.Now().Add(2 * time.Hour),
			Reason:    "Doctor appointment",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/manager/barbers/"+validUUID+"/blocks", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("successfully deletes block with 204 No Content", func(t *testing.T) {
		h := handler.NewScheduleBlockHandler(&mockScheduleBlockManager{})
		r := chi.NewRouter()
		r.Use(middleware.AuthMiddleware(jwtSecret))
		r.Delete("/manager/barbers/blocks/{block_id}", h.DeleteBlock)

		req := httptest.NewRequest("DELETE", "/manager/barbers/blocks/"+validBlockUUID, nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d: %s", w.Code, w.Body.String())
		}
	})
}
