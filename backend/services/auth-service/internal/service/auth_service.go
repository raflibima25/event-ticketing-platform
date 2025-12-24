package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/utility"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
	ErrHashPassword       = errors.New("failed to hash password")
)

// AuthService defines interface for authentication business logic
type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error)
	GetUserByID(ctx context.Context, userID string) (*response.UserResponse, error)
}

// authService implements AuthService interface
type authService struct {
	userRepo   repository.UserRepository
	jwtUtil    *utility.JWTUtil
	bcryptCost int
}

// NewAuthService creates new auth service instance
func NewAuthService(userRepo repository.UserRepository, jwtUtil *utility.JWTUtil, bcryptCost int) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtUtil:    jwtUtil,
		bcryptCost: bcryptCost,
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
	accessToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.Role)
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
	accessToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateToken(user.ID, user.Email, user.Role)
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
