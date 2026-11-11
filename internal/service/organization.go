package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

var nonDigitsRegex = regexp.MustCompile(`\D`)

// OrganizationService manages organization profile operations.
type OrganizationService struct {
	orgRepo repository.OrganizationRepository
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(orgRepo repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo}
}

// GetProfile retrieves the corporate profile of an organization.
func (s *OrganizationService) GetProfile(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error) {
	if orgID == "" || !domain.IsValidUUID(orgID) {
		return nil, domain.ErrInvalidUUID
	}

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return toProfileResponse(org), nil
}

// UpdateProfile validates and updates corporate profile details.
func (s *OrganizationService) UpdateProfile(ctx context.Context, orgID string, input domain.UpdateOrganizationInput) (*domain.OrganizationProfileResponse, error) {
	if orgID == "" || !domain.IsValidUUID(orgID) {
		return nil, domain.ErrInvalidUUID
	}

	input.Name = strings.TrimSpace(input.Name)
	if len(input.Name) < 2 || len(input.Name) > 120 {
		return nil, fmt.Errorf("%w: organization name must be between 2 and 120 characters", domain.ErrValidationFailed)
	}

	if input.TradeName != nil {
		trimmed := strings.TrimSpace(*input.TradeName)
		input.TradeName = &trimmed
	}

	if input.Document != nil {
		cleanDoc := nonDigitsRegex.ReplaceAllString(*input.Document, "")
		if cleanDoc != "" && len(cleanDoc) != 11 && len(cleanDoc) != 14 {
			return nil, domain.ErrInvalidDocument
		}
		input.Document = &cleanDoc
	}

	if input.Phone != nil {
		trimmedPhone := strings.TrimSpace(*input.Phone)
		input.Phone = &trimmedPhone
	}

	if input.LogoURL != nil {
		trimmedLogo := strings.TrimSpace(*input.LogoURL)
		if trimmedLogo != "" {
			lower := strings.ToLower(trimmedLogo)
			if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
				return nil, domain.ErrInvalidLogoURL
			}
		}
		input.LogoURL = &trimmedLogo
	}

	updatedOrg, err := s.orgRepo.Update(ctx, orgID, input)
	if err != nil {
		return nil, err
	}

	return toProfileResponse(updatedOrg), nil
}

func toProfileResponse(org *domain.Organization) *domain.OrganizationProfileResponse {
	return &domain.OrganizationProfileResponse{
		ID:                 org.ID,
		Name:               org.Name,
		TradeName:          org.TradeName,
		Document:           org.Document,
		Phone:              org.Phone,
		LogoURL:            org.LogoURL,
		Plan:               org.Plan,
		SubscriptionStatus: org.SubscriptionStatus,
		IsActive:           org.IsActive,
		CreatedAt:          org.CreatedAt,
		UpdatedAt:          org.UpdatedAt,
	}
}

