package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresDevRepo provides database access for developer testing discovery endpoints.
type PostgresDevRepo struct {
	db *pgxpool.Pool
}

// NewPostgresDevRepo creates a new PostgresDevRepo.
func NewPostgresDevRepo(db *pgxpool.Pool) *PostgresDevRepo {
	return &PostgresDevRepo{db: db}
}

// ListSuperAdmins retrieves all active SUPER_ADMIN users.
func (r *PostgresDevRepo) ListSuperAdmins(ctx context.Context) ([]domain.DevSuperAdmin, error) {
	query := `
		SELECT id, name, email, role
		FROM users
		WHERE role = 'SUPER_ADMIN' AND is_active = TRUE
		ORDER BY name ASC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list super admins query: %w", err)
	}
	defer rows.Close()

	admins := make([]domain.DevSuperAdmin, 0)
	for rows.Next() {
		var a domain.DevSuperAdmin
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Role); err != nil {
			return nil, fmt.Errorf("scan super admin: %w", err)
		}
		admins = append(admins, a)
	}

	return admins, rows.Err()
}

// ListTenants retrieves paginated tenants with optional search and scenario filter.
func (r *PostgresDevRepo) ListTenants(ctx context.Context, search, scenario string, limit, offset int) ([]domain.DevTenant, int, error) {
	query := `
		SELECT 
			o.id,
			COALESCE(NULLIF(o.trade_name, ''), o.name) AS trade_name,
			o.plan,
			o.subscription_status,
			COALESCE(EXTRACT(DAY FROM NOW() - o.past_due_since)::INT, 0) AS days_past_due,
			u.name AS owner_name,
			u.email AS owner_email,
			COUNT(*) OVER() AS total_count
		FROM organizations o
		LEFT JOIN LATERAL (
			SELECT name, email
			FROM users
			WHERE organization_id = o.id AND role = 'ORG_OWNER'
			ORDER BY created_at ASC
			LIMIT 1
		) u ON TRUE
		WHERE ($1 = '' OR o.name ILIKE '%' || $1 || '%' OR o.trade_name ILIKE '%' || $1 || '%' OR o.document ILIKE '%' || $1 || '%')
		  AND (
			  $2 = '' OR
			  ($2 = 'past_due' AND o.subscription_status = 'PAST_DUE')
		  )
		ORDER BY o.created_at DESC
		LIMIT $3 OFFSET $4;
	`

	rows, err := r.db.Query(ctx, query, search, scenario, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list dev tenants query: %w", err)
	}
	defer rows.Close()

	tenants := make([]domain.DevTenant, 0)
	var totalCount int

	for rows.Next() {
		var t domain.DevTenant
		var ownerName, ownerEmail sql.NullString
		if err := rows.Scan(
			&t.ID,
			&t.TradeName,
			&t.Plan,
			&t.Status,
			&t.DaysPastDue,
			&ownerName,
			&ownerEmail,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan dev tenant: %w", err)
		}

		if ownerName.Valid && ownerEmail.Valid {
			t.Owner = &domain.DevTenantOwner{
				Name:  ownerName.String,
				Email: ownerEmail.String,
			}
		}

		tenants = append(tenants, t)
	}

	return tenants, totalCount, rows.Err()
}

// ListTenantUsers retrieves all users belonging to a specific organization.
func (r *PostgresDevRepo) ListTenantUsers(ctx context.Context, tenantID string) ([]domain.DevTenantUser, error) {
	query := `
		SELECT 
			u.id,
			u.name,
			u.email,
			u.role,
			COALESCE(bu.name, '') AS unit_name
		FROM users u
		LEFT JOIN barbershop_units bu ON bu.id = u.unit_id
		WHERE u.organization_id = $1
		ORDER BY 
			CASE u.role 
				WHEN 'ORG_OWNER' THEN 1 
				WHEN 'UNIT_MANAGER' THEN 2 
				WHEN 'RECEPTIONIST' THEN 3 
				WHEN 'BARBER' THEN 4 
				ELSE 5 
			END, 
			u.name ASC;
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list tenant users query: %w", err)
	}
	defer rows.Close()

	users := make([]domain.DevTenantUser, 0)
	for rows.Next() {
		var u domain.DevTenantUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.UnitName); err != nil {
			return nil, fmt.Errorf("scan tenant user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// ListUsersByRole retrieves paginated users filtered by role.
func (r *PostgresDevRepo) ListUsersByRole(ctx context.Context, role string, limit, offset int) ([]domain.DevUser, int, error) {
	query := `
		SELECT 
			u.id,
			u.name,
			u.email,
			u.role,
			COALESCE(NULLIF(o.trade_name, ''), o.name, 'Plataforma') AS tenant_name,
			COUNT(*) OVER() AS total_count
		FROM users u
		LEFT JOIN organizations o ON o.id = u.organization_id
		WHERE ($1 = '' OR u.role = $1)
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.Query(ctx, query, role, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users by role query: %w", err)
	}
	defer rows.Close()

	users := make([]domain.DevUser, 0)
	var totalCount int

	for rows.Next() {
		var u domain.DevUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.TenantName, &totalCount); err != nil {
			return nil, 0, fmt.Errorf("scan user by role: %w", err)
		}
		users = append(users, u)
	}

	return users, totalCount, rows.Err()
}
