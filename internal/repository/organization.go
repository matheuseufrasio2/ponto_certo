package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// OrganizationRepository defines data operations for organizations/tenants.
type OrganizationRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	Update(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.Organization, error)
}

// PostgresOrganizationRepository implements OrganizationRepository using PostgreSQL.
type PostgresOrganizationRepository struct {
	db *pgxpool.Pool
}

// NewPostgresOrganizationRepository creates a new organization repository.
func NewPostgresOrganizationRepository(db *pgxpool.Pool) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}

// GetByID retrieves organization profile details by UUID.
func (r *PostgresOrganizationRepository) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	query := `
		SELECT 
			id,
			name,
			COALESCE(trade_name, '') AS trade_name,
			COALESCE(document, '') AS document,
			COALESCE(phone, '') AS phone,
			COALESCE(logo_url, '') AS logo_url,
			plan,
			monthly_fee_in_cents,
			COALESCE(subscription_status, 'ACTIVE') AS subscription_status,
			is_active,
			created_at,
			updated_at
		FROM organizations
		WHERE id = $1
		LIMIT 1;
	`

	var org domain.Organization
	var statusStr string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.TradeName,
		&org.Document,
		&org.Phone,
		&org.LogoURL,
		&org.Plan,
		&org.MonthlyFeeInCents,
		&statusStr,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("get organization by id: %w", err)
	}

	org.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
	return &org, nil
}

// Update persists changes to organization corporate profile attributes.
func (r *PostgresOrganizationRepository) Update(ctx context.Context, id string, input domain.UpdateOrganizationInput) (*domain.Organization, error) {
	query := `
		UPDATE organizations
		SET 
			name = $1,
			trade_name = $2,
			document = $3,
			phone = $4,
			logo_url = $5,
			updated_at = NOW()
		WHERE id = $6
		RETURNING 
			id,
			name,
			COALESCE(trade_name, ''),
			COALESCE(document, ''),
			COALESCE(phone, ''),
			COALESCE(logo_url, ''),
			plan,
			monthly_fee_in_cents,
			COALESCE(subscription_status, 'ACTIVE'),
			is_active,
			created_at,
			updated_at;
	`

	var tradeName, document, phone, logoURL *string
	if input.TradeName != nil {
		tradeName = input.TradeName
	}
	if input.Document != nil {
		document = input.Document
	}
	if input.Phone != nil {
		phone = input.Phone
	}
	if input.LogoURL != nil {
		logoURL = input.LogoURL
	}

	var org domain.Organization
	var statusStr string

	err := r.db.QueryRow(ctx, query,
		input.Name,
		tradeName,
		document,
		phone,
		logoURL,
		id,
	).Scan(
		&org.ID,
		&org.Name,
		&org.TradeName,
		&org.Document,
		&org.Phone,
		&org.LogoURL,
		&org.Plan,
		&org.MonthlyFeeInCents,
		&statusStr,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("update organization profile: %w", err)
	}

	org.SubscriptionStatus = domain.SubscriptionStatus(statusStr)
	return &org, nil
}
