package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Common domain errors for the barbershop system
var (
	ErrNotFound                     = errors.New("resource not found")
	ErrBarberNotFound               = errors.New("barber not found")
	ErrCustomerNotFound             = errors.New("customer not found")
	ErrClientNotFound               = errors.New("client not found")
	ErrServiceNotFound              = errors.New("service not found")
	ErrBillingNotFound              = errors.New("billing record not found")
	ErrUserNotFound                 = errors.New("user not found")
	ErrOrganizationNotFound         = errors.New("organization not found")
	ErrUnitNotFound                 = errors.New("unit not found")
	ErrProductNotFound              = errors.New("product not found")
	ErrAppointmentNotFound          = errors.New("appointment not found")
	ErrAppointmentConflict          = errors.New("barber already has an appointment scheduled at this time")
	ErrInvalidTimeSlot              = errors.New("invalid appointment time slot")
	ErrInvalidPaymentMethod         = errors.New("invalid payment method")
	ErrInvalidBillingStatus         = errors.New("invalid billing status")
	ErrInvalidRole                  = errors.New("invalid user role")
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrForbidden                    = errors.New("access forbidden")
	ErrInvalidProductPrice          = errors.New("product price must be greater than or equal to zero")
	ErrInvalidProductStock          = errors.New("product stock quantity must be greater than or equal to zero")
	ErrInvalidAppointmentTransition = errors.New("invalid appointment status transition")
	ErrInvalidAppointmentTime       = errors.New("appointment start time must be before end time")
	ErrBillingAlreadyProcessed      = errors.New("appointment has already been paid and processed")
	ErrBarberHasFutureAppointments  = errors.New("cannot change barber status to off duty or vacation while future appointments exist")
	ErrInvalidBarberStatus          = errors.New("invalid barber operational status")
	ErrCannotImpersonateSelf        = errors.New("cannot impersonate yourself")
	ErrCannotImpersonateSuperAdmin  = errors.New("cannot impersonate another super admin")
	ErrImpersonationActionBlocked   = errors.New("action blocked during user impersonation")
	ErrTenantInactive               = errors.New("organization blocked due to registration or financial pending issues")
	ErrCannotDeactivateSuperAdmin   = errors.New("super admin cannot be deactivated")
	ErrCannotDeactivateSelf         = errors.New("cannot change your own active status")
	ErrInsufficientRoleHierarchy    = errors.New("insufficient role hierarchy to change this user status")
	ErrScheduleBlockConflict        = errors.New("schedule block conflicts with existing appointments")
	ErrScheduleBlockNotFound        = errors.New("schedule block not found")
	ErrEmployeeLeaveNotFound        = errors.New("employee leave not found")
	ErrInvalidDateRange             = errors.New("invalid date range: start date must be before or equal to end date")
	ErrEmployeeLeaveConflict        = errors.New("employee already has a leave scheduled for this period")
)

// BarberStatus defines live operational shift state for a barber.
type BarberStatus string

const (
	BarberStatusAvailable BarberStatus = "AVAILABLE"
	BarberStatusBusy      BarberStatus = "BUSY"
	BarberStatusLunch     BarberStatus = "LUNCH"
	BarberStatusOffDuty   BarberStatus = "OFF_DUTY"
	BarberStatusVacation  BarberStatus = "VACATION"
)

// IsValid validates whether the barber status is recognized.
func (s BarberStatus) IsValid() bool {
	switch s {
	case BarberStatusAvailable, BarberStatusBusy, BarberStatusLunch, BarberStatusOffDuty, BarberStatusVacation:
		return true
	default:
		return false
	}
}

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

// SubscriptionStatus defines the SaaS billing and platform access state for an organization.
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusPastDue   SubscriptionStatus = "PAST_DUE"
	SubscriptionStatusTrial     SubscriptionStatus = "TRIAL"
	SubscriptionStatusCancelled SubscriptionStatus = "CANCELLED"
)

// Organization represents the top-level tenant (brand, franchise or single barbershop).
type Organization struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	TradeName          string             `json:"trade_name,omitempty"`
	Document           string             `json:"document,omitempty"`
	Plan               string             `json:"plan"`
	MonthlyFeeInCents  int64              `json:"monthly_fee_in_cents"`
	SubscriptionStatus SubscriptionStatus `json:"subscription_status"`
	IsActive           bool               `json:"is_active"`
	UnitsCount         int                `json:"units_count"`
	ActiveUsersCount   int                `json:"active_users_count"`
	TotalUsersCount    int                `json:"total_users_count"`
	LastActivityAt     *time.Time         `json:"last_activity_at,omitempty"`
	PastDueSince       *time.Time         `json:"past_due_since,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// BarbershopUnit represents a physical branch or location belonging to an organization.
type BarbershopUnit struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Phone          string    `json:"phone,omitempty"`
	Address        string    `json:"address,omitempty"`
	ChairCount     int       `json:"chair_count"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// User represents an authenticated system actor with specific RBAC role.
type User struct {
	ID                   string              `json:"id"`
	OrganizationID       *string             `json:"organization_id,omitempty"` // NULL for SUPER_ADMIN
	UnitID               *string             `json:"unit_id,omitempty"`         // NULL for SUPER_ADMIN & ORG_OWNER
	Name                 string              `json:"name"`
	Email                string              `json:"email"`
	PasswordHash         string              `json:"-"`
	Role                 Role                `json:"role"`
	IsActive             bool                `json:"is_active"`
	OrganizationIsActive *bool               `json:"organization_is_active,omitempty"`
	SubscriptionStatus   *SubscriptionStatus `json:"subscription_status,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// OrganizationEmployee represents a staff member (manager, receptionist, barber) in an organization.
type OrganizationEmployee struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UnitID         *string   `json:"unit_id,omitempty"`
	UnitName       string    `json:"unit_name,omitempty"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Role           Role      `json:"role"`
	IsActive       bool      `json:"is_active"`
	BarberID       *string   `json:"barber_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ToggleUserStatusInput models payload for toggling user active status.
type ToggleUserStatusInput struct {
	IsActive bool `json:"is_active"`
}

// ToggleUserStatusResponse carries updated user status and any cascaded organization status.
type ToggleUserStatusResponse struct {
	UserID               string  `json:"user_id"`
	IsActive             bool    `json:"is_active"`
	Role                 Role    `json:"role"`
	OrganizationID       *string `json:"organization_id,omitempty"`
	OrganizationIsActive *bool   `json:"organization_is_active,omitempty"`
}

// OwnerSubscription defines subscription status and commercial tier for an organization owner.
type OwnerSubscription struct {
	OrganizationID     string             `json:"organization_id"`
	Name               string             `json:"name"`
	TradeName          string             `json:"trade_name"`
	Document           string             `json:"document"`
	Plan               string             `json:"plan"`
	MonthlyFeeInCents  int64              `json:"monthly_fee_in_cents"`
	SubscriptionStatus SubscriptionStatus `json:"subscription_status"`
	IsActive           bool               `json:"is_active"`
	SuspensionReason   string             `json:"suspension_reason"` // NONE, FINANCIAL_PENDING, SUBSCRIPTION_CANCELLED, MANUAL_SUSPENSION
}

// Barber represents a barber professional linked to an optional user login.
type Barber struct {
	ID                   string       `json:"id"`
	UnitID               string       `json:"unit_id,omitempty"`
	UserID               *string      `json:"user_id,omitempty"` // Linked if barber has system login
	Name                 string       `json:"name"`
	Phone                string       `json:"phone,omitempty"`
	CommissionPercentage int          `json:"commission_percentage"`
	Status               BarberStatus `json:"status"`
	IsActive             bool         `json:"is_active"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
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
	ServiceCategoryCuidadosEspeciais ServiceCategory = "CUIDADOS_ESPECIAIS"
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
	ID              string    `json:"id"`
	BarberID        string    `json:"barber_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Reason          string    `json:"reason"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

// CashClosingSummary aggregates financial close data by payment method for a unit and date.
type CashClosingSummary struct {
	Date               time.Time               `json:"date"`
	UnitID             string                  `json:"unit_id"`
	TotalAmountInCents int64                   `json:"total_amount_in_cents"`
	TotalTransactions  int                     `json:"total_transactions"`
	ByPaymentMethod    map[PaymentMethod]int64 `json:"by_payment_method"`
}

// ManagerDashboardMetrics provides real-time operational KPIs for unit managers.
type ManagerDashboardMetrics struct {
	TodayRevenueInCents       int64 `json:"today_revenue_in_cents"`
	ChairOccupancyRatePercent int   `json:"chair_occupancy_rate_percent"`
	TodayAppointmentsCount    int   `json:"today_appointments_count"`
	ActiveBarbersCount        int   `json:"active_barbers_count"`
	WaitingClientsCount       int   `json:"waiting_clients_count"`
}

// UpdateBarberStatusInput carries new operational shift status for a barber.
type UpdateBarberStatusInput struct {
	Status BarberStatus `json:"status"`
}

// Validate verifies that status matches recognized operational states.
func (i UpdateBarberStatusInput) Validate() error {
	if !i.Status.IsValid() {
		return ErrInvalidBarberStatus
	}
	return nil
}

// ToggleServiceInput defines parameters for enabling or pausing a service at unit level.
type ToggleServiceInput struct {
	IsActive bool `json:"is_active"`
}

// OwnerExecutiveKPIs aggregates organization-wide financial KPIs and performance.
type OwnerExecutiveKPIs struct {
	GrossRevenueInCents     int64   `json:"gross_revenue_in_cents"`
	NetRevenueInCents       int64   `json:"net_revenue_in_cents"`
	TotalCommissionsInCents int64   `json:"total_commissions_in_cents"`
	AverageTicketInCents    int64   `json:"average_ticket_in_cents"`
	TotalAppointments       int     `json:"total_appointments"`
	RetentionRatePercent    float64 `json:"retention_rate_percent"`
}

// UnitRevenueComparison summarizes revenue per unit for organization comparisons.
type UnitRevenueComparison struct {
	UnitID         string `json:"unit_id"`
	UnitName       string `json:"unit_name"`
	RevenueInCents int64  `json:"revenue_in_cents"`
}

// ServiceMixAnalytics details sales volume and revenue by service.
type ServiceMixAnalytics struct {
	ServiceID      string `json:"service_id"`
	ServiceName    string `json:"service_name"`
	Category       string `json:"category"`
	Count          int    `json:"count"`
	RevenueInCents int64  `json:"revenue_in_cents"`
}

// PaginatedBillings carries a slice of billings with pagination metadata.
type PaginatedBillings struct {
	Items   []Billing `json:"items"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	PerPage int       `json:"per_page"`
}

// SaaSMetrics represents global platform health and recurring financial KPIs.
type SaaSMetrics struct {
	MRRInCents              int64   `json:"mrr_in_cents"`
	ARRInCents              int64   `json:"arr_in_cents"`
	ActiveTenants           int     `json:"active_tenants"`
	ChurnRatePercent        float64 `json:"churn_rate_percent"`
	PastDueTenantsCount     int     `json:"past_due_tenants_count"`
	TotalUnitsAcrossTenants int     `json:"total_units_across_tenants"`
}

// CreateTenantInput defines parameters for onboarding a new barbershop franchise.
type CreateTenantInput struct {
	Name              string `json:"name"`
	TradeName         string `json:"trade_name"`
	Document          string `json:"document"`
	Plan              string `json:"plan"` // BASIC, PRO, ENTERPRISE
	MonthlyFeeInCents int64  `json:"monthly_fee_in_cents"`
	AdminName         string `json:"admin_name"`
	AdminEmail        string `json:"admin_email"`
	AdminPassword     string `json:"admin_password"`
}

// ImpersonateResponse carries access credentials for a super admin diagnostic session.
type ImpersonateResponse struct {
	Token          string `json:"token"`
	TargetUserID   string `json:"target_user_id"`
	TargetUserName string `json:"target_user_name"`
	TargetRole     string `json:"target_role"`
	ExpiresIn      int    `json:"expires_in"` // Ex: 1800 for 30min
}

// PaginatedOrganizations carries a list of organizations with pagination metadata.
type PaginatedOrganizations struct {
	Items   []Organization `json:"items"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
}

// TenantOverview aggregates detailed engagement, financial health, and operational capacity for a tenant.
type TenantOverview struct {
	Organization           Organization `json:"organization"`
	GMVLast30DaysInCents   int64        `json:"gmv_last_30_days_in_cents"`
	AppointmentsLast30Days int          `json:"appointments_last_30_days"`
	TotalBarbersCount      int          `json:"total_barbers_count"`
	DaysOverdue            int          `json:"days_overdue"`
}

// UpdateTenantPlanInput defines parameters for upgrading/downgrading a tenant's subscription.
type UpdateTenantPlanInput struct {
	Plan              string `json:"plan"`
	MonthlyFeeInCents int64  `json:"monthly_fee_in_cents"`
}

// Validate verifies plan and fee constraints.
func (i UpdateTenantPlanInput) Validate() error {
	if i.Plan != "BASIC" && i.Plan != "PRO" && i.Plan != "ENTERPRISE" {
		return errors.New("plan must be BASIC, PRO, or ENTERPRISE")
	}
	if i.MonthlyFeeInCents < 0 {
		return errors.New("monthly fee must be greater than or equal to zero")
	}
	return nil
}

// ResetTenantPasswordInput defines optional payload for administrative password resets.
type ResetTenantPasswordInput struct {
	NewPassword string `json:"new_password,omitempty"`
}

// ResetPasswordResponse carries the result of an administrative password reset.
type ResetPasswordResponse struct {
	UserID            string `json:"user_id"`
	UserEmail         string `json:"user_email"`
	TemporaryPassword string `json:"temporary_password"`
}

// DevSuperAdmin represents a super admin user for the developer authentication modal.
type DevSuperAdmin struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// DevTenantOwner represents the primary owner of a tenant for dev discovery.
type DevTenantOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// DevTenant represents a tenant summary for the dev modal.
type DevTenant struct {
	ID          string          `json:"id"`
	TradeName   string          `json:"trade_name"`
	Plan        string          `json:"plan"`
	Status      string          `json:"status"`
	DaysPastDue int             `json:"days_past_due"`
	Owner       *DevTenantOwner `json:"owner,omitempty"`
}

// DevTenantsResponse represents the paginated response for dev tenants.
type DevTenantsResponse struct {
	Tenants []DevTenant `json:"tenants"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int         `json:"total"`
	HasMore bool        `json:"has_more"`
}

// DevTenantUser represents a user belonging to a specific tenant.
type DevTenantUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UnitName string `json:"unit_name"`
}

// DevUser represents a user categorized by role for dev discovery.
type DevUser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	TenantName string `json:"tenant_name"`
}

// DevUsersResponse represents the paginated response for dev users.
type DevUsersResponse struct {
	Users   []DevUser `json:"users"`
	Page    int       `json:"page"`
	Limit   int       `json:"limit"`
	Total   int       `json:"total"`
	HasMore bool      `json:"has_more"`
}

// IsValidUUID validates that a string is a standard 36-character canonical UUID (8-4-4-4-12 hex).
func IsValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}
	for i, r := range u {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
	}
	return true
}

// AuditLog represents an immutable compliance log entry.
type AuditLog struct {
	ID             string          `json:"id"`
	OrganizationID *string         `json:"organization_id,omitempty"`
	UnitID         *string         `json:"unit_id,omitempty"`
	ActorID        *string         `json:"actor_id,omitempty"`
	ActorRole      Role            `json:"actor_role"`
	Action         string          `json:"action"`
	TargetEntity   string          `json:"target_entity"`
	TargetID       string          `json:"target_id"`
	Details        json.RawMessage `json:"details" swaggertype:"object"`
	IPAddress      string          `json:"ip_address,omitempty"`
	UserAgent      string          `json:"user_agent,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// LeaveType defines accepted types of employee leaves.
type LeaveType string

const (
	LeaveTypeVacation     LeaveType = "VACATION"
	LeaveTypeMedicalLeave LeaveType = "MEDICAL_LEAVE"
	LeaveTypePersonal     LeaveType = "PERSONAL"
	LeaveTypeOther        LeaveType = "OTHER"
)

// IsValid validates whether the leave type is supported.
func (l LeaveType) IsValid() bool {
	switch l {
	case LeaveTypeVacation, LeaveTypeMedicalLeave, LeaveTypePersonal, LeaveTypeOther:
		return true
	default:
		return false
	}
}

// UserLeaveContext models user organizational and role data for leave management.
type UserLeaveContext struct {
	UserName       string  `json:"user_name"`
	UserRole       Role    `json:"user_role"`
	OrganizationID string  `json:"organization_id"`
	UnitID         *string `json:"unit_id,omitempty"`
	BarberID       *string `json:"barber_id,omitempty"`
}

// EmployeeLeave models employee vacations and leaves across all roles.
type EmployeeLeave struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	UnitID          *string   `json:"unit_id,omitempty"`
	UserID          string    `json:"user_id"`
	UserName        string    `json:"user_name,omitempty"`
	UserRole        Role      `json:"user_role,omitempty"`
	LeaveType       LeaveType `json:"leave_type"`
	StartDate       string    `json:"start_date"` // YYYY-MM-DD
	EndDate         string    `json:"end_date"`   // YYYY-MM-DD
	Reason          string    `json:"reason"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateScheduleBlockInput models payload for creating a schedule block for a barber.
type CreateScheduleBlockInput struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

// Validate checks required fields and chronological consistency.
func (i CreateScheduleBlockInput) Validate() error {
	if i.StartTime.IsZero() || i.EndTime.IsZero() {
		return errors.New("start_time and end_time are required")
	}
	if !i.StartTime.Before(i.EndTime) {
		return ErrInvalidAppointmentTime
	}
	if strings.TrimSpace(i.Reason) == "" {
		return errors.New("reason is required")
	}
	return nil
}

// CreateEmployeeLeaveInput models payload for scheduling employee leave or vacation.
type CreateEmployeeLeaveInput struct {
	UserID    string    `json:"user_id"`
	UnitID    *string   `json:"unit_id,omitempty"`
	LeaveType LeaveType `json:"leave_type"`
	StartDate string    `json:"start_date"` // YYYY-MM-DD
	EndDate   string    `json:"end_date"`   // YYYY-MM-DD
	Reason    string    `json:"reason"`
}

// Validate checks date ranges and required fields.
func (i CreateEmployeeLeaveInput) Validate() error {
	if strings.TrimSpace(i.UserID) == "" {
		return errors.New("user_id is required")
	}
	if !i.LeaveType.IsValid() {
		return errors.New("invalid leave_type")
	}
	start, err := time.Parse("2006-01-02", i.StartDate)
	if err != nil {
		return errors.New("invalid start_date format, must be YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", i.EndDate)
	if err != nil {
		return errors.New("invalid end_date format, must be YYYY-MM-DD")
	}
	if start.After(end) {
		return ErrInvalidDateRange
	}
	if strings.TrimSpace(i.Reason) == "" {
		return errors.New("reason is required")
	}
	return nil
}
