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

// PostgresScheduleBlockRepository manages barber schedule blocks in PostgreSQL.
type PostgresScheduleBlockRepository struct {
	db *pgxpool.Pool
}

// NewPostgresScheduleBlockRepository instantiates a new schedule block repository.
func NewPostgresScheduleBlockRepository(db *pgxpool.Pool) *PostgresScheduleBlockRepository {
	return &PostgresScheduleBlockRepository{db: db}
}

// GetBarberInfo fetches the unit_id and organization_id of a given barber.
func (r *PostgresScheduleBlockRepository) GetBarberInfo(ctx context.Context, barberID string) (unitID, orgID string, err error) {
	query := `
		SELECT b.unit_id, u.organization_id
		FROM barbers b
		JOIN barbershop_units u ON u.id = b.unit_id
		WHERE b.id = $1
		LIMIT 1;
	`
	if err := r.db.QueryRow(ctx, query, barberID).Scan(&unitID, &orgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", domain.ErrBarberNotFound
		}
		return "", "", fmt.Errorf("fetch barber info: %w", err)
	}
	return unitID, orgID, nil
}

// HasConflictWithAppointments checks if there are active scheduled appointments overlapping the interval.
func (r *PostgresScheduleBlockRepository) HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM appointments
			WHERE barber_id = $1
			  AND status = 'SCHEDULED'
			  AND start_time < $3
			  AND end_time > $2
		);
	`
	var conflict bool
	if err := r.db.QueryRow(ctx, query, barberID, start, end).Scan(&conflict); err != nil {
		return false, fmt.Errorf("check schedule block conflict with appointments: %w", err)
	}
	return conflict, nil
}

// Create inserts a schedule block and an optional audit log atomically.
func (r *PostgresScheduleBlockRepository) Create(ctx context.Context, block *domain.BarberScheduleBlock, audit *domain.AuditLog) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO barber_schedule_blocks (
			barber_id, start_time, end_time, reason, created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`
	if err := tx.QueryRow(
		ctx, query,
		block.BarberID, block.StartTime, block.EndTime, block.Reason, block.CreatedByUserID,
	).Scan(&block.ID, &block.CreatedAt, &block.UpdatedAt); err != nil {
		return fmt.Errorf("insert schedule block: %w", err)
	}

	if audit != nil {
		audit.TargetID = block.ID
		auditQuery := `
			INSERT INTO audit_logs (
				organization_id, unit_id, actor_id, actor_role, action, target_entity, target_id,
				details, ip_address, user_agent, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
		`
		detailsJSON := audit.Details
		if len(detailsJSON) == 0 {
			detailsJSON = []byte("{}")
		}
		if _, err := tx.Exec(ctx, auditQuery,
			audit.OrganizationID,
			audit.UnitID,
			audit.ActorID,
			string(audit.ActorRole),
			audit.Action,
			audit.TargetEntity,
			audit.TargetID,
			detailsJSON,
			audit.IPAddress,
			audit.UserAgent,
		); err != nil {
			return fmt.Errorf("insert audit log in create schedule block: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ListByBarber retrieves schedule blocks for a barber in an optional interval.
func (r *PostgresScheduleBlockRepository) ListByBarber(ctx context.Context, barberID string, from, to time.Time) ([]domain.BarberScheduleBlock, error) {
	var query string
	var args []any

	if !from.IsZero() && !to.IsZero() {
		query = `
			SELECT id, barber_id, start_time, end_time, reason, created_by_user_id, created_at, updated_at
			FROM barber_schedule_blocks
			WHERE barber_id = $1
			  AND start_time < $3
			  AND end_time > $2
			ORDER BY start_time ASC;
		`
		args = []any{barberID, from, to}
	} else {
		query = `
			SELECT id, barber_id, start_time, end_time, reason, created_by_user_id, created_at, updated_at
			FROM barber_schedule_blocks
			WHERE barber_id = $1
			ORDER BY start_time ASC;
		`
		args = []any{barberID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schedule blocks: %w", err)
	}
	defer rows.Close()

	var list []domain.BarberScheduleBlock
	for rows.Next() {
		var b domain.BarberScheduleBlock
		if err := rows.Scan(
			&b.ID,
			&b.BarberID,
			&b.StartTime,
			&b.EndTime,
			&b.Reason,
			&b.CreatedByUserID,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan schedule block: %w", err)
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

// GetByID finds a schedule block by its ID.
func (r *PostgresScheduleBlockRepository) GetByID(ctx context.Context, blockID string) (*domain.BarberScheduleBlock, error) {
	query := `
		SELECT id, barber_id, start_time, end_time, reason, created_by_user_id, created_at, updated_at
		FROM barber_schedule_blocks
		WHERE id = $1
		LIMIT 1;
	`
	var b domain.BarberScheduleBlock
	if err := r.db.QueryRow(ctx, query, blockID).Scan(
		&b.ID,
		&b.BarberID,
		&b.StartTime,
		&b.EndTime,
		&b.Reason,
		&b.CreatedByUserID,
		&b.CreatedAt,
		&b.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleBlockNotFound
		}
		return nil, fmt.Errorf("get schedule block by id: %w", err)
	}
	return &b, nil
}

// Delete removes a schedule block by ID and inserts an audit log atomically.
func (r *PostgresScheduleBlockRepository) Delete(ctx context.Context, blockID string, audit *domain.AuditLog) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(ctx, `DELETE FROM barber_schedule_blocks WHERE id = $1;`, blockID)
	if err != nil {
		return fmt.Errorf("delete schedule block: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrScheduleBlockNotFound
	}

	if audit != nil {
		auditQuery := `
			INSERT INTO audit_logs (
				organization_id, unit_id, actor_id, actor_role, action, target_entity, target_id,
				details, ip_address, user_agent, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
		`
		detailsJSON := audit.Details
		if len(detailsJSON) == 0 {
			detailsJSON = []byte("{}")
		}
		if _, err := tx.Exec(ctx, auditQuery,
			audit.OrganizationID,
			audit.UnitID,
			audit.ActorID,
			string(audit.ActorRole),
			audit.Action,
			audit.TargetEntity,
			audit.TargetID,
			detailsJSON,
			audit.IPAddress,
			audit.UserAgent,
		); err != nil {
			return fmt.Errorf("insert audit log in delete schedule block: %w", err)
		}
	}

	return tx.Commit(ctx)
}
