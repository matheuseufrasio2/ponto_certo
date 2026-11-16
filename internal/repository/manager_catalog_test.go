package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

func TestPostgresManagerCatalogRepository_UUIDGuards(t *testing.T) {
	repo := repository.NewPostgresManagerCatalogRepository(nil)
	ctx := context.Background()
	validUUID := "11111111-1111-1111-1111-111111111111"

	t.Run("ListServicesByUnit invalid UUID", func(t *testing.T) {
		_, err := repo.ListServicesByUnit(ctx, "invalid-uuid", nil, nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("GetServiceByID invalid UUID", func(t *testing.T) {
		_, err := repo.GetServiceByID(ctx, "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("CreateService invalid unit UUID", func(t *testing.T) {
		_, err := repo.CreateService(ctx, "invalid-uuid", domain.CreateServiceInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("UpdateService invalid UUIDs", func(t *testing.T) {
		_, err := repo.UpdateService(ctx, "invalid-uuid", validUUID, domain.UpdateServiceInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, err = repo.UpdateService(ctx, validUUID, "invalid-uuid", domain.UpdateServiceInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for serviceID, got %v", err)
		}
	})

	t.Run("DeleteService invalid UUIDs", func(t *testing.T) {
		_, err := repo.DeleteService(ctx, "invalid-uuid", validUUID)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, err = repo.DeleteService(ctx, validUUID, "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for serviceID, got %v", err)
		}
	})

	t.Run("ListProductsByUnit invalid UUID", func(t *testing.T) {
		_, err := repo.ListProductsByUnit(ctx, "invalid-uuid", nil, nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("GetProductByID invalid UUID", func(t *testing.T) {
		_, err := repo.GetProductByID(ctx, "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("CreateProduct invalid unit UUID", func(t *testing.T) {
		_, err := repo.CreateProduct(ctx, "invalid-uuid", domain.CreateProductInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("UpdateProduct invalid UUIDs", func(t *testing.T) {
		_, err := repo.UpdateProduct(ctx, "invalid-uuid", validUUID, domain.UpdateProductInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, err = repo.UpdateProduct(ctx, validUUID, "invalid-uuid", domain.UpdateProductInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for productID, got %v", err)
		}
	})

	t.Run("DeleteProduct invalid UUIDs", func(t *testing.T) {
		_, err := repo.DeleteProduct(ctx, "invalid-uuid", validUUID)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, err = repo.DeleteProduct(ctx, validUUID, "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for productID, got %v", err)
		}
	})

	t.Run("AdjustProductStock invalid UUIDs", func(t *testing.T) {
		_, _, err := repo.AdjustProductStock(ctx, "invalid-uuid", validUUID, validUUID, domain.AdjustProductStockInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, _, err = repo.AdjustProductStock(ctx, validUUID, "invalid-uuid", validUUID, domain.AdjustProductStockInput{})
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for productID, got %v", err)
		}
	})

	t.Run("ListProductMovements invalid UUIDs", func(t *testing.T) {
		_, err := repo.ListProductMovements(ctx, "invalid-uuid", validUUID, 10)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
		}

		_, err = repo.ListProductMovements(ctx, validUUID, "invalid-uuid", 10)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID for productID, got %v", err)
		}
	})
}
