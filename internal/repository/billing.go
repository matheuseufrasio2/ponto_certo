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

// BillingRepository defines operations for billing transactions and checkouts.
type BillingRepository interface {
	CreateCheckout(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error)
}

// PostgresBillingRepository implements BillingRepository using PostgreSQL transactions.
type PostgresBillingRepository struct {
	db *pgxpool.Pool
}

// NewPostgresBillingRepository instantiates a new billing repository.
func NewPostgresBillingRepository(db *pgxpool.Pool) *PostgresBillingRepository {
	return &PostgresBillingRepository{db: db}
}

// CreateCheckout performs atomic checkout for an appointment: creates a Billing record and marks appointment as completed and paid.
func (r *PostgresBillingRepository) CreateCheckout(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin checkout tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Lock appointment row
	lockQuery := `
		SELECT id, unit_id, barber_id, client_id, service_id, price_in_cents, status, payment_status
		FROM appointments
		WHERE id = $1
		FOR UPDATE;
	`
	var (
		aptID, unitID, barberID, clientID, serviceID string
		priceInCents                                 int64
		status                                       domain.AppointmentStatus
		paymentStatus                                domain.PaymentStatus
	)
	err = tx.QueryRow(ctx, lockQuery, input.AppointmentID).Scan(
		&aptID,
		&unitID,
		&barberID,
		&clientID,
		&serviceID,
		&priceInCents,
		&status,
		&paymentStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("lock appointment: %w", err)
	}

	if paymentStatus == domain.PaymentStatusPaid {
		return nil, domain.ErrBillingAlreadyProcessed
	}

	// 2. Insert into billings table
	insertBillingQuery := `
		INSERT INTO billings (
			unit_id, date, barber_id, client_id, service_id, appointment_id,
			quantity, amount_in_cents, payment_method, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, 1, $7, $8, 'PAID', $9, NOW(), NOW())
		RETURNING id, unit_id, date, barber_id, client_id, service_id, appointment_id,
		          quantity, amount_in_cents, payment_method, status, notes, created_at, updated_at;
	`
	var (
		b          domain.Billing
		notes      *string
		bServiceID *string
		bAptID     *string
	)
	today := time.Now().UTC()
	err = tx.QueryRow(
		ctx,
		insertBillingQuery,
		unitID,
		today,
		barberID,
		clientID,
		serviceID,
		input.AppointmentID,
		priceInCents,
		input.PaymentMethod,
		input.Notes,
	).Scan(
		&b.ID,
		&b.UnitID,
		&b.Date,
		&b.BarberID,
		&b.ClientID,
		&bServiceID,
		&bAptID,
		&b.Quantity,
		&b.AmountInCents,
		&b.PaymentMethod,
		&b.Status,
		&notes,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert billing record: %w", err)
	}
	if notes != nil {
		b.Notes = *notes
	}
	b.ServiceID = bServiceID
	b.AppointmentID = bAptID

	// 3. Update appointment to COMPLETED and PAID
	updateAptQuery := `
		UPDATE appointments
		SET status = 'COMPLETED', payment_status = 'PAID', updated_at = NOW()
		WHERE id = $1;
	`
	_, err = tx.Exec(ctx, updateAptQuery, input.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("update appointment status on checkout: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit checkout tx: %w", err)
	}

	return &b, nil
}
