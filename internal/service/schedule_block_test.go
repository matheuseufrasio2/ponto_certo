package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockScheduleBlockRepo struct {
	getBarberInfoFn               func(ctx context.Context, barberID string) (unitID, orgID string, err error)
	hasConflictWithAppointmentsFn func(ctx context.Context, barberID string, start, end time.Time) (bool, error)
	createFn                      func(ctx context.Context, block *domain.BarberScheduleBlock, audit *domain.AuditLog) error
	listByBarberFn                func(ctx context.Context, barberID string, from, to time.Time) ([]domain.BarberScheduleBlock, error)
	getByIDFn                     func(ctx context.Context, blockID string) (*domain.BarberScheduleBlock, error)
	deleteFn                      func(ctx context.Context, blockID string, audit *domain.AuditLog) error
}

func (m *mockScheduleBlockRepo) GetBarberInfo(ctx context.Context, barberID string) (unitID, orgID string, err error) {
	if m.getBarberInfoFn != nil {
		return m.getBarberInfoFn(ctx, barberID)
	}
	return "unit-1", "org-1", nil
}

func (m *mockScheduleBlockRepo) HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error) {
	if m.hasConflictWithAppointmentsFn != nil {
		return m.hasConflictWithAppointmentsFn(ctx, barberID, start, end)
	}
	return false, nil
}

func (m *mockScheduleBlockRepo) Create(ctx context.Context, block *domain.BarberScheduleBlock, audit *domain.AuditLog) error {
	if m.createFn != nil {
		return m.createFn(ctx, block, audit)
	}
	block.ID = "block-1"
	return nil
}

func (m *mockScheduleBlockRepo) ListByBarber(ctx context.Context, barberID string, from, to time.Time) ([]domain.BarberScheduleBlock, error) {
	if m.listByBarberFn != nil {
		return m.listByBarberFn(ctx, barberID, from, to)
	}
	return nil, nil
}

func (m *mockScheduleBlockRepo) GetByID(ctx context.Context, blockID string) (*domain.BarberScheduleBlock, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, blockID)
	}
	return &domain.BarberScheduleBlock{ID: blockID, BarberID: "b-1", Reason: "Doctor"}, nil
}

func (m *mockScheduleBlockRepo) Delete(ctx context.Context, blockID string, audit *domain.AuditLog) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, blockID, audit)
	}
	return nil
}

func TestScheduleBlockService_CreateBlock(t *testing.T) {
	ctx := context.Background()
	unit1 := "unit-1"
	unit2 := "unit-2"
	mgrUnit1 := &service.UserClaims{UserID: "mgr-1", Role: domain.RoleUnitManager, UnitID: &unit1}
	now := time.Now()

	t.Run("rejects caller without claims", func(t *testing.T) {
		svc := service.NewScheduleBlockService(&mockScheduleBlockRepo{})
		_, err := svc.CreateBlock(ctx, "b-1", domain.CreateScheduleBlockInput{}, nil)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects caller lacking unit access", func(t *testing.T) {
		repo := &mockScheduleBlockRepo{
			getBarberInfoFn: func(ctx context.Context, barberID string) (unitID, orgID string, err error) {
				return unit2, "org-1", nil
			},
		}
		svc := service.NewScheduleBlockService(repo)
		input := domain.CreateScheduleBlockInput{
			StartTime: now.Add(time.Hour),
			EndTime:   now.Add(2 * time.Hour),
			Reason:    "Doctor",
		}
		_, err := svc.CreateBlock(ctx, "b-1", input, mgrUnit1)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("rejects block with conflicting appointments", func(t *testing.T) {
		repo := &mockScheduleBlockRepo{
			hasConflictWithAppointmentsFn: func(ctx context.Context, barberID string, start, end time.Time) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewScheduleBlockService(repo)
		input := domain.CreateScheduleBlockInput{
			StartTime: now.Add(time.Hour),
			EndTime:   now.Add(2 * time.Hour),
			Reason:    "Doctor",
		}
		_, err := svc.CreateBlock(ctx, "b-1", input, mgrUnit1)
		if err != domain.ErrScheduleBlockConflict {
			t.Fatalf("expected ErrScheduleBlockConflict, got %v", err)
		}
	})

	t.Run("successfully creates block with audit log", func(t *testing.T) {
		var receivedAudit *domain.AuditLog
		repo := &mockScheduleBlockRepo{
			createFn: func(ctx context.Context, block *domain.BarberScheduleBlock, audit *domain.AuditLog) error {
				block.ID = "block-10"
				receivedAudit = audit
				return nil
			},
		}
		svc := service.NewScheduleBlockService(repo)
		input := domain.CreateScheduleBlockInput{
			StartTime: now.Add(time.Hour),
			EndTime:   now.Add(2 * time.Hour),
			Reason:    "Doctor appointment",
		}
		block, err := svc.CreateBlock(ctx, "b-1", input, mgrUnit1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if block.ID != "block-10" {
			t.Errorf("expected block ID block-10, got %s", block.ID)
		}
		if receivedAudit == nil || receivedAudit.Action != "SCHEDULE_BLOCK_CREATED" {
			t.Errorf("expected audit action SCHEDULE_BLOCK_CREATED, got %+v", receivedAudit)
		}
	})
}

func TestScheduleBlockService_DeleteBlock(t *testing.T) {
	ctx := context.Background()
	unit1 := "unit-1"
	mgrUnit1 := &service.UserClaims{UserID: "mgr-1", Role: domain.RoleUnitManager, UnitID: &unit1}

	t.Run("successfully deletes block with audit log", func(t *testing.T) {
		var receivedAudit *domain.AuditLog
		repo := &mockScheduleBlockRepo{
			deleteFn: func(ctx context.Context, blockID string, audit *domain.AuditLog) error {
				receivedAudit = audit
				return nil
			},
		}
		svc := service.NewScheduleBlockService(repo)
		err := svc.DeleteBlock(ctx, "block-1", mgrUnit1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedAudit == nil || receivedAudit.Action != "SCHEDULE_BLOCK_DELETED" {
			t.Errorf("expected audit action SCHEDULE_BLOCK_DELETED, got %+v", receivedAudit)
		}
	})
}
