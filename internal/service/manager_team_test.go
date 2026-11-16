package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

type mockManagerTeamRepo struct {
	members      []domain.TeamMemberItem
	workingHours []domain.BarberWorkingHour
	createdItem  *domain.TeamMemberItem
	createErr    error
	resolveID    string
	resolveErr   error
	upsertErr    error
}

func (m *mockManagerTeamRepo) ListTeamMembers(ctx context.Context, unitID string) ([]domain.TeamMemberItem, error) {
	return m.members, nil
}

func (m *mockManagerTeamRepo) CreateTeamMember(ctx context.Context, orgID, unitID string, input domain.CreateTeamMemberInput, passwordHash string, audit *domain.AuditLog) (*domain.TeamMemberItem, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createdItem != nil {
		return m.createdItem, nil
	}
	barberID := "barber-1111"
	return &domain.TeamMemberItem{
		ID:        "user-1111",
		BarberID:  &barberID,
		UnitID:    unitID,
		Name:      input.Name,
		Email:     input.Email,
		Phone:     input.Phone,
		Role:      domain.Role(input.Role),
		IsActive:  true,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockManagerTeamRepo) ResolveBarberID(ctx context.Context, unitID, identifier string) (string, error) {
	if m.resolveErr != nil {
		return "", m.resolveErr
	}
	if m.resolveID != "" {
		return m.resolveID, nil
	}
	return identifier, nil
}

func (m *mockManagerTeamRepo) GetBarberWorkingHours(ctx context.Context, barberID string) ([]domain.BarberWorkingHour, error) {
	return m.workingHours, nil
}

func (m *mockManagerTeamRepo) UpsertBarberWorkingHours(ctx context.Context, barberID string, hours []domain.UpdateWorkingHoursDayInput, audit *domain.AuditLog) error {
	return m.upsertErr
}

func TestManagerTeamService_Authorization(t *testing.T) {
	repo := &mockManagerTeamRepo{}
	svc := NewManagerTeamService(repo)
	unitID := "unit-1111"
	otherUnit := "unit-2222"
	orgID := "org-1"

	t.Run("nil claims returns ErrUnauthorized", func(t *testing.T) {
		_, err := svc.ListTeamMembers(context.Background(), unitID, nil)
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("barber role returns ErrForbidden", func(t *testing.T) {
		claims := &UserClaims{Role: domain.RoleBarber, UnitID: &unitID}
		_, err := svc.ListTeamMembers(context.Background(), unitID, claims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("manager of different unit returns ErrForbidden", func(t *testing.T) {
		claims := &UserClaims{Role: domain.RoleUnitManager, UnitID: &otherUnit}
		_, err := svc.ListTeamMembers(context.Background(), unitID, claims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("super admin is authorized", func(t *testing.T) {
		claims := &UserClaims{Role: domain.RoleSuperAdmin}
		_, err := svc.ListTeamMembers(context.Background(), unitID, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("org owner is authorized", func(t *testing.T) {
		claims := &UserClaims{Role: domain.RoleOrgOwner, OrganizationID: &orgID}
		_, err := svc.ListTeamMembers(context.Background(), unitID, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("matching unit manager is authorized", func(t *testing.T) {
		claims := &UserClaims{Role: domain.RoleUnitManager, UnitID: &unitID}
		_, err := svc.ListTeamMembers(context.Background(), unitID, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestManagerTeamService_CreateTeamMember(t *testing.T) {
	unitID := "unit-1111"
	orgID := "org-1"
	claims := &UserClaims{
		Role:           domain.RoleUnitManager,
		UnitID:         &unitID,
		UserID:         "manager-1",
		OrganizationID: &orgID,
	}

	t.Run("valid input creates team member", func(t *testing.T) {
		repo := &mockManagerTeamRepo{}
		svc := NewManagerTeamService(repo)
		comm := 60
		input := domain.CreateTeamMemberInput{
			Name:                 "Carlos Barbeiro",
			Email:                "carlos@barbearia.com",
			Phone:                "11977778888",
			Password:             "senhaForte123",
			Role:                 domain.TeamRoleBarber,
			CommissionPercentage: &comm,
		}

		item, err := svc.CreateTeamMember(context.Background(), unitID, input, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Name != input.Name || item.Email != input.Email {
			t.Fatalf("unexpected item data: %+v", item)
		}
	})

	t.Run("invalid input fails validation", func(t *testing.T) {
		repo := &mockManagerTeamRepo{}
		svc := NewManagerTeamService(repo)
		input := domain.CreateTeamMemberInput{
			Name:     "",
			Password: "123",
		}
		_, err := svc.CreateTeamMember(context.Background(), unitID, input, claims)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("duplicate email propagates error", func(t *testing.T) {
		repo := &mockManagerTeamRepo{createErr: domain.ErrManagerEmailTaken}
		svc := NewManagerTeamService(repo)
		input := domain.CreateTeamMemberInput{
			Name:     "Carlos Barbeiro",
			Email:    "carlos@barbearia.com",
			Phone:    "11977778888",
			Password: "senhaForte123",
			Role:     domain.TeamRoleReceptionist,
		}
		_, err := svc.CreateTeamMember(context.Background(), unitID, input, claims)
		if !errors.Is(err, domain.ErrManagerEmailTaken) {
			t.Fatalf("expected ErrManagerEmailTaken, got %v", err)
		}
	})
}

func TestManagerTeamService_WorkingHours(t *testing.T) {
	unitID := "unit-1111"
	claims := &UserClaims{Role: domain.RoleUnitManager, UnitID: &unitID}

	t.Run("GetWorkingHours resolves barber and returns schedule", func(t *testing.T) {
		repo := &mockManagerTeamRepo{
			resolveID: "barber-real-id",
			workingHours: []domain.BarberWorkingHour{
				{DayOfWeek: 1, StartTime: "09:00", EndTime: "19:00"},
			},
		}
		svc := NewManagerTeamService(repo)
		hours, err := svc.GetWorkingHours(context.Background(), unitID, "user-or-barber-id", claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hours) != 1 || hours[0].StartTime != "09:00" {
			t.Fatalf("unexpected hours returned: %+v", hours)
		}
	})

	t.Run("UpdateWorkingHours validates and upserts schedule", func(t *testing.T) {
		repo := &mockManagerTeamRepo{
			resolveID: "barber-real-id",
		}
		svc := NewManagerTeamService(repo)

		var schedule []domain.UpdateWorkingHoursDayInput
		for day := 0; day < 7; day++ {
			schedule = append(schedule, domain.UpdateWorkingHoursDayInput{
				DayOfWeek: day,
				IsOff:     day == 0,
				StartTime: "09:00",
				EndTime:   "19:00",
			})
		}
		input := domain.UpdateWorkingHoursInput{Schedule: schedule}

		_, err := svc.UpdateWorkingHours(context.Background(), unitID, "barber-real-id", input, claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
