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
}
