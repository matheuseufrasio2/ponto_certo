package domain

import (
	"errors"
	"time"
)

// Common domain errors for the barbershop system
var (
	ErrNotFound                   = errors.New("resource not found")
	ErrBarberNotFound             = errors.New("barber not found")
	ErrCustomerNotFound           = errors.New("customer not found")
	ErrClientNotFound             = errors.New("client not found")
	ErrServiceNotFound            = errors.New("service not found")
	ErrBillingNotFound            = errors.New("billing record not found")
	ErrUserNotFound               = errors.New("user not found")
	ErrOrganizationNotFound       = errors.New("organization not found")
	ErrUnitNotFound               = errors.New("unit not found")
	ErrProductNotFound            = errors.New("product not found")
	ErrAppointmentNotFound        = errors.New("appointment not found")
	ErrAppointmentConflict        = errors.New("barber already has an appointment scheduled at this time")
	ErrInvalidTimeSlot            = errors.New("invalid appointment time slot")
	ErrInvalidPaymentMethod       = errors.New("invalid payment method")
	ErrInvalidBillingStatus       = errors.New("invalid billing status")
	ErrInvalidRole                = errors.New("invalid user role")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrForbidden                  = errors.New("access forbidden")
	ErrInvalidProductPrice        = errors.New("product price must be greater than or equal to zero")
	ErrInvalidProductStock        = errors.New("product stock quantity must be greater than or equal to zero")
	ErrInvalidAppointmentTransition = errors.New("invalid appointment status transition")
	ErrInvalidAppointmentTime     = errors.New("appointment start time must be before end time")
	ErrBillingAlreadyProcessed    = errors.New("appointment has already been paid and processed")
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

// ServiceCategory defines categories for barbershop services.
type ServiceCategory string

const (
	ServiceCategoryCabelo            ServiceCategory = "CABELO"
	ServiceCategoryBarba             ServiceCategory = "BARBA"
	ServiceCategoryCabeloEBarba      ServiceCategory = "CABELO_E_BARBA"
	ServiceCategoryCombos            ServiceCategory = "COMBOS"
	ServiceCategoryTratamentos       ServiceCategory = "TRATAMENTOS_CAPILARES"
	ServiceCategoryCuidadosEspeciais  ServiceCategory = "CUIDADOS_ESPECIAIS"
	ServiceCategoryQuimicas          ServiceCategory = "QUIMICAS_CAPILARES"
	ServiceCategoryVisagismo         ServiceCategory = "VISAGISMO"
	ServiceCategoryEsteticaFacial    ServiceCategory = "ESTETICA_FACIAL"
)

// IsValid validates whether the service category is recognized.
func (s ServiceCategory) IsValid() bool {
	switch s {
	case ServiceCategoryCabelo,
		ServiceCategoryBarba,
		ServiceCategoryCabeloEBarba,
		ServiceCategoryCombos,
		ServiceCategoryTratamentos,
		ServiceCategoryCuidadosEspeciais,
		ServiceCategoryQuimicas,
		ServiceCategoryVisagismo,
		ServiceCategoryEsteticaFacial:
		return true
	default:
		return false
	}
}

// Service represents a barbershop service (e.g. Haircut, Shave, Combo).
type Service struct {
	ID              string          `json:"id"`
	UnitID          string          `json:"unit_id,omitempty"`
	Name            string          `json:"name"`
	Category        ServiceCategory `json:"category,omitempty"`
	Description     string          `json:"description,omitempty"`
	DurationMinutes int             `json:"duration_minutes"`
	PriceInCents    int64           `json:"price_in_cents"` // Integer cents avoids floating-point issues
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ProductCategory defines taxonomy for retail and professional products sold in barbershop.
type ProductCategory string

const (
	ProductCategoryKits              ProductCategory = "KITS"
	ProductCategoryShampooHigiene    ProductCategory = "SHAMPOO_HIGIENE"
	ProductCategoryFinalizadores     ProductCategory = "FINALIZADORES"
	ProductCategoryBarbeariaCuidados ProductCategory = "BARBEARIA_CUIDADOS"
	ProductCategoryAcessorios        ProductCategory = "ACESSORIOS"
	ProductCategorySkincare          ProductCategory = "SKINCARE"
	ProductCategoryBebidasMerch      ProductCategory = "BEBIDAS_MERCHANDISING"
)

// IsValid validates whether the product category is recognized.
func (p ProductCategory) IsValid() bool {
	switch p {
	case ProductCategoryKits,
		ProductCategoryShampooHigiene,
		ProductCategoryFinalizadores,
		ProductCategoryBarbeariaCuidados,
		ProductCategoryAcessorios,
		ProductCategorySkincare,
		ProductCategoryBebidasMerch:
		return true
	default:
		return false
	}
}

// Product represents a retail or professional supply item sold/managed in a unit.
type Product struct {
	ID            string          `json:"id"`
	UnitID        string          `json:"unit_id"`
	Name          string          `json:"name"`
	Category      ProductCategory `json:"category"`
	Description   string          `json:"description,omitempty"`
	PriceInCents  int64           `json:"price_in_cents"` // Sale price to client
	CostInCents   int64           `json:"cost_in_cents"`  // Cost of acquisition
	StockQuantity int             `json:"stock_quantity"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Validate verifies product field constraints.
func (p Product) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if !p.Category.IsValid() {
		return errors.New("invalid product category")
	}
	if p.PriceInCents < 0 {
		return ErrInvalidProductPrice
	}
	if p.CostInCents < 0 {
		return errors.New("product cost must be greater than or equal to zero")
	}
	if p.StockQuantity < 0 {
		return ErrInvalidProductStock
	}
	return nil
}

// UnitOperatingHour represents operating schedule for a unit per day of week (0=Sunday .. 6=Saturday).
type UnitOperatingHour struct {
	ID        string    `json:"id"`
	UnitID    string    `json:"unit_id"`
	DayOfWeek int       `json:"day_of_week"`
	OpenTime  string    `json:"open_time"`
	CloseTime string    `json:"close_time"`
	IsClosed  bool      `json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BarberWorkingHour represents standard weekly shift schedule for a barber.
type BarberWorkingHour struct {
	ID             string    `json:"id"`
	BarberID       string    `json:"barber_id"`
	DayOfWeek      int       `json:"day_of_week"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	BreakStartTime *string   `json:"break_start_time,omitempty"`
	BreakEndTime   *string   `json:"break_end_time,omitempty"`
	IsOff          bool      `json:"is_off"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BarberScheduleBlock represents temporary agenda block (e.g. time off, doctor, vacation).
type BarberScheduleBlock struct {
	ID        string    `json:"id"`
	BarberID  string    `json:"barber_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BarberService associates a barber with a service they are qualified to perform.
type BarberService struct {
	BarberID  string    `json:"barber_id"`
	ServiceID string    `json:"service_id"`
	CreatedAt time.Time `json:"created_at"`
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

// Billing represents a financial billing transaction (service, product, or appointment checkout).
type Billing struct {
	ID            string        `json:"id"`
	UnitID        string        `json:"unit_id,omitempty"`
	Date          time.Time     `json:"date"`
	BarberID      string        `json:"barber_id"`
	ClientID      string        `json:"client_id"`
	ServiceID     *string       `json:"service_id,omitempty"`
	ProductID     *string       `json:"product_id,omitempty"`
	AppointmentID *string       `json:"appointment_id,omitempty"`
	Quantity      int           `json:"quantity"`
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
	StatusNoShow    AppointmentStatus = "NO_SHOW"
)

// PaymentStatus defines whether an appointment has been paid.
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusPaid    PaymentStatus = "PAID"
)

// Appointment represents a customer booking with a barber.
type Appointment struct {
	ID            string            `json:"id"`
	UnitID        string            `json:"unit_id,omitempty"`
	BarberID      string            `json:"barber_id"`
	ClientID      string            `json:"client_id"`
	CustomerID    string            `json:"customer_id,omitempty"` // Legacy alias
	ServiceID     string            `json:"service_id"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Status        AppointmentStatus `json:"status"`
	PaymentStatus PaymentStatus     `json:"payment_status"`
	PriceInCents  int64             `json:"price_in_cents"`
	Notes         string            `json:"notes,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// CanTransitionTo enforces the appointment state machine:
// SCHEDULED -> COMPLETED, CANCELLED, NO_SHOW.
// Terminal states cannot transition to other states.
func (a Appointment) CanTransitionTo(target AppointmentStatus) bool {
	if a.Status == StatusScheduled {
		return target == StatusCompleted || target == StatusCancelled || target == StatusNoShow
	}
	return false
}

// Validate verifies appointment invariants.
func (a Appointment) Validate() error {
	if a.UnitID == "" {
		return errors.New("unit_id is required")
	}
	if a.BarberID == "" {
		return errors.New("barber_id is required")
	}
	if a.ClientID == "" && a.CustomerID == "" {
		return errors.New("client_id is required")
	}
	if a.ServiceID == "" {
		return errors.New("service_id is required")
	}
	if !a.StartTime.Before(a.EndTime) {
		return ErrInvalidAppointmentTime
	}
	if a.PriceInCents < 0 {
		return errors.New("price_in_cents must be greater than or equal to zero")
	}
	return nil
}

// HasScheduleConflict evaluates whether two appointment time intervals overlap.
// Invariant: newStart < existingEnd AND newEnd > existingStart.
func HasScheduleConflict(startA, endA, startB, endB time.Time) bool {
	return startA.Before(endB) && endA.After(startB)
}

// AppointmentDetails enriches an Appointment with related entity names for UI presentation.
type AppointmentDetails struct {
	Appointment
	ClientName  string `json:"client_name"`
	ClientPhone string `json:"client_phone,omitempty"`
	BarberName  string `json:"barber_name"`
	ServiceName string `json:"service_name"`
	TimeSlot    string `json:"time_slot"`
}

// CreateAppointmentInput holds payload data for booking an appointment.
type CreateAppointmentInput struct {
	UnitID      string    `json:"unit_id"`
	BarberID    string    `json:"barber_id"`
	ClientID    string    `json:"client_id,omitempty"`
	ClientName  string    `json:"client_name,omitempty"`
	ClientPhone string    `json:"client_phone,omitempty"`
	ServiceID   string    `json:"service_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Notes       string    `json:"notes,omitempty"`
}

// CheckoutInput holds parameters required to liquidate an appointment.
type CheckoutInput struct {
	AppointmentID string        `json:"appointment_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Notes         string        `json:"notes,omitempty"`
}

