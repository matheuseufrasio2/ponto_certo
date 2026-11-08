package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockEmployeeLeaveRepo struct {
	getUserLeaveContextFn         func(ctx context.Context, userID string) (*domain.UserLeaveContext, error)
	hasConflictForUserFn          func(ctx context.Context, userID, startDate, endDate string) (bool, error)
	hasConflictWithAppointmentsFn func(ctx context.Context, barberID string, start, end time.Time) (bool, error)
	createFn                      func(ctx context.Context, leave *domain.EmployeeLeave, block *domain.BarberScheduleBlock, setVacation bool, barberID *string, audit *domain.AuditLog) error
	listByOrganizationFn          func(ctx context.Context, orgID string, unitID *string, status string) ([]domain.EmployeeLeave, error)
	getByIDFn                     func(ctx context.Context, leaveID string) (*domain.EmployeeLeave, error)
	deleteFn                      func(ctx context.Context, leaveID string, audit *domain.AuditLog) error
}

func (m *mockEmployeeLeaveRepo) GetUserLeaveContext(ctx context.Context, userID string) (*domain.UserLeaveContext, error) {
	if m.getUserLeaveContextFn != nil {
		return m.getUserLeaveContextFn(ctx, userID)
	}
	unit1 := "unit-1"
	return &domain.UserLeaveContext{
		UserName:       "Employee One",
		UserRole:       domain.RoleReceptionist,
		OrganizationID: "org-1",
		UnitID:         &unit1,
	}, nil
}

func (m *mockEmployeeLeaveRepo) HasConflictForUser(ctx context.Context, userID, startDate, endDate string) (bool, error) {
	if m.hasConflictForUserFn != nil {
		return m.hasConflictForUserFn(ctx, userID, startDate, endDate)
	}
	return false, nil
}

func (m *mockEmployeeLeaveRepo) HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error) {
	if m.hasConflictWithAppointmentsFn != nil {
		return m.hasConflictWithAppointmentsFn(ctx, barberID, start, end)
	}
	return false, nil
}

func (m *mockEmployeeLeaveRepo) Create(
	ctx context.Context,
	leave *domain.EmployeeLeave,
	block *domain.BarberScheduleBlock,
	setVacation bool,
	barberID *string,
	audit *domain.AuditLog,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, leave, block, setVacation, barberID, audit)
	}
	leave.ID = "leave-1"
	return nil
}

func (m *mockEmployeeLeaveRepo) ListByOrganization(ctx context.Context, orgID string, unitID *string, status string) ([]domain.EmployeeLeave, error) {
	if m.listByOrganizationFn != nil {
		return m.listByOrganizationFn(ctx, orgID, unitID, status)
	}
	return nil, nil
}

func (m *mockEmployeeLeaveRepo) GetByID(ctx context.Context, leaveID string) (*domain.EmployeeLeave, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, leaveID)
	}
	unit1 := "unit-1"
	return &domain.EmployeeLeave{
		ID:             leaveID,
		OrganizationID: "org-1",
		UnitID:         &unit1,
		UserID:         "user-1",
	}, nil
}

func (m *mockEmployeeLeaveRepo) Delete(ctx context.Context, leaveID string, audit *domain.AuditLog) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, leaveID, audit)
	}
	return nil
}

func TestEmployeeLeaveService_CreateLeave(t *testing.T) {
	ctx := context.Background()
	org1 := "org-1"
	org2 := "org-2"
	unit1 := "unit-1"
	ownerOrg1 := &service.UserClaims{UserID: "owner-1", Role: domain.RoleOrgOwner, OrganizationID: &org1}
	ownerOrg2 := &service.UserClaims{UserID: "owner-2", Role: domain.RoleOrgOwner, OrganizationID: &org2}

	validInput := domain.CreateEmployeeLeaveInput{
		UserID:    "user-1",
		LeaveType: domain.LeaveTypeVacation,
		StartDate: "2026-12-01",
		EndDate:   "2026-12-15",
		Reason:    "Annual vacation",
	}

	t.Run("rejects caller without claims", func(t *testing.T) {
		svc := service.NewEmployeeLeaveService(&mockEmployeeLeaveRepo{})
		_, err := svc.CreateLeave(ctx, validInput, nil)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("rejects owner from different organization", func(t *testing.T) {
		svc := service.NewEmployeeLeaveService(&mockEmployeeLeaveRepo{})
		_, err := svc.CreateLeave(ctx, validInput, ownerOrg2)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("rejects overlapping leave for user", func(t *testing.T) {
		repo := &mockEmployeeLeaveRepo{
			hasConflictForUserFn: func(ctx context.Context, userID, startDate, endDate string) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewEmployeeLeaveService(repo)
		_, err := svc.CreateLeave(ctx, validInput, ownerOrg1)
		if err != domain.ErrEmployeeLeaveConflict {
			t.Fatalf("expected ErrEmployeeLeaveConflict, got %v", err)
		}
	})

	t.Run("rejects barber vacation when appointment conflicts exist", func(t *testing.T) {
		barberID := "barber-1"
		repo := &mockEmployeeLeaveRepo{
			getUserLeaveContextFn: func(ctx context.Context, userID string) (*domain.UserLeaveContext, error) {
				return &domain.UserLeaveContext{
					UserName:       "Barber Bob",
					UserRole:       domain.RoleBarber,
					OrganizationID: org1,
					UnitID:         &unit1,
					BarberID:       &barberID,
				}, nil
			},
			hasConflictWithAppointmentsFn: func(ctx context.Context, bID string, start, end time.Time) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewEmployeeLeaveService(repo)
		_, err := svc.CreateLeave(ctx, validInput, ownerOrg1)
		if err != domain.ErrScheduleBlockConflict {
			t.Fatalf("expected ErrScheduleBlockConflict, got %v", err)
		}
	})

	t.Run("successfully creates employee leave and records audit log", func(t *testing.T) {
		var receivedAudit *domain.AuditLog
		repo := &mockEmployeeLeaveRepo{
			createFn: func(ctx context.Context, leave *domain.EmployeeLeave, block *domain.BarberScheduleBlock, setVacation bool, barberID *string, audit *domain.AuditLog) error {
				leave.ID = "leave-123"
				receivedAudit = audit
				return nil
			},
		}
		svc := service.NewEmployeeLeaveService(repo)
		leave, err := svc.CreateLeave(ctx, validInput, ownerOrg1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if leave.ID != "leave-123" {
			t.Errorf("expected leave ID leave-123, got %s", leave.ID)
		}
		if receivedAudit == nil || receivedAudit.Action != "EMPLOYEE_LEAVE_CREATED" {
			t.Errorf("expected audit action EMPLOYEE_LEAVE_CREATED, got %+v", receivedAudit)
		}
	})
}

func TestEmployeeLeaveService_CancelLeave(t *testing.T) {
	ctx := context.Background()
	org1 := "org-1"
	ownerOrg1 := &service.UserClaims{UserID: "owner-1", Role: domain.RoleOrgOwner, OrganizationID: &org1}

	t.Run("successfully cancels leave with audit log", func(t *testing.T) {
		var receivedAudit *domain.AuditLog
		repo := &mockEmployeeLeaveRepo{
			deleteFn: func(ctx context.Context, leaveID string, audit *domain.AuditLog) error {
				receivedAudit = audit
				return nil
			},
		}
		svc := service.NewEmployeeLeaveService(repo)
		err := svc.CancelLeave(ctx, "leave-1", ownerOrg1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedAudit == nil || receivedAudit.Action != "EMPLOYEE_LEAVE_CANCELLED" {
			t.Errorf("expected audit action EMPLOYEE_LEAVE_CANCELLED, got %+v", receivedAudit)
		}
	})
}
