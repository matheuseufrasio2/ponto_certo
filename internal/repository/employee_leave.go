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

// PostgresEmployeeLeaveRepository manages employee leaves and vacations in PostgreSQL.
type PostgresEmployeeLeaveRepository struct {
	db *pgxpool.Pool
}

// NewPostgresEmployeeLeaveRepository instantiates a new employee leave repository.
func NewPostgresEmployeeLeaveRepository(db *pgxpool.Pool) *PostgresEmployeeLeaveRepository {
	return &PostgresEmployeeLeaveRepository{db: db}
}

// GetUserLeaveContext fetches user role, organization, unit, and optional barber ID.
func (r *PostgresEmployeeLeaveRepository) GetUserLeaveContext(ctx context.Context, userID string) (*domain.UserLeaveContext, error) {
	query := `
		SELECT 
			u.name,
			u.role,
			u.organization_id,
			u.unit_id,
			b.id AS barber_id
		FROM users u
		LEFT JOIN barbers b ON b.user_id = u.id
		WHERE u.id = $1
		LIMIT 1;
	`
	var c domain.UserLeaveContext
	var roleStr string
	var orgID *string
	if err := r.db.QueryRow(ctx, query, userID).Scan(
		&c.UserName,
		&roleStr,
		&orgID,
		&c.UnitID,
		&c.BarberID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("fetch user leave context: %w", err)
	}
	if orgID == nil {
		return nil, domain.ErrOrganizationNotFound
	}
	c.OrganizationID = *orgID
	c.UserRole = domain.Role(roleStr)
	return &c, nil
}

// HasConflictForUser checks if the user already has a leave scheduled overlapping the range.
func (r *PostgresEmployeeLeaveRepository) HasConflictForUser(ctx context.Context, userID, startDate, endDate string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM employee_leaves
			WHERE user_id = $1
			  AND start_date <= $3::date
			  AND end_date >= $2::date
		);
	`
	var conflict bool
	if err := r.db.QueryRow(ctx, query, userID, startDate, endDate).Scan(&conflict); err != nil {
		return false, fmt.Errorf("check employee leave conflict: %w", err)
	}
	return conflict, nil
}

// HasConflictWithAppointments checks if there are active scheduled appointments overlapping the interval.
func (r *PostgresEmployeeLeaveRepository) HasConflictWithAppointments(ctx context.Context, barberID string, start, end time.Time) (bool, error) {
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
		return false, fmt.Errorf("check appointment conflict for barber leave: %w", err)
	}
	return conflict, nil
}

// Create inserts employee leave and cascades to schedule blocks/status atomically if barber.
func (r *PostgresEmployeeLeaveRepository) Create(
	ctx context.Context,
	leave *domain.EmployeeLeave,
	block *domain.BarberScheduleBlock,
	setVacation bool,
	barberID *string,
	audit *domain.AuditLog,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO employee_leaves (
			organization_id, unit_id, user_id, leave_type, start_date, end_date, reason,
			created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`
	if err := tx.QueryRow(
		ctx, query,
		leave.OrganizationID, leave.UnitID, leave.UserID, string(leave.LeaveType),
		leave.StartDate, leave.EndDate, leave.Reason, leave.CreatedByUserID,
	).Scan(&leave.ID, &leave.CreatedAt, &leave.UpdatedAt); err != nil {
		return fmt.Errorf("insert employee leave: %w", err)
	}

	if block != nil && barberID != nil && *barberID != "" {
		blockQuery := `
			INSERT INTO barber_schedule_blocks (
				barber_id, start_time, end_time, reason, created_by_user_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, NOW(), NOW());
		`
		if _, err := tx.Exec(
			ctx, blockQuery,
			*barberID, block.StartTime, block.EndTime, block.Reason, block.CreatedByUserID,
		); err != nil {
			return fmt.Errorf("insert barber schedule block for leave: %w", err)
		}
	}

	if setVacation && barberID != nil && *barberID != "" {
		if _, err := tx.Exec(ctx, `UPDATE barbers SET status = 'VACATION', updated_at = NOW() WHERE id = $1;`, *barberID); err != nil {
			return fmt.Errorf("update barber status to vacation: %w", err)
		}
	}

	if audit != nil {
		audit.TargetID = leave.ID
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
			return fmt.Errorf("insert audit log in create employee leave: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ListByOrganization lists leaves for an organization with optional unit and status filters.
func (r *PostgresEmployeeLeaveRepository) ListByOrganization(
	ctx context.Context,
	orgID string,
	unitID *string,
	status string,
) ([]domain.EmployeeLeave, error) {
	query := `
		SELECT 
			l.id,
			l.organization_id,
			l.unit_id,
			l.user_id,
			u.name AS user_name,
			u.role AS user_role,
			l.leave_type,
			l.start_date::text,
			l.end_date::text,
			l.reason,
			l.created_by_user_id,
			l.created_at,
			l.updated_at
		FROM employee_leaves l
		JOIN users u ON u.id = l.user_id
		WHERE l.organization_id = $1
		  AND ($2::uuid IS NULL OR l.unit_id = $2::uuid)
		  AND (
			  $3 = '' OR $3 = 'all'
			  OR ($3 = 'active' AND CURRENT_DATE BETWEEN l.start_date AND l.end_date)
			  OR ($3 = 'upcoming' AND l.start_date > CURRENT_DATE)
			  OR ($3 = 'past' AND l.end_date < CURRENT_DATE)
		  )
		ORDER BY l.start_date DESC;
	`

	rows, err := r.db.Query(ctx, query, orgID, unitID, status)
	if err != nil {
		return nil, fmt.Errorf("query employee leaves: %w", err)
	}
	defer rows.Close()

	var list []domain.EmployeeLeave
	for rows.Next() {
		var l domain.EmployeeLeave
		var roleStr string
		var leaveTypeStr string
		if err := rows.Scan(
			&l.ID,
			&l.OrganizationID,
			&l.UnitID,
			&l.UserID,
			&l.UserName,
			&roleStr,
			&leaveTypeStr,
			&l.StartDate,
			&l.EndDate,
			&l.Reason,
			&l.CreatedByUserID,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan employee leave: %w", err)
		}
		l.UserRole = domain.Role(roleStr)
		l.LeaveType = domain.LeaveType(leaveTypeStr)
		list = append(list, l)
	}
	return list, rows.Err()
}

// GetByID finds an employee leave by ID.
func (r *PostgresEmployeeLeaveRepository) GetByID(ctx context.Context, leaveID string) (*domain.EmployeeLeave, error) {
	query := `
		SELECT 
			l.id,
			l.organization_id,
			l.unit_id,
			l.user_id,
			u.name AS user_name,
			u.role AS user_role,
			l.leave_type,
			l.start_date::text,
			l.end_date::text,
			l.reason,
			l.created_by_user_id,
			l.created_at,
			l.updated_at
		FROM employee_leaves l
		JOIN users u ON u.id = l.user_id
		WHERE l.id = $1
		LIMIT 1;
	`
	var l domain.EmployeeLeave
	var roleStr, leaveTypeStr string
	if err := r.db.QueryRow(ctx, query, leaveID).Scan(
		&l.ID,
		&l.OrganizationID,
		&l.UnitID,
		&l.UserID,
		&l.UserName,
		&roleStr,
		&leaveTypeStr,
		&l.StartDate,
		&l.EndDate,
		&l.Reason,
		&l.CreatedByUserID,
		&l.CreatedAt,
		&l.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeLeaveNotFound
		}
		return nil, fmt.Errorf("get employee leave by id: %w", err)
	}
	l.UserRole = domain.Role(roleStr)
	l.LeaveType = domain.LeaveType(leaveTypeStr)
	return &l, nil
}

// Delete removes an employee leave and writes audit log atomically.
func (r *PostgresEmployeeLeaveRepository) Delete(ctx context.Context, leaveID string, audit *domain.AuditLog) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var userID string
	err = tx.QueryRow(ctx, `DELETE FROM employee_leaves WHERE id = $1 RETURNING user_id;`, leaveID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEmployeeLeaveNotFound
		}
		return fmt.Errorf("delete employee leave: %w", err)
	}

	// If user is a barber and currently has status VACATION, restore to AVAILABLE
	var barberID *string
	_ = tx.QueryRow(ctx, `SELECT id FROM barbers WHERE user_id = $1;`, userID).Scan(&barberID)
	if barberID != nil {
		_, _ = tx.Exec(ctx, `UPDATE barbers SET status = 'AVAILABLE', updated_at = NOW() WHERE id = $1 AND status = 'VACATION';`, *barberID)
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
			return fmt.Errorf("insert audit log in delete employee leave: %w", err)
		}
	}

	return tx.Commit(ctx)
}
