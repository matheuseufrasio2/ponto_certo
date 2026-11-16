package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresManagerUnitSettingsRepository handles database access for unit settings and operating hours.
type PostgresManagerUnitSettingsRepository struct {
	db       *pgxpool.Pool
	auditLog *PostgresAuditLogRepository
}

// NewPostgresManagerUnitSettingsRepository instantiates a new PostgresManagerUnitSettingsRepository.
func NewPostgresManagerUnitSettingsRepository(db *pgxpool.Pool) *PostgresManagerUnitSettingsRepository {
	return &PostgresManagerUnitSettingsRepository{
		db:       db,
		auditLog: NewPostgresAuditLogRepository(db),
	}
}

// GetUnitSettings returns the branch contact details along with its 7-day operating hours.
func (r *PostgresManagerUnitSettingsRepository) GetUnitSettings(ctx context.Context, unitID string) (*domain.UnitSettingsResponse, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	unitQuery := `
		SELECT
			id::TEXT,
			name,
			COALESCE(phone, ''),
			COALESCE(address, ''),
			chair_count
		FROM barbershop_units
		WHERE id = $1;
	`

	var (
		res        domain.UnitSettingsResponse
		phoneStr   string
		addressStr string
	)

	err := r.db.QueryRow(ctx, unitQuery, unitID).Scan(
		&res.UnitID,
		&res.UnitName,
		&phoneStr,
		&addressStr,
		&res.ChairCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, fmt.Errorf("get unit profile: %w", err)
	}

	if phoneStr != "" {
		res.Phone = &phoneStr
	}
	if addressStr != "" {
		res.Address = &addressStr
	}

	hours, err := r.GetUnitOperatingHours(ctx, unitID)
	if err != nil {
		return nil, err
	}
	res.OperatingHours = hours

	return &res, nil
}

// GetUnitOperatingHours fetches the 7-day operating schedule for the unit.
func (r *PostgresManagerUnitSettingsRepository) GetUnitOperatingHours(ctx context.Context, unitID string) ([]domain.UnitOperatingHour, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	hoursQuery := `
		SELECT
			id::TEXT,
			unit_id::TEXT,
			day_of_week,
			TO_CHAR(open_time, 'HH24:MI'),
			TO_CHAR(close_time, 'HH24:MI'),
			is_closed,
			created_at,
			updated_at
		FROM unit_operating_hours
		WHERE unit_id = $1
		ORDER BY day_of_week;
	`

	rows, err := r.db.Query(ctx, hoursQuery, unitID)
	if err != nil {
		return nil, fmt.Errorf("get unit operating hours: %w", err)
	}
	defer rows.Close()

	var list []domain.UnitOperatingHour
	for rows.Next() {
		var oh domain.UnitOperatingHour
		if err := rows.Scan(
			&oh.ID,
			&oh.UnitID,
			&oh.DayOfWeek,
			&oh.OpenTime,
			&oh.CloseTime,
			&oh.IsClosed,
			&oh.CreatedAt,
			&oh.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan unit operating hour row: %w", err)
		}
		list = append(list, oh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unit operating hours: %w", err)
	}

	// If no rows found, generate default 7 days (Sunday closed, Mon-Sat 09:00 - 20:00)
	if len(list) == 0 {
		now := time.Now().UTC()
		for day := 0; day < 7; day++ {
			isClosed := day == 0
			list = append(list, domain.UnitOperatingHour{
				ID:        fmt.Sprintf("default-%s-%d", unitID, day),
				UnitID:    unitID,
				DayOfWeek: day,
				OpenTime:  "09:00",
				CloseTime: "20:00",
				IsClosed:  isClosed,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
	}

	return list, nil
}

// UpdateUnitSettings updates the phone and address of the branch, along with upserting operating hours.
func (r *PostgresManagerUnitSettingsRepository) UpdateUnitSettings(
	ctx context.Context,
	unitID string,
	input domain.UpdateUnitSettingsInput,
	audit *domain.AuditLog,
) (*domain.UnitSettingsResponse, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update unit settings tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Update barbershop_units (phone, address)
	updateUnitQuery := `
		UPDATE barbershop_units
		SET
			phone = COALESCE($2, phone),
			address = COALESCE($3, address),
			updated_at = NOW()
		WHERE id = $1;
	`

	var (
		phoneVal   *string
		addressVal *string
	)
	if input.Phone != nil {
		p := strings.TrimSpace(*input.Phone)
		phoneVal = &p
	}
	if input.Address != nil {
		a := strings.TrimSpace(*input.Address)
		addressVal = &a
	}

	cmdTag, err := tx.Exec(ctx, updateUnitQuery, unitID, phoneVal, addressVal)
	if err != nil {
		return nil, fmt.Errorf("update unit contact details: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, domain.ErrUnitNotFound
	}

	// 2. Upsert operating hours if provided
	hours := input.NormalizeHours()
	if len(hours) > 0 {
		upsertHoursQuery := `
			INSERT INTO unit_operating_hours (
				unit_id, day_of_week, open_time, close_time, is_closed, updated_at
			) VALUES (
				$1, $2, $3::TIME, $4::TIME, $5, NOW()
			)
			ON CONFLICT (unit_id, day_of_week)
			DO UPDATE SET
				open_time = EXCLUDED.open_time,
				close_time = EXCLUDED.close_time,
				is_closed = EXCLUDED.is_closed,
				updated_at = NOW();
		`

		for _, d := range hours {
			openTime := strings.TrimSpace(d.OpenTime)
			if openTime == "" {
				openTime = "09:00"
			}
			closeTime := strings.TrimSpace(d.CloseTime)
			if closeTime == "" {
				closeTime = "20:00"
			}

			if _, err := tx.Exec(ctx, upsertHoursQuery, unitID, d.DayOfWeek, openTime, closeTime, d.IsClosed); err != nil {
				return nil, fmt.Errorf("upsert unit operating hour day %d: %w", d.DayOfWeek, err)
			}
		}
	}

	// 3. Record audit log
	if audit != nil && r.auditLog != nil {
		if err := r.auditLog.RecordTx(ctx, tx, *audit); err != nil {
			return nil, fmt.Errorf("record unit settings audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update unit settings tx: %w", err)
	}

	return r.GetUnitSettings(ctx, unitID)
}
