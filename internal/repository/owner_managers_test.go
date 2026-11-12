package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestPostgresOwnerManagersRepository_InvalidUUIDGuards(t *testing.T) {
	repo := NewPostgresOwnerManagersRepository(nil)
	validUUID := "11111111-1111-1111-1111-111111111111"

	t.Run("UnitBelongsToOrg returns ErrInvalidUUID for bad unitID", func(t *testing.T) {
		ok, err := repo.UnitBelongsToOrg(context.Background(), "not-a-uuid", validUUID)
		if ok {
			t.Fatalf("expected false, got true")
		}
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("UnitBelongsToOrg returns ErrInvalidUUID for bad orgID", func(t *testing.T) {
		ok, err := repo.UnitBelongsToOrg(context.Background(), validUUID, "bad-org")
		if ok {
			t.Fatalf("expected false, got true")
		}
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("ListManagers returns ErrInvalidUUID for bad orgID", func(t *testing.T) {
		_, err := repo.ListManagers(context.Background(), "bad-org")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("CreateManager returns ErrInvalidUUID for bad orgID or unitID", func(t *testing.T) {
		_, err := repo.CreateManager(context.Background(), "bad-org", validUUID, domain.CreateManagerInput{}, "hash", nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}

		_, err = repo.CreateManager(context.Background(), validUUID, "bad-unit", domain.CreateManagerInput{}, "hash", nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})
}
