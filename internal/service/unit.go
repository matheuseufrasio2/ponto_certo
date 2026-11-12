package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

const (
	defaultUnitChairCount = 4
	minUnitNameRunes      = 2
	maxUnitNameRunes      = 120
	maxUnitPhoneRunes     = 20
	maxUnitAddressRunes   = 255
)

// UnitRepository defines data persistence operations required by UnitService.
type UnitRepository interface {
	ListByOrg(ctx context.Context, orgID string, status string) ([]domain.BarbershopUnit, error)
	ExistsByNameInOrg(ctx context.Context, orgID string, name string, excludeUnitID string) (bool, error)
	Create(ctx context.Context, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error)
	Update(ctx context.Context, unitID string, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error)
	ToggleStatus(ctx context.Context, unitID string, orgID string, isActive bool) (*domain.BarbershopUnit, error)
}

// UnitService orchestrates barbershop branch lifecycle management with multi-tenant scoping and auditing.
type UnitService struct {
	repo          UnitRepository
	auditRecorder AuditLogRecorder
}

// NewUnitService instantiates a UnitService.
func NewUnitService(repo UnitRepository) *UnitService {
	return &UnitService{repo: repo}
}

// SetAuditRecorder assigns the audit log recorder for unit governance events.
func (s *UnitService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

// ListUnits returns branches belonging to the caller's organization, optionally filtered by status.
func (s *UnitService) ListUnits(ctx context.Context, claims *UserClaims, status string) ([]domain.BarbershopUnit, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return nil, err
	}

	normStatus := strings.ToLower(strings.TrimSpace(status))
	switch normStatus {
	case "", "all", "active", "inactive":
	default:
		return nil, fmt.Errorf("%w: invalid status filter, expected active, inactive, or all", domain.ErrValidationFailed)
	}

	return s.repo.ListByOrg(ctx, orgID, normStatus)
}

// CreateUnit validates input fields, ensures name uniqueness within the organization, and persists a new branch.
func (s *UnitService) CreateUnit(ctx context.Context, claims *UserClaims, input domain.CreateUnitInput) (*domain.BarbershopUnit, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return nil, err
	}

	name, phone, address, err := normalizeAndValidateUnitFields(input.Name, input.Phone, input.Address)
	if err != nil {
		return nil, err
	}

	chairCount := defaultUnitChairCount
	if input.ChairCount != nil {
		if *input.ChairCount <= 0 {
			return nil, domain.ErrInvalidChairCount
		}
		chairCount = *input.ChairCount
	}

	exists, err := s.repo.ExistsByNameInOrg(ctx, orgID, name, "")
	if err != nil {
		return nil, fmt.Errorf("verify unit name uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrUnitNameAlreadyExists
	}

	unit, err := s.repo.Create(ctx, orgID, name, phone, address, chairCount)
	if err != nil {
		return nil, err
	}

	s.recordUnitAudit(ctx, claims, orgID, unit.ID, "UNIT_CREATED", map[string]any{
		"name":        unit.Name,
		"chair_count": unit.ChairCount,
		"is_active":   unit.IsActive,
	})

	return unit, nil
}

// UpdateUnit validates input fields, verifies uniqueness excluding the target unit, and persists updates.
func (s *UnitService) UpdateUnit(ctx context.Context, claims *UserClaims, unitID string, input domain.UpdateUnitInput) (*domain.BarbershopUnit, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return nil, err
	}
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	name, phone, address, err := normalizeAndValidateUnitFields(input.Name, input.Phone, input.Address)
	if err != nil {
		return nil, err
	}

	if input.ChairCount <= 0 {
		return nil, domain.ErrInvalidChairCount
	}

	exists, err := s.repo.ExistsByNameInOrg(ctx, orgID, name, unitID)
	if err != nil {
		return nil, fmt.Errorf("verify unit name uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrUnitNameAlreadyExists
	}

	unit, err := s.repo.Update(ctx, unitID, orgID, name, phone, address, input.ChairCount)
	if err != nil {
		return nil, err
	}

	s.recordUnitAudit(ctx, claims, orgID, unit.ID, "UNIT_UPDATED", map[string]any{
		"name":        unit.Name,
		"chair_count": unit.ChairCount,
	})

	return unit, nil
}

// ToggleUnitStatus activates or pauses a branch within the caller's organization.
func (s *UnitService) ToggleUnitStatus(ctx context.Context, claims *UserClaims, unitID string, input domain.ToggleUnitStatusInput) (*domain.BarbershopUnit, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return nil, err
	}
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	unit, err := s.repo.ToggleStatus(ctx, unitID, orgID, input.IsActive)
	if err != nil {
		return nil, err
	}

	s.recordUnitAudit(ctx, claims, orgID, unit.ID, "UNIT_STATUS_TOGGLED", map[string]any{
		"name":      unit.Name,
		"is_active": unit.IsActive,
	})

	return unit, nil
}

func resolveOrgIDFromClaims(claims *UserClaims) (string, error) {
	if claims == nil {
		return "", domain.ErrUnauthorized
	}
	if claims.OrganizationID == nil || *claims.OrganizationID == "" {
		return "", domain.ErrForbidden
	}
	if !domain.IsValidUUID(*claims.OrganizationID) {
		return "", domain.ErrInvalidUUID
	}
	return *claims.OrganizationID, nil
}

func normalizeAndValidateUnitFields(rawName string, rawPhone *string, rawAddress *string) (string, *string, *string, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		return "", nil, nil, domain.ErrUnitNameRequired
	}
	nameLen := utf8.RuneCountInString(name)
	if nameLen < minUnitNameRunes || nameLen > maxUnitNameRunes {
		return "", nil, nil, fmt.Errorf("%w: unit name must be between 2 and 120 characters", domain.ErrValidationFailed)
	}

	var phone *string
	if rawPhone != nil {
		trimmed := strings.TrimSpace(*rawPhone)
		if utf8.RuneCountInString(trimmed) > maxUnitPhoneRunes {
			return "", nil, nil, fmt.Errorf("%w: phone must be at most 20 characters", domain.ErrValidationFailed)
		}
		phone = &trimmed
	}

	var address *string
	if rawAddress != nil {
		trimmed := strings.TrimSpace(*rawAddress)
		if utf8.RuneCountInString(trimmed) > maxUnitAddressRunes {
			return "", nil, nil, fmt.Errorf("%w: address must be at most 255 characters", domain.ErrValidationFailed)
		}
		address = &trimmed
	}

	return name, phone, address, nil
}

func (s *UnitService) recordUnitAudit(ctx context.Context, claims *UserClaims, orgID, unitID, action string, detailsMap map[string]any) {
	if s.auditRecorder == nil {
		return
	}
	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims != nil && claims.UserID != "" {
		actorID = &claims.UserID
	}
	var actorRole domain.Role
	if claims != nil {
		actorRole = claims.Role
	}
	details, _ := json.Marshal(detailsMap)
	_ = s.auditRecorder.Record(ctx, domain.AuditLog{
		OrganizationID: &orgID,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      actorRole,
		Action:         action,
		TargetEntity:   "barbershop_units",
		TargetID:       unitID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	})
}
