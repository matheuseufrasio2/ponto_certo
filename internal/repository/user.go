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
		SELECT id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1;
	`

	var user domain.User
	var roleStr string

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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	user.Role = domain.Role(roleStr)
	return &user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1;
	`

	var user domain.User
	var roleStr string

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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	user.Role = domain.Role(roleStr)
	return &user, nil
}
