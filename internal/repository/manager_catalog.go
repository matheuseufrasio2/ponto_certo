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

// ManagerCatalogRepository defines operations for manager catalog and stock management.
type ManagerCatalogRepository interface {
	ListServicesByUnit(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string) ([]domain.Service, error)
	GetServiceByID(ctx context.Context, id string) (*domain.Service, error)
	CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput) (*domain.Service, error)
	UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput) (*domain.Service, error)
	DeleteService(ctx context.Context, unitID, serviceID string) (softDeleted bool, err error)
	ListProductsByUnit(ctx context.Context, unitID string, category *domain.ProductCategory, status *string) ([]domain.Product, error)
	GetProductByID(ctx context.Context, id string) (*domain.Product, error)
	CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput) (*domain.Product, error)
	UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput) (*domain.Product, error)
	DeleteProduct(ctx context.Context, unitID, productID string) (softDeleted bool, err error)
	AdjustProductStock(ctx context.Context, unitID, productID, userID string, input domain.AdjustProductStockInput) (*domain.Product, *domain.InventoryMovement, error)
	ListProductMovements(ctx context.Context, unitID, productID string, limit int) ([]domain.InventoryMovement, error)
}

// PostgresManagerCatalogRepository implements ManagerCatalogRepository using PostgreSQL.
type PostgresManagerCatalogRepository struct {
	db *pgxpool.Pool
}

// NewPostgresManagerCatalogRepository instantiates a new repository.
func NewPostgresManagerCatalogRepository(db *pgxpool.Pool) *PostgresManagerCatalogRepository {
	return &PostgresManagerCatalogRepository{db: db}
}

// ListServicesByUnit returns services associated with a unit with optional category and status filters.
func (r *PostgresManagerCatalogRepository) ListServicesByUnit(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string) ([]domain.Service, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT id, COALESCE(unit_id::text, ''), name, category, COALESCE(description, ''), duration_minutes, price_in_cents, is_active, created_at, updated_at
		FROM services
		WHERE (unit_id = $1 OR unit_id IS NULL)
	`
	args := []any{unitID}
	argIdx := 2

	if category != nil && *category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *category)
		argIdx++
	}

	if status != nil && *status != "" {
		switch strings.ToLower(*status) {
		case "active":
			query += " AND is_active = TRUE"
		case "inactive":
			query += " AND is_active = FALSE"
		}
	}

	query += " ORDER BY name ASC;"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query services by unit: %w", err)
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		var desc, uID string
		if err := rows.Scan(
			&s.ID,
			&uID,
			&s.Name,
			&s.Category,
			&desc,
			&s.DurationMinutes,
			&s.PriceInCents,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		s.UnitID = uID
		s.Description = desc
		services = append(services, s)
	}
	return services, rows.Err()
}

// GetServiceByID returns a single service by its primary key.
func (r *PostgresManagerCatalogRepository) GetServiceByID(ctx context.Context, id string) (*domain.Service, error) {
	if !domain.IsValidUUID(id) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT id, COALESCE(unit_id::text, ''), name, category, COALESCE(description, ''), duration_minutes, price_in_cents, is_active, created_at, updated_at
		FROM services
		WHERE id = $1;
	`
	var s domain.Service
	var desc, uID string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&uID,
		&s.Name,
		&s.Category,
		&desc,
		&s.DurationMinutes,
		&s.PriceInCents,
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
	s.UnitID = uID
	s.Description = desc
	return &s, nil
}

// CreateService creates a new service in the specified unit after validating uniqueness of its name.
func (r *PostgresManagerCatalogRepository) CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput) (*domain.Service, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	trimmedName := strings.TrimSpace(input.Name)
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM services
			WHERE unit_id = $1 AND LOWER(TRIM(name)) = LOWER($2)
		);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, checkQuery, unitID, trimmedName).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check service name duplicate: %w", err)
	}
	if exists {
		return nil, domain.ErrServiceNameAlreadyExists
	}

	insertQuery := `
		INSERT INTO services (unit_id, name, category, description, duration_minutes, price_in_cents, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		RETURNING id, unit_id, name, category, COALESCE(description, ''), duration_minutes, price_in_cents, is_active, created_at, updated_at;
	`
	var s domain.Service
	var desc, uID string
	err := r.db.QueryRow(ctx, insertQuery, unitID, trimmedName, input.Category, input.Description, input.DurationMinutes, input.PriceInCents).Scan(
		&s.ID,
		&uID,
		&s.Name,
		&s.Category,
		&desc,
		&s.DurationMinutes,
		&s.PriceInCents,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert service: %w", err)
	}
	s.UnitID = uID
	s.Description = desc
	return &s, nil
}

// UpdateService updates an existing service belonging to the unit.
func (r *PostgresManagerCatalogRepository) UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput) (*domain.Service, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(serviceID) {
		return nil, domain.ErrInvalidUUID
	}

	trimmedName := strings.TrimSpace(input.Name)
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM services
			WHERE unit_id = $1 AND LOWER(TRIM(name)) = LOWER($2) AND id != $3
		);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, checkQuery, unitID, trimmedName, serviceID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check service name duplicate on update: %w", err)
	}
	if exists {
		return nil, domain.ErrServiceNameAlreadyExists
	}

	updateQuery := `
		UPDATE services
		SET name = $1,
		    category = $2,
		    description = $3,
		    duration_minutes = $4,
		    price_in_cents = $5,
		    is_active = COALESCE($6, is_active),
		    updated_at = NOW()
		WHERE id = $7 AND unit_id = $8
		RETURNING id, unit_id, name, category, COALESCE(description, ''), duration_minutes, price_in_cents, is_active, created_at, updated_at;
	`
	var s domain.Service
	var desc, uID string
	err := r.db.QueryRow(ctx, updateQuery, trimmedName, input.Category, input.Description, input.DurationMinutes, input.PriceInCents, input.IsActive, serviceID, unitID).Scan(
		&s.ID,
		&uID,
		&s.Name,
		&s.Category,
		&desc,
		&s.DurationMinutes,
		&s.PriceInCents,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("update service: %w", err)
	}
	s.UnitID = uID
	s.Description = desc
	return &s, nil
}

// DeleteService safely removes a service. If linked appointments or billings exist, performs soft-delete (is_active = false).
func (r *PostgresManagerCatalogRepository) DeleteService(ctx context.Context, unitID, serviceID string) (bool, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(serviceID) {
		return false, domain.ErrInvalidUUID
	}

	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM services WHERE id = $1 AND unit_id = $2);", serviceID, unitID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("verify service exists: %w", err)
	}
	if !exists {
		return false, domain.ErrServiceNotFound
	}

	checkRefsQuery := `
		SELECT (
			(SELECT COUNT(*) FROM appointments WHERE service_id = $1) +
			(SELECT COUNT(*) FROM billings WHERE service_id = $1)
		);
	`
	var refCount int
	if err := r.db.QueryRow(ctx, checkRefsQuery, serviceID).Scan(&refCount); err != nil {
		return false, fmt.Errorf("check service references: %w", err)
	}

	if refCount > 0 {
		_, err := r.db.Exec(ctx, "UPDATE services SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND unit_id = $2;", serviceID, unitID)
		if err != nil {
			return false, fmt.Errorf("soft delete service: %w", err)
		}
		return true, nil
	}

	_, err = r.db.Exec(ctx, "DELETE FROM services WHERE id = $1 AND unit_id = $2;", serviceID, unitID)
	if err != nil {
		return false, fmt.Errorf("hard delete service: %w", err)
	}
	return false, nil
}

// ListProductsByUnit returns products associated with a unit with optional category and status filters.
func (r *PostgresManagerCatalogRepository) ListProductsByUnit(ctx context.Context, unitID string, category *domain.ProductCategory, status *string) ([]domain.Product, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at
		FROM products
		WHERE unit_id = $1
	`
	args := []any{unitID}
	argIdx := 2

	if category != nil && *category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *category)
		argIdx++
	}

	if status != nil && *status != "" {
		switch strings.ToLower(*status) {
		case "active":
			query += " AND is_active = TRUE"
		case "inactive":
			query += " AND is_active = FALSE"
		case "low_stock":
			query += " AND stock_quantity <= 5"
		}
	}

	query += " ORDER BY name ASC;"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products by unit: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var desc string
		if err := rows.Scan(
			&p.ID,
			&p.UnitID,
			&p.Name,
			&p.Category,
			&desc,
			&p.PriceInCents,
			&p.CostInCents,
			&p.StockQuantity,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		p.Description = desc
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetProductByID returns a single product by primary key.
func (r *PostgresManagerCatalogRepository) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	if !domain.IsValidUUID(id) {
		return nil, domain.ErrInvalidUUID
	}

	query := `
		SELECT id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at
		FROM products
		WHERE id = $1;
	`
	var p domain.Product
	var desc string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.UnitID,
		&p.Name,
		&p.Category,
		&desc,
		&p.PriceInCents,
		&p.CostInCents,
		&p.StockQuantity,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("query product by id: %w", err)
	}
	p.Description = desc
	return &p, nil
}

// CreateProduct creates a new retail product in the unit.
func (r *PostgresManagerCatalogRepository) CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput) (*domain.Product, error) {
	if !domain.IsValidUUID(unitID) {
		return nil, domain.ErrInvalidUUID
	}

	trimmedName := strings.TrimSpace(input.Name)
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM products
			WHERE unit_id = $1 AND LOWER(TRIM(name)) = LOWER($2)
		);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, checkQuery, unitID, trimmedName).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check product name duplicate: %w", err)
	}
	if exists {
		return nil, domain.ErrProductNameAlreadyExists
	}

	insertQuery := `
		INSERT INTO products (unit_id, name, category, description, price_in_cents, cost_in_cents, stock_quantity, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at;
	`
	var p domain.Product
	var desc string
	err := r.db.QueryRow(ctx, insertQuery, unitID, trimmedName, input.Category, input.Description, input.PriceInCents, input.CostInCents, input.StockQuantity).Scan(
		&p.ID,
		&p.UnitID,
		&p.Name,
		&p.Category,
		&desc,
		&p.PriceInCents,
		&p.CostInCents,
		&p.StockQuantity,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}
	p.Description = desc
	return &p, nil
}

// UpdateProduct updates an existing product in the unit.
func (r *PostgresManagerCatalogRepository) UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput) (*domain.Product, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(productID) {
		return nil, domain.ErrInvalidUUID
	}

	trimmedName := strings.TrimSpace(input.Name)
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM products
			WHERE unit_id = $1 AND LOWER(TRIM(name)) = LOWER($2) AND id != $3
		);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, checkQuery, unitID, trimmedName, productID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check product name duplicate on update: %w", err)
	}
	if exists {
		return nil, domain.ErrProductNameAlreadyExists
	}

	updateQuery := `
		UPDATE products
		SET name = $1,
		    category = $2,
		    description = $3,
		    price_in_cents = $4,
		    cost_in_cents = $5,
		    is_active = COALESCE($6, is_active),
		    updated_at = NOW()
		WHERE id = $7 AND unit_id = $8
		RETURNING id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at;
	`
	var p domain.Product
	var desc string
	err := r.db.QueryRow(ctx, updateQuery, trimmedName, input.Category, input.Description, input.PriceInCents, input.CostInCents, input.IsActive, productID, unitID).Scan(
		&p.ID,
		&p.UnitID,
		&p.Name,
		&p.Category,
		&desc,
		&p.PriceInCents,
		&p.CostInCents,
		&p.StockQuantity,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("update product: %w", err)
	}
	p.Description = desc
	return &p, nil
}

// DeleteProduct safely removes a product. If movement history exists, performs soft-delete (is_active = false).
func (r *PostgresManagerCatalogRepository) DeleteProduct(ctx context.Context, unitID, productID string) (bool, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(productID) {
		return false, domain.ErrInvalidUUID
	}

	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND unit_id = $2);", productID, unitID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("verify product exists: %w", err)
	}
	if !exists {
		return false, domain.ErrProductNotFound
	}

	var movementCount int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM product_inventory_movements WHERE product_id = $1;", productID).Scan(&movementCount)
	if err != nil {
		return false, fmt.Errorf("check product inventory movements: %w", err)
	}

	if movementCount > 0 {
		_, err := r.db.Exec(ctx, "UPDATE products SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND unit_id = $2;", productID, unitID)
		if err != nil {
			return false, fmt.Errorf("soft delete product: %w", err)
		}
		return true, nil
	}

	_, err = r.db.Exec(ctx, "DELETE FROM products WHERE id = $1 AND unit_id = $2;", productID, unitID)
	if err != nil {
		return false, fmt.Errorf("hard delete product: %w", err)
	}
	return false, nil
}

// AdjustProductStock atomically adjusts product balance and creates an audit movement record using FOR UPDATE row locking.
func (r *PostgresManagerCatalogRepository) AdjustProductStock(ctx context.Context, unitID, productID, userID string, input domain.AdjustProductStockInput) (*domain.Product, *domain.InventoryMovement, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(productID) {
		return nil, nil, domain.ErrInvalidUUID
	}
	if userID != "" && !domain.IsValidUUID(userID) {
		return nil, nil, domain.ErrInvalidUUID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin adjust stock tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	selectQuery := `
		SELECT id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at
		FROM products
		WHERE id = $1 AND unit_id = $2
		FOR UPDATE;
	`
	var p domain.Product
	var desc string
	err = tx.QueryRow(ctx, selectQuery, productID, unitID).Scan(
		&p.ID,
		&p.UnitID,
		&p.Name,
		&p.Category,
		&desc,
		&p.PriceInCents,
		&p.CostInCents,
		&p.StockQuantity,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrProductNotFound
		}
		return nil, nil, fmt.Errorf("lock product for stock adjustment: %w", err)
	}
	p.Description = desc

	prevStock := p.StockQuantity
	var newStock int
	switch input.MovementType {
	case domain.InventoryMovementEntry:
		newStock = prevStock + input.Quantity
	case domain.InventoryMovementExit:
		newStock = prevStock - input.Quantity
		if newStock < 0 {
			return nil, nil, domain.ErrInvalidProductStock
		}
	case domain.InventoryMovementAdjustment:
		newStock = input.Quantity
		if newStock < 0 {
			return nil, nil, domain.ErrInvalidProductStock
		}
	default:
		return nil, nil, domain.ErrValidationFailed
	}

	updateQuery := `
		UPDATE products
		SET stock_quantity = $1, updated_at = NOW()
		WHERE id = $2 AND unit_id = $3
		RETURNING id, unit_id, name, category, COALESCE(description, ''), price_in_cents, cost_in_cents, stock_quantity, is_active, created_at, updated_at;
	`
	err = tx.QueryRow(ctx, updateQuery, newStock, productID, unitID).Scan(
		&p.ID,
		&p.UnitID,
		&p.Name,
		&p.Category,
		&desc,
		&p.PriceInCents,
		&p.CostInCents,
		&p.StockQuantity,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("update product stock quantity: %w", err)
	}
	p.Description = desc

	insertMovQuery := `
		INSERT INTO product_inventory_movements (product_id, unit_id, user_id, movement_type, quantity, previous_stock, new_stock, reason)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8)
		RETURNING id, product_id, unit_id, user_id, movement_type, quantity, previous_stock, new_stock, reason, created_at;
	`
	var mov domain.InventoryMovement
	var uID *string
	err = tx.QueryRow(ctx, insertMovQuery, productID, unitID, userID, input.MovementType, input.Quantity, prevStock, newStock, input.Reason).Scan(
		&mov.ID,
		&mov.ProductID,
		&mov.UnitID,
		&uID,
		&mov.MovementType,
		&mov.Quantity,
		&mov.PreviousStock,
		&mov.NewStock,
		&mov.Reason,
		&mov.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("insert inventory movement: %w", err)
	}
	mov.UserID = uID

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit adjust stock tx: %w", err)
	}

	return &p, &mov, nil
}

// ListProductMovements returns historical inventory movements for a product ordered by most recent first.
func (r *PostgresManagerCatalogRepository) ListProductMovements(ctx context.Context, unitID, productID string, limit int) ([]domain.InventoryMovement, error) {
	if !domain.IsValidUUID(unitID) || !domain.IsValidUUID(productID) {
		return nil, domain.ErrInvalidUUID
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT m.id, m.product_id, m.unit_id, m.user_id, COALESCE(u.name, ''), m.movement_type, m.quantity, m.previous_stock, m.new_stock, m.reason, m.created_at
		FROM product_inventory_movements m
		LEFT JOIN users u ON u.id = m.user_id
		WHERE m.unit_id = $1 AND m.product_id = $2
		ORDER BY m.created_at DESC
		LIMIT $3;
	`
	rows, err := r.db.Query(ctx, query, unitID, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("query product movements: %w", err)
	}
	defer rows.Close()

	var movements []domain.InventoryMovement
	for rows.Next() {
		var m domain.InventoryMovement
		var uID *string
		var uName string
		if err := rows.Scan(
			&m.ID,
			&m.ProductID,
			&m.UnitID,
			&uID,
			&uName,
			&m.MovementType,
			&m.Quantity,
			&m.PreviousStock,
			&m.NewStock,
			&m.Reason,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product movement: %w", err)
		}
		m.UserID = uID
		if uName != "" {
			m.UserName = &uName
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}
