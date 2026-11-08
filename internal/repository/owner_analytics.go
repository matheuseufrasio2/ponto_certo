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

// BillingExportRow models a flattened billing record for CSV stream export.
type BillingExportRow struct {
	ID            string
	Date          string
	UnitName      string
	BarberName    string
	ClientName    string
	ServiceName   string
	AmountInCents int64
	PaymentMethod string
	Status        string
	CreatedAt     string
}

// OwnerAnalyticsRepository defines SQL methods for executive network analytics.
type OwnerAnalyticsRepository interface {
	GetExecutiveKPIs(ctx context.Context, orgID string, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error)
	GetRevenueComparisonByUnit(ctx context.Context, orgID string, start, end time.Time) ([]domain.UnitRevenueComparison, error)
	GetServicesMix(ctx context.Context, orgID string, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error)
	ListConsolidatedBillings(ctx context.Context, orgID string, unitID *string, paymentMethod *string, limit, offset int) ([]domain.Billing, int, error)
	GetExportBillingsData(ctx context.Context, orgID string, unitID *string) ([]BillingExportRow, error)
	GetOwnerSubscription(ctx context.Context, orgID string) (*domain.OwnerSubscription, error)
	ListEmployees(ctx context.Context, orgID string, status string, unitID *string) ([]domain.OrganizationEmployee, error)
}

// PostgresOwnerAnalyticsRepository implements OwnerAnalyticsRepository using pgxpool.
type PostgresOwnerAnalyticsRepository struct {
	db *pgxpool.Pool
}

// NewPostgresOwnerAnalyticsRepository instantiates a new owner analytics repository.
func NewPostgresOwnerAnalyticsRepository(db *pgxpool.Pool) *PostgresOwnerAnalyticsRepository {
	return &PostgresOwnerAnalyticsRepository{db: db}
}

// GetExecutiveKPIs computes gross revenue, net revenue, total commissions, and average ticket.
func (r *PostgresOwnerAnalyticsRepository) GetExecutiveKPIs(ctx context.Context, orgID string, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error) {
	query := `
		SELECT 
			COALESCE(SUM(b.amount_in_cents), 0) AS gross_revenue,
			COALESCE(SUM(ROUND(b.amount_in_cents * (COALESCE(bar.commission_percentage, 50)::NUMERIC / 100))), 0) AS total_commissions,
			COUNT(b.id) AS total_appointments
		FROM billings b
		JOIN barbershop_units u ON b.unit_id = u.id
		LEFT JOIN barbers bar ON b.barber_id = bar.id
		WHERE u.organization_id = $1
		  AND ($2::UUID IS NULL OR b.unit_id = $2)
		  AND b.date >= $3 AND b.date <= $4
		  AND b.status = 'PAID';
	`

	kpis := &domain.OwnerExecutiveKPIs{}
	var grossRev, totalComm int64
	var totalApts int

	err := r.db.QueryRow(ctx, query, orgID, unitID, start, end).Scan(&grossRev, &totalComm, &totalApts)
	if err != nil {
		return nil, fmt.Errorf("query executive kpis: %w", err)
	}

	kpis.GrossRevenueInCents = grossRev
	kpis.TotalCommissionsInCents = totalComm
	kpis.NetRevenueInCents = grossRev - totalComm
	kpis.TotalAppointments = totalApts

	if totalApts > 0 {
		kpis.AverageTicketInCents = grossRev / int64(totalApts)
	} else {
		kpis.AverageTicketInCents = 0
	}

	// Retention rate: repeat clients / total unique clients in period
	retentionQuery := `
		WITH client_counts AS (
			SELECT b.client_id, COUNT(*) AS cnt
			FROM billings b
			JOIN barbershop_units u ON b.unit_id = u.id
			WHERE u.organization_id = $1
			  AND ($2::UUID IS NULL OR b.unit_id = $2)
			  AND b.date >= $3 AND b.date <= $4
			  AND b.status = 'PAID'
			GROUP BY b.client_id
		)
		SELECT 
			COALESCE(COUNT(*), 0) AS total_clients,
			COALESCE(COUNT(*) FILTER (WHERE cnt > 1), 0) AS repeat_clients
		FROM client_counts;
	`
	var totalClients, repeatClients int
	if err := r.db.QueryRow(ctx, retentionQuery, orgID, unitID, start, end).Scan(&totalClients, &repeatClients); err == nil && totalClients > 0 {
		kpis.RetentionRatePercent = float64(repeatClients*100) / float64(totalClients)
	} else {
		kpis.RetentionRatePercent = 0.0
	}

	return kpis, nil
}

// GetRevenueComparisonByUnit aggregates revenue per unit for network-wide comparisons.
func (r *PostgresOwnerAnalyticsRepository) GetRevenueComparisonByUnit(ctx context.Context, orgID string, start, end time.Time) ([]domain.UnitRevenueComparison, error) {
	query := `
		SELECT 
			u.id::TEXT AS unit_id,
			u.name AS unit_name,
			COALESCE(SUM(b.amount_in_cents), 0) AS revenue_in_cents
		FROM barbershop_units u
		LEFT JOIN billings b ON u.id = b.unit_id 
			AND b.status = 'PAID' 
			AND b.date >= $2 AND b.date <= $3
		WHERE u.organization_id = $1
		GROUP BY u.id, u.name
		ORDER BY revenue_in_cents DESC;
	`

	rows, err := r.db.Query(ctx, query, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query revenue by unit: %w", err)
	}
	defer rows.Close()

	var comparisons []domain.UnitRevenueComparison
	for rows.Next() {
		var item domain.UnitRevenueComparison
		if err := rows.Scan(&item.UnitID, &item.UnitName, &item.RevenueInCents); err != nil {
			return nil, fmt.Errorf("scan revenue by unit: %w", err)
		}
		comparisons = append(comparisons, item)
	}

	return comparisons, rows.Err()
}

// GetServicesMix ranks services by volume and revenue.
func (r *PostgresOwnerAnalyticsRepository) GetServicesMix(ctx context.Context, orgID string, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error) {
	query := `
		SELECT 
			s.id::TEXT AS service_id,
			s.name AS service_name,
			s.category AS category,
			COUNT(b.id) AS count,
			COALESCE(SUM(b.amount_in_cents), 0) AS revenue_in_cents
		FROM services s
		JOIN billings b ON s.id = b.service_id
		JOIN barbershop_units u ON b.unit_id = u.id
		WHERE u.organization_id = $1
		  AND ($2::UUID IS NULL OR b.unit_id = $2)
		  AND b.date >= $3 AND b.date <= $4
		  AND b.status = 'PAID'
		GROUP BY s.id, s.name, s.category
		ORDER BY revenue_in_cents DESC;
	`

	rows, err := r.db.Query(ctx, query, orgID, unitID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query services mix: %w", err)
	}
	defer rows.Close()

	var mix []domain.ServiceMixAnalytics
	for rows.Next() {
		var item domain.ServiceMixAnalytics
		if err := rows.Scan(&item.ServiceID, &item.ServiceName, &item.Category, &item.Count, &item.RevenueInCents); err != nil {
			return nil, fmt.Errorf("scan service mix row: %w", err)
		}
		mix = append(mix, item)
	}

	return mix, rows.Err()
}

// ListConsolidatedBillings retrieves paginated billing transactions with total count.
func (r *PostgresOwnerAnalyticsRepository) ListConsolidatedBillings(ctx context.Context, orgID string, unitID *string, paymentMethod *string, limit, offset int) ([]domain.Billing, int, error) {
	query := `
		SELECT 
			b.id::TEXT,
			COALESCE(b.unit_id::TEXT, ''),
			b.date,
			b.barber_id::TEXT,
			b.client_id::TEXT,
			b.service_id::TEXT,
			b.product_id::TEXT,
			b.appointment_id::TEXT,
			COALESCE(b.quantity, 1),
			b.amount_in_cents,
			b.payment_method,
			b.status,
			COALESCE(b.notes, ''),
			b.created_at,
			b.updated_at,
			COUNT(*) OVER() AS total_count
		FROM billings b
		JOIN barbershop_units u ON b.unit_id = u.id
		WHERE u.organization_id = $1
		  AND ($2::UUID IS NULL OR b.unit_id = $2)
		  AND ($3::TEXT IS NULL OR b.payment_method = $3)
		ORDER BY b.date DESC, b.created_at DESC
		LIMIT $4 OFFSET $5;
	`

	rows, err := r.db.Query(ctx, query, orgID, unitID, paymentMethod, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query consolidated billings: %w", err)
	}
	defer rows.Close()

	var billings []domain.Billing
	var totalCount int

	for rows.Next() {
		var b domain.Billing
		var serviceID, productID, appointmentID *string

		if err := rows.Scan(
			&b.ID,
			&b.UnitID,
			&b.Date,
			&b.BarberID,
			&b.ClientID,
			&serviceID,
			&productID,
			&appointmentID,
			&b.Quantity,
			&b.AmountInCents,
			&b.PaymentMethod,
			&b.Status,
			&b.Notes,
			&b.CreatedAt,
			&b.UpdatedAt,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan billing row: %w", err)
		}

		b.ServiceID = serviceID
		b.ProductID = productID
		b.AppointmentID = appointmentID
		billings = append(billings, b)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate billing rows: %w", err)
	}

	return billings, totalCount, nil
}

// GetExportBillingsData fetches rows formatted for CSV export.
func (r *PostgresOwnerAnalyticsRepository) GetExportBillingsData(ctx context.Context, orgID string, unitID *string) ([]BillingExportRow, error) {
	query := `
		SELECT 
			b.id::TEXT,
			b.date::TEXT,
			COALESCE(u.name, 'N/A') AS unit_name,
			COALESCE(bar.name, 'N/A') AS barber_name,
			COALESCE(c.name, 'N/A') AS client_name,
			COALESCE(s.name, 'N/A') AS service_name,
			b.amount_in_cents,
			b.payment_method,
			b.status,
			b.created_at::TEXT
		FROM billings b
		JOIN barbershop_units u ON b.unit_id = u.id
		LEFT JOIN barbers bar ON b.barber_id = bar.id
		LEFT JOIN clients c ON b.client_id = c.id
		LEFT JOIN services s ON b.service_id = s.id
		WHERE u.organization_id = $1
		  AND ($2::UUID IS NULL OR b.unit_id = $2)
		ORDER BY b.date DESC, b.created_at DESC;
	`

	rows, err := r.db.Query(ctx, query, orgID, unitID)
	if err != nil {
		return nil, fmt.Errorf("query export billings data: %w", err)
	}
	defer rows.Close()

	var exportRows []BillingExportRow
	for rows.Next() {
		var row BillingExportRow
		if err := rows.Scan(
			&row.ID,
			&row.Date,
			&row.UnitName,
			&row.BarberName,
			&row.ClientName,
			&row.ServiceName,
			&row.AmountInCents,
			&row.PaymentMethod,
			&row.Status,
			&row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan export row: %w", err)
		}
		exportRows = append(exportRows, row)
	}

	return exportRows, rows.Err()
}

// GetOwnerSubscription queries the organization subscription and commercial tier for the owner.
func (r *PostgresOwnerAnalyticsRepository) GetOwnerSubscription(ctx context.Context, orgID string) (*domain.OwnerSubscription, error) {
	query := `
		SELECT 
			id::TEXT, 
			name, 
			trade_name, 
			document, 
			plan, 
			monthly_fee_in_cents, 
			COALESCE(subscription_status, 'ACTIVE'), 
			is_active
		FROM organizations
		WHERE id = $1
		LIMIT 1;
	`
	var sub domain.OwnerSubscription
	var planStr, statusStr string
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&sub.OrganizationID,
		&sub.Name,
		&sub.TradeName,
		&sub.Document,
		&planStr,
		&sub.MonthlyFeeInCents,
		&statusStr,
		&sub.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("query owner subscription: %w", err)
	}
	sub.Plan = planStr
	sub.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
	sub.SuspensionReason = calculateSuspensionReason(sub.IsActive, sub.SubscriptionStatus)
	return &sub, nil
}

func calculateSuspensionReason(isActive bool, subStatus domain.SubscriptionStatus) string {
	if isActive {
		return "NONE"
	}
	switch subStatus {
	case domain.SubscriptionStatusPastDue:
		return "FINANCIAL_PAST_DUE"
	case domain.SubscriptionStatusCancelled:
		return "SUBSCRIPTION_CANCELLED"
	default:
		return "MANUAL_SUSPENSION"
	}
}

// ListEmployees retrieves all staff members belonging to an organization, filtered by active/inactive status and optional unit.
func (r *PostgresOwnerAnalyticsRepository) ListEmployees(ctx context.Context, orgID string, status string, unitID *string) ([]domain.OrganizationEmployee, error) {
	normStatus := strings.ToLower(strings.TrimSpace(status))
	targetUnitID := ""
	if unitID != nil {
		targetUnitID = strings.TrimSpace(*unitID)
	}

	query := `
		SELECT 
			u.id::TEXT,
			u.organization_id::TEXT,
			COALESCE(u.unit_id::TEXT, ''),
			COALESCE(bu.name, ''),
			u.name,
			u.email,
			u.role,
			u.is_active,
			COALESCE(b.id::TEXT, ''),
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN barbershop_units bu ON bu.id = u.unit_id
		LEFT JOIN barbers b ON b.user_id = u.id
		WHERE u.organization_id = $1
		  AND ($2 = '' OR u.unit_id::TEXT = $2)
		  AND (
			($3 = 'all') OR
			($3 = 'inactive' AND u.is_active = FALSE) OR
			(($3 = 'active' OR $3 = '') AND u.is_active = TRUE)
		  )
		ORDER BY u.name ASC;
	`

	rows, err := r.db.Query(ctx, query, orgID, targetUnitID, normStatus)
	if err != nil {
		return nil, fmt.Errorf("query employees: %w", err)
	}
	defer rows.Close()

	var employees []domain.OrganizationEmployee
	for rows.Next() {
		var emp domain.OrganizationEmployee
		var unitIDStr, barberIDStr, roleStr string
		if err := rows.Scan(
			&emp.ID,
			&emp.OrganizationID,
			&unitIDStr,
			&emp.UnitName,
			&emp.Name,
			&emp.Email,
			&roleStr,
			&emp.IsActive,
			&barberIDStr,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		if unitIDStr != "" {
			emp.UnitID = &unitIDStr
		}
		if barberIDStr != "" {
			emp.BarberID = &barberIDStr
		}
		emp.Role = domain.Role(roleStr)
		employees = append(employees, emp)
	}

	return employees, rows.Err()
}
