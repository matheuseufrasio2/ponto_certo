package service

import (
	"context"
	"fmt"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// CatalogResponse contains lists of active barbers and services for a given unit.
type CatalogResponse struct {
	Barbers  []domain.Barber  `json:"barbers"`
	Services []domain.Service `json:"services"`
}

// CatalogService provides business logic for catalog lookups.
type CatalogService struct {
	catalogRepo repository.CatalogRepository
}

// NewCatalogService instantiates CatalogService.
func NewCatalogService(catalogRepo repository.CatalogRepository) *CatalogService {
	return &CatalogService{catalogRepo: catalogRepo}
}

// GetUnitCatalog fetches active barbers and services for a unit.
func (s *CatalogService) GetUnitCatalog(ctx context.Context, unitID string) (*CatalogResponse, error) {
	if unitID == "" {
		return nil, domain.ErrUnitNotFound
	}

	barbers, err := s.catalogRepo.ListBarbersByUnit(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("list barbers: %w", err)
	}

	services, err := s.catalogRepo.ListServicesByUnit(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	return &CatalogResponse{
		Barbers:  barbers,
		Services: services,
	}, nil
}
