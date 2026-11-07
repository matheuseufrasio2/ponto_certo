package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// TenantAdminRepository defines data access methods for SaaS multi-tenancy and platform administration.
type TenantAdminRepository interface {
	GetSaaSPlatformKPIs(ctx context.Context) (*domain.SaaSMetrics, error)
	ListTenants(ctx context.Context, search string, limit, offset int) ([]domain.Organization, int, error)
	CreateTenantWithAdmin(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error)
	ToggleTenantActiveStatus(ctx context.Context, tenantID string, isActive bool) error
	GetTenantUsers(ctx context.Context, tenantID string) ([]domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}

// PostgresTenantAdminRepository implements TenantAdminRepository with PostgreSQL.
type PostgresTenantAdminRepository struct {
	db *pgxpool.Pool
}

// NewPostgresTenantAdminRepository instantiates a new tenant admin repository.
func NewPostgresTenantAdminRepository(db *pgxpool.Pool) *PostgresTenantAdminRepository {
	return &PostgresTenantAdminRepository{db: db}
}

// GetSaaSPlatformKPIs aggregates platform-level metrics including MRR, ARR, and active franchises.
func (r *PostgresTenantAdminRepository) GetSaaSPlatformKPIs(ctx context.Context) (*domain.SaaSMetrics, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN subscription_status = 'ACTIVE' THEN monthly_fee_in_cents ELSE 0 END), 0) AS mrr_in_cents,
			COUNT(CASE WHEN is_active = TRUE THEN 1 END) AS active_tenants,
			COUNT(CASE WHEN subscription_status = 'PAST_DUE' THEN 1 END) AS past_due_count,
			(SELECT COUNT(*) FROM barbershop_units WHERE is_active = TRUE) AS total_units
		FROM organizations;
	`

	metrics := &domain.SaaSMetrics{}
	err := r.db.QueryRow(ctx, query).Scan(
		&metrics.MRRInCents,
		&metrics.ActiveTenants,
		&metrics.PastDueTenantsCount,
		&metrics.TotalUnitsAcrossTenants,
	)
	if err != nil {
		return nil, fmt.Errorf("query saas platform kpis: %w", err)
	}

	metrics.ARRInCents = metrics.MRRInCents * 12
	metrics.ChurnRatePercent = 2.1 // Platform benchmark

	return metrics, nil
}

// ListTenants retrieves organizations matching search query with total count.
func (r *PostgresTenantAdminRepository) ListTenants(ctx context.Context, search string, limit, offset int) ([]domain.Organization, int, error) {
	query := `
		SELECT 
			id::TEXT,
			name,
			COALESCE(trade_name, ''),
			COALESCE(document, ''),
			COALESCE(plan, 'BASIC'),
			COALESCE(monthly_fee_in_cents, 19900),
			COALESCE(subscription_status, 'ACTIVE'),
			is_active,
			created_at,
			updated_at,
			COUNT(*) OVER() AS total_count
		FROM organizations
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR trade_name ILIKE '%' || $1 || '%' OR document ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.Query(ctx, query, search, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []domain.Organization
	var totalCount int

	for rows.Next() {
		var o domain.Organization
		var statusStr string
		if err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.TradeName,
			&o.Document,
			&o.Plan,
			&o.MonthlyFeeInCents,
			&statusStr,
			&o.IsActive,
			&o.CreatedAt,
			&o.UpdatedAt,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan tenant row: %w", err)
		}
		o.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
		tenants = append(tenants, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tenant rows: %w", err)
	}

	return tenants, totalCount, nil
}

// CreateTenantWithAdmin creates organization, default headquarters unit, and primary owner in a single transaction.
func (r *PostgresTenantAdminRepository) CreateTenantWithAdmin(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	plan := input.Plan
	if plan == "" {
		plan = "PRO"
	}
	fee := input.MonthlyFeeInCents
	if fee <= 0 {
		fee = 24900
	}

	tradeName := input.TradeName
	if tradeName == "" {
		tradeName = input.Name
	}

	orgQuery := `
		INSERT INTO organizations (name, trade_name, document, plan, monthly_fee_in_cents, subscription_status, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE', TRUE, NOW(), NOW())
		RETURNING id::TEXT, name, trade_name, document, plan, monthly_fee_in_cents, subscription_status, is_active, created_at, updated_at;
	`

	var org domain.Organization
	var statusStr string
	err = tx.QueryRow(ctx, orgQuery, input.Name, tradeName, input.Document, plan, fee).Scan(
		&org.ID,
		&org.Name,
		&org.TradeName,
		&org.Document,
		&org.Plan,
		&org.MonthlyFeeInCents,
		&statusStr,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert organization: %w", err)
	}
	org.SubscriptionStatus = domain.SubscriptionStatus(statusStr)

	// Create default headquarters unit
	unitQuery := `
		INSERT INTO barbershop_units (organization_id, name, phone, address, is_active, created_at, updated_at)
		VALUES ($1, $2, '11999990000', 'Matriz Principal', TRUE, NOW(), NOW())
		RETURNING id::TEXT;
	`
	var unitID string
	err = tx.QueryRow(ctx, unitQuery, org.ID, fmt.Sprintf("%s - Matriz", tradeName)).Scan(&unitID)
	if err != nil {
		return nil, fmt.Errorf("insert initial unit: %w", err)
	}

	// Create primary owner administrator user
	userQuery := `
		INSERT INTO users (organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW(), NOW());
	`
	_, err = tx.Exec(ctx, userQuery, org.ID, unitID, input.AdminName, input.AdminEmail, passwordHash, domain.RoleOrgOwner)
	if err != nil {
		return nil, fmt.Errorf("insert owner user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tenant creation: %w", err)
	}

	return &org, nil
}

// ToggleTenantActiveStatus toggles active state and cascades login blocking if disabled.
func (r *PostgresTenantAdminRepository) ToggleTenantActiveStatus(ctx context.Context, tenantID string, isActive bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	updateOrgQuery := `UPDATE organizations SET is_active = $1, updated_at = NOW() WHERE id = $2;`
	tag, err := tx.Exec(ctx, updateOrgQuery, isActive, tenantID)
	if err != nil {
		return fmt.Errorf("update organization is_active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}

	// Cascade update users under this organization
	updateUsersQuery := `UPDATE users SET is_active = $1, updated_at = NOW() WHERE organization_id = $2;`
	if _, err := tx.Exec(ctx, updateUsersQuery, isActive, tenantID); err != nil {
		return fmt.Errorf("cascade update users is_active: %w", err)
	}

	return tx.Commit(ctx)
}

// GetTenantUsers returns active users belonging to a tenant organization for impersonation.
func (r *PostgresTenantAdminRepository) GetTenantUsers(ctx context.Context, tenantID string) ([]domain.User, error) {
	query := `
		SELECT 
			id::TEXT,
			COALESCE(organization_id::TEXT, ''),
			COALESCE(unit_id::TEXT, ''),
			name,
			email,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE organization_id = $1 AND is_active = TRUE
		ORDER BY name ASC;
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query tenant users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var orgIDStr, unitIDStr, roleStr string
		if err := rows.Scan(
			&u.ID,
			&orgIDStr,
			&unitIDStr,
			&u.Name,
			&u.Email,
			&roleStr,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tenant user: %w", err)
		}
		if orgIDStr != "" {
			u.OrganizationID = &orgIDStr
		}
		if unitIDStr != "" {
			u.UnitID = &unitIDStr
		}
		u.Role = domain.Role(roleStr)
		users = append(users, u)
	}

	return users, rows.Err()
}

// GetUserByID retrieves a single user by ID for impersonation verification.
func (r *PostgresTenantAdminRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT 
			id::TEXT,
			COALESCE(organization_id::TEXT, ''),
			COALESCE(unit_id::TEXT, ''),
			name,
			email,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE id = $1;
	`

	var u domain.User
	var orgIDStr, unitIDStr, roleStr string
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&u.ID,
		&orgIDStr,
		&unitIDStr,
		&u.Name,
		&u.Email,
		&roleStr,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	if orgIDStr != "" {
		u.OrganizationID = &orgIDStr
	}
	if unitIDStr != "" {
		u.UnitID = &unitIDStr
	}
	u.Role = domain.Role(roleStr)

	return &u, nil
}
