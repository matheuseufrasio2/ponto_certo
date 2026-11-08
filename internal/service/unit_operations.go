package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// HasUnitAccess checks whether the claim bearer has authority over the target unit.
func (c *UserClaims) HasUnitAccess(targetUnitID string) bool {
	if c == nil {
		return false
	}
	if c.Role == domain.RoleSuperAdmin || c.Role == domain.RoleOrgOwner {
		return true
	}
	return c.UnitID != nil && *c.UnitID == targetUnitID
}

// UnitOperationsService coordinates branch management operations with strict RBAC.
type UnitOperationsService struct {
	repo          repository.UnitOperationsRepository
	auditRecorder AuditLogRecorder
}

// NewUnitOperationsService instantiates a unit operations service.
func NewUnitOperationsService(repo repository.UnitOperationsRepository) *UnitOperationsService {
	return &UnitOperationsService{repo: repo}
}

// SetAuditRecorder assigns the audit log recorder.
func (s *UnitOperationsService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

// GetManagerMetrics retrieves live branch KPIs, verifying unit authorization.
func (s *UnitOperationsService) GetManagerMetrics(ctx context.Context, unitID string, date time.Time, claims *UserClaims) (*domain.ManagerDashboardMetrics, error) {
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.GetManagerLiveMetrics(ctx, unitID, date)
}

// GetDailyCashClosing returns aggregated financial closing data by payment method.
func (s *UnitOperationsService) GetDailyCashClosing(ctx context.Context, unitID string, date time.Time, claims *UserClaims) (*domain.CashClosingSummary, error) {
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.GetDailyCashClosing(ctx, unitID, date)
}

// UpdateBarberStatus validates input, verifies that off-duty/vacation shift has no future appointments, and updates status.
func (s *UnitOperationsService) UpdateBarberStatus(ctx context.Context, unitID, barberID string, input domain.UpdateBarberStatusInput, claims *UserClaims) error {
	if !claims.HasUnitAccess(unitID) {
		return domain.ErrForbidden
	}

	if err := input.Validate(); err != nil {
		return err
	}

	// Invariant: cannot shift to OFF_DUTY or VACATION if barber has future scheduled appointments
	if input.Status == domain.BarberStatusOffDuty || input.Status == domain.BarberStatusVacation {
		futureCount, err := s.repo.CountFutureAppointmentsForBarber(ctx, barberID, time.Now())
		if err != nil {
			return fmt.Errorf("verify future appointments: %w", err)
		}
		if futureCount > 0 {
			return domain.ErrBarberHasFutureAppointments
		}
	}

	var changedBy *string
	if claims != nil && claims.UserID != "" {
		changedBy = &claims.UserID
	}

	if err := s.repo.UpdateBarberStatus(ctx, barberID, input.Status, changedBy); err != nil {
		return err
	}

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		details, _ := json.Marshal(map[string]any{
			"new_status": input.Status,
		})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			UnitID:       &unitID,
			ActorID:      changedBy,
			ActorRole:    claims.Role,
			Action:       "BARBER_STATUS_CHANGE",
			TargetEntity: "barbers",
			TargetID:     barberID,
			Details:      details,
			IPAddress:    meta.IPAddress,
			UserAgent:    meta.UserAgent,
		})
	}

	return nil
}

// ToggleService toggles availability of a service for a branch.
func (s *UnitOperationsService) ToggleService(ctx context.Context, unitID, serviceID string, input domain.ToggleServiceInput, claims *UserClaims) error {
	if !claims.HasUnitAccess(unitID) {
		return domain.ErrForbidden
	}
	if err := s.repo.ToggleServiceUnitAvailability(ctx, unitID, serviceID, input.IsActive); err != nil {
		return err
	}

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		var actorID *string
		if claims != nil && claims.UserID != "" {
			actorID = &claims.UserID
		}
		details, _ := json.Marshal(map[string]any{
			"is_active": input.IsActive,
		})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			UnitID:       &unitID,
			ActorID:      actorID,
			ActorRole:    claims.Role,
			Action:       "TOGGLE_SERVICE_UNIT",
			TargetEntity: "services",
			TargetID:     serviceID,
			Details:      details,
			IPAddress:    meta.IPAddress,
			UserAgent:    meta.UserAgent,
		})
	}

	return nil
}

// ListBarbers retrieves barbers for a branch, optionally filtered by status (active, inactive, all).
func (s *UnitOperationsService) ListBarbers(ctx context.Context, unitID string, status string, claims *UserClaims) ([]domain.Barber, error) {
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListBarbersByUnit(ctx, unitID, status)
}
