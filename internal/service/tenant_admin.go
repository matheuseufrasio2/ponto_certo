package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuditLogRecorder records compliance and security audit logs.
type AuditLogRecorder interface {
	Record(ctx context.Context, log domain.AuditLog) error
}

// TenantAdminService orchestrates SaaS multi-tenancy operations and diagnostic impersonation.
type TenantAdminService struct {
	repo          repository.TenantAdminRepository
	jwtSecret     string
	auditRecorder AuditLogRecorder
}

// NewTenantAdminService instantiates a new TenantAdminService.
func NewTenantAdminService(repo repository.TenantAdminRepository, jwtSecret string) *TenantAdminService {
	return &TenantAdminService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// SetAuditRecorder assigns the audit log recorder.
func (s *TenantAdminService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

func requireSuperAdmin(claims *UserClaims) error {
	if claims == nil {
		return domain.ErrUnauthorized
	}
	if claims.Role != domain.RoleSuperAdmin {
		return domain.ErrForbidden
	}
	return nil
}

// GetSaaSMetrics returns platform recurring revenue and health KPIs.
func (s *TenantAdminService) GetSaaSMetrics(ctx context.Context, claims *UserClaims) (*domain.SaaSMetrics, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	return s.repo.GetSaaSPlatformKPIs(ctx)
}

// ListTenants returns paginated organizations matching search, status, and plan filters.
func (s *TenantAdminService) ListTenants(ctx context.Context, claims *UserClaims, search, status, plan string, page, perPage int) (*domain.PaginatedOrganizations, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	items, total, err := s.repo.ListTenants(ctx, search, status, plan, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedOrganizations{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// CreateTenant validates input, hashes owner password, and creates organization in a single transaction.
func (s *TenantAdminService) CreateTenant(ctx context.Context, claims *UserClaims, input domain.CreateTenantInput) (*domain.Organization, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}

	if input.Name == "" {
		return nil, errors.New("organization name is required")
	}
	if input.AdminEmail == "" || input.AdminPassword == "" {
		return nil, errors.New("admin email and password are required")
	}
	if len(input.AdminPassword) < 6 {
		return nil, errors.New("admin password must be at least 6 characters long")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	return s.repo.CreateTenantWithAdmin(ctx, input, string(hash))
}

// ToggleTenantStatus activates or suspends an organization and cascades to user sessions.
func (s *TenantAdminService) ToggleTenantStatus(ctx context.Context, claims *UserClaims, tenantID string, isActive bool) error {
	if err := requireSuperAdmin(claims); err != nil {
		return err
	}
	if err := s.repo.ToggleTenantActiveStatus(ctx, tenantID, isActive); err != nil {
		return err
	}

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		details, _ := json.Marshal(map[string]any{"is_active": isActive})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			OrganizationID: &tenantID,
			ActorID:        &claims.UserID,
			ActorRole:      claims.Role,
			Action:         "TOGGLE_TENANT_STATUS",
			TargetEntity:   "organizations",
			TargetID:       tenantID,
			Details:        details,
			IPAddress:      meta.IPAddress,
			UserAgent:      meta.UserAgent,
		})
	}

	return nil
}

// GetTenantUsers returns users belonging to a tenant organization for impersonation filtered by status.
func (s *TenantAdminService) GetTenantUsers(ctx context.Context, claims *UserClaims, tenantID string, status string) ([]domain.User, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	return s.repo.GetTenantUsers(ctx, tenantID, status)
}

// GetTenantUnits returns physical units/branches belonging to a tenant organization.
func (s *TenantAdminService) GetTenantUnits(ctx context.Context, claims *UserClaims, tenantID string) ([]domain.BarbershopUnit, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}
	return s.repo.GetTenantUnits(ctx, tenantID)
}

// GetTenantOverview aggregates tenant engagement, 30-day GMV, appointments, and days overdue.
func (s *TenantAdminService) GetTenantOverview(ctx context.Context, claims *UserClaims, tenantID string) (*domain.TenantOverview, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}
	return s.repo.GetTenantOverview(ctx, tenantID)
}

// UpdateTenantPlan modifies an organization's subscription tier and monthly billing rate.
func (s *TenantAdminService) UpdateTenantPlan(ctx context.Context, claims *UserClaims, tenantID string, input domain.UpdateTenantPlanInput) (*domain.Organization, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	org, err := s.repo.UpdateTenantPlan(ctx, tenantID, input.Plan, input.MonthlyFeeInCents)
	if err != nil {
		return nil, err
	}

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		details, _ := json.Marshal(map[string]any{
			"plan":                 input.Plan,
			"monthly_fee_in_cents": input.MonthlyFeeInCents,
		})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			OrganizationID: &tenantID,
			ActorID:        &claims.UserID,
			ActorRole:      claims.Role,
			Action:         "UPDATE_TENANT_PLAN",
			TargetEntity:   "organizations",
			TargetID:       tenantID,
			Details:        details,
			IPAddress:      meta.IPAddress,
			UserAgent:      meta.UserAgent,
		})
	}

	return org, nil
}

// ResetTenantPassword assists franchise owners who lost system access by setting a new or temporary password.
func (s *TenantAdminService) ResetTenantPassword(ctx context.Context, claims *UserClaims, tenantID string, input domain.ResetTenantPasswordInput) (*domain.ResetPasswordResponse, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}

	owner, err := s.repo.GetPrimaryOwnerByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	passwordToSet := input.NewPassword
	if passwordToSet == "" {
		passwordToSet = fmt.Sprintf("Ponto@%d", time.Now().UnixNano()%1000000)
	}
	if len(passwordToSet) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(passwordToSet), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(ctx, owner.ID, string(hash)); err != nil {
		return nil, err
	}

	slog.Warn("superadmin reset tenant owner password",
		"admin_id", claims.UserID,
		"tenant_id", tenantID,
		"owner_id", owner.ID,
		"owner_email", owner.Email,
	)

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		details, _ := json.Marshal(map[string]any{"owner_id": owner.ID, "owner_email": owner.Email})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			OrganizationID: &tenantID,
			ActorID:        &claims.UserID,
			ActorRole:      claims.Role,
			Action:         "RESET_TENANT_PASSWORD",
			TargetEntity:   "users",
			TargetID:       owner.ID,
			Details:        details,
			IPAddress:      meta.IPAddress,
			UserAgent:      meta.UserAgent,
		})
	}

	return &domain.ResetPasswordResponse{
		UserID:            owner.ID,
		UserEmail:         owner.Email,
		TemporaryPassword: passwordToSet,
	}, nil
}

// ImpersonateUser generates a temporary JWT token allowing a Super Admin to access the system as a specific user.
func (s *TenantAdminService) ImpersonateUser(ctx context.Context, claims *UserClaims, targetUserID string) (*domain.ImpersonateResponse, error) {
	if err := requireSuperAdmin(claims); err != nil {
		return nil, err
	}

	if claims.UserID == targetUserID {
		return nil, domain.ErrCannotImpersonateSelf
	}

	targetUser, err := s.repo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	if !targetUser.IsActive {
		return nil, errors.New("cannot impersonate an inactive user")
	}

	if targetUser.Role == domain.RoleSuperAdmin {
		return nil, domain.ErrCannotImpersonateSuperAdmin
	}

	ttl := 1800 // 30 minutes
	now := time.Now().UTC()
	tokenClaims := &UserClaims{
		UserID:         targetUser.ID,
		Email:          targetUser.Email,
		Role:           targetUser.Role,
		OrganizationID: targetUser.OrganizationID,
		UnitID:         targetUser.UnitID,
		IsImpersonated: true,
		ImpersonatedBy: &claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   targetUser.ID,
			Issuer:    "ponto_certo_api_impersonation",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign impersonation token: %w", err)
	}

	slog.Warn("superadmin impersonation token issued",
		"admin_id", claims.UserID,
		"admin_email", claims.Email,
		"target_user_id", targetUser.ID,
		"target_user_name", targetUser.Name,
		"target_role", targetUser.Role,
		"target_org_id", targetUser.OrganizationID,
	)

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		details, _ := json.Marshal(map[string]any{
			"target_user_id":    targetUser.ID,
			"target_user_email": targetUser.Email,
			"target_role":       targetUser.Role,
		})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			OrganizationID: targetUser.OrganizationID,
			UnitID:         targetUser.UnitID,
			ActorID:        &claims.UserID,
			ActorRole:      claims.Role,
			Action:         "IMPERSONATE_USER",
			TargetEntity:   "users",
			TargetID:       targetUser.ID,
			Details:        details,
			IPAddress:      meta.IPAddress,
			UserAgent:      meta.UserAgent,
		})
	}

	return &domain.ImpersonateResponse{
		Token:          tokenString,
		TargetUserID:   targetUser.ID,
		TargetUserName: targetUser.Name,
		TargetRole:     string(targetUser.Role),
		ExpiresIn:      ttl,
	}, nil
}
