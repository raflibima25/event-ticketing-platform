package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/raflibima25/event-ticketing-platform/backend/pkg/cache"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/utility"
)

// Password reset token expiry duration
const PasswordResetTokenExpiry = 1 * time.Hour

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailExists         = errors.New("email already registered")
	ErrHashPassword        = errors.New("failed to hash password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidTokenType    = errors.New("invalid token type")
	ErrPasswordMismatch    = errors.New("current password is incorrect")
	ErrInvalidResetToken   = errors.New("invalid or expired reset token")
)

// AuthService defines interface for authentication business logic
type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error)
	GetUserByID(ctx context.Context, userID string) (*response.UserResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*response.TokenRefreshResponse, error)
	ChangePassword(ctx context.Context, userID string, req *request.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *request.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error
}

// authService implements AuthService interface
type authService struct {
	userRepo          repository.UserRepository
	passwordResetRepo repository.PasswordResetRepository
	jwtUtil           *utility.JWTUtil
	cache             cache.RedisClient // For future features: token blacklist, rate limiting
	bcryptCost        int
}

// NewAuthService creates new auth service instance
func NewAuthService(
	userRepo repository.UserRepository,
	passwordResetRepo repository.PasswordResetRepository,
	jwtUtil *utility.JWTUtil,
	redisClient cache.RedisClient,
	bcryptCost int,
) AuthService {
	return &authService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		jwtUtil:           jwtUtil,
		cache:             redisClient,
		bcryptCost:        bcryptCost,
	}
}

// Register handles user registration
func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, ErrHashPassword
	}

	// Create user entity
	user := &entity.User{
		Email:           req.Email,
		PasswordHash:    string(hashedPassword),
		FullName:        req.FullName,
		Role:            req.Role,
		IsEmailVerified: false,
	}

	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Build response
	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtUtil.GetExpiryDuration().Seconds()),
		User:         s.mapUserToResponse(user),
	}, nil
}

// Login handles user authentication
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Build response
	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtUtil.GetExpiryDuration().Seconds()),
		User:         s.mapUserToResponse(user),
	}, nil
}

// GetUserByID retrieves user information by ID
func (s *authService) GetUserByID(ctx context.Context, userID string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userResponse := s.mapUserToResponse(user)
	return &userResponse, nil
}

// mapUserToResponse converts entity.User to response.UserResponse
func (s *authService) mapUserToResponse(user *entity.User) response.UserResponse {
	return response.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FullName:        user.FullName,
		Phone:           user.Phone,
		Role:            user.Role,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.CreatedAt,
	}
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (*response.TokenRefreshResponse, error) {
	// Validate refresh token
	claims, err := s.jwtUtil.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Verify token type is refresh
	if claims.TokenType != utility.TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	// Verify user still exists
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new access token only (not a new refresh token)
	accessToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &response.TokenRefreshResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtUtil.GetExpiryDuration().Seconds()),
	}, nil
}

// ChangePassword changes the password for an authenticated user
func (s *authService) ChangePassword(ctx context.Context, userID string, req *request.ChangePasswordRequest) error {
	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return repository.ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrPasswordMismatch
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return ErrHashPassword
	}

	// Update password in database
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ForgotPassword initiates password reset flow by generating a reset token
func (s *authService) ForgotPassword(ctx context.Context, req *request.ForgotPasswordRequest) error {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Don't reveal if email exists or not for security
			// Return success even if email doesn't exist
			log.Printf("Password reset requested for non-existent email: %s", req.Email)
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete any existing tokens for this user
	if err := s.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		log.Printf("Failed to delete existing tokens: %v", err)
		// Continue anyway
	}

	// Create new reset token
	resetToken, err := s.passwordResetRepo.Create(ctx, user.ID, PasswordResetTokenExpiry)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// TODO: Send email with reset link via notification service
	// For now, log the token (in production, this should be sent via email)
	log.Printf("Password reset token created for user %s: %s (expires: %s)",
		user.Email, resetToken.Token, resetToken.ExpiresAt.Format(time.RFC3339))

	return nil
}

// ResetPassword resets the password using a valid reset token
func (s *authService) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error {
	// Validate reset token
	resetToken, err := s.passwordResetRepo.GetByToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) ||
			errors.Is(err, repository.ErrTokenExpired) ||
			errors.Is(err, repository.ErrTokenUsed) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("failed to get reset token: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return ErrHashPassword
	}

	// Update password in database
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		log.Printf("Failed to mark reset token as used: %v", err)
		// Password was already updated, so this is non-critical
	}

	return nil
}
