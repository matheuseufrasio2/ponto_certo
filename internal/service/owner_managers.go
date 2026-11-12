package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// OwnerManagersRepository defines data access for owner-managed unit managers.
type OwnerManagersRepository interface {
	UnitBelongsToOrg(ctx context.Context, unitID, orgID string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	ListManagers(ctx context.Context, orgID string) ([]domain.ManagerListItem, error)
	CreateManager(ctx context.Context, orgID, unitID string, input domain.CreateManagerInput, passwordHash string, audit *domain.AuditLog) (*domain.ManagerListItem, error)
}

// OwnerManagersService enforces business rules for unit manager lifecycle.
type OwnerManagersService struct {
	repo OwnerManagersRepository
}

// NewOwnerManagersService instantiates a new OwnerManagersService.
func NewOwnerManagersService(repo OwnerManagersRepository) *OwnerManagersService {
	return &OwnerManagersService{repo: repo}
}

func resolveOwnerOrgForManagers(claims *UserClaims) (string, error) {
	orgID, err := resolveOrgIDFromClaims(claims)
	if err != nil {
		return "", err
	}
	if claims.Role != domain.RoleOrgOwner && claims.Role != domain.RoleSuperAdmin {
		return "", domain.ErrForbidden
	}
	return orgID, nil
}

// ListManagers returns all UNIT_MANAGER users under the caller's organization.
func (s *OwnerManagersService) ListManagers(ctx context.Context, claims *UserClaims) ([]domain.ManagerListItem, error) {
	orgID, err := resolveOwnerOrgForManagers(claims)
	if err != nil {
		return nil, err
	}
	return s.repo.ListManagers(ctx, orgID)
}

// CreateManager validates, hashes the password and creates a new UNIT_MANAGER user.
func (s *OwnerManagersService) CreateManager(ctx context.Context, claims *UserClaims, input domain.CreateManagerInput) (*domain.ManagerListItem, error) {
	orgID, err := resolveOwnerOrgForManagers(claims)
	if err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Normalize email before uniqueness check.
	email := strings.ToLower(strings.TrimSpace(input.Email))
	unitID := strings.TrimSpace(input.UnitID)

	ok, err := s.repo.UnitBelongsToOrg(ctx, unitID, orgID)
	if err != nil {
		return nil, fmt.Errorf("check unit org: %w", err)
	}
	if !ok {
		return nil, domain.ErrUnitNotInOrg
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, domain.ErrManagerEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}
	details, _ := json.Marshal(map[string]any{
		"manager_name":  strings.TrimSpace(input.Name),
		"manager_email": email,
		"unit_id":       unitID,
	})
	audit := &domain.AuditLog{
		OrganizationID: &orgID,
		UnitID:         &unitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         "MANAGER_CREATED",
		TargetEntity:   "users",
		TargetID:       unitID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	return s.repo.CreateManager(ctx, orgID, unitID, input, string(hash), audit)
}
