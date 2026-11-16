package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

func TestPostgresManagerTeamRepository_UUIDGuards(t *testing.T) {
	repo := NewPostgresManagerTeamRepository(nil)
	validUUID := "11111111-1111-1111-1111-111111111111"

	t.Run("ListTeamMembers invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.ListTeamMembers(context.Background(), "invalid-uuid")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("CreateTeamMember invalid orgID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.CreateTeamMember(context.Background(), "invalid-org", validUUID, domain.CreateTeamMemberInput{}, "hash", nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("CreateTeamMember invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.CreateTeamMember(context.Background(), validUUID, "invalid-unit", domain.CreateTeamMemberInput{}, "hash", nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("ResolveBarberID invalid unitID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.ResolveBarberID(context.Background(), "invalid-unit", validUUID)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("ResolveBarberID invalid identifier returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.ResolveBarberID(context.Background(), validUUID, "invalid-identifier")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("GetBarberWorkingHours invalid barberID returns ErrInvalidUUID", func(t *testing.T) {
		_, err := repo.GetBarberWorkingHours(context.Background(), "invalid-barber")
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})

	t.Run("UpsertBarberWorkingHours invalid barberID returns ErrInvalidUUID", func(t *testing.T) {
		err := repo.UpsertBarberWorkingHours(context.Background(), "invalid-barber", nil, nil)
		if !errors.Is(err, domain.ErrInvalidUUID) {
			t.Fatalf("expected ErrInvalidUUID, got %v", err)
		}
	})
}
