package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresOwnerCommissionsRepository implements multi-scope commission updates in PostgreSQL.
type PostgresOwnerCommissionsRepository struct {
	db *pgxpool.Pool
}

// NewPostgresOwnerCommissionsRepository creates a repository for owner commission rules.
func NewPostgresOwnerCommissionsRepository(db *pgxpool.Pool) *PostgresOwnerCommissionsRepository {
	return &PostgresOwnerCommissionsRepository{db: db}
}

// UnitBelongsToOrg checks whether the unit belongs to the specified organization.
func (r *PostgresOwnerCommissionsRepository) UnitBelongsToOrg(ctx context.Context, unitID, orgID string) (bool, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(orgID) {
		return false, domain.ErrInvalidUUID
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM barbershop_units WHERE id = $1 AND organization_id = $2);`
	if err := r.db.QueryRow(ctx, query, unitID, orgID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check unit org ownership: %w", err)
	}
	return exists, nil
}

// BarberBelongsToOrg checks whether the barber belongs to any unit inside the organization.
func (r *PostgresOwnerCommissionsRepository) BarberBelongsToOrg(ctx context.Context, barberID, orgID string) (bool, error) {
	if !domain.IsValidUUID(barberID) || !domain.IsValidUUID(orgID) {
		return false, domain.ErrInvalidUUID
	}

	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM barbers b
			INNER JOIN barbershop_units u ON u.id = b.unit_id
			WHERE b.id = $1 AND u.organization_id = $2
		);
	`
	if err := r.db.QueryRow(ctx, query, barberID, orgID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check barber org ownership: %w", err)
	}
	return exists, nil
}

// UpdateAllOrganization sets commission_percentage for every barber across all units of the organization.
func (r *PostgresOwnerCommissionsRepository) UpdateAllOrganization(ctx context.Context, orgID string, percentage int) (int, error) {
	if !domain.IsValidUUID(orgID) {
		return 0, domain.ErrInvalidUUID
	}

	query := `
		UPDATE barbers
		SET commission_percentage = $1, updated_at = NOW()
		WHERE unit_id IN (SELECT id FROM barbershop_units WHERE organization_id = $2);
	`
	tag, err := r.db.Exec(ctx, query, percentage, orgID)
	if err != nil {
		return 0, fmt.Errorf("update all organization commissions: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// UpdateByUnit sets commission_percentage for all barbers in a single unit belonging to the organization.
func (r *PostgresOwnerCommissionsRepository) UpdateByUnit(ctx context.Context, orgID, unitID string, percentage int) (int, error) {
	if !domain.IsValidUUID(orgID) || !domain.IsValidUUID(unitID) {
		return 0, domain.ErrInvalidUUID
	}

	query := `
		UPDATE barbers b
		SET commission_percentage = $1, updated_at = NOW()
		FROM barbershop_units u
		WHERE b.unit_id = u.id AND b.unit_id = $2 AND u.organization_id = $3;
	`
	tag, err := r.db.Exec(ctx, query, percentage, unitID, orgID)
	if err != nil {
		return 0, fmt.Errorf("update unit commissions: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// UpdateByBarber sets commission_percentage for a single barber inside the organization.
func (r *PostgresOwnerCommissionsRepository) UpdateByBarber(ctx context.Context, orgID, barberID string, percentage int) (int, error) {
	if !domain.IsValidUUID(orgID) || !domain.IsValidUUID(barberID) {
		return 0, domain.ErrInvalidUUID
	}

	query := `
		UPDATE barbers b
		SET commission_percentage = $1, updated_at = NOW()
		FROM barbershop_units u
		WHERE b.unit_id = u.id AND b.id = $2 AND u.organization_id = $3;
	`
	tag, err := r.db.Exec(ctx, query, percentage, barberID, orgID)
	if err != nil {
		return 0, fmt.Errorf("update barber commission: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
