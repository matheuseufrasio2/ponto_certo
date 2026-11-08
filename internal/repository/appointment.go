package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// AppointmentRepository defines data access methods for appointments.
type AppointmentRepository interface {
	Create(ctx context.Context, apt *domain.Appointment) error
	GetByID(ctx context.Context, id string) (*domain.Appointment, error)
	ListByUnitAndDate(ctx context.Context, unitID string, date time.Time) ([]domain.AppointmentDetails, error)
	ListByBarberAndDate(ctx context.Context, barberID string, date time.Time) ([]domain.AppointmentDetails, error)
	UpdateStatus(ctx context.Context, id string, status domain.AppointmentStatus) error
	CheckConflict(ctx context.Context, barberID string, start, end time.Time, excludeID string) (bool, error)
}

// PostgresAppointmentRepository implements AppointmentRepository using PostgreSQL.
type PostgresAppointmentRepository struct {
	db *pgxpool.Pool
}

// NewPostgresAppointmentRepository instantiates a new appointment repository.
func NewPostgresAppointmentRepository(db *pgxpool.Pool) *PostgresAppointmentRepository {
	return &PostgresAppointmentRepository{db: db}
}

// Create inserts a new appointment record.
func (r *PostgresAppointmentRepository) Create(ctx context.Context, apt *domain.Appointment) error {
	query := `
		INSERT INTO appointments (
			unit_id, barber_id, client_id, service_id,
			start_time, end_time, status, payment_status,
			price_in_cents, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`
	return r.db.QueryRow(
		ctx,
		query,
		apt.UnitID,
		apt.BarberID,
		apt.ClientID,
		apt.ServiceID,
		apt.StartTime,
		apt.EndTime,
		apt.Status,
		apt.PaymentStatus,
		apt.PriceInCents,
		apt.Notes,
	).Scan(&apt.ID, &apt.CreatedAt, &apt.UpdatedAt)
}

// GetByID finds an appointment by ID.
func (r *PostgresAppointmentRepository) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	query := `
		SELECT id, unit_id, barber_id, client_id, service_id,
		       start_time, end_time, status, payment_status,
		       price_in_cents, notes, created_at, updated_at
		FROM appointments
		WHERE id = $1;
	`
	var a domain.Appointment
	var notes *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.UnitID,
		&a.BarberID,
		&a.ClientID,
		&a.ServiceID,
		&a.StartTime,
		&a.EndTime,
		&a.Status,
		&a.PaymentStatus,
		&a.PriceInCents,
		&notes,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("query appointment by id: %w", err)
	}
	if notes != nil {
		a.Notes = *notes
	}
	return &a, nil
}

// CheckConflict returns true if the barber has overlapping active appointments.
func (r *PostgresAppointmentRepository) CheckConflict(ctx context.Context, barberID string, start, end time.Time, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM appointments
			WHERE barber_id = $1
			  AND status = 'SCHEDULED'
			  AND ($4 = '' OR id != $4::UUID)
			  AND start_time < $3
			  AND end_time > $2
		);
	`
	var conflict bool
	err := r.db.QueryRow(ctx, query, barberID, start, end, excludeID).Scan(&conflict)
	if err != nil {
		return false, fmt.Errorf("check appointment conflict: %w", err)
	}
	return conflict, nil
}

// ListByUnitAndDate lists all appointments for a given unit within a day.
func (r *PostgresAppointmentRepository) ListByUnitAndDate(ctx context.Context, unitID string, date time.Time) ([]domain.AppointmentDetails, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			a.id, a.unit_id, a.barber_id, a.client_id, a.service_id,
			a.start_time, a.end_time, a.status, a.payment_status,
			a.price_in_cents, a.notes, a.created_at, a.updated_at,
			c.name AS client_name, c.phone AS client_phone,
			b.name AS barber_name,
			s.name AS service_name
		FROM appointments a
		JOIN clients c ON a.client_id = c.id
		JOIN barbers b ON a.barber_id = b.id
		JOIN services s ON a.service_id = s.id
		WHERE a.unit_id = $1
		  AND a.start_time >= $2
		  AND a.start_time < $3
		ORDER BY a.start_time ASC;
	`
	rows, err := r.db.Query(ctx, query, unitID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("query appointments by unit: %w", err)
	}
	defer rows.Close()

	var list []domain.AppointmentDetails
	for rows.Next() {
		var item domain.AppointmentDetails
		var notes, clientPhone *string
		if err := rows.Scan(
			&item.ID,
			&item.UnitID,
			&item.BarberID,
			&item.ClientID,
			&item.ServiceID,
			&item.StartTime,
			&item.EndTime,
			&item.Status,
			&item.PaymentStatus,
			&item.PriceInCents,
			&notes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ClientName,
			&clientPhone,
			&item.BarberName,
			&item.ServiceName,
		); err != nil {
			return nil, fmt.Errorf("scan appointment details: %w", err)
		}
		if notes != nil {
			item.Notes = *notes
		}
		if clientPhone != nil {
			item.ClientPhone = *clientPhone
		}
		item.TimeSlot = item.StartTime.Format("15:04")
		list = append(list, item)
	}
	return list, rows.Err()
}

// ListByBarberAndDate lists all appointments for a barber within a day.
func (r *PostgresAppointmentRepository) ListByBarberAndDate(ctx context.Context, barberID string, date time.Time) ([]domain.AppointmentDetails, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			a.id, a.unit_id, a.barber_id, a.client_id, a.service_id,
			a.start_time, a.end_time, a.status, a.payment_status,
			a.price_in_cents, a.notes, a.created_at, a.updated_at,
			c.name AS client_name, c.phone AS client_phone,
			b.name AS barber_name,
			s.name AS service_name
		FROM appointments a
		JOIN clients c ON a.client_id = c.id
		JOIN barbers b ON a.barber_id = b.id
		JOIN services s ON a.service_id = s.id
		WHERE a.barber_id = $1
		  AND a.start_time >= $2
		  AND a.start_time < $3
		ORDER BY a.start_time ASC;
	`
	rows, err := r.db.Query(ctx, query, barberID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("query appointments by barber: %w", err)
	}
	defer rows.Close()

	var list []domain.AppointmentDetails
	for rows.Next() {
		var item domain.AppointmentDetails
		var notes, clientPhone *string
		if err := rows.Scan(
			&item.ID,
			&item.UnitID,
			&item.BarberID,
			&item.ClientID,
			&item.ServiceID,
			&item.StartTime,
			&item.EndTime,
			&item.Status,
			&item.PaymentStatus,
			&item.PriceInCents,
			&notes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ClientName,
			&clientPhone,
			&item.BarberName,
			&item.ServiceName,
		); err != nil {
			return nil, fmt.Errorf("scan barber queue: %w", err)
		}
		if notes != nil {
			item.Notes = *notes
		}
		if clientPhone != nil {
			item.ClientPhone = *clientPhone
		}
		item.TimeSlot = item.StartTime.Format("15:04")
		list = append(list, item)
	}
	return list, rows.Err()
}

// UpdateStatus sets a new status for an appointment.
func (r *PostgresAppointmentRepository) UpdateStatus(ctx context.Context, id string, status domain.AppointmentStatus) error {
	query := `
		UPDATE appointments
		SET status = $1, updated_at = NOW()
		WHERE id = $2;
	`
	cmd, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update appointment status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrAppointmentNotFound
	}
	return nil
}
