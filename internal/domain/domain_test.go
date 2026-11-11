package domain_test

import (
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     domain.Role
		expected bool
	}{
		{name: "super admin", role: domain.RoleSuperAdmin, expected: true},
		{name: "org owner", role: domain.RoleOrgOwner, expected: true},
		{name: "unit manager", role: domain.RoleUnitManager, expected: true},
		{name: "receptionist", role: domain.RoleReceptionist, expected: true},
		{name: "barber", role: domain.RoleBarber, expected: true},
		{name: "invalid role", role: domain.Role("INVALID"), expected: false},
		{name: "empty role", role: domain.Role(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.expected {
				t.Errorf("Role.IsValid() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestPaymentMethod_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		method   domain.PaymentMethod
		expected bool
	}{
		{name: "pix", method: domain.PaymentMethodPix, expected: true},
		{name: "credit card", method: domain.PaymentMethodCreditCard, expected: true},
		{name: "debit card", method: domain.PaymentMethodDebitCard, expected: true},
		{name: "cash", method: domain.PaymentMethodCash, expected: true},
		{name: "other", method: domain.PaymentMethodOther, expected: true},
		{name: "card legacy", method: domain.PaymentMethodCard, expected: true},
		{name: "invalid method", method: domain.PaymentMethod("CRYPTO"), expected: false},
		{name: "empty method", method: domain.PaymentMethod(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.IsValid(); got != tt.expected {
				t.Errorf("PaymentMethod.IsValid() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestProductCategory_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cat      domain.ProductCategory
		expected bool
	}{
		{name: "kits", cat: domain.ProductCategoryKits, expected: true},
		{name: "shampoo higiene", cat: domain.ProductCategoryShampooHigiene, expected: true},
		{name: "finalizadores", cat: domain.ProductCategoryFinalizadores, expected: true},
		{name: "barbearia cuidados", cat: domain.ProductCategoryBarbeariaCuidados, expected: true},
		{name: "acessorios", cat: domain.ProductCategoryAcessorios, expected: true},
		{name: "skincare", cat: domain.ProductCategorySkincare, expected: true},
		{name: "bebidas merch", cat: domain.ProductCategoryBebidasMerch, expected: true},
		{name: "invalid", cat: domain.ProductCategory("FOOD"), expected: false},
		{name: "empty", cat: domain.ProductCategory(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cat.IsValid(); got != tt.expected {
				t.Errorf("ProductCategory.IsValid() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestServiceCategory_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cat      domain.ServiceCategory
		expected bool
	}{
		{name: "cabelo", cat: domain.ServiceCategoryCabelo, expected: true},
		{name: "barba", cat: domain.ServiceCategoryBarba, expected: true},
		{name: "cabelo e barba", cat: domain.ServiceCategoryCabeloEBarba, expected: true},
		{name: "combos", cat: domain.ServiceCategoryCombos, expected: true},
		{name: "tratamentos", cat: domain.ServiceCategoryTratamentos, expected: true},
		{name: "cuidados especiais", cat: domain.ServiceCategoryCuidadosEspeciais, expected: true},
		{name: "quimicas", cat: domain.ServiceCategoryQuimicas, expected: true},
		{name: "visagismo", cat: domain.ServiceCategoryVisagismo, expected: true},
		{name: "estetica facial", cat: domain.ServiceCategoryEsteticaFacial, expected: true},
		{name: "invalid", cat: domain.ServiceCategory("TATTOO"), expected: false},
		{name: "empty", cat: domain.ServiceCategory(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cat.IsValid(); got != tt.expected {
				t.Errorf("ServiceCategory.IsValid() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestProduct_Validate(t *testing.T) {
	validProduct := domain.Product{
		ID:            "prod-1",
		UnitID:        "unit-1",
		Name:          "Pomada Modeladora Efeito Matte",
		Category:      domain.ProductCategoryFinalizadores,
		PriceInCents:  4500,
		CostInCents:   2000,
		StockQuantity: 10,
		IsActive:      true,
	}

	t.Run("valid product", func(t *testing.T) {
		if err := validProduct.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty name fails", func(t *testing.T) {
		p := validProduct
		p.Name = ""
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("invalid category fails", func(t *testing.T) {
		p := validProduct
		p.Category = "UNKNOWN"
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid category, got nil")
		}
	})

	t.Run("negative price fails", func(t *testing.T) {
		p := validProduct
		p.PriceInCents = -100
		if err := p.Validate(); err == nil {
			t.Error("expected error for negative price, got nil")
		}
	})

	t.Run("negative cost fails", func(t *testing.T) {
		p := validProduct
		p.CostInCents = -50
		if err := p.Validate(); err == nil {
			t.Error("expected error for negative cost, got nil")
		}
	})

	t.Run("negative stock fails", func(t *testing.T) {
		p := validProduct
		p.StockQuantity = -1
		if err := p.Validate(); err == nil {
			t.Error("expected error for negative stock, got nil")
		}
	})
}

func TestAppointment_Validate(t *testing.T) {
	now := time.Now().UTC()
	validAppt := domain.Appointment{
		ID:            "appt-1",
		UnitID:        "unit-1",
		BarberID:      "barber-1",
		ClientID:      "client-1",
		ServiceID:     "service-1",
		StartTime:     now,
		EndTime:       now.Add(30 * time.Minute),
		Status:        domain.StatusScheduled,
		PaymentStatus: domain.PaymentStatusPending,
		PriceInCents:  6000,
	}

	t.Run("valid appointment", func(t *testing.T) {
		if err := validAppt.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing unit_id fails", func(t *testing.T) {
		a := validAppt
		a.UnitID = ""
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing unit_id, got nil")
		}
	})

	t.Run("missing barber_id fails", func(t *testing.T) {
		a := validAppt
		a.BarberID = ""
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing barber_id, got nil")
		}
	})

	t.Run("missing client_id fails", func(t *testing.T) {
		a := validAppt
		a.ClientID = ""
		a.CustomerID = ""
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing client_id, got nil")
		}
	})

	t.Run("missing service_id fails", func(t *testing.T) {
		a := validAppt
		a.ServiceID = ""
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing service_id, got nil")
		}
	})

	t.Run("start time not before end time fails", func(t *testing.T) {
		a := validAppt
		a.EndTime = a.StartTime
		if err := a.Validate(); err == nil {
			t.Error("expected error when start == end, got nil")
		}

		a.EndTime = a.StartTime.Add(-10 * time.Minute)
		if err := a.Validate(); err == nil {
			t.Error("expected error when start > end, got nil")
		}
	})

	t.Run("negative price fails", func(t *testing.T) {
		a := validAppt
		a.PriceInCents = -1
		if err := a.Validate(); err == nil {
			t.Error("expected error for negative price, got nil")
		}
	})
}

func TestAppointment_CanTransitionTo(t *testing.T) {
	appt := domain.Appointment{
		Status: domain.StatusScheduled,
	}

	// Valid transitions from SCHEDULED
	if !appt.CanTransitionTo(domain.StatusCompleted) {
		t.Error("expected SCHEDULED -> COMPLETED to be allowed")
	}
	if !appt.CanTransitionTo(domain.StatusCancelled) {
		t.Error("expected SCHEDULED -> CANCELLED to be allowed")
	}
	if !appt.CanTransitionTo(domain.StatusNoShow) {
		t.Error("expected SCHEDULED -> NO_SHOW to be allowed")
	}
	if appt.CanTransitionTo(domain.StatusScheduled) {
		t.Error("expected SCHEDULED -> SCHEDULED to be disallowed")
	}

	// Terminal states cannot transition
	completedAppt := domain.Appointment{Status: domain.StatusCompleted}
	if completedAppt.CanTransitionTo(domain.StatusScheduled) ||
		completedAppt.CanTransitionTo(domain.StatusCancelled) ||
		completedAppt.CanTransitionTo(domain.StatusNoShow) {
		t.Error("expected COMPLETED to be terminal")
	}

	cancelledAppt := domain.Appointment{Status: domain.StatusCancelled}
	if cancelledAppt.CanTransitionTo(domain.StatusCompleted) {
		t.Error("expected CANCELLED to be terminal")
	}

	noShowAppt := domain.Appointment{Status: domain.StatusNoShow}
	if noShowAppt.CanTransitionTo(domain.StatusCompleted) {
		t.Error("expected NO_SHOW to be terminal")
	}
}

func TestHasScheduleConflict(t *testing.T) {
	base := time.Date(2026, 11, 10, 14, 0, 0, 0, time.UTC)
	end := base.Add(30 * time.Minute) // 14:00 to 14:30

	tests := []struct {
		name      string
		startB    time.Time
		endB      time.Time
		conflicts bool
	}{
		{
			name:      "exact overlap",
			startB:    base,
			endB:      end,
			conflicts: true,
		},
		{
			name:      "starts before and ends during",
			startB:    base.Add(-15 * time.Minute), // 13:45
			endB:      base.Add(15 * time.Minute),  // 14:15
			conflicts: true,
		},
		{
			name:      "starts during and ends after",
			startB:    base.Add(15 * time.Minute), // 14:15
			endB:      base.Add(45 * time.Minute), // 14:45
			conflicts: true,
		},
		{
			name:      "enclosing interval",
			startB:    base.Add(-15 * time.Minute), // 13:45
			endB:      base.Add(45 * time.Minute),  // 14:45
			conflicts: true,
		},
		{
			name:      "completely inside interval",
			startB:    base.Add(5 * time.Minute),  // 14:05
			endB:      base.Add(25 * time.Minute), // 14:25
			conflicts: true,
		},
		{
			name:      "immediately adjacent before (no conflict)",
			startB:    base.Add(-30 * time.Minute), // 13:30
			endB:      base,                        // 14:00
			conflicts: false,
		},
		{
			name:      "immediately adjacent after (no conflict)",
			startB:    end,                       // 14:30
			endB:      end.Add(30 * time.Minute), // 15:00
			conflicts: false,
		},
		{
			name:      "completely before (no conflict)",
			startB:    base.Add(-60 * time.Minute), // 13:00
			endB:      base.Add(-30 * time.Minute), // 13:30
			conflicts: false,
		},
		{
			name:      "completely after (no conflict)",
			startB:    end.Add(30 * time.Minute), // 15:00
			endB:      end.Add(60 * time.Minute), // 15:30
			conflicts: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.HasScheduleConflict(base, end, tt.startB, tt.endB)
			if got != tt.conflicts {
				t.Errorf("HasScheduleConflict() = %v, expected %v", got, tt.conflicts)
			}
		})
	}
}

func TestOrganizationAndUnit_Instantiation(t *testing.T) {
	now := time.Now().UTC()

	org := domain.Organization{
		ID:        "org-1",
		Name:      "Barbearia Ponto Certo",
		TradeName: "Ponto Certo Matriz",
		Document:  "12.345.678/0001-90",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if org.Name != "Barbearia Ponto Certo" || !org.IsActive {
		t.Errorf("unexpected organization values: %+v", org)
	}

	unit := domain.BarbershopUnit{
		ID:             "unit-1",
		OrganizationID: org.ID,
		Name:           "Unidade Centro",
		Phone:          "11999999999",
		Address:        "Rua Central, 100",
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if unit.OrganizationID != org.ID || unit.Name != "Unidade Centro" {
		t.Errorf("unexpected unit values: %+v", unit)
	}

	opHours := domain.UnitOperatingHour{
		ID:        "uoh-1",
		UnitID:    unit.ID,
		DayOfWeek: 1, // Monday
		OpenTime:  "10:00:00",
		CloseTime: "22:00:00",
		IsClosed:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if opHours.DayOfWeek != 1 || opHours.OpenTime != "10:00:00" || opHours.CloseTime != "22:00:00" {
		t.Errorf("unexpected operating hours values: %+v", opHours)
	}
}

func TestUserAndBarber_Link(t *testing.T) {
	now := time.Now().UTC()
	orgID := "org-1"
	unitID := "unit-1"

	user := domain.User{
		ID:             "user-1",
		OrganizationID: &orgID,
		UnitID:         &unitID,
		Name:           "Pedro Barbeiro",
		Email:          "pedro@barbearia.com",
		Role:           domain.RoleBarber,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if user.Role != domain.RoleBarber || *user.UnitID != unitID {
		t.Errorf("unexpected user values: %+v", user)
	}

	barber := domain.Barber{
		ID:                   "barber-1",
		UnitID:               unitID,
		UserID:               &user.ID,
		Name:                 user.Name,
		CommissionPercentage: 60,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if barber.UserID == nil || *barber.UserID != user.ID || barber.CommissionPercentage != 60 {
		t.Errorf("unexpected barber values: %+v", barber)
	}

	breakStart := "13:00:00"
	breakEnd := "14:00:00"
	workingHour := domain.BarberWorkingHour{
		ID:             "bwh-1",
		BarberID:       barber.ID,
		DayOfWeek:      2, // Tuesday
		StartTime:      "10:00:00",
		EndTime:        "19:00:00",
		BreakStartTime: &breakStart,
		BreakEndTime:   &breakEnd,
		IsOff:          false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if workingHour.DayOfWeek != 2 || *workingHour.BreakStartTime != "13:00:00" {
		t.Errorf("unexpected working hour values: %+v", workingHour)
	}
}

func TestBarberStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.BarberStatus
		expected bool
	}{
		{name: "available", status: domain.BarberStatusAvailable, expected: true},
		{name: "busy", status: domain.BarberStatusBusy, expected: true},
		{name: "lunch", status: domain.BarberStatusLunch, expected: true},
		{name: "off duty", status: domain.BarberStatusOffDuty, expected: true},
		{name: "vacation", status: domain.BarberStatusVacation, expected: true},
		{name: "invalid status", status: domain.BarberStatus("SLEEPING"), expected: false},
		{name: "empty status", status: domain.BarberStatus(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("BarberStatus.IsValid() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestUpdateBarberStatusInput_Validate(t *testing.T) {
	valid := domain.UpdateBarberStatusInput{Status: domain.BarberStatusLunch}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected valid input, got error: %v", err)
	}

	invalid := domain.UpdateBarberStatusInput{Status: domain.BarberStatus("INVALID")}
	if err := invalid.Validate(); err != domain.ErrInvalidBarberStatus {
		t.Errorf("expected ErrInvalidBarberStatus, got: %v", err)
	}
}

func TestUpdateTenantPlanInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       domain.UpdateTenantPlanInput
		expectError bool
	}{
		{
			name:        "valid basic plan",
			input:       domain.UpdateTenantPlanInput{Plan: "BASIC", MonthlyFeeInCents: 19900},
			expectError: false,
		},
		{
			name:        "valid pro plan",
			input:       domain.UpdateTenantPlanInput{Plan: "PRO", MonthlyFeeInCents: 29900},
			expectError: false,
		},
		{
			name:        "valid enterprise plan",
			input:       domain.UpdateTenantPlanInput{Plan: "ENTERPRISE", MonthlyFeeInCents: 59900},
			expectError: false,
		},
		{
			name:        "invalid plan name",
			input:       domain.UpdateTenantPlanInput{Plan: "GOLD", MonthlyFeeInCents: 19900},
			expectError: true,
		},
		{
			name:        "negative fee fails",
			input:       domain.UpdateTenantPlanInput{Plan: "PRO", MonthlyFeeInCents: -100},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLeaveType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		lt       domain.LeaveType
		expected bool
	}{
		{"vacation", domain.LeaveTypeVacation, true},
		{"medical_leave", domain.LeaveTypeMedicalLeave, true},
		{"personal", domain.LeaveTypePersonal, true},
		{"other", domain.LeaveTypeOther, true},
		{"invalid", domain.LeaveType("UNPAID"), false},
		{"empty", domain.LeaveType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lt.IsValid(); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCreateScheduleBlockInput_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		input       domain.CreateScheduleBlockInput
		expectError bool
	}{
		{
			name: "valid schedule block",
			input: domain.CreateScheduleBlockInput{
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
				Reason:    "Dentist appointment",
			},
			expectError: false,
		},
		{
			name: "zero start time fails",
			input: domain.CreateScheduleBlockInput{
				EndTime: now.Add(2 * time.Hour),
				Reason:  "Doctor",
			},
			expectError: true,
		},
		{
			name: "start time after end time fails",
			input: domain.CreateScheduleBlockInput{
				StartTime: now.Add(2 * time.Hour),
				EndTime:   now.Add(time.Hour),
				Reason:    "Doctor",
			},
			expectError: true,
		},
		{
			name: "empty reason fails",
			input: domain.CreateScheduleBlockInput{
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
				Reason:    "  ",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreateEmployeeLeaveInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       domain.CreateEmployeeLeaveInput
		expectError bool
	}{
		{
			name: "valid employee leave",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				LeaveType: domain.LeaveTypeVacation,
				StartDate: "2026-12-01",
				EndDate:   "2026-12-20",
				Reason:    "End of year vacation",
			},
			expectError: false,
		},
		{
			name: "empty user_id fails",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "",
				LeaveType: domain.LeaveTypeVacation,
				StartDate: "2026-12-01",
				EndDate:   "2026-12-20",
				Reason:    "Vacation",
			},
			expectError: true,
		},
		{
			name: "invalid leave_type fails",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				LeaveType: domain.LeaveType("INVALID"),
				StartDate: "2026-12-01",
				EndDate:   "2026-12-20",
				Reason:    "Vacation",
			},
			expectError: true,
		},
		{
			name: "invalid start_date format fails",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				LeaveType: domain.LeaveTypeVacation,
				StartDate: "01/12/2026",
				EndDate:   "2026-12-20",
				Reason:    "Vacation",
			},
			expectError: true,
		},
		{
			name: "start_date after end_date fails",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				LeaveType: domain.LeaveTypeVacation,
				StartDate: "2026-12-25",
				EndDate:   "2026-12-10",
				Reason:    "Vacation",
			},
			expectError: true,
		},
		{
			name: "empty reason fails",
			input: domain.CreateEmployeeLeaveInput{
				UserID:    "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				LeaveType: domain.LeaveTypeVacation,
				StartDate: "2026-12-01",
				EndDate:   "2026-12-20",
				Reason:    "  ",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnitDomainErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected domain.ErrorCode
	}{
		{name: "unit not found", err: domain.ErrUnitNotFound, expected: domain.ErrCodeUnitNotFound},
		{name: "unit name required", err: domain.ErrUnitNameRequired, expected: domain.ErrCodeUnitNameRequired},
		{name: "invalid chair count", err: domain.ErrInvalidChairCount, expected: domain.ErrCodeInvalidChairCount},
		{name: "unit name already exists", err: domain.ErrUnitNameAlreadyExists, expected: domain.ErrCodeUnitNameAlreadyExists},
		{name: "unit inactive", err: domain.ErrUnitInactive, expected: domain.ErrCodeUnitInactive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.CodeForError(tc.err); got != tc.expected {
				t.Errorf("expected error code %q, got %q", tc.expected, got)
			}
		})
	}
}
