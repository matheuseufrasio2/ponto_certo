package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// OwnerAnalyticsService handles executive analytical governance and report exports.
type OwnerAnalyticsService struct {
	repo repository.OwnerAnalyticsRepository
}

// NewOwnerAnalyticsService instantiates a new owner analytics service.
func NewOwnerAnalyticsService(repo repository.OwnerAnalyticsRepository) *OwnerAnalyticsService {
	return &OwnerAnalyticsService{repo: repo}
}

// resolveOrgID extracts and validates the target organization ID based on user claims.
func (s *OwnerAnalyticsService) resolveOrgID(claims *UserClaims, targetOrgID *string) (string, error) {
	if claims == nil {
		return "", domain.ErrUnauthorized
	}

	if claims.Role == domain.RoleSuperAdmin {
		if targetOrgID != nil && *targetOrgID != "" {
			return *targetOrgID, nil
		}
		if claims.OrganizationID != nil && *claims.OrganizationID != "" {
			return *claims.OrganizationID, nil
		}
		return "", fmt.Errorf("organization_id is required for super admin: %w", domain.ErrForbidden)
	}

	if claims.Role == domain.RoleOrgOwner {
		if claims.OrganizationID == nil || *claims.OrganizationID == "" {
			return "", domain.ErrForbidden
		}
		return *claims.OrganizationID, nil
	}

	return "", domain.ErrForbidden
}

// GetExecutiveKPIs computes consolidated financial and operational KPIs for the organization or branch.
func (s *OwnerAnalyticsService) GetExecutiveKPIs(ctx context.Context, claims *UserClaims, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
	orgID, err := s.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}

	if unitID != nil && *unitID != "" && !claims.HasUnitAccess(*unitID) {
		return nil, domain.ErrForbidden
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date cannot be before start date: %w", domain.ErrInvalidTimeSlot)
	}

	return s.repo.GetExecutiveKPIs(ctx, orgID, unitID, start, end)
}

// GetRevenueComparisonByUnit returns revenue comparisons across all branches in the organization.
func (s *OwnerAnalyticsService) GetRevenueComparisonByUnit(ctx context.Context, claims *UserClaims, start, end time.Time) ([]domain.UnitRevenueComparison, error) {
	orgID, err := s.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date cannot be before start date: %w", domain.ErrInvalidTimeSlot)
	}

	return s.repo.GetRevenueComparisonByUnit(ctx, orgID, start, end)
}

// GetServicesMix ranks top-performing services across the organization or for a specific branch.
func (s *OwnerAnalyticsService) GetServicesMix(ctx context.Context, claims *UserClaims, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error) {
	orgID, err := s.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}

	if unitID != nil && *unitID != "" && !claims.HasUnitAccess(*unitID) {
		return nil, domain.ErrForbidden
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date cannot be before start date: %w", domain.ErrInvalidTimeSlot)
	}

	return s.repo.GetServicesMix(ctx, orgID, unitID, start, end)
}

// ListConsolidatedBillings retrieves paginated billing transactions.
func (s *OwnerAnalyticsService) ListConsolidatedBillings(ctx context.Context, claims *UserClaims, unitID *string, paymentMethod *string, page, perPage int) (*domain.PaginatedBillings, error) {
	orgID, err := s.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}

	if unitID != nil && *unitID != "" && !claims.HasUnitAccess(*unitID) {
		return nil, domain.ErrForbidden
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	items, total, err := s.repo.ListConsolidatedBillings(ctx, orgID, unitID, paymentMethod, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedBillings{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// ExportBillingsCSV streams billing records formatted as CSV to the provided writer.
func (s *OwnerAnalyticsService) ExportBillingsCSV(ctx context.Context, claims *UserClaims, unitID *string, w io.Writer) error {
	orgID, err := s.resolveOrgID(claims, nil)
	if err != nil {
		return err
	}

	if unitID != nil && *unitID != "" && !claims.HasUnitAccess(*unitID) {
		return domain.ErrForbidden
	}

	rows, err := s.repo.GetExportBillingsData(ctx, orgID, unitID)
	if err != nil {
		return fmt.Errorf("fetch export data: %w", err)
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	header := []string{"ID", "Data", "Unidade", "Barbeiro", "Cliente", "Serviço", "Valor (R$)", "Forma de Pagamento", "Status", "Criado Em"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, r := range rows {
		amountFormatted := fmt.Sprintf("%.2f", float64(r.AmountInCents)/100.0)
		record := []string{
			r.ID,
			r.Date,
			r.UnitName,
			r.BarberName,
			r.ClientName,
			r.ServiceName,
			amountFormatted,
			r.PaymentMethod,
			r.Status,
			r.CreatedAt,
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
