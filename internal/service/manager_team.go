package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// ManagerTeamRepository defines required data access methods for team management.
type ManagerTeamRepository interface {
	ListTeamMembers(ctx context.Context, unitID string) ([]domain.TeamMemberItem, error)
	CreateTeamMember(ctx context.Context, orgID, unitID string, input domain.CreateTeamMemberInput, passwordHash string, audit *domain.AuditLog) (*domain.TeamMemberItem, error)
	ResolveBarberID(ctx context.Context, unitID, identifier string) (string, error)
	GetBarberWorkingHours(ctx context.Context, barberID string) ([]domain.BarberWorkingHour, error)
	UpsertBarberWorkingHours(ctx context.Context, barberID string, hours []domain.UpdateWorkingHoursDayInput, audit *domain.AuditLog) error
}

// ManagerTeamService coordinates business logic for team administration and barber shifts.
type ManagerTeamService struct {
	repo          ManagerTeamRepository
	auditRecorder AuditLogRecorder
}

// NewManagerTeamService instantiates a new ManagerTeamService.
func NewManagerTeamService(repo ManagerTeamRepository) *ManagerTeamService {
	return &ManagerTeamService{repo: repo}
}

// SetAuditRecorder wires the audit log recorder.
func (s *ManagerTeamService) SetAuditRecorder(rec AuditLogRecorder) {
	s.auditRecorder = rec
}

// authorizeUnitManager ensures caller has authority over the requested branch unit.
func (s *ManagerTeamService) authorizeUnitManager(unitID string, claims *UserClaims) error {
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

// ListTeamMembers returns all staff members assigned to the branch unit.
func (s *ManagerTeamService) ListTeamMembers(ctx context.Context, unitID string, claims *UserClaims) ([]domain.TeamMemberItem, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}
	return s.repo.ListTeamMembers(ctx, unitID)
}

// CreateTeamMember registers a new barber or receptionist with login credentials and default shifts.
func (s *ManagerTeamService) CreateTeamMember(ctx context.Context, unitID string, input domain.CreateTeamMemberInput, claims *UserClaims) (*domain.TeamMemberItem, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash team member password: %w", err)
	}

	detailsMap := map[string]any{
		"name":  strings.TrimSpace(input.Name),
		"email": strings.TrimSpace(strings.ToLower(input.Email)),
		"role":  string(input.Role),
	}
	if input.Role == domain.TeamRoleBarber && input.CommissionPercentage != nil {
		detailsMap["commission_percentage"] = *input.CommissionPercentage
	}
	detailsJSON, _ := json.Marshal(detailsMap)

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	var orgID string
	var orgIDPtr *string
	var role domain.Role
	if claims != nil {
		if claims.UserID != "" {
			actorID = &claims.UserID
		}
		if claims.OrganizationID != nil {
			orgID = *claims.OrganizationID
			orgIDPtr = claims.OrganizationID
		}
		role = claims.Role
	}

	audit := &domain.AuditLog{
		OrganizationID: orgIDPtr,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      role,
		Action:         "TEAM_MEMBER_CREATED",
		TargetEntity:   "users",
		Details:        detailsJSON,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	return s.repo.CreateTeamMember(ctx, orgID, unitID, input, string(hash), audit)
}

// GetWorkingHours fetches the weekly shift schedule for the specified barber.
func (s *ManagerTeamService) GetWorkingHours(ctx context.Context, unitID, identifier string, claims *UserClaims) ([]domain.BarberWorkingHour, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}

	barberID, err := s.repo.ResolveBarberID(ctx, unitID, identifier)
	if err != nil {
		return nil, err
	}

	return s.repo.GetBarberWorkingHours(ctx, barberID)
}

// UpdateWorkingHours updates weekly shifts for a barber.
func (s *ManagerTeamService) UpdateWorkingHours(ctx context.Context, unitID, identifier string, input domain.UpdateWorkingHoursInput, claims *UserClaims) ([]domain.BarberWorkingHour, error) {
	if err := s.authorizeUnitManager(unitID, claims); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	barberID, err := s.repo.ResolveBarberID(ctx, unitID, identifier)
	if err != nil {
		return nil, err
	}

	hours := input.Normalize()

	detailsJSON, _ := json.Marshal(map[string]any{
		"barber_id": barberID,
		"days":      len(hours),
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
		Action:         "BARBER_WORKING_HOURS_UPDATED",
		TargetEntity:   "barbers",
		TargetID:       barberID,
		Details:        detailsJSON,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	if err := s.repo.UpsertBarberWorkingHours(ctx, barberID, hours, audit); err != nil {
		return nil, err
	}

	return s.repo.GetBarberWorkingHours(ctx, barberID)
}
