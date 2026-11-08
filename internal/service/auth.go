package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type UserClaims struct {
	UserID               string      `json:"user_id"`
	Email                string      `json:"email"`
	Role                 domain.Role `json:"role"`
	OrganizationID       *string     `json:"organization_id,omitempty"`
	UnitID               *string     `json:"unit_id,omitempty"`
	OrganizationIsActive bool        `json:"organization_is_active"`
	SubscriptionStatus   string      `json:"subscription_status,omitempty"`
	IsImpersonated       bool        `json:"is_impersonated,omitempty"`
	ImpersonatedBy       *string     `json:"impersonated_by,omitempty"`
	jwt.RegisteredClaims
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type"` // always Bearer
	ExpiresIn int64        `json:"expires_in"` // in seconds
	User      UserResponse `json:"user"`
}

type UserResponse struct {
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	Email                string      `json:"email"`
	Role                 domain.Role `json:"role"`
	OrganizationID       *string     `json:"organization_id,omitempty"`
	UnitID               *string     `json:"unit_id,omitempty"`
	OrganizationIsActive bool        `json:"organization_is_active"`
	SubscriptionStatus   string      `json:"subscription_status,omitempty"`
}

type AuthService struct {
	userRepo   UserRepository
	jwtSecret  []byte
	jwtExpires time.Duration
}

func NewAuthService(userRepo UserRepository, jwtSecret string, jwtExpires time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		jwtExpires: jwtExpires,
	}
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if input.Email == "" || input.Password == "" {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, fmt.Errorf("auth service get by email: %w", err)
	}

	if !user.IsActive {
		return nil, domain.ErrForbidden
	}

	orgIsActive := true
	if user.OrganizationIsActive != nil {
		orgIsActive = *user.OrganizationIsActive
	}

	if !orgIsActive {
		if user.Role != domain.RoleOrgOwner && user.Role != domain.RoleSuperAdmin {
			return nil, domain.ErrTenantInactive
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.jwtExpires)

	var subStatus string
	if user.SubscriptionStatus != nil {
		subStatus = string(*user.SubscriptionStatus)
	}

	claims := UserClaims{
		UserID:               user.ID,
		Email:                user.Email,
		Role:                 user.Role,
		OrganizationID:       user.OrganizationID,
		UnitID:               user.UnitID,
		OrganizationIsActive: orgIsActive,
		SubscriptionStatus:   subStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString((s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt token: %w", err)
	}

	return &LoginOutput{
		Token:     signedToken,
		TokenType: "Bearer",
		ExpiresIn: int64(s.jwtExpires.Seconds()),
		User: UserResponse{
			ID:                   user.ID,
			Name:                 user.Name,
			Email:                user.Email,
			Role:                 user.Role,
			OrganizationID:       user.OrganizationID,
			UnitID:               user.UnitID,
			OrganizationIsActive: orgIsActive,
			SubscriptionStatus:   subStatus,
		},
	}, nil

}

func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}
