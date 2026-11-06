package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// mockUserRepo simulates database queries in memory for isolated unit testing.
type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func TestAuthService_Login(t *testing.T) {
	rawPassword := "senha123"
	hash, err := service.GeneratePasswordHash(rawPassword)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	testUser := &domain.User{
		ID:           "user-123",
		Name:         "Carlos Admin",
		Email:        "carlos@barbearia.com",
		PasswordHash: hash,
		Role:         domain.RoleOrgOwner,
		IsActive:     true,
	}

	t.Run("successful login generates valid jwt", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		output, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != nil {
			t.Fatalf("expected successful login, got error: %v", err)
		}

		if output.Token == "" || output.TokenType != "Bearer" {
			t.Errorf("expected bearer token, got: %+v", output)
		}

		if output.User.Email != testUser.Email {
			t.Errorf("expected email %s, got %s", testUser.Email, output.User.Email)
		}

		if output.ExpiresIn != 3600 {
			t.Errorf("expected expiresIn 3600, got %d", output.ExpiresIn)
		}
	})

	t.Run("login with wrong password returns unauthorized", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha_errada",
		})

		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("login with inactive user returns forbidden", func(t *testing.T) {
		inactiveUser := *testUser
		inactiveUser.IsActive = false
		repo := &mockUserRepo{user: &inactiveUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "carlos@barbearia.com",
			Password: "senha123",
		})

		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("login with empty credentials returns unauthorized", func(t *testing.T) {
		repo := &mockUserRepo{user: testUser}
		authSvc := service.NewAuthService(repo, "test-secret-key-12345", 1*time.Hour)

		_, err := authSvc.Login(context.Background(), service.LoginInput{
			Email:    "",
			Password: "",
		})

		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})
}

