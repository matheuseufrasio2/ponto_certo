package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// EmployeeLeaveRepository defines data access methods for employee vacations and leaves.
type EmployeeLeaveRepository interface {
	GetUserLeaveContext(ctx context.Context, userID string) (*domain.UserLeaveContext, error)
	HasConflictForUser(ctx context.Context, userID, startDate, endDate string) (bool, error)
	HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error)
	Create(
		ctx context.Context,
		leave *domain.EmployeeLeave,
		block *domain.BarberScheduleBlock,
		setVacation bool,
		barberID *string,
		audit *domain.AuditLog,
	) error
	ListByOrganization(ctx context.Context, orgID string, unitID *string, status string) ([]domain.EmployeeLeave, error)
	GetByID(ctx context.Context, leaveID string) (*domain.EmployeeLeave, error)
	Delete(ctx context.Context, leaveID string, audit *domain.AuditLog) error
}

// EmployeeLeaveService orchestrates employee leaves across roles with auditing and barber agenda synchronization.
type EmployeeLeaveService struct {
	repo EmployeeLeaveRepository
}

// NewEmployeeLeaveService instantiates a new employee leave service.
func NewEmployeeLeaveService(repo EmployeeLeaveRepository) *EmployeeLeaveService {
	return &EmployeeLeaveService{repo: repo}
}

// CreateLeave validates permissions, conflicts, and cascades vacations to barber schedules with audit logging.
func (s *EmployeeLeaveService) CreateLeave(
	ctx context.Context,
	input domain.CreateEmployeeLeaveInput,
	claims *UserClaims,
) (*domain.EmployeeLeave, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	userCtx, err := s.repo.GetUserLeaveContext(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	switch claims.Role {
	case domain.RoleSuperAdmin:
		// SuperAdmin can manage leaves across all tenants
	case domain.RoleOrgOwner:
		if claims.OrganizationID == nil || *claims.OrganizationID != userCtx.OrganizationID {
			return nil, domain.ErrForbidden
		}
	case domain.RoleUnitManager:
		if claims.OrganizationID == nil || *claims.OrganizationID != userCtx.OrganizationID {
			return nil, domain.ErrForbidden
		}
		if claims.UnitID == nil || userCtx.UnitID == nil || *claims.UnitID != *userCtx.UnitID {
			return nil, domain.ErrForbidden
		}
	default:
		return nil, domain.ErrForbidden
	}

	// Check if employee already has an active or scheduled leave overlapping
	hasConflict, err := s.repo.HasConflictForUser(ctx, input.UserID, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, domain.ErrEmployeeLeaveConflict
	}

	startT, _ := time.Parse("2006-01-02", input.StartDate)
	endT, _ := time.Parse("2006-01-02", input.EndDate)
	startUTC := time.Date(startT.Year(), startT.Month(), startT.Day(), 0, 0, 0, 0, time.UTC)
	endUTC := time.Date(endT.Year(), endT.Month(), endT.Day(), 23, 59, 59, 999999999, time.UTC)

	// If employee is a barber, check appointment conflicts across the entire vacation duration
	var scheduleBlock *domain.BarberScheduleBlock
	setVacation := false
	if userCtx.BarberID != nil && *userCtx.BarberID != "" {
		hasApptConflict, err := s.repo.HasConflictWithAppointments(ctx, *userCtx.BarberID, startUTC, endUTC)
		if err != nil {
			return nil, err
		}
		if hasApptConflict {
			return nil, domain.ErrScheduleBlockConflict
		}

		nowDate := time.Now().UTC().Format("2006-01-02")
		if input.StartDate <= nowDate && nowDate <= input.EndDate {
			setVacation = true
		}

		var createdBy *string
		if claims.UserID != "" {
			createdBy = &claims.UserID
		}

		scheduleBlock = &domain.BarberScheduleBlock{
			BarberID:        *userCtx.BarberID,
			StartTime:       startUTC,
			EndTime:         endUTC,
			Reason:          "Vacation / Leave: " + input.Reason,
			CreatedByUserID: createdBy,
		}
	}

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}

	details, _ := json.Marshal(map[string]any{
		"target_user_id":   input.UserID,
		"target_user_name": userCtx.UserName,
		"target_role":      userCtx.UserRole,
		"leave_type":       input.LeaveType,
		"start_date":       input.StartDate,
		"end_date":         input.EndDate,
		"reason":           input.Reason,
		"is_barber":        userCtx.BarberID != nil,
	})

	audit := &domain.AuditLog{
		OrganizationID: &userCtx.OrganizationID,
		UnitID:         userCtx.UnitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         "EMPLOYEE_LEAVE_CREATED",
		TargetEntity:   "employee_leaves",
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	leave := &domain.EmployeeLeave{
		OrganizationID:  userCtx.OrganizationID,
		UnitID:          userCtx.UnitID,
		UserID:          input.UserID,
		UserName:        userCtx.UserName,
		UserRole:        userCtx.UserRole,
		LeaveType:       input.LeaveType,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Reason:          input.Reason,
		CreatedByUserID: actorID,
	}

	if err := s.repo.Create(ctx, leave, scheduleBlock, setVacation, userCtx.BarberID, audit); err != nil {
		return nil, err
	}

	return leave, nil
}

// ListLeaves lists leaves for the caller's organization.
func (s *EmployeeLeaveService) ListLeaves(
	ctx context.Context,
	claims *UserClaims,
	unitID *string,
	status string,
) ([]domain.EmployeeLeave, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}

	var orgID string
	if claims.Role == domain.RoleSuperAdmin {
		if claims.OrganizationID != nil {
			orgID = *claims.OrganizationID
		}
	} else if claims.Role == domain.RoleOrgOwner || claims.Role == domain.RoleUnitManager {
		if claims.OrganizationID == nil || *claims.OrganizationID == "" {
			return nil, domain.ErrForbidden
		}
		orgID = *claims.OrganizationID
	} else {
		return nil, domain.ErrForbidden
	}

	// Managers can only list their unit
	if claims.Role == domain.RoleUnitManager {
		unitID = claims.UnitID
	}

	return s.repo.ListByOrganization(ctx, orgID, unitID, status)
}

// CancelLeave deletes an employee leave and records an audit log.
func (s *EmployeeLeaveService) CancelLeave(ctx context.Context, leaveID string, claims *UserClaims) error {
	if claims == nil {
		return domain.ErrUnauthorized
	}
	if leaveID == "" {
		return errors.New("leave id is required")
	}

	leave, err := s.repo.GetByID(ctx, leaveID)
	if err != nil {
		return err
	}

	switch claims.Role {
	case domain.RoleSuperAdmin:
	case domain.RoleOrgOwner:
		if claims.OrganizationID == nil || *claims.OrganizationID != leave.OrganizationID {
			return domain.ErrForbidden
		}
	case domain.RoleUnitManager:
		if claims.OrganizationID == nil || *claims.OrganizationID != leave.OrganizationID {
			return domain.ErrForbidden
		}
		if claims.UnitID == nil || leave.UnitID == nil || *claims.UnitID != *leave.UnitID {
			return domain.ErrForbidden
		}
	default:
		return domain.ErrForbidden
	}

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}

	details, _ := json.Marshal(map[string]any{
		"target_user_id": leave.UserID,
		"start_date":     leave.StartDate,
		"end_date":       leave.EndDate,
		"reason":         leave.Reason,
	})

	audit := &domain.AuditLog{
		OrganizationID: &leave.OrganizationID,
		UnitID:         leave.UnitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         "EMPLOYEE_LEAVE_CANCELLED",
		TargetEntity:   "employee_leaves",
		TargetID:       leaveID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	return s.repo.Delete(ctx, leaveID, audit)
}
