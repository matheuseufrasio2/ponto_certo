package domain

import (
	"errors"
	"time"
)

// Common domain errors for the barbershop
var (
	ErrNotFound              = errors.New("resource not found")
	ErrBarberNotFound        = errors.New("barber not found")
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrClientNotFound        = errors.New("client not found")
	ErrServiceNotFound       = errors.New("service not found")
	ErrBillingNotFound       = errors.New("billing record not found")
	ErrAppointmentConflict   = errors.New("barber already has an appointment scheduled at this time")
	ErrInvalidTimeSlot       = errors.New("invalid appointment time slot")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrInvalidBillingStatus  = errors.New("invalid billing status")
)

// Barber represents a barber professional.
type Barber struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Client represents a barbershop customer/client.
type Client struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Service represents a barbershop service (e.g. Haircut, Shave, Combo).
type Service struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceInCents    int64     `json:"price_in_cents"` // Price in cents avoids floating-point precision issues
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PaymentMethod defines accepted payment methods for billing.
type PaymentMethod string

const (
	PaymentMethodCard  PaymentMethod = "CARD"
	PaymentMethodCash  PaymentMethod = "CASH"
	PaymentMethodPix   PaymentMethod = "PIX"
	PaymentMethodOther PaymentMethod = "OTHER"
)

// BillingStatus defines billing lifecycle states.
type BillingStatus string

const (
	BillingStatusPaid      BillingStatus = "PAID"
	BillingStatusCancelled BillingStatus = "CANCELLED"
)

// Billing represents a financial billing transaction.
type Billing struct {
	ID            string        `json:"id"`
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
