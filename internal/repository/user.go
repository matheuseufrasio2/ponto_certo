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
