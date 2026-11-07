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

// UnitOperationsRepository defines SQL interactions for unit-level operational governance.
type UnitOperationsRepository interface {
	GetDailyCashClosing(ctx context.Context, unitID string, date time.Time) (*domain.CashClosingSummary, error)
	GetManagerLiveMetrics(ctx context.Context, unitID string, date time.Time) (*domain.ManagerDashboardMetrics, error)
	UpdateBarberStatus(ctx context.Context, barberID string, status domain.BarberStatus, changedByUserID *string) error
	ToggleServiceUnitAvailability(ctx context.Context, unitID, serviceID string, isActive bool) error
	ListBarbersByUnit(ctx context.Context, unitID string) ([]domain.Barber, error)
	CountFutureAppointmentsForBarber(ctx context.Context, barberID string, fromTime time.Time) (int, error)
}

// PostgresUnitOperationsRepository implements UnitOperationsRepository with pgxpool.
type PostgresUnitOperationsRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUnitOperationsRepository instantiates a new unit operations repository.
func NewPostgresUnitOperationsRepository(db *pgxpool.Pool) *PostgresUnitOperationsRepository {
	return &PostgresUnitOperationsRepository{db: db}
}

// GetDailyCashClosing aggregates paid billings for a specific unit and date by payment method.
func (r *PostgresUnitOperationsRepository) GetDailyCashClosing(ctx context.Context, unitID string, date time.Time) (*domain.CashClosingSummary, error) {
	query := `
		SELECT 
			payment_method,
			COUNT(*) AS tx_count,
			COALESCE(SUM(amount_in_cents), 0) AS total_cents
		FROM billings
		WHERE unit_id = $1 
		  AND date = $2 
		  AND status = 'PAID'
		GROUP BY payment_method;
	`

	rows, err := r.db.Query(ctx, query, unitID, date)
	if err != nil {
		return nil, fmt.Errorf("query daily cash closing: %w", err)
	}
	defer rows.Close()

	summary := &domain.CashClosingSummary{
		Date:   date,
		UnitID: unitID,
		ByPaymentMethod: map[domain.PaymentMethod]int64{
			domain.PaymentMethodPix:        0,
			domain.PaymentMethodCreditCard: 0,
			domain.PaymentMethodDebitCard:  0,
			domain.PaymentMethodCash:       0,
			domain.PaymentMethodOther:      0,
		},
	}

	for rows.Next() {
		var method domain.PaymentMethod
		var txCount int
		var totalCents int64

		if err := rows.Scan(&method, &txCount, &totalCents); err != nil {
			return nil, fmt.Errorf("scan cash closing row: %w", err)
		}

		summary.ByPaymentMethod[method] = totalCents
		summary.TotalAmountInCents += totalCents
		summary.TotalTransactions += txCount
	}

	return summary, rows.Err()
}

// GetManagerLiveMetrics computes real-time operational statistics for the unit manager dashboard.
func (r *PostgresUnitOperationsRepository) GetManagerLiveMetrics(ctx context.Context, unitID string, date time.Time) (*domain.ManagerDashboardMetrics, error) {
	metrics := &domain.ManagerDashboardMetrics{}

	// 1. Today Revenue from PAID billings
	revQuery := `
		SELECT COALESCE(SUM(amount_in_cents), 0)
		FROM billings
		WHERE unit_id = $1 AND date = $2 AND status = 'PAID';
	`
	if err := r.db.QueryRow(ctx, revQuery, unitID, date).Scan(&metrics.TodayRevenueInCents); err != nil {
		return nil, fmt.Errorf("query today revenue: %w", err)
	}

	// 2. Today Appointments count (non-cancelled)
	aptsQuery := `
		SELECT COUNT(*)
		FROM appointments
		WHERE unit_id = $1 AND DATE(start_time) = $2 AND status != 'CANCELLED';
	`
	if err := r.db.QueryRow(ctx, aptsQuery, unitID, date).Scan(&metrics.TodayAppointmentsCount); err != nil {
		return nil, fmt.Errorf("query today appointments count: %w", err)
	}

	// 3. Active barbers on duty / available / busy
	barbersQuery := `
		SELECT COUNT(*)
		FROM barbers
		WHERE unit_id = $1 AND is_active = TRUE AND status IN ('AVAILABLE', 'BUSY');
	`
	if err := r.db.QueryRow(ctx, barbersQuery, unitID).Scan(&metrics.ActiveBarbersCount); err != nil {
		return nil, fmt.Errorf("query active barbers: %w", err)
	}

	// 4. Waiting clients (appointments scheduled right now)
	waitingQuery := `
		SELECT COUNT(*)
		FROM appointments
		WHERE unit_id = $1 
		  AND DATE(start_time) = $2 
		  AND status = 'SCHEDULED' 
		  AND start_time <= NOW() 
		  AND end_time >= NOW();
	`
	if err := r.db.QueryRow(ctx, waitingQuery, unitID, date).Scan(&metrics.WaitingClientsCount); err != nil {
		return nil, fmt.Errorf("query waiting clients: %w", err)
	}

	// 5. Chair Occupancy Rate Percent
	var chairCount int
	chairQuery := `SELECT COALESCE(chair_count, 4) FROM barbershop_units WHERE id = $1;`
	if err := r.db.QueryRow(ctx, chairQuery, unitID).Scan(&chairCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			chairCount = 4
		} else {
			return nil, fmt.Errorf("query unit chair count: %w", err)
		}
	}

	if chairCount > 0 && metrics.WaitingClientsCount > 0 {
		metrics.ChairOccupancyRatePercent = (metrics.WaitingClientsCount * 100) / chairCount
		if metrics.ChairOccupancyRatePercent > 100 {
			metrics.ChairOccupancyRatePercent = 100
		}
	} else {
		metrics.ChairOccupancyRatePercent = 0
	}

	return metrics, nil
}

// UpdateBarberStatus updates barber operational shift status and writes an audit log.
func (r *PostgresUnitOperationsRepository) UpdateBarberStatus(ctx context.Context, barberID string, status domain.BarberStatus, changedByUserID *string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var prevStatus domain.BarberStatus
	fetchQuery := `SELECT status FROM barbers WHERE id = $1 FOR UPDATE;`
	if err := tx.QueryRow(ctx, fetchQuery, barberID).Scan(&prevStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrBarberNotFound
		}
		return fmt.Errorf("fetch previous barber status: %w", err)
	}

	updateQuery := `UPDATE barbers SET status = $1, updated_at = NOW() WHERE id = $2;`
	tag, err := tx.Exec(ctx, updateQuery, status, barberID)
	if err != nil {
		return fmt.Errorf("update barber status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBarberNotFound
	}

	logQuery := `
		INSERT INTO barber_status_logs (barber_id, previous_status, new_status, changed_by_user_id, created_at)
		VALUES ($1, $2, $3, $4, NOW());
	`
	if _, err := tx.Exec(ctx, logQuery, barberID, prevStatus, status, changedByUserID); err != nil {
		return fmt.Errorf("insert barber status log: %w", err)
	}

	return tx.Commit(ctx)
}

// ToggleServiceUnitAvailability pauses or activates a service at the unit level.
func (r *PostgresUnitOperationsRepository) ToggleServiceUnitAvailability(ctx context.Context, unitID, serviceID string, isActive bool) error {
	query := `
		UPDATE services 
		SET is_active = $1, updated_at = NOW() 
		WHERE id = $2 AND (unit_id = $3 OR unit_id IS NULL);
	`
	tag, err := r.db.Exec(ctx, query, isActive, serviceID, unitID)
	if err != nil {
		return fmt.Errorf("toggle service availability: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrServiceNotFound
	}
	return nil
}

// ListBarbersByUnit retrieves all barbers for a branch including their live shift status.
func (r *PostgresUnitOperationsRepository) ListBarbersByUnit(ctx context.Context, unitID string) ([]domain.Barber, error) {
	query := `
		SELECT id, unit_id, user_id, name, phone, commission_percentage, status, is_active, created_at, updated_at
		FROM barbers
		WHERE unit_id = $1 AND is_active = TRUE
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("query barbers by unit: %w", err)
	}
	defer rows.Close()

	var barbers []domain.Barber
	for rows.Next() {
		var b domain.Barber
		var phone *string
		if err := rows.Scan(
			&b.ID,
			&b.UnitID,
			&b.UserID,
			&b.Name,
			&phone,
			&b.CommissionPercentage,
			&b.Status,
			&b.IsActive,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan barber: %w", err)
		}
		if phone != nil {
			b.Phone = *phone
		}
		barbers = append(barbers, b)
	}

	return barbers, rows.Err()
}

// CountFutureAppointmentsForBarber counts pending scheduled appointments starting at or after fromTime.
func (r *PostgresUnitOperationsRepository) CountFutureAppointmentsForBarber(ctx context.Context, barberID string, fromTime time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM appointments
		WHERE barber_id = $1 AND status = 'SCHEDULED' AND start_time >= $2;
	`
	var count int
	if err := r.db.QueryRow(ctx, query, barberID, fromTime).Scan(&count); err != nil {
		return 0, fmt.Errorf("count future appointments for barber: %w", err)
	}
	return count, nil
}
