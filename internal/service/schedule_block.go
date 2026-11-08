package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// ScheduleBlockRepository defines data access methods for managing barber agenda blocks.
type ScheduleBlockRepository interface {
	GetBarberInfo(ctx context.Context, barberID string) (unitID, orgID string, err error)
	HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error)
	Create(ctx context.Context, block *domain.BarberScheduleBlock, audit *domain.AuditLog) error
	ListByBarber(ctx context.Context, barberID string, from, to time.Time) ([]domain.BarberScheduleBlock, error)
	GetByID(ctx context.Context, blockID string) (*domain.BarberScheduleBlock, error)
	Delete(ctx context.Context, blockID string, audit *domain.AuditLog) error
}

// ScheduleBlockService enforces business rules and auditing for agenda blockages.
type ScheduleBlockService struct {
	repo ScheduleBlockRepository
}

// NewScheduleBlockService instantiates a new schedule block service.
func NewScheduleBlockService(repo ScheduleBlockRepository) *ScheduleBlockService {
	return &ScheduleBlockService{repo: repo}
}

// CreateBlock validates appointment conflicts and inserts a schedule block with audit logging.
func (s *ScheduleBlockService) CreateBlock(
	ctx context.Context,
	barberID string,
	input domain.CreateScheduleBlockInput,
	claims *UserClaims,
) (*domain.BarberScheduleBlock, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if barberID == "" {
		return nil, errors.New("barber id is required")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	unitID, orgID, err := s.repo.GetBarberInfo(ctx, barberID)
	if err != nil {
		return nil, err
	}

	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}

	hasConflict, err := s.repo.HasConflictWithAppointments(ctx, barberID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, domain.ErrScheduleBlockConflict
	}

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}

	details, _ := json.Marshal(map[string]any{
		"barber_id":  barberID,
		"start_time": input.StartTime.Format(time.RFC3339),
		"end_time":   input.EndTime.Format(time.RFC3339),
		"reason":     input.Reason,
	})

	audit := &domain.AuditLog{
		OrganizationID: &orgID,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         "SCHEDULE_BLOCK_CREATED",
		TargetEntity:   "barber_schedule_blocks",
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	block := &domain.BarberScheduleBlock{
		BarberID:        barberID,
		StartTime:       input.StartTime,
		EndTime:         input.EndTime,
		Reason:          input.Reason,
		CreatedByUserID: actorID,
	}

	if err := s.repo.Create(ctx, block, audit); err != nil {
		return nil, err
	}

	return block, nil
}

// ListBlocks retrieves schedule blocks for a barber within the caller's unit access scope.
func (s *ScheduleBlockService) ListBlocks(
	ctx context.Context,
	barberID string,
	from, to time.Time,
	claims *UserClaims,
) ([]domain.BarberScheduleBlock, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if barberID == "" {
		return nil, errors.New("barber id is required")
	}

	unitID, _, err := s.repo.GetBarberInfo(ctx, barberID)
	if err != nil {
		return nil, err
	}

	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}

	return s.repo.ListByBarber(ctx, barberID, from, to)
}

// DeleteBlock removes a schedule block after verifying unit access and records an audit log.
func (s *ScheduleBlockService) DeleteBlock(ctx context.Context, blockID string, claims *UserClaims) error {
	if claims == nil {
		return domain.ErrUnauthorized
	}
	if blockID == "" {
		return errors.New("block id is required")
	}

	block, err := s.repo.GetByID(ctx, blockID)
	if err != nil {
		return err
	}

	unitID, orgID, err := s.repo.GetBarberInfo(ctx, block.BarberID)
	if err != nil {
		return err
	}

	if !claims.HasUnitAccess(unitID) {
		return domain.ErrForbidden
	}

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}

	details, _ := json.Marshal(map[string]any{
		"barber_id":  block.BarberID,
		"start_time": block.StartTime.Format(time.RFC3339),
		"end_time":   block.EndTime.Format(time.RFC3339),
		"reason":     block.Reason,
	})

	audit := &domain.AuditLog{
		OrganizationID: &orgID,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         "SCHEDULE_BLOCK_DELETED",
		TargetEntity:   "barber_schedule_blocks",
		TargetID:       blockID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	return s.repo.Delete(ctx, blockID, audit)
}
