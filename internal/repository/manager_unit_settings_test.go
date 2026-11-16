package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestPostgresManagerUnitSettingsRepository_UUIDGuards(t *testing.T) {
	repo := NewPostgresManagerUnitSettingsRepository(nil)

	t.Run("GetUnitSettings invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.GetUnitSettings(context.Background(), "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("GetUnitOperatingHours invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.GetUnitOperatingHours(context.Background(), "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("UpdateUnitSettings invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.UpdateUnitSettings(context.Background(), "invalid-uuid", domain.UpdateUnitSettingsInput{}, nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})
}
