package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockUserRepoForHandler struct {
	user *domain.User
	err  error
}

func (m *mockUserRepoForHandler) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepoForHandler) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func TestAuthHandler_Login(t *testing.T) {
	rawPassword := "senha123"
	hash, err := service.GeneratePasswordHash(rawPassword)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}

	testUser := &domain.User{
		ID:           "user-1",
		Name:         "Admin Test",
		Email:        "admin@test.com",
		PasswordHash: hash,
		Role:         domain.RoleSuperAdmin,
		IsActive:     true,
	}

	repo := &mockUserRepoForHandler{user: testUser}
	authSvc := service.NewAuthService(repo, "secret-key", time.Hour)
	authHdl := handler.NewAuthHandler(authSvc)

	t.Run("valid login returns 200 with token", func(t *testing.T) {
		payload := service.LoginInput{
			Email:    "admin@test.com",
			Password: "senha123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		authHdl.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var res service.LoginOutput
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.Token == "" || res.TokenType != "Bearer" {
			t.Errorf("expected bearer token, got %+v", res)
		}
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		payload := service.LoginInput{
			Email:    "admin@test.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		authHdl.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})
}
