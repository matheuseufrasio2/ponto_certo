package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// PostgresAuditLogRepository handles immutable audit log records in PostgreSQL.
type PostgresAuditLogRepository struct {
	db *pgxpool.Pool
}

// NewPostgresAuditLogRepository instantiates a new PostgresAuditLogRepository.
func NewPostgresAuditLogRepository(db *pgxpool.Pool) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

// RecordTx inserts an audit log entry within an existing database transaction.
func (r *PostgresAuditLogRepository) RecordTx(ctx context.Context, tx pgx.Tx, log domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			organization_id,
			unit_id,
			actor_id,
			actor_role,
			action,
			target_entity,
			target_id,
			details,
			ip_address,
			user_agent,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
	`
	detailsJSON := log.Details
	if len(detailsJSON) == 0 {
		detailsJSON = []byte("{}")
	}

	_, err := tx.Exec(ctx, query,
		log.OrganizationID,
		log.UnitID,
		log.ActorID,
		string(log.ActorRole),
		log.Action,
		log.TargetEntity,
		log.TargetID,
		detailsJSON,
		log.IPAddress,
		log.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("insert audit log in tx: %w", err)
	}
	return nil
}

// Record inserts an audit log entry in a standalone query.
func (r *PostgresAuditLogRepository) Record(ctx context.Context, log domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			organization_id,
			unit_id,
			actor_id,
			actor_role,
			action,
			target_entity,
			target_id,
			details,
			ip_address,
			user_agent,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
	`
	detailsJSON := log.Details
	if len(detailsJSON) == 0 {
		detailsJSON = []byte("{}")
	}

	_, err := r.db.Exec(ctx, query,
		log.OrganizationID,
		log.UnitID,
		log.ActorID,
		string(log.ActorRole),
		log.Action,
		log.TargetEntity,
		log.TargetID,
		detailsJSON,
		log.IPAddress,
		log.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// List retrieves audit logs paginated, optionally filtered by organization.
func (r *PostgresAuditLogRepository) List(ctx context.Context, orgID *string, limit, offset int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var query string
	var args []any

	if orgID != nil && *orgID != "" {
		query = `
			SELECT 
				id, organization_id, unit_id, actor_id, actor_role, action, target_entity, target_id,
				details, ip_address, user_agent, created_at
			FROM audit_logs
			WHERE organization_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3;
		`
		args = []any{*orgID, limit, offset}
	} else {
		query = `
			SELECT 
				id, organization_id, unit_id, actor_id, actor_role, action, target_entity, target_id,
				details, ip_address, user_agent, created_at
			FROM audit_logs
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2;
		`
		args = []any{limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var roleStr string
		var ipAddr, uAgent *string

		if err := rows.Scan(
			&l.ID,
			&l.OrganizationID,
			&l.UnitID,
			&l.ActorID,
			&roleStr,
			&l.Action,
			&l.TargetEntity,
			&l.TargetID,
			&l.Details,
			&ipAddr,
			&uAgent,
			&l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log row: %w", err)
		}
		l.ActorRole = domain.Role(roleStr)
		if ipAddr != nil {
			l.IPAddress = *ipAddr
		}
		if uAgent != nil {
			l.UserAgent = *uAgent
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit logs rows error: %w", err)
	}

	return logs, nil
}
