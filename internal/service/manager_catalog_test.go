package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type mockManagerCatalogRepo struct {
	services      map[string]domain.Service
	products      map[string]domain.Product
	movements     []domain.InventoryMovement
	softDeleteSvc bool
	softDeletePrd bool
	errToReturn   error
}

func newMockManagerCatalogRepo() *mockManagerCatalogRepo {
	return &mockManagerCatalogRepo{
		services:  make(map[string]domain.Service),
		products:  make(map[string]domain.Product),
		movements: make([]domain.InventoryMovement, 0),
	}
}

func (m *mockManagerCatalogRepo) ListServicesByUnit(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string) ([]domain.Service, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	var res []domain.Service
	for _, s := range m.services {
		if s.UnitID == unitID {
			res = append(res, s)
		}
	}
	return res, nil
}

func (m *mockManagerCatalogRepo) GetServiceByID(ctx context.Context, id string) (*domain.Service, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	s, ok := m.services[id]
	if !ok {
		return nil, domain.ErrServiceNotFound
	}
	return &s, nil
}

func (m *mockManagerCatalogRepo) CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput) (*domain.Service, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	s := domain.Service{
		ID:              "svc-1234-5678-9012",
		UnitID:          unitID,
		Name:            input.Name,
		Category:        input.Category,
		Description:     input.Description,
		DurationMinutes: input.DurationMinutes,
		PriceInCents:    input.PriceInCents,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	m.services[s.ID] = s
	return &s, nil
}

func (m *mockManagerCatalogRepo) UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput) (*domain.Service, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	s, ok := m.services[serviceID]
	if !ok {
		return nil, domain.ErrServiceNotFound
	}
	s.Name = input.Name
	s.Category = input.Category
	s.Description = input.Description
	s.DurationMinutes = input.DurationMinutes
	s.PriceInCents = input.PriceInCents
	if input.IsActive != nil {
		s.IsActive = *input.IsActive
	}
	m.services[serviceID] = s
	return &s, nil
}

func (m *mockManagerCatalogRepo) DeleteService(ctx context.Context, unitID, serviceID string) (bool, error) {
	if m.errToReturn != nil {
		return false, m.errToReturn
	}
	if _, ok := m.services[serviceID]; !ok {
		return false, domain.ErrServiceNotFound
	}
	delete(m.services, serviceID)
	return m.softDeleteSvc, nil
}

func (m *mockManagerCatalogRepo) ListProductsByUnit(ctx context.Context, unitID string, category *domain.ProductCategory, status *string) ([]domain.Product, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	var res []domain.Product
	for _, p := range m.products {
		if p.UnitID == unitID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (m *mockManagerCatalogRepo) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return &p, nil
}

func (m *mockManagerCatalogRepo) CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput) (*domain.Product, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	p := domain.Product{
		ID:            "prd-1234-5678-9012",
		UnitID:        unitID,
		Name:          input.Name,
		Category:      input.Category,
		Description:   input.Description,
		PriceInCents:  input.PriceInCents,
		CostInCents:   input.CostInCents,
		StockQuantity: input.StockQuantity,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	m.products[p.ID] = p
	return &p, nil
}

func (m *mockManagerCatalogRepo) UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput) (*domain.Product, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	p, ok := m.products[productID]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	p.Name = input.Name
	p.Category = input.Category
	p.Description = input.Description
	p.PriceInCents = input.PriceInCents
	p.CostInCents = input.CostInCents
	if input.IsActive != nil {
		p.IsActive = *input.IsActive
	}
	m.products[productID] = p
	return &p, nil
}

func (m *mockManagerCatalogRepo) DeleteProduct(ctx context.Context, unitID, productID string) (bool, error) {
	if m.errToReturn != nil {
		return false, m.errToReturn
	}
	if _, ok := m.products[productID]; !ok {
		return false, domain.ErrProductNotFound
	}
	delete(m.products, productID)
	return m.softDeletePrd, nil
}

func (m *mockManagerCatalogRepo) AdjustProductStock(ctx context.Context, unitID, productID, userID string, input domain.AdjustProductStockInput) (*domain.Product, *domain.InventoryMovement, error) {
	if m.errToReturn != nil {
		return nil, nil, m.errToReturn
	}
	p, ok := m.products[productID]
	if !ok {
		return nil, nil, domain.ErrProductNotFound
	}
	prev := p.StockQuantity
	var next int
	switch input.MovementType {
	case domain.InventoryMovementEntry:
		next = prev + input.Quantity
	case domain.InventoryMovementExit:
		next = prev - input.Quantity
		if next < 0 {
			return nil, nil, domain.ErrInvalidProductStock
		}
	case domain.InventoryMovementAdjustment:
		next = input.Quantity
		if next < 0 {
			return nil, nil, domain.ErrInvalidProductStock
		}
	}
	p.StockQuantity = next
	m.products[productID] = p

	uID := userID
	mov := domain.InventoryMovement{
		ID:            "mov-1234-5678-9012",
		ProductID:     productID,
		UnitID:        unitID,
		UserID:        &uID,
		MovementType:  input.MovementType,
		Quantity:      input.Quantity,
		PreviousStock: prev,
		NewStock:      next,
		Reason:        input.Reason,
		CreatedAt:     time.Now().UTC(),
	}
	m.movements = append(m.movements, mov)
	return &p, &mov, nil
}

func (m *mockManagerCatalogRepo) ListProductMovements(ctx context.Context, unitID, productID string, limit int) ([]domain.InventoryMovement, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	return m.movements, nil
}

type mockAuditRecorder struct {
	recordedLogs []domain.AuditLog
}

func (a *mockAuditRecorder) Record(ctx context.Context, log domain.AuditLog) error {
	a.recordedLogs = append(a.recordedLogs, log)
	return nil
}

func TestManagerCatalogService_RBAC(t *testing.T) {
	repo := newMockManagerCatalogRepo()
	svc := service.NewManagerCatalogService(repo)
	ctx := context.Background()

	unit1 := "11111111-1111-1111-1111-111111111111"
	unit2 := "22222222-2222-2222-2222-222222222222"

	managerClaims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unit1,
	}

	t.Run("nil claims returns unauthorized", func(t *testing.T) {
		_, err := svc.ListServices(ctx, unit1, nil, nil, nil)
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("manager cannot access another unit", func(t *testing.T) {
		_, err := svc.ListServices(ctx, unit2, nil, nil, managerClaims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		_, err = svc.CreateService(ctx, unit2, domain.CreateServiceInput{}, managerClaims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		_, err = svc.CreateProduct(ctx, unit2, domain.CreateProductInput{}, managerClaims)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestManagerCatalogService_ServiceOperations(t *testing.T) {
	repo := newMockManagerCatalogRepo()
	audit := &mockAuditRecorder{}
	svc := service.NewManagerCatalogService(repo)
	svc.SetAuditRecorder(audit)
	ctx := context.Background()

	unitID := "11111111-1111-1111-1111-111111111111"
	managerClaims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("CreateService validation and audit", func(t *testing.T) {
		// invalid input
		_, err := svc.CreateService(ctx, unitID, domain.CreateServiceInput{
			Name: "",
		}, managerClaims)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		// valid input
		created, err := svc.CreateService(ctx, unitID, domain.CreateServiceInput{
			Name:            "Barba Tradicional",
			Category:        domain.ServiceCategoryBarba,
			DurationMinutes: 25,
			PriceInCents:    3500,
		}, managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.Name != "Barba Tradicional" {
			t.Fatalf("expected Barba Tradicional, got %s", created.Name)
		}

		// verify audit log
		if len(audit.recordedLogs) != 1 || audit.recordedLogs[0].Action != "CREATE_SERVICE" {
			t.Fatalf("expected CREATE_SERVICE audit log, got %+v", audit.recordedLogs)
		}
	})

	t.Run("UpdateService validation and audit", func(t *testing.T) {
		updated, err := svc.UpdateService(ctx, unitID, "svc-1234-5678-9012", domain.UpdateServiceInput{
			Name:            "Barba Terapia Completa",
			Category:        domain.ServiceCategoryBarba,
			DurationMinutes: 35,
			PriceInCents:    4500,
		}, managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.PriceInCents != 4500 {
			t.Fatalf("expected price 4500, got %d", updated.PriceInCents)
		}

		if len(audit.recordedLogs) < 2 || audit.recordedLogs[1].Action != "UPDATE_SERVICE" {
			t.Fatalf("expected UPDATE_SERVICE audit log, got %+v", audit.recordedLogs)
		}
	})

	t.Run("DeleteService audit", func(t *testing.T) {
		softDeleted, err := svc.DeleteService(ctx, unitID, "svc-1234-5678-9012", managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if softDeleted {
			t.Fatalf("expected hard delete, got softDeleted=true")
		}
		if len(audit.recordedLogs) < 3 || audit.recordedLogs[2].Action != "DELETE_SERVICE" {
			t.Fatalf("expected DELETE_SERVICE audit log, got %+v", audit.recordedLogs)
		}
	})
}

func TestManagerCatalogService_ProductAndStockOperations(t *testing.T) {
	repo := newMockManagerCatalogRepo()
	audit := &mockAuditRecorder{}
	svc := service.NewManagerCatalogService(repo)
	svc.SetAuditRecorder(audit)
	ctx := context.Background()

	unitID := "11111111-1111-1111-1111-111111111111"
	managerClaims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	t.Run("CreateProduct validation and audit", func(t *testing.T) {
		created, err := svc.CreateProduct(ctx, unitID, domain.CreateProductInput{
			Name:          "Óleo para Barba",
			Category:      domain.ProductCategoryBarbeariaCuidados,
			PriceInCents:  4000,
			CostInCents:   1500,
			StockQuantity: 10,
		}, managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.StockQuantity != 10 {
			t.Fatalf("expected stock 10, got %d", created.StockQuantity)
		}

		if len(audit.recordedLogs) != 1 || audit.recordedLogs[0].Action != "CREATE_PRODUCT" {
			t.Fatalf("expected CREATE_PRODUCT audit log, got %+v", audit.recordedLogs)
		}
	})

	t.Run("AdjustProductStock validation, negative stock block and audit", func(t *testing.T) {
		// negative stock exit
		_, _, err := svc.AdjustStock(ctx, unitID, "prd-1234-5678-9012", domain.AdjustProductStockInput{
			MovementType: domain.InventoryMovementExit,
			Quantity:     20, // previous was 10
			Reason:       "Venda avulsa",
		}, managerClaims)
		if !errors.Is(err, domain.ErrInvalidProductStock) {
			t.Fatalf("expected ErrInvalidProductStock, got %v", err)
		}

		// successful entry
		prod, mov, err := svc.AdjustStock(ctx, unitID, "prd-1234-5678-9012", domain.AdjustProductStockInput{
			MovementType: domain.InventoryMovementEntry,
			Quantity:     5,
			Reason:       "Recebimento de lote",
		}, managerClaims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prod.StockQuantity != 15 || mov.NewStock != 15 {
			t.Fatalf("expected stock 15, got %d (mov: %d)", prod.StockQuantity, mov.NewStock)
		}

		// verify audit log
		lastLog := audit.recordedLogs[len(audit.recordedLogs)-1]
		if lastLog.Action != "ADJUST_STOCK" || lastLog.TargetID != "prd-1234-5678-9012" {
			t.Fatalf("expected ADJUST_STOCK audit log, got %+v", lastLog)
		}
	})
}
