package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// UnitAvailabilityChecker verifies whether a branch exists and is active for booking.
type UnitAvailabilityChecker interface {
	IsUnitActive(ctx context.Context, unitID string) (bool, error)
}

// UnitOperatingHoursProvider retrieves the unit operating schedule.
type UnitOperatingHoursProvider interface {
	GetUnitOperatingHours(ctx context.Context, unitID string) ([]domain.UnitOperatingHour, error)
}

// BarberWorkingHoursProvider retrieves the barber weekly schedule.
type BarberWorkingHoursProvider interface {
	GetBarberWorkingHours(ctx context.Context, barberID string) ([]domain.BarberWorkingHour, error)
}

// AppointmentService coordinates appointment scheduling and lifecycle operations.
type AppointmentService struct {
	appointmentRepo     repository.AppointmentRepository
	catalogRepo         repository.CatalogRepository
	unitChecker         UnitAvailabilityChecker
	unitHoursProvider   UnitOperatingHoursProvider
	barberHoursProvider BarberWorkingHoursProvider
}

// NewAppointmentService instantiates AppointmentService.
func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	catalogRepo repository.CatalogRepository,
	unitChecker UnitAvailabilityChecker,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		catalogRepo:     catalogRepo,
		unitChecker:     unitChecker,
	}
}

// SetOperatingHoursProvider wires the unit operating schedule provider.
func (s *AppointmentService) SetOperatingHoursProvider(provider UnitOperatingHoursProvider) {
	s.unitHoursProvider = provider
}

// SetBarberWorkingHoursProvider wires the barber shift schedule provider.
func (s *AppointmentService) SetBarberWorkingHoursProvider(provider BarberWorkingHoursProvider) {
	s.barberHoursProvider = provider
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

	if s.unitChecker != nil {
		isActive, err := s.unitChecker.IsUnitActive(ctx, input.UnitID)
		if err != nil {
			if errors.Is(err, domain.ErrUnitNotFound) {
				return nil, domain.ErrUnitNotFound
			}
			return nil, fmt.Errorf("verify unit availability: %w", err)
		}
		if !isActive {
			return nil, domain.ErrUnitInactive
		}
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

	// 4. Validate unit operating hours
	if err := s.validateOperatingHours(ctx, input.UnitID, startTime, endTime); err != nil {
		return nil, err
	}

	// 5. Validate barber working hours & shifts
	if err := s.validateBarberWorkingHours(ctx, input.BarberID, startTime, endTime); err != nil {
		return nil, err
	}

	// 6. Check for overlapping appointments with this barber
	conflict, err := s.appointmentRepo.CheckConflict(ctx, input.BarberID, startTime, endTime, "")
	if err != nil {
		return nil, fmt.Errorf("verify conflict: %w", err)
	}
	if conflict {
		return nil, domain.ErrAppointmentConflict
	}

	// 7. Build domain appointment
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

func normalizeTimeStr(t string) string {
	trimmed := strings.TrimSpace(t)
	if len(trimmed) == 5 {
		return trimmed + ":00"
	}
	return trimmed
}

func (s *AppointmentService) validateOperatingHours(ctx context.Context, unitID string, start, end time.Time) error {
	if s.unitHoursProvider == nil {
		return nil
	}
	hours, err := s.unitHoursProvider.GetUnitOperatingHours(ctx, unitID)
	if err != nil {
		return fmt.Errorf("verify unit operating hours: %w", err)
	}
	if len(hours) == 0 {
		return nil
	}

	dayOfWeek := int(start.Weekday())
	var dayRule *domain.UnitOperatingHour
	for i := range hours {
		if hours[i].DayOfWeek == dayOfWeek {
			dayRule = &hours[i]
			break
		}
	}
	if dayRule == nil {
		return nil
	}

	if dayRule.IsClosed {
		return domain.ErrUnitClosed
	}

	apptStartTime := start.Format("15:04:00")
	apptEndTime := end.Format("15:04:00")
	openTime := normalizeTimeStr(dayRule.OpenTime)
	closeTime := normalizeTimeStr(dayRule.CloseTime)

	if apptStartTime < openTime || apptEndTime > closeTime {
		return domain.ErrOutsideOperatingHours
	}

	return nil
}

func (s *AppointmentService) validateBarberWorkingHours(ctx context.Context, barberID string, start, end time.Time) error {
	if s.barberHoursProvider == nil {
		return nil
	}
	hours, err := s.barberHoursProvider.GetBarberWorkingHours(ctx, barberID)
	if err != nil {
		return fmt.Errorf("verify barber working hours: %w", err)
	}
	if len(hours) == 0 {
		return nil
	}

	dayOfWeek := int(start.Weekday())
	var dayRule *domain.BarberWorkingHour
	for i := range hours {
		if hours[i].DayOfWeek == dayOfWeek {
			dayRule = &hours[i]
			break
		}
	}
	if dayRule == nil {
		return nil
	}

	if dayRule.IsOff {
		return domain.ErrBarberOffDuty
	}

	apptStartTime := start.Format("15:04:00")
	apptEndTime := end.Format("15:04:00")
	shiftStart := normalizeTimeStr(dayRule.StartTime)
	shiftEnd := normalizeTimeStr(dayRule.EndTime)

	if apptStartTime < shiftStart || apptEndTime > shiftEnd {
		return domain.ErrBarberOffDuty
	}

	if dayRule.BreakStartTime != nil && dayRule.BreakEndTime != nil {
		breakStart := normalizeTimeStr(*dayRule.BreakStartTime)
		breakEnd := normalizeTimeStr(*dayRule.BreakEndTime)
		if breakStart != "" && breakEnd != "" {
			if apptStartTime < breakEnd && apptEndTime > breakStart {
				return domain.ErrBarberOnBreak
			}
		}
	}

	return nil
}
