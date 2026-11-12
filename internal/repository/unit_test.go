package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestPostgresUnitRepository_IsUnitActive_InvalidUUID(t *testing.T) {
	repo := NewPostgresUnitRepository(nil)
	active, err := repo.IsUnitActive(context.Background(), "not-a-uuid")
	if active {
		t.Fatalf("expected active=false for invalid uuid")
	}
	if !errors.Is(err, domain.ErrUnitNotFound) {
		t.Fatalf("expected ErrUnitNotFound, got %v", err)
	}
}
