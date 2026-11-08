package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// DevRepository defines the consumer-driven interface for dev discovery.
type DevRepository interface {
	ListSuperAdmins(ctx context.Context) ([]domain.DevSuperAdmin, error)
	ListTenants(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error)
	ListTenantUsers(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error)
	ListUsersByRole(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error)
}

// DevService orchestrates business logic for developer debugging and auth presets.
type DevService struct {
	devRepo DevRepository
}

// NewDevService creates a new DevService instance.
func NewDevService(devRepo DevRepository) *DevService {
	return &DevService{devRepo: devRepo}
}

// GetSuperAdmins retrieves all super admin personas.
func (s *DevService) GetSuperAdmins(ctx context.Context) ([]domain.DevSuperAdmin, error) {
	return s.devRepo.ListSuperAdmins(ctx)
}

// GetTenants retrieves paginated tenants filtered by query and scenario.
func (s *DevService) GetTenants(ctx context.Context, search, scenario string, page, limit int) (*domain.DevTenantsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	tenants, total, err := s.devRepo.ListTenants(ctx, strings.TrimSpace(search), strings.TrimSpace(scenario), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("dev service list tenants: %w", err)
	}

	hasMore := (page * limit) < total

	return &domain.DevTenantsResponse{
		Tenants: tenants,
		Page:    page,
		Limit:   limit,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

// GetTenantUsers retrieves all users for a given tenant ID.
func (s *DevService) GetTenantUsers(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error) {
	if !domain.IsValidUUID(tenantID) {
		return nil, domain.ErrNotFound
	}

	users, err := s.devRepo.ListTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("dev service list tenant users: %w", err)
	}

	return users, nil
}

// GetUsersByRole retrieves paginated users filtered by role.
func (s *DevService) GetUsersByRole(ctx context.Context, role string, page, limit int) (*domain.DevUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	users, total, err := s.devRepo.ListUsersByRole(ctx, strings.TrimSpace(role), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("dev service list users by role: %w", err)
	}

	hasMore := (page * limit) < total

	return &domain.DevUsersResponse{
		Users:   users,
		Page:    page,
		Limit:   limit,
		Total:   total,
		HasMore: hasMore,
	}, nil
}
