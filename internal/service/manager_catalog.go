package service

import (
	"context"
	"encoding/json"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// ManagerCatalogRepository defines operations needed by ManagerCatalogService.
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

// ManagerCatalogService orchestrates catalog and inventory management for branch managers.
type ManagerCatalogService struct {
	repo          ManagerCatalogRepository
	auditRecorder AuditLogRecorder
}

// NewManagerCatalogService creates a new ManagerCatalogService.
func NewManagerCatalogService(repo ManagerCatalogRepository) *ManagerCatalogService {
	return &ManagerCatalogService{repo: repo}
}

// SetAuditRecorder sets the compliance audit log recorder.
func (s *ManagerCatalogService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

func (s *ManagerCatalogService) recordAudit(ctx context.Context, claims *UserClaims, unitID, action, entity, targetID string, details any) {
	if s.auditRecorder == nil {
		return
	}
	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	var role domain.Role
	if claims != nil {
		if claims.UserID != "" {
			actorID = &claims.UserID
		}
		role = claims.Role
	}

	payload, _ := json.Marshal(details)
	_ = s.auditRecorder.Record(ctx, domain.AuditLog{
		UnitID:       &unitID,
		ActorID:      actorID,
		ActorRole:    role,
		Action:       action,
		TargetEntity: entity,
		TargetID:     targetID,
		Details:      payload,
		IPAddress:    meta.IPAddress,
		UserAgent:    meta.UserAgent,
	})
}

// ListServices lists all services configured for the unit.
func (s *ManagerCatalogService) ListServices(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string, claims *UserClaims) ([]domain.Service, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListServicesByUnit(ctx, unitID, category, status)
}

// GetService returns a service by ID if authorized.
func (s *ManagerCatalogService) GetService(ctx context.Context, unitID, serviceID string, claims *UserClaims) (*domain.Service, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}

	svc, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if svc.UnitID != "" && svc.UnitID != unitID {
		return nil, domain.ErrForbidden
	}
	return svc, nil
}

// CreateService creates a new service for the unit with audit logging.
func (s *ManagerCatalogService) CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput, claims *UserClaims) (*domain.Service, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	svc, err := s.repo.CreateService(ctx, unitID, input)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, claims, unitID, "CREATE_SERVICE", "services", svc.ID, map[string]any{
		"name":             svc.Name,
		"category":         svc.Category,
		"price_in_cents":   svc.PriceInCents,
		"duration_minutes": svc.DurationMinutes,
	})

	return svc, nil
}

// UpdateService updates a unit service with audit logging.
func (s *ManagerCatalogService) UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput, claims *UserClaims) (*domain.Service, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	svc, err := s.repo.UpdateService(ctx, unitID, serviceID, input)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, claims, unitID, "UPDATE_SERVICE", "services", serviceID, map[string]any{
		"name":             svc.Name,
		"category":         svc.Category,
		"price_in_cents":   svc.PriceInCents,
		"duration_minutes": svc.DurationMinutes,
		"is_active":        svc.IsActive,
	})

	return svc, nil
}

// DeleteService removes or deactivates a service with audit logging.
func (s *ManagerCatalogService) DeleteService(ctx context.Context, unitID, serviceID string, claims *UserClaims) (bool, error) {
	if claims == nil {
		return false, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return false, domain.ErrForbidden
	}

	softDeleted, err := s.repo.DeleteService(ctx, unitID, serviceID)
	if err != nil {
		return false, err
	}

	s.recordAudit(ctx, claims, unitID, "DELETE_SERVICE", "services", serviceID, map[string]any{
		"soft_deleted": softDeleted,
	})

	return softDeleted, nil
}

// ListProducts lists products of a unit with optional filters.
func (s *ManagerCatalogService) ListProducts(ctx context.Context, unitID string, category *domain.ProductCategory, status *string, claims *UserClaims) ([]domain.Product, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListProductsByUnit(ctx, unitID, category, status)
}

// GetProduct returns a single product by ID.
func (s *ManagerCatalogService) GetProduct(ctx context.Context, unitID, productID string, claims *UserClaims) (*domain.Product, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}

	prod, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if prod.UnitID != unitID {
		return nil, domain.ErrForbidden
	}
	return prod, nil
}

// CreateProduct creates a new retail product in the unit.
func (s *ManagerCatalogService) CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput, claims *UserClaims) (*domain.Product, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	prod, err := s.repo.CreateProduct(ctx, unitID, input)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, claims, unitID, "CREATE_PRODUCT", "products", prod.ID, map[string]any{
		"name":           prod.Name,
		"category":       prod.Category,
		"price_in_cents": prod.PriceInCents,
		"cost_in_cents":  prod.CostInCents,
		"stock_quantity": prod.StockQuantity,
	})

	return prod, nil
}

// UpdateProduct updates retail product details.
func (s *ManagerCatalogService) UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput, claims *UserClaims) (*domain.Product, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	prod, err := s.repo.UpdateProduct(ctx, unitID, productID, input)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, claims, unitID, "UPDATE_PRODUCT", "products", productID, map[string]any{
		"name":           prod.Name,
		"category":       prod.Category,
		"price_in_cents": prod.PriceInCents,
		"cost_in_cents":  prod.CostInCents,
		"is_active":      prod.IsActive,
	})

	return prod, nil
}

// DeleteProduct removes or deactivates a product with audit logging.
func (s *ManagerCatalogService) DeleteProduct(ctx context.Context, unitID, productID string, claims *UserClaims) (bool, error) {
	if claims == nil {
		return false, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return false, domain.ErrForbidden
	}

	softDeleted, err := s.repo.DeleteProduct(ctx, unitID, productID)
	if err != nil {
		return false, err
	}

	s.recordAudit(ctx, claims, unitID, "DELETE_PRODUCT", "products", productID, map[string]any{
		"soft_deleted": softDeleted,
	})

	return softDeleted, nil
}

// AdjustStock adjusts product stock balance and creates an audit movement log.
func (s *ManagerCatalogService) AdjustStock(ctx context.Context, unitID, productID string, input domain.AdjustProductStockInput, claims *UserClaims) (*domain.Product, *domain.InventoryMovement, error) {
	if claims == nil {
		return nil, nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, nil, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	var userID string
	if claims.UserID != "" {
		userID = claims.UserID
	}

	prod, mov, err := s.repo.AdjustProductStock(ctx, unitID, productID, userID, input)
	if err != nil {
		return nil, nil, err
	}

	s.recordAudit(ctx, claims, unitID, "ADJUST_STOCK", "products", productID, map[string]any{
		"movement_id":    mov.ID,
		"movement_type":  mov.MovementType,
		"quantity":       mov.Quantity,
		"previous_stock": mov.PreviousStock,
		"new_stock":      mov.NewStock,
		"reason":         mov.Reason,
	})

	return prod, mov, nil
}

// ListProductMovements lists inventory movement history for a product.
func (s *ManagerCatalogService) ListProductMovements(ctx context.Context, unitID, productID string, limit int, claims *UserClaims) ([]domain.InventoryMovement, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if !claims.HasUnitAccess(unitID) {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListProductMovements(ctx, unitID, productID, limit)
}
