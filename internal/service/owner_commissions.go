package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// OwnerCommissionsRepository defines data access operations for multi-scope commission rules.
type OwnerCommissionsRepository interface {
	UnitBelongsToOrg(ctx context.Context, unitID, orgID string) (bool, error)
	BarberBelongsToOrg(ctx context.Context, barberID, orgID string) (bool, error)
	UpdateAllOrganization(ctx context.Context, orgID string, percentage int) (int, error)
	UpdateByUnit(ctx context.Context, orgID, unitID string, percentage int) (int, error)
	UpdateByBarber(ctx context.Context, orgID, barberID string, percentage int) (int, error)
}

// OwnerCommissionsService orchestrates hierarchical barber commission updates for owners.
type OwnerCommissionsService struct {
	repo          OwnerCommissionsRepository
	auditRecorder AuditLogRecorder
}

// NewOwnerCommissionsService constructs an OwnerCommissionsService instance.
func NewOwnerCommissionsService(repo OwnerCommissionsRepository) *OwnerCommissionsService {
	return &OwnerCommissionsService{repo: repo}
}

// SetAuditRecorder attaches an audit log recorder to OwnerCommissionsService.
func (s *OwnerCommissionsService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

// UpdateCommissions applies percentage updates to ALL_ORGANIZATION, UNIT, or BARBER scope.
func (s *OwnerCommissionsService) UpdateCommissions(
	ctx context.Context,
	claims *UserClaims,
	input domain.UpdateCommissionsInput,
) (*domain.UpdateCommissionsResponse, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return nil, err
	}
	if claims.Role != domain.RoleOrgOwner && claims.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	var affected int
	var unitIDPtr *string
	targetID := orgID

	switch input.Scope {
	case domain.CommissionScopeAllOrganization:
		affected, err = s.repo.UpdateAllOrganization(ctx, orgID, input.Percentage)
		if err != nil {
			return nil, err
		}

	case domain.CommissionScopeUnit:
		unitID := strings.TrimSpace(input.UnitID)
		belongs, checkErr := s.repo.UnitBelongsToOrg(ctx, unitID, orgID)
		if checkErr != nil {
			return nil, checkErr
		}
		if !belongs {
			return nil, domain.ErrUnitNotFound
		}
		affected, err = s.repo.UpdateByUnit(ctx, orgID, unitID, input.Percentage)
		if err != nil {
			return nil, err
		}
		unitIDPtr = &unitID
		targetID = unitID

	case domain.CommissionScopeBarber:
		barberID := strings.TrimSpace(input.BarberID)
		belongs, checkErr := s.repo.BarberBelongsToOrg(ctx, barberID, orgID)
		if checkErr != nil {
			return nil, checkErr
		}
		if !belongs {
			return nil, domain.ErrBarberNotInOrg
		}
		affected, err = s.repo.UpdateByBarber(ctx, orgID, barberID, input.Percentage)
		if err != nil {
			return nil, err
		}
		targetID = barberID
	}

	s.recordCommissionAudit(ctx, claims, orgID, unitIDPtr, targetID, input, affected)

	return &domain.UpdateCommissionsResponse{
		Scope:           input.Scope,
		Percentage:      input.Percentage,
		BarbersAffected: affected,
	}, nil
}

func (s *OwnerCommissionsService) recordCommissionAudit(
	ctx context.Context,
	claims *UserClaims,
	orgID string,
	unitID *string,
	targetID string,
	input domain.UpdateCommissionsInput,
	affected int,
) {
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

	detailsMap := map[string]any{
		"scope":            input.Scope,
		"percentage":       input.Percentage,
		"barbers_affected": affected,
	}
	if trimmedUnit := strings.TrimSpace(input.UnitID); trimmedUnit != "" {
		detailsMap["unit_id"] = trimmedUnit
	}
	if trimmedBarber := strings.TrimSpace(input.BarberID); trimmedBarber != "" {
		detailsMap["barber_id"] = trimmedBarber
	}
	details, _ := json.Marshal(detailsMap)

	_ = s.auditRecorder.Record(ctx, domain.AuditLog{
		OrganizationID: &orgID,
		UnitID:         unitID,
		ActorID:        actorID,
		ActorRole:      actorRole,
		Action:         "COMMISSIONS_UPDATED",
		TargetEntity:   "barbers",
		TargetID:       targetID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	})
}
