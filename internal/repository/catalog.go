package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// CatalogRepository defines operations for fetching barbers, services and clients.
type CatalogRepository interface {
	ListBarbersByUnit(ctx context.Context, unitID string) ([]domain.Barber, error)
	ListServicesByUnit(ctx context.Context, unitID string) ([]domain.Service, error)
	GetServiceByID(ctx context.Context, id string) (*domain.Service, error)
	GetBarberByUserID(ctx context.Context, userID string) (*domain.Barber, error)
	GetOrCreateClient(ctx context.Context, unitID, name, phone string) (*domain.Client, error)
}

// PostgresCatalogRepository implements CatalogRepository using PostgreSQL.
type PostgresCatalogRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCatalogRepository instantiates a new catalog repository.
func NewPostgresCatalogRepository(db *pgxpool.Pool) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

// ListBarbersByUnit returns all active barbers assigned to a unit.
func (r *PostgresCatalogRepository) ListBarbersByUnit(ctx context.Context, unitID string) ([]domain.Barber, error) {
	query := `
		SELECT id, unit_id, user_id, name, phone, commission_percentage, is_active, created_at, updated_at
		FROM barbers
		WHERE unit_id = $1 AND is_active = TRUE
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("query barbers by unit: %w", err)
	}
	defer rows.Close()

	var barbers []domain.Barber
	for rows.Next() {
		var b domain.Barber
		var phone *string
		if err := rows.Scan(
			&b.ID,
			&b.UnitID,
			&b.UserID,
			&b.Name,
			&phone,
			&b.CommissionPercentage,
			&b.IsActive,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan barber: %w", err)
		}
		if phone != nil {
			b.Phone = *phone
		}
		barbers = append(barbers, b)
	}
	return barbers, rows.Err()
}

// ListServicesByUnit returns all active services available to a unit.
func (r *PostgresCatalogRepository) ListServicesByUnit(ctx context.Context, unitID string) ([]domain.Service, error) {
	query := `
		SELECT id, unit_id, name, description, duration_minutes, price_in_cents, category, is_active, created_at, updated_at
		FROM services
		WHERE (unit_id IS NULL OR unit_id = $1) AND is_active = TRUE
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("query services by unit: %w", err)
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		var desc *string
		var uID *string
		if err := rows.Scan(
			&s.ID,
			&uID,
			&s.Name,
			&desc,
			&s.DurationMinutes,
			&s.PriceInCents,
			&s.Category,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		if desc != nil {
			s.Description = *desc
		}
		if uID != nil {
			s.UnitID = *uID
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

// GetServiceByID returns a service by its primary key.
func (r *PostgresCatalogRepository) GetServiceByID(ctx context.Context, id string) (*domain.Service, error) {
	query := `
		SELECT id, unit_id, name, description, duration_minutes, price_in_cents, category, is_active, created_at, updated_at
		FROM services
		WHERE id = $1;
	`
	var s domain.Service
	var desc *string
	var uID *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&uID,
		&s.Name,
		&desc,
		&s.DurationMinutes,
		&s.PriceInCents,
		&s.Category,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("query service by id: %w", err)
	}
	if desc != nil {
		s.Description = *desc
	}
	if uID != nil {
		s.UnitID = *uID
	}
	return &s, nil
}

// GetBarberByUserID finds the barber record linked to a system user login.
func (r *PostgresCatalogRepository) GetBarberByUserID(ctx context.Context, userID string) (*domain.Barber, error) {
	query := `
		SELECT id, unit_id, user_id, name, phone, commission_percentage, is_active, created_at, updated_at
		FROM barbers
		WHERE user_id = $1;
	`
	var b domain.Barber
	var phone *string
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&b.ID,
		&b.UnitID,
		&b.UserID,
		&b.Name,
		&phone,
		&b.CommissionPercentage,
		&b.IsActive,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBarberNotFound
		}
		return nil, fmt.Errorf("query barber by user id: %w", err)
	}
	if phone != nil {
		b.Phone = *phone
	}
	return &b, nil
}

// GetOrCreateClient retrieves an existing client by phone or inserts a new one.
func (r *PostgresCatalogRepository) GetOrCreateClient(ctx context.Context, unitID, name, phone string) (*domain.Client, error) {
	if phone != "" {
		query := `
			SELECT id, unit_id, name, phone, email, created_at, updated_at
			FROM clients
			WHERE unit_id = $1 AND phone = $2
			LIMIT 1;
		`
		var c domain.Client
		var cPhone, cEmail *string
		err := r.db.QueryRow(ctx, query, unitID, phone).Scan(
			&c.ID,
			&c.UnitID,
			&c.Name,
			&cPhone,
			&cEmail,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err == nil {
			if cPhone != nil {
				c.Phone = *cPhone
			}
			if cEmail != nil {
				c.Email = *cEmail
			}
			return &c, nil
		}
	}

	// Insert new client
	insertQuery := `
		INSERT INTO clients (unit_id, name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, unit_id, name, phone, email, created_at, updated_at;
	`
	var c domain.Client
	var cPhone, cEmail *string
	err := r.db.QueryRow(ctx, insertQuery, unitID, name, phone).Scan(
		&c.ID,
		&c.UnitID,
		&c.Name,
		&cPhone,
		&cEmail,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert client: %w", err)
	}
	if cPhone != nil {
		c.Phone = *cPhone
	}
	if cEmail != nil {
		c.Email = *cEmail
	}
	return &c, nil
}
