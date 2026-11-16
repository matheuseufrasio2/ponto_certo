package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type stubManagerCatalogManager struct {
	services       []domain.Service
	service        *domain.Service
	products       []domain.Product
	product        *domain.Product
	movements      []domain.InventoryMovement
	movement       *domain.InventoryMovement
	softDeletedSvc bool
	softDeletedPrd bool
	errToReturn    error
}

func (s *stubManagerCatalogManager) ListServices(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string, claims *service.UserClaims) ([]domain.Service, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.services, nil
}

func (s *stubManagerCatalogManager) GetService(ctx context.Context, unitID, serviceID string, claims *service.UserClaims) (*domain.Service, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.service, nil
}

func (s *stubManagerCatalogManager) CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput, claims *service.UserClaims) (*domain.Service, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.service, nil
}

func (s *stubManagerCatalogManager) UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput, claims *service.UserClaims) (*domain.Service, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.service, nil
}

func (s *stubManagerCatalogManager) DeleteService(ctx context.Context, unitID, serviceID string, claims *service.UserClaims) (bool, error) {
	if s.errToReturn != nil {
		return false, s.errToReturn
	}
	return s.softDeletedSvc, nil
}

func (s *stubManagerCatalogManager) ListProducts(ctx context.Context, unitID string, category *domain.ProductCategory, status *string, claims *service.UserClaims) ([]domain.Product, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.products, nil
}

func (s *stubManagerCatalogManager) GetProduct(ctx context.Context, unitID, productID string, claims *service.UserClaims) (*domain.Product, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.product, nil
}

func (s *stubManagerCatalogManager) CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput, claims *service.UserClaims) (*domain.Product, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.product, nil
}

func (s *stubManagerCatalogManager) UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput, claims *service.UserClaims) (*domain.Product, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.product, nil
}

func (s *stubManagerCatalogManager) DeleteProduct(ctx context.Context, unitID, productID string, claims *service.UserClaims) (bool, error) {
	if s.errToReturn != nil {
		return false, s.errToReturn
	}
	return s.softDeletedPrd, nil
}

func (s *stubManagerCatalogManager) AdjustStock(ctx context.Context, unitID, productID string, input domain.AdjustProductStockInput, claims *service.UserClaims) (*domain.Product, *domain.InventoryMovement, error) {
	if s.errToReturn != nil {
		return nil, nil, s.errToReturn
	}
	return s.product, s.movement, nil
}

func (s *stubManagerCatalogManager) ListProductMovements(ctx context.Context, unitID, productID string, limit int, claims *service.UserClaims) ([]domain.InventoryMovement, error) {
	if s.errToReturn != nil {
		return nil, s.errToReturn
	}
	return s.movements, nil
}

func TestManagerCatalogHandler_Services(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	serviceID := "22222222-2222-2222-2222-222222222222"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	dummyService := &domain.Service{
		ID:              serviceID,
		UnitID:          unitID,
		Name:            "Barba Terapia",
		Category:        domain.ServiceCategoryBarba,
		DurationMinutes: 30,
		PriceInCents:    4500,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	t.Run("GET /services 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			services: []domain.Service{*dummyService},
		})
		r := chi.NewRouter()
		r.Get("/services", h.ListServices)

		req := httptest.NewRequest(http.MethodGet, "/services", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("POST /services 201 Created", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			service: dummyService,
		})
		r := chi.NewRouter()
		r.Post("/services", h.CreateService)

		body := domain.CreateServiceInput{
			Name:            "Barba Terapia",
			Category:        domain.ServiceCategoryBarba,
			DurationMinutes: 30,
			PriceInCents:    4500,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("POST /services 409 Conflict on duplicate name", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			errToReturn: domain.ErrServiceNameAlreadyExists,
		})
		r := chi.NewRouter()
		r.Post("/services", h.CreateService)

		body := domain.CreateServiceInput{
			Name:            "Barba Terapia",
			Category:        domain.ServiceCategoryBarba,
			DurationMinutes: 30,
			PriceInCents:    4500,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("GET /services/{id} 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			service: dummyService,
		})
		r := chi.NewRouter()
		r.Get("/services/{id}", h.GetService)

		req := httptest.NewRequest(http.MethodGet, "/services/"+serviceID, nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET /services/{id} 404 Not Found", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			errToReturn: domain.ErrServiceNotFound,
		})
		r := chi.NewRouter()
		r.Get("/services/{id}", h.GetService)

		req := httptest.NewRequest(http.MethodGet, "/services/"+serviceID, nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("PUT /services/{id} 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			service: dummyService,
		})
		r := chi.NewRouter()
		r.Put("/services/{id}", h.UpdateService)

		body := domain.UpdateServiceInput{
			Name:            "Barba Terapia VIP",
			Category:        domain.ServiceCategoryBarba,
			DurationMinutes: 40,
			PriceInCents:    5000,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/services/"+serviceID, bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("DELETE /services/{id} 200 OK on soft delete", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			softDeletedSvc: true,
		})
		r := chi.NewRouter()
		r.Delete("/services/{id}", h.DeleteService)

		req := httptest.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("DELETE /services/{id} 204 No Content on hard delete", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			softDeletedSvc: false,
		})
		r := chi.NewRouter()
		r.Delete("/services/{id}", h.DeleteService)

		req := httptest.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestManagerCatalogHandler_ProductsAndStock(t *testing.T) {
	unitID := "11111111-1111-1111-1111-111111111111"
	productID := "33333333-3333-3333-3333-333333333333"
	claims := &service.UserClaims{
		UserID: "user-mgr-1",
		Role:   domain.RoleUnitManager,
		UnitID: &unitID,
	}

	dummyProduct := &domain.Product{
		ID:            productID,
		UnitID:        unitID,
		Name:          "Pomada Modeladora",
		Category:      domain.ProductCategoryFinalizadores,
		PriceInCents:  5500,
		CostInCents:   2200,
		StockQuantity: 15,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	dummyMovement := &domain.InventoryMovement{
		ID:            "mov-1",
		ProductID:     productID,
		UnitID:        unitID,
		MovementType:  domain.InventoryMovementEntry,
		Quantity:      5,
		PreviousStock: 10,
		NewStock:      15,
		Reason:        "Lote novo",
		CreatedAt:     time.Now().UTC(),
	}

	t.Run("GET /products 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			products: []domain.Product{*dummyProduct},
		})
		r := chi.NewRouter()
		r.Get("/products", h.ListProducts)

		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("POST /products 201 Created", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			product: dummyProduct,
		})
		r := chi.NewRouter()
		r.Post("/products", h.CreateProduct)

		body := domain.CreateProductInput{
			Name:          "Pomada Modeladora",
			Category:      domain.ProductCategoryFinalizadores,
			PriceInCents:  5500,
			CostInCents:   2200,
			StockQuantity: 15,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("PATCH /products/{id}/stock 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			product:  dummyProduct,
			movement: dummyMovement,
		})
		r := chi.NewRouter()
		r.Patch("/products/{id}/stock", h.AdjustStock)

		body := domain.AdjustProductStockInput{
			MovementType: domain.InventoryMovementEntry,
			Quantity:     5,
			Reason:       "Lote novo",
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/products/"+productID+"/stock", bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("PATCH /products/{id}/stock 400 Bad Request on negative stock", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			errToReturn: domain.ErrInvalidProductStock,
		})
		r := chi.NewRouter()
		r.Patch("/products/{id}/stock", h.AdjustStock)

		body := domain.AdjustProductStockInput{
			MovementType: domain.InventoryMovementExit,
			Quantity:     99,
			Reason:       "Saída excessiva",
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/products/"+productID+"/stock", bytes.NewBuffer(raw))
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("GET /products/{id}/movements 200 OK", func(t *testing.T) {
		h := handler.NewManagerCatalogHandler(&stubManagerCatalogManager{
			movements: []domain.InventoryMovement{*dummyMovement},
		})
		r := chi.NewRouter()
		r.Get("/products/{id}/movements", h.ListProductMovements)

		req := httptest.NewRequest(http.MethodGet, "/products/"+productID+"/movements", nil)
		req = req.WithContext(middleware.WithUser(req.Context(), claims))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}
