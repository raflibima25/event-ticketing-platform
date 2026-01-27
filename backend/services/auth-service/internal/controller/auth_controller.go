package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	sharedresponse "github.com/raflibima25/event-ticketing-platform/backend/pkg/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/service"
)

// AuthController handles HTTP requests for authentication
type AuthController struct {
	authService service.AuthService
}

// NewAuthController creates new auth controller instance
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles user registration request
// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "Registration data"
// @Success 201 {object} response.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req request.RegisterRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	authResponse, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		// Handle specific errors
		if errors.Is(err, service.ErrEmailExists) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrEmailAlreadyExists
		} else if errors.Is(err, service.ErrHashPassword) {
			statusCode = http.StatusInternalServerError
			errorMessage = message.ErrHashPassword
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, sharedresponse.Success(message.MsgRegisterSuccess, authResponse))
}

// Login handles user login request
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login credentials"
// @Success 200 {object} response.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	authResponse, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		// Handle specific errors
		if errors.Is(err, service.ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
			errorMessage = message.ErrInvalidCredentials
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgLoginSuccess, authResponse))
}

// GetProfile retrieves current user profile
// @Summary Get user profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/auth/profile [get]
func (c *AuthController) GetProfile(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, sharedresponse.Error(message.ErrUnauthorized, nil))
		return
	}

	// Call service
	userResponse, err := c.authService.GetUserByID(ctx.Request.Context(), userID.(string))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, repository.ErrUserNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrUserNotFound
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, sharedresponse.Success("Profile retrieved successfully", userResponse))
}

// RefreshToken handles token refresh request
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.TokenRefreshResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req request.RefreshTokenRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	tokenResponse, err := c.authService.RefreshAccessToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized
		errorMessage := message.ErrInvalidToken

		if errors.Is(err, service.ErrInvalidRefreshToken) || errors.Is(err, service.ErrInvalidTokenType) {
			statusCode = http.StatusUnauthorized
			errorMessage = message.ErrInvalidToken
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgTokenRefreshed, tokenResponse))
}

// ChangePassword handles password change request for authenticated users
// @Summary Change password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/change-password [post]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req request.ChangePasswordRequest

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, sharedresponse.Error(message.ErrUnauthorized, nil))
		return
	}

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	err := c.authService.ChangePassword(ctx.Request.Context(), userID.(string), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrPasswordMismatch) {
			statusCode = http.StatusBadRequest
			errorMessage = "Current password is incorrect"
		} else if errors.Is(err, repository.ErrUserNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrUserNotFound
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, sharedresponse.Success("Password changed successfully", nil))
}

// ForgotPassword handles forgot password request
// @Summary Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ForgotPasswordRequest true "Email address"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req request.ForgotPasswordRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	err := c.authService.ForgotPassword(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, sharedresponse.Error(message.ErrInternalServer, err.Error()))
		return
	}

	// Always return success (don't reveal if email exists)
	ctx.JSON(http.StatusOK, sharedresponse.Success("If the email exists, a password reset link has been sent", nil))
}

// ResetPassword handles password reset with token
// @Summary Reset password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req request.ResetPasswordRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Call service
	err := c.authService.ResetPassword(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrInvalidResetToken) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrInvalidToken
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, sharedresponse.Success("Password reset successfully", nil))
}

// Health check endpoint
func (c *AuthController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}
