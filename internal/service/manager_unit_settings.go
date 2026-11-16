package service

import (
	"context"
	"encoding/json"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// ManagerUnitSettingsRepository defines data access methods for branch operating rules.
type ManagerUnitSettingsRepository interface {
	GetUnitSettings(ctx context.Context, unitID string) (*domain.UnitSettingsResponse, error)
	UpdateUnitSettings(ctx context.Context, unitID string, input domain.UpdateUnitSettingsInput, audit *domain.AuditLog) (*domain.UnitSettingsResponse, error)
	GetUnitOperatingHours(ctx context.Context, unitID string) ([]domain.UnitOperatingHour, error)
}

// ManagerUnitSettingsService coordinates business logic for unit rules and schedule governance.
type ManagerUnitSettingsService struct {
	repo          ManagerUnitSettingsRepository
	auditRecorder AuditLogRecorder
}

// NewManagerUnitSettingsService instantiates a new ManagerUnitSettingsService.
func NewManagerUnitSettingsService(repo ManagerUnitSettingsRepository) *ManagerUnitSettingsService {
	return &ManagerUnitSettingsService{repo: repo}
}

// SetAuditRecorder wires the audit log recorder.
func (s *ManagerUnitSettingsService) SetAuditRecorder(rec AuditLogRecorder) {
	s.auditRecorder = rec
}

// authorizeUnitManager ensures caller has authority over the requested branch unit.
func (s *ManagerUnitSettingsService) authorizeUnitManager(unitID string, claims *UserClaims) error {
	if claims == nil {
		return domain.ErrUnauthorized
	}

	switch claims.Role {
	case domain.RoleSuperAdmin, domain.RoleOrgOwner:
		return nil
	case domain.RoleUnitManager:
		if claims.UnitID == nil || *claims.UnitID != unitID {
			return domain.ErrForbidden
		}
		return nil
	default:
		return domain.ErrForbidden
	}
}

// GetUnitSettings fetches contact details and the 7-day schedule for the branch.
func (s *ManagerUnitSettingsService) GetUnitSettings(ctx context.Context, unitID string, claims *UserClaims) (*domain.UnitSettingsResponse, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}
	return s.repo.GetUnitSettings(ctx, unitID)
}

// UpdateUnitSettings saves contact updates and the 7-day schedule for the branch.
func (s *ManagerUnitSettingsService) UpdateUnitSettings(ctx context.Context, unitID string, input domain.UpdateUnitSettingsInput, claims *UserClaims) (*domain.UnitSettingsResponse, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	hours := input.NormalizeHours()
	detailsJSON, _ := json.Marshal(map[string]any{
		"unit_id": unitID,
		"days":    len(hours),
	})

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	var orgIDPtr *string
	var role domain.Role
	if claims != nil {
		if claims.UserID != "" {
			actorID = &claims.UserID
		}
		orgIDPtr = claims.OrganizationID
		role = claims.Role
	}

	audit := &domain.AuditLog{
		OrganizationID: orgIDPtr,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      role,
		Action:         "UNIT_SETTINGS_UPDATED",
		TargetEntity:   "barbershop_units",
		TargetID:       unitID,
		Details:        detailsJSON,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	return s.repo.UpdateUnitSettings(ctx, unitID, input, audit)
}
