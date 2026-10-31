package domain

import (
	"errors"
	"time"
)

// Common domain errors for the barbershop system
var (
	ErrNotFound             = errors.New("resource not found")
	ErrBarberNotFound       = errors.New("barber not found")
	ErrCustomerNotFound     = errors.New("customer not found")
	ErrClientNotFound       = errors.New("client not found")
	ErrServiceNotFound      = errors.New("service not found")
	ErrBillingNotFound      = errors.New("billing record not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUnitNotFound         = errors.New("unit not found")
	ErrAppointmentConflict  = errors.New("barber already has an appointment scheduled at this time")
	ErrInvalidTimeSlot      = errors.New("invalid appointment time slot")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrInvalidBillingStatus = errors.New("invalid billing status")
	ErrInvalidRole          = errors.New("invalid user role")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("access forbidden")
)

// Role defines system actor roles for RBAC.
type Role string

const (
	RoleSuperAdmin   Role = "SUPER_ADMIN"
	RoleOrgOwner     Role = "ORG_OWNER"
	RoleUnitManager  Role = "UNIT_MANAGER"
	RoleReceptionist Role = "RECEPTIONIST"
	RoleBarber       Role = "BARBER"
)

// IsValid validates whether the role matches accepted RBAC taxonomy.
func (r Role) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleOrgOwner, RoleUnitManager, RoleReceptionist, RoleBarber:
		return true
	default:
		return false
	}
}

// Organization represents the top-level tenant (brand, franchise or single barbershop).
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TradeName string    `json:"trade_name,omitempty"`
	Document  string    `json:"document,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BarbershopUnit represents a physical branch or location belonging to an organization.
type BarbershopUnit struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Phone          string    `json:"phone,omitempty"`
	Address        string    `json:"address,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// User represents an authenticated system actor with specific RBAC role.
type User struct {
	ID             string    `json:"id"`
	OrganizationID *string   `json:"organization_id,omitempty"` // NULL for SUPER_ADMIN
	UnitID         *string   `json:"unit_id,omitempty"`         // NULL for SUPER_ADMIN & ORG_OWNER
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	Role           Role      `json:"role"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Barber represents a barber professional linked to an optional user login.
type Barber struct {
	ID                   string    `json:"id"`
	UnitID               string    `json:"unit_id,omitempty"`
	UserID               *string   `json:"user_id,omitempty"` // Linked if barber has system login
	Name                 string    `json:"name"`
	Phone                string    `json:"phone,omitempty"`
	CommissionPercentage int       `json:"commission_percentage"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Client represents a barbershop customer/client.
type Client struct {
	ID        string    `json:"id"`
	UnitID    string    `json:"unit_id,omitempty"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Service represents a barbershop service (e.g. Haircut, Shave, Combo).
type Service struct {
	ID              string    `json:"id"`
	UnitID          string    `json:"unit_id,omitempty"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceInCents    int64     `json:"price_in_cents"` // Integer cents avoids floating-point issues
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PaymentMethod defines accepted payment methods for billing.
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
	PaymentMethodDebitCard  PaymentMethod = "DEBIT_CARD"
	PaymentMethodPix        PaymentMethod = "PIX"
	PaymentMethodCash       PaymentMethod = "CASH"
	PaymentMethodOther      PaymentMethod = "OTHER"
	PaymentMethodCard       PaymentMethod = "CARD" // Kept for backward compatibility
)

// IsValid validates whether the payment method is recognized.
func (p PaymentMethod) IsValid() bool {
	switch p {
	case PaymentMethodCreditCard, PaymentMethodDebitCard, PaymentMethodPix, PaymentMethodCash, PaymentMethodOther, PaymentMethodCard:
		return true
	default:
		return false
	}
}

// BillingStatus defines billing lifecycle states.
type BillingStatus string

const (
	BillingStatusPaid      BillingStatus = "PAID"
	BillingStatusCancelled BillingStatus = "CANCELLED"
)

// Billing represents a financial billing transaction.
type Billing struct {
	ID            string        `json:"id"`
	UnitID        string        `json:"unit_id,omitempty"`
	Date          time.Time     `json:"date"`
	BarberID      string        `json:"barber_id"`
	ClientID      string        `json:"client_id"`
	ServiceID     string        `json:"service_id"`
	AmountInCents int64         `json:"amount_in_cents"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Status        BillingStatus `json:"status"`
	Notes         string        `json:"notes,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// AppointmentStatus defines the possible states of an appointment.
type AppointmentStatus string

const (
	StatusScheduled AppointmentStatus = "SCHEDULED"
	StatusCompleted AppointmentStatus = "COMPLETED"
	StatusCancelled AppointmentStatus = "CANCELLED"
)

// Appointment represents a customer booking with a barber.
type Appointment struct {
	ID         string            `json:"id"`
	UnitID     string            `json:"unit_id,omitempty"`
	BarberID   string            `json:"barber_id"`
	CustomerID string            `json:"customer_id"`
	ServiceID  string            `json:"service_id"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Status     AppointmentStatus `json:"status"`
	Notes      string            `json:"notes,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
