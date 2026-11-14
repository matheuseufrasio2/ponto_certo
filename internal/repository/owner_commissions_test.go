package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

func TestPostgresOwnerCommissionsRepository_UUIDGuards(t *testing.T) {
	repo := repository.NewPostgresOwnerCommissionsRepository(nil)
	ctx := context.Background()
	validUUID := "11111111-1111-1111-1111-111111111111"

	if _, err := repo.UnitBelongsToOrg(ctx, "invalid", validUUID); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for unitID, got %v", err)
	}
	if _, err := repo.UnitBelongsToOrg(ctx, validUUID, "invalid"); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for orgID, got %v", err)
	}
	if _, err := repo.BarberBelongsToOrg(ctx, "invalid", validUUID); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for barberID, got %v", err)
	}
	if _, err := repo.BarberBelongsToOrg(ctx, validUUID, "invalid"); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for orgID, got %v", err)
	}
	if _, err := repo.UpdateAllOrganization(ctx, "invalid", 50); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for orgID in UpdateAllOrganization, got %v", err)
	}
	if _, err := repo.UpdateByUnit(ctx, validUUID, "invalid", 50); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for unitID in UpdateByUnit, got %v", err)
	}
	if _, err := repo.UpdateByBarber(ctx, validUUID, "invalid", 50); !errors.Is(err, domain.ErrInvalidUUID) {
		t.Fatalf("expected ErrInvalidUUID for barberID in UpdateByBarber, got %v", err)
	}
}
