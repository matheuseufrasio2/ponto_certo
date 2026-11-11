package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

const unitSelectColumns = `
	id::TEXT,
	organization_id::TEXT,
	name,
	COALESCE(phone, ''),
	COALESCE(address, ''),
	COALESCE(chair_count, 4),
	is_active,
	created_at,
	updated_at
`

type rowScanner interface {
	Scan(dest ...any) error
}

// PostgresUnitRepository handles PostgreSQL persistence for barbershop units/branches.
type PostgresUnitRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUnitRepository instantiates a new PostgresUnitRepository.
func NewPostgresUnitRepository(db *pgxpool.Pool) *PostgresUnitRepository {
	return &PostgresUnitRepository{db: db}
}

// ListByOrg returns barbershop units belonging to an organization, optionally filtered by active status.
func (r *PostgresUnitRepository) ListByOrg(ctx context.Context, orgID string, status string) ([]domain.BarbershopUnit, error) {
	normStatus := strings.ToLower(strings.TrimSpace(status))
	query := fmt.Sprintf(`
		SELECT %s
		FROM barbershop_units
		WHERE organization_id = $1
		  AND (
			($2 = 'all' OR $2 = '') OR
			($2 = 'active' AND is_active = TRUE) OR
			($2 = 'inactive' AND is_active = FALSE)
		  )
		ORDER BY name ASC;
	`, unitSelectColumns)

	rows, err := r.db.Query(ctx, query, orgID, normStatus)
	if err != nil {
		return nil, fmt.Errorf("list units by organization: %w", err)
	}
	defer rows.Close()

	units := make([]domain.BarbershopUnit, 0)
	for rows.Next() {
		u, err := scanUnit(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unit row: %w", err)
		}
		units = append(units, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unit rows: %w", err)
	}

	return units, nil
}

// GetByIDAndOrg fetches a unit verifying strict ownership by organization_id.
func (r *PostgresUnitRepository) GetByIDAndOrg(ctx context.Context, unitID string, orgID string) (*domain.BarbershopUnit, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM barbershop_units
		WHERE id = $1 AND organization_id = $2
		LIMIT 1;
	`, unitSelectColumns)

	u, err := scanUnit(r.db.QueryRow(ctx, query, unitID, orgID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, fmt.Errorf("get unit by id and org: %w", err)
	}
	return u, nil
}

// ExistsByNameInOrg verifies case-insensitive unit name uniqueness within an organization.
func (r *PostgresUnitRepository) ExistsByNameInOrg(ctx context.Context, orgID string, name string, excludeUnitID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM barbershop_units
			WHERE organization_id = $1
			  AND LOWER(TRIM(name)) = LOWER(TRIM($2))
			  AND ($3 = '' OR id != $3::UUID)
		);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, query, orgID, name, excludeUnitID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check unit name uniqueness: %w", err)
	}
	return exists, nil
}

// Create inserts a new active barbershop unit for an organization.
func (r *PostgresUnitRepository) Create(ctx context.Context, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
	query := fmt.Sprintf(`
		INSERT INTO barbershop_units (organization_id, name, phone, address, chair_count, is_active)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, TRUE)
		RETURNING %s;
	`, unitSelectColumns)

	u, err := scanUnit(r.db.QueryRow(ctx, query, orgID, name, derefString(phone), derefString(address), chairCount))
	if err != nil {
		return nil, fmt.Errorf("create unit: %w", err)
	}
	return u, nil
}

// Update modifies unit details scoped by unit_id and organization_id.
func (r *PostgresUnitRepository) Update(ctx context.Context, unitID string, orgID string, name string, phone *string, address *string, chairCount int) (*domain.BarbershopUnit, error) {
	query := fmt.Sprintf(`
		UPDATE barbershop_units
		SET
			name = $1,
			phone = NULLIF($2, ''),
			address = NULLIF($3, ''),
			chair_count = $4,
			updated_at = NOW()
		WHERE id = $5 AND organization_id = $6
		RETURNING %s;
	`, unitSelectColumns)

	u, err := scanUnit(r.db.QueryRow(ctx, query, name, derefString(phone), derefString(address), chairCount, unitID, orgID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, fmt.Errorf("update unit: %w", err)
	}
	return u, nil
}

// ToggleStatus activates or pauses a barbershop unit scoped by unit_id and organization_id.
func (r *PostgresUnitRepository) ToggleStatus(ctx context.Context, unitID string, orgID string, isActive bool) (*domain.BarbershopUnit, error) {
	query := fmt.Sprintf(`
		UPDATE barbershop_units
		SET
			is_active = $1,
			updated_at = NOW()
		WHERE id = $2 AND organization_id = $3
		RETURNING %s;
	`, unitSelectColumns)

	u, err := scanUnit(r.db.QueryRow(ctx, query, isActive, unitID, orgID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, fmt.Errorf("toggle unit status: %w", err)
	}
	return u, nil
}

// IsUnitActive returns whether the specified unit exists and is currently active.
func (r *PostgresUnitRepository) IsUnitActive(ctx context.Context, unitID string) (bool, error) {
	if !domain.IsValidUUID(unitID) {
		return false, domain.ErrUnitNotFound
	}
	query := `SELECT is_active FROM barbershop_units WHERE id = $1 LIMIT 1;`
	var isActive bool
	err := r.db.QueryRow(ctx, query, unitID).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrUnitNotFound
		}
		return false, fmt.Errorf("check unit active status: %w", err)
	}
	return isActive, nil
}

func scanUnit(scanner rowScanner) (*domain.BarbershopUnit, error) {
	var u domain.BarbershopUnit
	err := scanner.Scan(
		&u.ID,
		&u.OrganizationID,
		&u.Name,
		&u.Phone,
		&u.Address,
		&u.ChairCount,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
