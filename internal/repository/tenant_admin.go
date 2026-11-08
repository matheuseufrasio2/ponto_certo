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

// TenantAdminRepository defines data access methods for SaaS multi-tenancy and platform administration.
type TenantAdminRepository interface {
	GetSaaSPlatformKPIs(ctx context.Context) (*domain.SaaSMetrics, error)
	ListTenants(ctx context.Context, search, status, plan string, limit, offset int) ([]domain.Organization, int, error)
	GetTenantOverview(ctx context.Context, tenantID string) (*domain.TenantOverview, error)
	UpdateTenantPlan(ctx context.Context, tenantID, plan string, feeInCents int64) (*domain.Organization, error)
	CreateTenantWithAdmin(ctx context.Context, input domain.CreateTenantInput, passwordHash string) (*domain.Organization, error)
	ToggleTenantActiveStatus(ctx context.Context, tenantID string, isActive bool) error
	GetTenantUsers(ctx context.Context, tenantID string, status string) ([]domain.User, error)
	GetTenantUnits(ctx context.Context, tenantID string) ([]domain.BarbershopUnit, error)
	GetPrimaryOwnerByTenantID(ctx context.Context, tenantID string) (*domain.User, error)
	UpdateUserPassword(ctx context.Context, userID, passwordHash string) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	IsOrganizationActive(ctx context.Context, orgID string) (bool, error)
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

// ListTenants retrieves organizations matching search, status, and plan filters with operational counters.
func (r *PostgresTenantAdminRepository) ListTenants(ctx context.Context, search, status, plan string, limit, offset int) ([]domain.Organization, int, error) {
	query := `
		SELECT 
			o.id::TEXT,
			o.name,
			COALESCE(o.trade_name, ''),
			COALESCE(o.document, ''),
			COALESCE(o.plan, 'BASIC'),
			COALESCE(o.monthly_fee_in_cents, 19900),
			COALESCE(o.subscription_status, 'ACTIVE'),
			o.is_active,
			(SELECT COUNT(*) FROM barbershop_units u WHERE u.organization_id = o.id)::INT AS units_count,
			(SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.is_active = TRUE)::INT AS active_users_count,
			(SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id)::INT AS total_users_count,
			(
				SELECT MAX(a.created_at)
				FROM appointments a
				JOIN barbershop_units bu ON a.unit_id = bu.id
				WHERE bu.organization_id = o.id
			) AS last_activity_at,
			o.past_due_since,
			o.created_at,
			o.updated_at,
			COUNT(*) OVER() AS total_count
		FROM organizations o
		WHERE ($1 = '' OR o.name ILIKE '%' || $1 || '%' OR o.trade_name ILIKE '%' || $1 || '%' OR o.document ILIKE '%' || $1 || '%')
		  AND (
			$2 = '' OR 
			($2 = 'BLOCKED' AND o.is_active = FALSE) OR
			($2 = 'INACTIVE' AND o.is_active = FALSE) OR
			($2 = 'ACTIVE' AND o.is_active = TRUE AND o.subscription_status = 'ACTIVE') OR
			(o.subscription_status = $2)
		  )
		  AND ($3 = '' OR UPPER(o.plan) = UPPER($3))
		ORDER BY o.created_at DESC
		LIMIT $4 OFFSET $5;
	`

	rows, err := r.db.Query(ctx, query, search, status, plan, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	tenants := make([]domain.Organization, 0)
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
			&o.UnitsCount,
			&o.ActiveUsersCount,
			&o.TotalUsersCount,
			&o.LastActivityAt,
			&o.PastDueSince,
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

// ToggleTenantActiveStatus toggles active state of the organization.
func (r *PostgresTenantAdminRepository) ToggleTenantActiveStatus(ctx context.Context, tenantID string, isActive bool) error {
	updateOrgQuery := `UPDATE organizations SET is_active = $1, updated_at = NOW() WHERE id = $2;`
	tag, err := r.db.Exec(ctx, updateOrgQuery, isActive, tenantID)
	if err != nil {
		return fmt.Errorf("update organization is_active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}

	return nil
}

// IsOrganizationActive checks if the organization is active in real-time.
func (r *PostgresTenantAdminRepository) IsOrganizationActive(ctx context.Context, orgID string) (bool, error) {
	query := `SELECT is_active FROM organizations WHERE id = $1 LIMIT 1;`
	var isActive bool
	err := r.db.QueryRow(ctx, query, orgID).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrOrganizationNotFound
		}
		return false, fmt.Errorf("check organization is_active: %w", err)
	}
	return isActive, nil
}

// GetTenantUsers returns users belonging to a tenant organization filtered by status (active, inactive, all).
func (r *PostgresTenantAdminRepository) GetTenantUsers(ctx context.Context, tenantID string, status string) ([]domain.User, error) {
	normStatus := strings.ToLower(strings.TrimSpace(status))
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
		WHERE organization_id = $1
		  AND (
			($2 = 'all') OR
			($2 = 'inactive' AND is_active = FALSE) OR
			(($2 = 'active' OR $2 = '') AND is_active = TRUE)
		  )
		ORDER BY name ASC;
	`

	rows, err := r.db.Query(ctx, query, tenantID, normStatus)
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

// GetTenantUnits returns all physical branches belonging to an organization.
func (r *PostgresTenantAdminRepository) GetTenantUnits(ctx context.Context, tenantID string) ([]domain.BarbershopUnit, error) {
	query := `
		SELECT 
			id::TEXT,
			organization_id::TEXT,
			name,
			COALESCE(phone, ''),
			COALESCE(address, ''),
			COALESCE(chair_count, 4),
			is_active,
			created_at,
			updated_at
		FROM barbershop_units
		WHERE organization_id = $1
		ORDER BY created_at ASC;
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query tenant units: %w", err)
	}
	defer rows.Close()

	units := make([]domain.BarbershopUnit, 0)
	for rows.Next() {
		var u domain.BarbershopUnit
		if err := rows.Scan(
			&u.ID,
			&u.OrganizationID,
			&u.Name,
			&u.Phone,
			&u.Address,
			&u.ChairCount,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tenant unit: %w", err)
		}
		units = append(units, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant unit rows: %w", err)
	}

	return units, nil
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

// GetTenantOverview aggregates tenant info, 30-day GMV, 30-day appointments, total barbers, and days overdue.
func (r *PostgresTenantAdminRepository) GetTenantOverview(ctx context.Context, tenantID string) (*domain.TenantOverview, error) {
	query := `
		SELECT 
			o.id::TEXT,
			o.name,
			COALESCE(o.trade_name, ''),
			COALESCE(o.document, ''),
			COALESCE(o.plan, 'BASIC'),
			COALESCE(o.monthly_fee_in_cents, 19900),
			COALESCE(o.subscription_status, 'ACTIVE'),
			o.is_active,
			(SELECT COUNT(*) FROM barbershop_units u WHERE u.organization_id = o.id)::INT AS units_count,
			(SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id AND u.is_active = TRUE)::INT AS active_users_count,
			(SELECT COUNT(*) FROM users u WHERE u.organization_id = o.id)::INT AS total_users_count,
			(
				SELECT MAX(a.created_at)
				FROM appointments a
				JOIN barbershop_units bu ON a.unit_id = bu.id
				WHERE bu.organization_id = o.id
			) AS last_activity_at,
			o.past_due_since,
			o.created_at,
			o.updated_at,
			COALESCE((
				SELECT SUM(b.total_amount_in_cents)
				FROM billings b
				JOIN barbershop_units bu ON b.unit_id = bu.id
				WHERE bu.organization_id = o.id
				  AND b.status = 'PAID'
				  AND b.created_at >= NOW() - INTERVAL '30 days'
			), 0) AS gmv_30d,
			COALESCE((
				SELECT COUNT(*)
				FROM appointments a
				JOIN barbershop_units bu ON a.unit_id = bu.id
				WHERE bu.organization_id = o.id
				  AND a.created_at >= NOW() - INTERVAL '30 days'
			), 0) AS appts_30d,
			COALESCE((
				SELECT COUNT(DISTINCT b.id)
				FROM barbers b
				JOIN barbershop_units bu ON b.unit_id = bu.id
				WHERE bu.organization_id = o.id
				  AND b.is_active = TRUE
			), 0) AS barbers_count
		FROM organizations o
		WHERE o.id = $1;
	`

	var o domain.Organization
	var statusStr string
	var overview domain.TenantOverview

	err := r.db.QueryRow(ctx, query, tenantID).Scan(
		&o.ID,
		&o.Name,
		&o.TradeName,
		&o.Document,
		&o.Plan,
		&o.MonthlyFeeInCents,
		&statusStr,
		&o.IsActive,
		&o.UnitsCount,
		&o.ActiveUsersCount,
		&o.TotalUsersCount,
		&o.LastActivityAt,
		&o.PastDueSince,
		&o.CreatedAt,
		&o.UpdatedAt,
		&overview.GMVLast30DaysInCents,
		&overview.AppointmentsLast30Days,
		&overview.TotalBarbersCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("query tenant overview: %w", err)
	}

	o.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
	overview.Organization = o

	if o.SubscriptionStatus == domain.SubscriptionStatusPastDue {
		refDate := o.UpdatedAt
		if o.PastDueSince != nil {
			refDate = *o.PastDueSince
		}
		days := int(time.Since(refDate).Hours() / 24)
		if days < 1 {
			days = 1
		}
		overview.DaysOverdue = days
	}

	return &overview, nil
}

// UpdateTenantPlan updates an organization's subscription plan and monthly fee.
func (r *PostgresTenantAdminRepository) UpdateTenantPlan(ctx context.Context, tenantID, plan string, feeInCents int64) (*domain.Organization, error) {
	query := `
		UPDATE organizations
		SET plan = $1, monthly_fee_in_cents = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING 
			id::TEXT,
			name,
			COALESCE(trade_name, ''),
			COALESCE(document, ''),
			plan,
			monthly_fee_in_cents,
			subscription_status,
			is_active,
			past_due_since,
			created_at,
			updated_at;
	`

	var o domain.Organization
	var statusStr string
	err := r.db.QueryRow(ctx, query, plan, feeInCents, tenantID).Scan(
		&o.ID,
		&o.Name,
		&o.TradeName,
		&o.Document,
		&o.Plan,
		&o.MonthlyFeeInCents,
		&statusStr,
		&o.IsActive,
		&o.PastDueSince,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("update tenant plan: %w", err)
	}

	o.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
	return &o, nil
}

// GetPrimaryOwnerByTenantID finds the primary ORG_OWNER user of a tenant.
func (r *PostgresTenantAdminRepository) GetPrimaryOwnerByTenantID(ctx context.Context, tenantID string) (*domain.User, error) {
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
		WHERE organization_id = $1 AND role = 'ORG_OWNER' AND is_active = TRUE
		ORDER BY created_at ASC
		LIMIT 1;
	`

	var u domain.User
	var orgIDStr, unitIDStr, roleStr string
	err := r.db.QueryRow(ctx, query, tenantID).Scan(
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
		return nil, fmt.Errorf("query primary owner: %w", err)
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

// UpdateUserPassword updates the password hash for a user.
func (r *PostgresTenantAdminRepository) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2;`
	tag, err := r.db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
