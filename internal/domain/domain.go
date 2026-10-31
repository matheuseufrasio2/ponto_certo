package domain

import (
	"errors"
	"time"
)

// Common domain errors for the barbershop
var (
	ErrNotFound            = errors.New("resource not found")
	ErrBarberNotFound      = errors.New("barber not found")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrServiceNotFound     = errors.New("service not found")
	ErrAppointmentConflict = errors.New("barber already has an appointment scheduled at this time")
	ErrInvalidTimeSlot     = errors.New("invalid appointment time slot")
)

// Barber represents a barber professional.
type Barber struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Service represents a barbershop service (e.g. Haircut, Shave, Combo).
type Service struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceInCents    int64     `json:"price_in_cents"` // Price in cents avoids floating-point precision issues
	CreatedAt       time.Time `json:"created_at"`
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
