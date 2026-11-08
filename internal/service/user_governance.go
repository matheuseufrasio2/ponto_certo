package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// UserGovernanceRepository defines data access methods for managing user lifecycle and status.
type UserGovernanceRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	ToggleActiveStatus(ctx context.Context, userID string, isActive bool, isBarber bool, ownerOrgID *string, auditLog *domain.AuditLog) error
}

// UserGovernanceService enforces hierarchical RBAC, multi-tenant boundaries, and status cascading.
type UserGovernanceService struct {
	repo UserGovernanceRepository
}

// NewUserGovernanceService instantiates a new user governance service.
func NewUserGovernanceService(repo UserGovernanceRepository) *UserGovernanceService {
	return &UserGovernanceService{repo: repo}
}

// ToggleUserStatus applies strict hierarchy and tenant/unit scoping before toggling active status.
func (s *UserGovernanceService) ToggleUserStatus(ctx context.Context, claims *UserClaims, targetUserID string, isActive bool) (*domain.ToggleUserStatusResponse, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized
	}
	if targetUserID == "" {
		return nil, errors.New("target user id is required")
	}

	// 1. Anti-suicídio: Ninguém pode alterar o próprio status de ativação
	if claims.UserID == targetUserID {
		return nil, domain.ErrCannotDeactivateSelf
	}

	// 2. Busca usuário alvo
	targetUser, err := s.repo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// 3. Super Admin é imutável: nunca pode ser desativado
	if targetUser.Role == domain.RoleSuperAdmin {
		return nil, domain.ErrCannotDeactivateSuperAdmin
	}

	// 4. Validação da Matriz Hierárquica e Escopo
	var ownerOrgID *string
	switch targetUser.Role {
	case domain.RoleOrgOwner:
		// Apenas SUPER_ADMIN pode alterar status de ORG_OWNER
		if claims.Role != domain.RoleSuperAdmin {
			return nil, domain.ErrInsufficientRoleHierarchy
		}
		ownerOrgID = targetUser.OrganizationID

	case domain.RoleUnitManager:
		// SUPER_ADMIN ou ORG_OWNER da mesma organização
		if claims.Role == domain.RoleSuperAdmin {
			break
		}
		if claims.Role == domain.RoleOrgOwner && claims.OrganizationID != nil && targetUser.OrganizationID != nil && *claims.OrganizationID == *targetUser.OrganizationID {
			break
		}
		return nil, domain.ErrInsufficientRoleHierarchy

	case domain.RoleReceptionist, domain.RoleBarber:
		// SUPER_ADMIN, ORG_OWNER da mesma organização, ou UNIT_MANAGER da mesma unidade
		if claims.Role == domain.RoleSuperAdmin {
			break
		}
		if claims.Role == domain.RoleOrgOwner && claims.OrganizationID != nil && targetUser.OrganizationID != nil && *claims.OrganizationID == *targetUser.OrganizationID {
			break
		}
		if claims.Role == domain.RoleUnitManager && claims.UnitID != nil && targetUser.UnitID != nil && *claims.UnitID == *targetUser.UnitID {
			break
		}
		return nil, domain.ErrInsufficientRoleHierarchy

	default:
		return nil, domain.ErrInsufficientRoleHierarchy
	}

	// 5. Executa atualização atômica no repositório com trilha de auditoria
	meta := domain.GetAuditMetadata(ctx)
	var actorID *string
	if claims.UserID != "" {
		actorID = &claims.UserID
	}
	action := "USER_DEACTIVATED"
	if isActive {
		action = "USER_ACTIVATED"
	}
	details, _ := json.Marshal(map[string]any{
		"target_user_name": targetUser.Name,
		"target_user_role": targetUser.Role,
		"target_email":     targetUser.Email,
		"new_is_active":    isActive,
	})
	audit := &domain.AuditLog{
		OrganizationID: targetUser.OrganizationID,
		UnitID:         targetUser.UnitID,
		ActorID:        actorID,
		ActorRole:      claims.Role,
		Action:         action,
		TargetEntity:   "users",
		TargetID:       targetUserID,
		Details:        details,
		IPAddress:      meta.IPAddress,
		UserAgent:      meta.UserAgent,
	}

	isBarber := targetUser.Role == domain.RoleBarber
	if err := s.repo.ToggleActiveStatus(ctx, targetUserID, isActive, isBarber, ownerOrgID, audit); err != nil {
		return nil, err
	}

	resp := &domain.ToggleUserStatusResponse{
		UserID:   targetUser.ID,
		IsActive: isActive,
		Role:     targetUser.Role,
	}
	if ownerOrgID != nil {
		resp.OrganizationID = ownerOrgID
		resp.OrganizationIsActive = &isActive
	}

	return resp, nil
}
