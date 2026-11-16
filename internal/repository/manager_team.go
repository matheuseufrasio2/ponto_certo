package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresManagerTeamRepository manages database operations for unit team members and barber schedules.
type PostgresManagerTeamRepository struct {
	db       *pgxpool.Pool
	auditLog *PostgresAuditLogRepository
}

// NewPostgresManagerTeamRepository instantiates a new PostgresManagerTeamRepository.
func NewPostgresManagerTeamRepository(db *pgxpool.Pool) *PostgresManagerTeamRepository {
	return &PostgresManagerTeamRepository{
		db:       db,
		auditLog: NewPostgresAuditLogRepository(db),
	}
}

// ListTeamMembers lists all team members (barbers and receptionists) belonging to a branch unit.
func (r *PostgresManagerTeamRepository) ListTeamMembers(ctx context.Context, unitID string) ([]domain.TeamMemberItem, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT
			u.id::TEXT,
			b.id::TEXT,
			u.unit_id::TEXT,
			u.name,
			u.email,
			COALESCE(u.phone, ''),
			u.role,
			b.commission_percentage,
			b.status,
			u.is_active,
			u.created_at
		FROM users u
		LEFT JOIN barbers b ON b.user_id = u.id AND b.unit_id = u.unit_id
		WHERE u.unit_id = $1 AND u.role IN ('BARBER', 'RECEPTIONIST')
		ORDER BY u.role, u.name;
	`

	rows, err := r.db.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	var members []domain.TeamMemberItem
	for rows.Next() {
		var (
			m            domain.TeamMemberItem
			barberID     *string
			commPct      *int
			barberStatus *domain.BarberStatus
		)

		if err := rows.Scan(
			&m.ID,
			&barberID,
			&m.UnitID,
			&m.Name,
			&m.Email,
			&m.Phone,
			&m.Role,
			&commPct,
			&barberStatus,
			&m.IsActive,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan team member row: %w", err)
		}

		m.BarberID = barberID
		m.CommissionPercentage = commPct
		m.Status = barberStatus
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members rows: %w", err)
	}

	return members, nil
}

// CreateTeamMember creates a new team member under the given unit and organization.
// If the role is BARBER, a corresponding record in the barbers table is created along with default working hours.
func (r *PostgresManagerTeamRepository) CreateTeamMember(
	ctx context.Context,
	orgID, unitID string,
	input domain.CreateTeamMemberInput,
	passwordHash string,
	audit *domain.AuditLog,
) (*domain.TeamMemberItem, error) {
	if !domain.IsValidUUID(orgID) || !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create team member tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Insert user
	userQuery := `
		INSERT INTO users (
			organization_id, unit_id, name, email, password_hash, role, phone, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, TRUE
		)
		RETURNING id::TEXT, created_at;
	`

	trimmedEmail := strings.TrimSpace(strings.ToLower(input.Email))
	trimmedPhone := strings.TrimSpace(input.Phone)
	trimmedName := strings.TrimSpace(input.Name)

	var (
		userID    string
		createdAt time.Time
	)

	err = tx.QueryRow(ctx, userQuery, orgID, unitID, trimmedName, trimmedEmail, passwordHash, string(input.Role), trimmedPhone).
		Scan(&userID, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrManagerEmailTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	item := &domain.TeamMemberItem{
		ID:        userID,
		UnitID:    unitID,
		Name:      trimmedName,
		Email:     trimmedEmail,
		Phone:     trimmedPhone,
		Role:      domain.Role(input.Role),
		IsActive:  true,
		CreatedAt: createdAt,
	}

	// 2. If role == BARBER, insert barber profile and default working hours
	if input.Role == domain.TeamRoleBarber {
		comm := 50
		if input.CommissionPercentage != nil {
			comm = *input.CommissionPercentage
		}

		barberQuery := `
			INSERT INTO barbers (
				unit_id, user_id, name, phone, commission_percentage, status, is_active
			) VALUES (
				$1, $2, $3, $4, $5, 'AVAILABLE', TRUE
			)
			RETURNING id::TEXT;
		`

		var barberID string
		err = tx.QueryRow(ctx, barberQuery, unitID, userID, trimmedName, trimmedPhone, comm).Scan(&barberID)
		if err != nil {
			return nil, fmt.Errorf("insert barber profile: %w", err)
		}

		item.BarberID = &barberID
		item.CommissionPercentage = &comm
		statusAvailable := domain.BarberStatusAvailable
		item.Status = &statusAvailable

		// Populate standard 7 days working hours for the barber (Sunday off, Mon-Sat 09:00 - 19:00, lunch 12:00 - 13:00)
		for day := 0; day < 7; day++ {
			isOff := day == 0
			start := "09:00:00"
			end := "19:00:00"
			var bStart, bEnd *string
			if !isOff {
				bs := "12:00:00"
				be := "13:00:00"
				bStart = &bs
				bEnd = &be
			}

			hoursQuery := `
				INSERT INTO barber_working_hours (
					barber_id, day_of_week, start_time, end_time, break_start_time, break_end_time, is_off
				) VALUES (
					$1, $2, $3::TIME, $4::TIME, $5::TIME, $6::TIME, $7
				);
			`
			if _, err := tx.Exec(ctx, hoursQuery, barberID, day, start, end, bStart, bEnd, isOff); err != nil {
				return nil, fmt.Errorf("insert initial barber working hours for day %d: %w", day, err)
			}
		}
	}

	// 3. Record audit log
	if audit != nil && r.auditLog != nil {
		if err := r.auditLog.RecordTx(ctx, tx, *audit); err != nil {
			return nil, fmt.Errorf("record team member audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit team member tx: %w", err)
	}

	return item, nil
}

// ResolveBarberID resolves a barber UUID from either a direct barber ID or user ID within the given unit.
func (r *PostgresManagerTeamRepository) ResolveBarberID(ctx context.Context, unitID, identifier string) (string, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(identifier) {
		return "", domain.ErrInvalidUUID
	}

	// First try matching as direct barber ID
	var barberID string
	err := r.db.QueryRow(ctx, "SELECT id::TEXT FROM barbers WHERE id = $1 AND unit_id = $2;", identifier, unitID).Scan(&barberID)
	if err == nil {
		return barberID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("query barber by id: %w", err)
	}

	// Next try matching as user ID
	err = r.db.QueryRow(ctx, "SELECT id::TEXT FROM barbers WHERE user_id = $1 AND unit_id = $2;", identifier, unitID).Scan(&barberID)
	if err == nil {
		return barberID, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrBarberNotFound
	}

	return "", fmt.Errorf("query barber by user_id: %w", err)
}

// GetBarberWorkingHours fetches weekly working hours for the specified barber.
func (r *PostgresManagerTeamRepository) GetBarberWorkingHours(ctx context.Context, barberID string) ([]domain.BarberWorkingHour, error) {
	if !domain.IsValidUUID(barberID) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT
			id::TEXT,
			barber_id::TEXT,
			day_of_week,
			TO_CHAR(start_time, 'HH24:MI'),
			TO_CHAR(end_time, 'HH24:MI'),
			TO_CHAR(break_start_time, 'HH24:MI'),
			TO_CHAR(break_end_time, 'HH24:MI'),
			is_off,
			created_at,
			updated_at
		FROM barber_working_hours
		WHERE barber_id = $1
		ORDER BY day_of_week;
	`

	rows, err := r.db.Query(ctx, query, barberID)
	if err != nil {
		return nil, fmt.Errorf("get barber working hours: %w", err)
	}
	defer rows.Close()

	var list []domain.BarberWorkingHour
	for rows.Next() {
		var (
			wh         domain.BarberWorkingHour
			bStartStr  *string
			bEndStr    *string
		)

		if err := rows.Scan(
			&wh.ID,
			&wh.BarberID,
			&wh.DayOfWeek,
			&wh.StartTime,
			&wh.EndTime,
			&bStartStr,
			&bEndStr,
			&wh.IsOff,
			&wh.CreatedAt,
			&wh.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan barber working hour row: %w", err)
		}

		wh.BreakStartTime = bStartStr
		wh.BreakEndTime = bEndStr
		list = append(list, wh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate barber working hour rows: %w", err)
	}

	// Fallback to default 7 days if no rows exist yet
	if len(list) == 0 {
		now := time.Now().UTC()
		for day := 0; day < 7; day++ {
			isOff := day == 0
			var bStart, bEnd *string
			if !isOff {
				bs := "12:00"
				be := "13:00"
				bStart = &bs
				bEnd = &be
			}
			list = append(list, domain.BarberWorkingHour{
				ID:             fmt.Sprintf("default-%s-%d", barberID, day),
				BarberID:       barberID,
				DayOfWeek:      day,
				StartTime:      "09:00",
				EndTime:        "19:00",
				BreakStartTime: bStart,
				BreakEndTime:   bEnd,
				IsOff:          isOff,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
		}
	}

	return list, nil
}

// UpsertBarberWorkingHours updates or inserts schedule entries for all 7 days of the week.
func (r *PostgresManagerTeamRepository) UpsertBarberWorkingHours(
	ctx context.Context,
	barberID string,
	hours []domain.UpdateWorkingHoursDayInput,
	audit *domain.AuditLog,
) error {
	if !domain.IsValidUUID(barberID) {
		return domain.ErrInvalidUUID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin upsert working hours tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO barber_working_hours (
			barber_id, day_of_week, start_time, end_time, break_start_time, break_end_time, is_off, updated_at
		) VALUES (
			$1, $2, $3::TIME, $4::TIME, $5::TIME, $6::TIME, $7, NOW()
		)
		ON CONFLICT (barber_id, day_of_week)
		DO UPDATE SET
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			break_start_time = EXCLUDED.break_start_time,
			break_end_time = EXCLUDED.break_end_time,
			is_off = EXCLUDED.is_off,
			updated_at = NOW();
	`

	for _, d := range hours {
		var bStart, bEnd *string
		if !d.IsOff && d.BreakStartTime != nil && d.BreakEndTime != nil {
			trimmedStart := strings.TrimSpace(*d.BreakStartTime)
			trimmedEnd := strings.TrimSpace(*d.BreakEndTime)
			if trimmedStart != "" && trimmedEnd != "" {
				bStart = &trimmedStart
				bEnd = &trimmedEnd
			}
		}

		startTime := strings.TrimSpace(d.StartTime)
		if startTime == "" {
			startTime = "09:00"
		}
		endTime := strings.TrimSpace(d.EndTime)
		if endTime == "" {
			endTime = "19:00"
		}

		if _, err := tx.Exec(ctx, query, barberID, d.DayOfWeek, startTime, endTime, bStart, bEnd, d.IsOff); err != nil {
			return fmt.Errorf("upsert barber working hours day %d: %w", d.DayOfWeek, err)
		}
	}

	if audit != nil && r.auditLog != nil {
		if err := r.auditLog.RecordTx(ctx, tx, *audit); err != nil {
			return fmt.Errorf("record working hours audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit upsert working hours tx: %w", err)
	}

	return nil
}
