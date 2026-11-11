package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockOrganizationRepo struct {
	getByIDFn func(ctx context.Context, id string) (*domain.Organization, error)
	updateFn   func(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.Organization, error)
}

func (m *mockOrganizationRepo) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockOrganizationRepo) Update(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.Organization, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, nil
}

func TestOrganizationService_GetProfile(t *testing.T) {
	validOrgID := "11111111-1111-1111-1111-111111111111"

	t.Run("success returns profile", func(t *testing.T) {
		repo := &mockOrganizationRepo{
			getByIDFn: func(ctx context.Context, id string) (*domain.Organization, error) {
				return &domain.Organization{
					ID:                 validOrgID,
					Name:               "Barbearia Matriz",
					TradeName:          "Matriz VIP",
					Document:           "12345678000199",
					Phone:              "11999998888",
					LogoURL:            "https://cdn.example.com/logo.png",
					Plan:               "PRO",
					SubscriptionStatus: domain.SubscriptionStatusActive,
					IsActive:           true,
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
				}, nil
			},
		}

		svc := service.NewOrganizationService(repo)
		profile, err := svc.GetProfile(context.Background(), validOrgID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.Name != "Barbearia Matriz" {
			t.Errorf("expected name 'Barbearia Matriz', got %s", profile.Name)
		}
		if profile.LogoURL != "https://cdn.example.com/logo.png" {
			t.Errorf("expected logo_url, got %s", profile.LogoURL)
		}
	})

	t.Run("fails on invalid uuid", func(t *testing.T) {
		svc := service.NewOrganizationService(&mockOrganizationRepo{})
		_, err := svc.GetProfile(context.Background(), "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Errorf("expected ErrInvalidUUID, got %v", err)
		}
	})
}

func TestOrganizationService_UpdateProfile(t *testing.T) {
	validOrgID := "11111111-1111-1111-1111-111111111111"
	validTradeName := "Barbearia Corleone"
	validDoc := "12.345.678/0001-90"
	validPhone := "(11) 98765-4321"
	validLogo := "https://imgbb.com/sample.png"

	t.Run("success sanitizes document and updates", func(t *testing.T) {
		repo := &mockOrganizationRepo{
			updateFn: func(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.Organization, error) {
				if *input.Document != "12345678000190" {
					t.Errorf("expected sanitized document without symbols, got %s", *input.Document)
				}
				return &domain.Organization{
					ID:                 id,
					Name:               input.Name,
					TradeName:          *input.TradeName,
					Document:           *input.Document,
					Phone:              *input.Phone,
					LogoURL:            *input.LogoURL,
					Plan:               "BASIC",
					SubscriptionStatus: domain.SubscriptionStatusActive,
					IsActive:           true,
				}, nil
			},
		}

		svc := service.NewOrganizationService(repo)
		res, err := svc.UpdateProfile(context.Background(), validOrgID, domain.UpdateOrganizationInput{
			Name:      "  Barbearia Oficial  ",
			TradeName: &validTradeName,
			Document:  &validDoc,
			Phone:     &validPhone,
			LogoURL:   &validLogo,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Name != "Barbearia Oficial" {
			t.Errorf("expected trimmed name, got %s", res.Name)
		}
	})

	t.Run("fails on short name", func(t *testing.T) {
		svc := service.NewOrganizationService(&mockOrganizationRepo{})
		_, err := svc.UpdateProfile(context.Background(), validOrgID, domain.UpdateOrganizationInput{
			Name: "A",
		})
		if !errors.Is(err, domain.ErrValidationFailed) {
			t.Errorf("expected ErrValidationFailed, got %v", err)
		}
	})

	t.Run("fails on invalid document length", func(t *testing.T) {
		svc := service.NewOrganizationService(&mockOrganizationRepo{})
		badDoc := "12345"
		_, err := svc.UpdateProfile(context.Background(), validOrgID, domain.UpdateOrganizationInput{
			Name:     "Barbearia Teste",
			Document: &badDoc,
		})
		if !errors.Is(err, domain.ErrInvalidDocument) {
			t.Errorf("expected ErrInvalidDocument, got %v", err)
		}
	})

	t.Run("fails on invalid logo url protocol", func(t *testing.T) {
		svc := service.NewOrganizationService(&mockOrganizationRepo{})
		badLogo := "ftp://example.com/logo.png"
		_, err := svc.UpdateProfile(context.Background(), validOrgID, domain.UpdateOrganizationInput{
			Name:    "Barbearia Teste",
			LogoURL: &badLogo,
		})
		if !errors.Is(err, domain.ErrInvalidLogoURL) {
			t.Errorf("expected ErrInvalidLogoURL, got %v", err)
		}
	})
}

