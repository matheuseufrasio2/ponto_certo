package service

import (
	"context"
	"fmt"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// AppointmentService coordinates appointment scheduling and lifecycle operations.
type AppointmentService struct {
	appointmentRepo repository.AppointmentRepository
	catalogRepo     repository.CatalogRepository
}

// NewAppointmentService instantiates AppointmentService.
func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	catalogRepo repository.CatalogRepository,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		catalogRepo:     catalogRepo,
	}
}

// CreateAppointment schedules a new customer appointment with validation and conflict checks.
func (s *AppointmentService) CreateAppointment(ctx context.Context, input domain.CreateAppointmentInput) (*domain.Appointment, error) {
	if input.UnitID == "" {
		return nil, domain.ErrUnitNotFound
	}
	if input.BarberID == "" {
		return nil, domain.ErrBarberNotFound
	}
	if input.ServiceID == "" {
		return nil, domain.ErrServiceNotFound
	}

	// 1. Resolve client
	clientID := input.ClientID
	if clientID == "" {
		if input.ClientName == "" {
			return nil, domain.ErrClientNotFound
		}
		client, err := s.catalogRepo.GetOrCreateClient(ctx, input.UnitID, input.ClientName, input.ClientPhone)
		if err != nil {
			return nil, fmt.Errorf("resolve client: %w", err)
		}
		clientID = client.ID
	}

	// 2. Fetch service to determine pricing and duration
	svc, err := s.catalogRepo.GetServiceByID(ctx, input.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}

	// 3. Compute times
	startTime := input.StartTime
	endTime := input.EndTime
	if endTime.IsZero() || !endTime.After(startTime) {
		duration := time.Duration(svc.DurationMinutes) * time.Minute
		if duration <= 0 {
			duration = 30 * time.Minute
		}
		endTime = startTime.Add(duration)
	}

	// 4. Check for overlapping appointments with this barber
	conflict, err := s.appointmentRepo.CheckConflict(ctx, input.BarberID, startTime, endTime, "")
	if err != nil {
		return nil, fmt.Errorf("verify conflict: %w", err)
	}
	if conflict {
		return nil, domain.ErrAppointmentConflict
	}

	// 5. Build domain appointment
	apt := &domain.Appointment{
		UnitID:        input.UnitID,
		BarberID:      input.BarberID,
		ClientID:      clientID,
		ServiceID:     input.ServiceID,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        domain.StatusScheduled,
		PaymentStatus: domain.PaymentStatusPending,
		PriceInCents:  svc.PriceInCents,
		Notes:         input.Notes,
	}

	if err := apt.Validate(); err != nil {
		return nil, err
	}

	if err := s.appointmentRepo.Create(ctx, apt); err != nil {
		return nil, fmt.Errorf("create appointment: %w", err)
	}

	return apt, nil
}

// ListUnitAppointments lists appointments for a unit on a given date.
func (s *AppointmentService) ListUnitAppointments(ctx context.Context, unitID string, date time.Time) ([]domain.AppointmentDetails, error) {
	if unitID == "" {
		return nil, domain.ErrUnitNotFound
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}
	return s.appointmentRepo.ListByUnitAndDate(ctx, unitID, date)
}

// ListBarberQueue lists today's appointments for the barber linked to the authenticated user.
func (s *AppointmentService) ListBarberQueue(ctx context.Context, barberUserID string, date time.Time) ([]domain.AppointmentDetails, error) {
	barber, err := s.catalogRepo.GetBarberByUserID(ctx, barberUserID)
	if err != nil {
		return nil, err
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}
	return s.appointmentRepo.ListByBarberAndDate(ctx, barber.ID, date)
}

// CompleteCut marks an appointment as completed by the barber.
func (s *AppointmentService) CompleteCut(ctx context.Context, appointmentID, barberUserID string) error {
	barber, err := s.catalogRepo.GetBarberByUserID(ctx, barberUserID)
	if err != nil {
		return err
	}

	apt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return err
	}

	if apt.BarberID != barber.ID {
		return domain.ErrForbidden
	}

	if !apt.CanTransitionTo(domain.StatusCompleted) {
		return domain.ErrInvalidAppointmentTransition
	}

	return s.appointmentRepo.UpdateStatus(ctx, appointmentID, domain.StatusCompleted)
}

