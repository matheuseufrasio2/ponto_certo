package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT 
			u.id, 
			u.organization_id, 
			u.unit_id, 
			u.name, 
			u.email, 
			u.password_hash, 
			u.role, 
			u.is_active, 
			u.created_at, 
			u.updated_at,
			o.is_active AS org_is_active,
			o.subscription_status AS org_subscription_status
		FROM users u
		LEFT JOIN organizations o ON o.id = u.organization_id
		WHERE u.email = $1
		LIMIT 1;
	`

	var user domain.User
	var roleStr string
	var orgIsActive *bool
	var orgSubStatus *string

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.UnitID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&roleStr,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&orgIsActive,
		&orgSubStatus,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	user.Role = domain.Role(roleStr)
	user.OrganizationIsActive = orgIsActive
	if orgSubStatus != nil {
		subStatus := domain.SubscriptionStatus(*orgSubStatus)
		user.SubscriptionStatus = &subStatus
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT 
			u.id, 
			u.organization_id, 
			u.unit_id, 
			u.name, 
			u.email, 
			u.password_hash, 
			u.role, 
			u.is_active, 
			u.created_at, 
			u.updated_at,
			o.is_active AS org_is_active,
			o.subscription_status AS org_subscription_status
		FROM users u
		LEFT JOIN organizations o ON o.id = u.organization_id
		WHERE u.id = $1
		LIMIT 1;
	`

	var user domain.User
	var roleStr string
	var orgIsActive *bool
	var orgSubStatus *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.UnitID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&roleStr,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&orgIsActive,
		&orgSubStatus,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	user.Role = domain.Role(roleStr)
	user.OrganizationIsActive = orgIsActive
	if orgSubStatus != nil {
		subStatus := domain.SubscriptionStatus(*orgSubStatus)
		user.SubscriptionStatus = &subStatus
	}
	return &user, nil
}

// ToggleActiveStatus updates user active status, cascading to barbers (status and is_active) and organizations if the user is an ORG_OWNER.
// If an auditLog is provided, it is inserted atomically within the same transaction.
func (r *PostgresUserRepository) ToggleActiveStatus(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	userQuery := `UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2;`
	tag, err := tx.Exec(ctx, userQuery, isActive, userID)
	if err != nil {
		return fmt.Errorf("update user active status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	if ownerOrgID != nil && *ownerOrgID != "" {
		orgQuery := `UPDATE organizations SET is_active = $1, updated_at = NOW() WHERE id = $2;`
		if _, err := tx.Exec(ctx, orgQuery, isActive, *ownerOrgID); err != nil {
			return fmt.Errorf("update organization active status: %w", err)
		}
	}

	if isBarber {
		barberStatus := domain.BarberStatusOffDuty
		if isActive {
			barberStatus = domain.BarberStatusAvailable
		}
		barberQuery := `
			UPDATE barbers 
			SET is_active = $1, status = $2, updated_at = NOW() 
			WHERE user_id = $3;
		`
		if _, err := tx.Exec(ctx, barberQuery, isActive, barberStatus, userID); err != nil {
			return fmt.Errorf("update barber active status: %w", err)
		}
	}

	if auditLog != nil {
		auditQuery := `
			INSERT INTO audit_logs (
				organization_id, unit_id, actor_id, actor_role, action, target_entity, target_id,
				details, ip_address, user_agent, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
		`
		detailsJSON := auditLog.Details
		if len(detailsJSON) == 0 {
			detailsJSON = []byte("{}")
		}
		if _, err := tx.Exec(ctx, auditQuery,
			auditLog.OrganizationID,
			auditLog.UnitID,
			auditLog.ActorID,
			string(auditLog.ActorRole),
			auditLog.Action,
			auditLog.TargetEntity,
			auditLog.TargetID,
			detailsJSON,
			auditLog.IPAddress,
			auditLog.UserAgent,
		); err != nil {
			return fmt.Errorf("insert audit log in toggle status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit user toggle active status: %w", err)
	}

	return nil
}
