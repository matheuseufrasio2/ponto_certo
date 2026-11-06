package service

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

type mockAppointmentRepo struct {
	appointments   map[string]*domain.Appointment
	conflictResult bool
}

func (m *mockAppointmentRepo) Create(ctx context.Context, apt *domain.Appointment) error {
	apt.ID = "apt-test-123"
	apt.CreatedAt = time.Now()
	apt.UpdatedAt = time.Now()
	if m.appointments == nil {
		m.appointments = make(map[string]*domain.Appointment)
	}
	m.appointments[apt.ID] = apt
	return nil
}

func (m *mockAppointmentRepo) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	if apt, ok := m.appointments[id]; ok {
		return apt, nil
	}
	return nil, domain.ErrAppointmentNotFound
}

func (m *mockAppointmentRepo) ListByUnitAndDate(ctx context.Context, unitID string, date time.Time) ([]domain.AppointmentDetails, error) {
	return []domain.AppointmentDetails{
		{
			Appointment: domain.Appointment{ID: "apt-1", UnitID: unitID},
			ClientName:  "João",
			BarberName:  "Marcos",
			ServiceName: "Corte",
			TimeSlot:    "14:00",
		},
	}, nil
}

func (m *mockAppointmentRepo) ListByBarberAndDate(ctx context.Context, barberID string, date time.Time) ([]domain.AppointmentDetails, error) {
	return []domain.AppointmentDetails{
		{
			Appointment: domain.Appointment{ID: "apt-1", BarberID: barberID},
			ClientName:  "João",
			BarberName:  "Marcos",
			ServiceName: "Corte",
			TimeSlot:    "14:00",
		},
	}, nil
}

func (m *mockAppointmentRepo) UpdateStatus(ctx context.Context, id string, status domain.AppointmentStatus) error {
	if apt, ok := m.appointments[id]; ok {
		apt.Status = status
		return nil
	}
	return domain.ErrAppointmentNotFound
}

func (m *mockAppointmentRepo) CheckConflict(ctx context.Context, barberID string, start, end time.Time, excludeID string) (bool, error) {
	return m.conflictResult, nil
}

type mockCatalogRepo struct {
	barbers  []domain.Barber
	services map[string]*domain.Service
}

func (m *mockCatalogRepo) ListBarbersByUnit(ctx context.Context, unitID string) ([]domain.Barber, error) {
	return m.barbers, nil
}

func (m *mockCatalogRepo) ListServicesByUnit(ctx context.Context, unitID string) ([]domain.Service, error) {
	var list []domain.Service
	for _, s := range m.services {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockCatalogRepo) GetServiceByID(ctx context.Context, id string) (*domain.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, domain.ErrServiceNotFound
}

func (m *mockCatalogRepo) GetBarberByUserID(ctx context.Context, userID string) (*domain.Barber, error) {
	for _, b := range m.barbers {
		if b.UserID != nil && *b.UserID == userID {
			return &b, nil
		}
	}
	return nil, domain.ErrBarberNotFound
}

func (m *mockCatalogRepo) GetOrCreateClient(ctx context.Context, unitID, name, phone string) (*domain.Client, error) {
	return &domain.Client{
		ID:     "client-test-1",
		UnitID: unitID,
		Name:   name,
		Phone:  phone,
	}, nil
}

func TestAppointmentService_CreateAppointment(t *testing.T) {
	aptRepo := &mockAppointmentRepo{}
	catalogRepo := &mockCatalogRepo{
		services: map[string]*domain.Service{
			"svc-1": {
				ID:              "svc-1",
				Name:            "Corte Degradê",
				PriceInCents:    6000,
				DurationMinutes: 30,
			},
		},
	}
	svc := NewAppointmentService(aptRepo, catalogRepo)

	ctx := context.Background()
	now := time.Now().Add(time.Hour)

	t.Run("successfully creates appointment with walk-in customer", func(t *testing.T) {
		apt, err := svc.CreateAppointment(ctx, domain.CreateAppointmentInput{
			UnitID:      "unit-1",
			BarberID:    "barber-1",
			ClientName:  "Lucas Cliente",
			ClientPhone: "11999998888",
			ServiceID:   "svc-1",
			StartTime:   now,
			Notes:       "Navalhado",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if apt.ID != "apt-test-123" {
			t.Errorf("expected apt ID apt-test-123, got %s", apt.ID)
		}
		if apt.PriceInCents != 6000 {
			t.Errorf("expected price 6000, got %d", apt.PriceInCents)
		}
		if !apt.EndTime.After(apt.StartTime) {
			t.Errorf("expected EndTime after StartTime")
		}
	})

	t.Run("fails when barber has schedule conflict", func(t *testing.T) {
		aptRepo.conflictResult = true
		_, err := svc.CreateAppointment(ctx, domain.CreateAppointmentInput{
			UnitID:     "unit-1",
			BarberID:   "barber-1",
			ClientID:   "client-1",
			ServiceID:  "svc-1",
			StartTime:  now,
		})
		if err != domain.ErrAppointmentConflict {
			t.Fatalf("expected ErrAppointmentConflict, got %v", err)
		}
	})
}

func TestAppointmentService_CompleteCut(t *testing.T) {
	barberUserID := "user-barber-1"
	aptRepo := &mockAppointmentRepo{
		appointments: map[string]*domain.Appointment{
			"apt-1": {
				ID:       "apt-1",
				BarberID: "barber-1",
				Status:   domain.StatusScheduled,
			},
		},
	}
	catalogRepo := &mockCatalogRepo{
		barbers: []domain.Barber{
			{
				ID:     "barber-1",
				UserID: &barberUserID,
				Name:   "Marcos Barbeiro",
			},
		},
	}
	svc := NewAppointmentService(aptRepo, catalogRepo)
	ctx := context.Background()

	t.Run("successfully completes scheduled cut", func(t *testing.T) {
		err := svc.CompleteCut(ctx, "apt-1", barberUserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if aptRepo.appointments["apt-1"].Status != domain.StatusCompleted {
			t.Errorf("expected status COMPLETED, got %s", aptRepo.appointments["apt-1"].Status)
		}
	})

	t.Run("rejects cut completion for another barber", func(t *testing.T) {
		otherBarberUser := "user-barber-2"
		catalogRepo.barbers = append(catalogRepo.barbers, domain.Barber{
			ID:     "barber-2",
			UserID: &otherBarberUser,
		})

		err := svc.CompleteCut(ctx, "apt-1", otherBarberUser)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
