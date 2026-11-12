package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresOwnerManagersRepository handles data access for owner-managed unit managers.
type PostgresOwnerManagersRepository struct {
	db       *pgxpool.Pool
	auditLog *PostgresAuditLogRepository
}

// NewPostgresOwnerManagersRepository instantiates a new PostgresOwnerManagersRepository.
func NewPostgresOwnerManagersRepository(db *pgxpool.Pool) *PostgresOwnerManagersRepository {
	return &PostgresOwnerManagersRepository{
		db:       db,
		auditLog: NewPostgresAuditLogRepository(db),
	}
}

// UnitBelongsToOrg checks whether a unit belongs to the given organization.
func (r *PostgresOwnerManagersRepository) UnitBelongsToOrg(ctx context.Context, unitID, orgID string) (bool, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(orgID) {
		return false, domain.ErrInvalidUUID
	}
	query := `SELECT EXISTS(SELECT 1 FROM barbershop_units WHERE id = $1 AND organization_id = $2);`
	var exists bool
	if err := r.db.QueryRow(ctx, query, unitID, orgID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check unit org membership: %w", err)
	}
	return exists, nil
}

// EmailExists checks whether an email is already registered in the users table.
func (r *PostgresOwnerManagersRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1));`
	var exists bool
	if err := r.db.QueryRow(ctx, query, strings.TrimSpace(email)).Scan(&exists); err != nil {
		return false, fmt.Errorf("check email existence: %w", err)
	}
	return exists, nil
}

// ListManagers returns all UNIT_MANAGER users under the given organization.
func (r *PostgresOwnerManagersRepository) ListManagers(ctx context.Context, orgID string) ([]domain.ManagerListItem, error) {
	if !domain.IsValidUUID(orgID) {
		return nil, domain.ErrInvalidUUID
	}
	query := `
		SELECT
			u.id::TEXT,
			u.name,
			u.email,
			COALESCE(u.phone, ''),
			u.unit_id::TEXT,
			bu.name AS unit_name,
			u.is_active,
			u.created_at
		FROM users u
		JOIN barbershop_units bu ON bu.id = u.unit_id
		WHERE u.organization_id = $1 AND u.role = 'UNIT_MANAGER'
		ORDER BY bu.name, u.name;
	`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list managers: %w", err)
	}
	defer rows.Close()

	var managers []domain.ManagerListItem
	for rows.Next() {
		var m domain.ManagerListItem
		if err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Email,
			&m.Phone,
			&m.UnitID,
			&m.UnitName,
			&m.IsActive,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan manager row: %w", err)
		}
		managers = append(managers, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list managers rows: %w", err)
	}

	if managers == nil {
		managers = []domain.ManagerListItem{}
	}

	return managers, nil
}

// CreateManager inserts a new UNIT_MANAGER user in a transaction with an audit log entry.
func (r *PostgresOwnerManagersRepository) CreateManager(
	ctx context.Context,
	orgID string,
	unitID string,
	input domain.CreateManagerInput,
	passwordHash string,
	audit *domain.AuditLog,
) (*domain.ManagerListItem, error) {
	if !domain.IsValidUUID(orgID) || !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Check email uniqueness inside the transaction for correctness.
	var emailExists bool
	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))
	emailCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = $1);`
	if err := tx.QueryRow(ctx, emailCheck, normalizedEmail).Scan(&emailExists); err != nil {
		return nil, fmt.Errorf("check email in tx: %w", err)
	}
	if emailExists {
		return nil, domain.ErrManagerEmailTaken
	}

	insertUser := `
		INSERT INTO users (organization_id, unit_id, name, email, phone, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'UNIT_MANAGER', TRUE, NOW(), NOW())
		RETURNING id::TEXT, created_at;
	`
	var userID string
	var manager domain.ManagerListItem

	err = tx.QueryRow(ctx, insertUser,
		orgID,
		unitID,
		strings.TrimSpace(input.Name),
		normalizedEmail,
		strings.TrimSpace(input.Phone),
		passwordHash,
	).Scan(&userID, &manager.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrManagerEmailTaken
		}
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "users_email_key") {
			return nil, domain.ErrManagerEmailTaken
		}
		return nil, fmt.Errorf("insert manager user: %w", err)
	}

	// Fetch unit name for the response.
	var unitName string
	if err := tx.QueryRow(ctx, `SELECT name FROM barbershop_units WHERE id = $1`, unitID).Scan(&unitName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, fmt.Errorf("fetch unit name: %w", err)
	}

	if audit != nil {
		audit.TargetID = userID
		audit.UnitID = &unitID
		if err := r.auditLog.RecordTx(ctx, tx, *audit); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create manager: %w", err)
	}

	manager.ID = userID
	manager.Name = strings.TrimSpace(input.Name)
	manager.Email = strings.ToLower(strings.TrimSpace(input.Email))
	manager.Phone = strings.TrimSpace(input.Phone)
	manager.UnitID = unitID
	manager.UnitName = unitName
	manager.IsActive = true

	return &manager, nil
}
