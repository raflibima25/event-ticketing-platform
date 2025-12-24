package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/payload/response"
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
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Message: message.ErrInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	// Call service
	authResponse, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServer error
		errorMessage := message.ErrInternalServer

		// Handle specific errors
		if errors.Is(err, service.ErrEmailExists) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrEmailAlreadyExists
		} else if errors.Is(err, service.ErrHashPassword) {
			statusCode = http.StatusInternalServer
			errorMessage = message.ErrHashPassword
		}

		ctx.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Message: errorMessage,
			Error:   err.Error(),
		})
		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: message.MsgRegisterSuccess,
		Data:    authResponse,
	})
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
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Message: message.ErrInvalidRequest,
			Error:   err.Error(),
		})
		return
	}

	// Call service
	authResponse, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServer
		errorMessage := message.ErrInternalServer

		// Handle specific errors
		if errors.Is(err, service.ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
			errorMessage = message.ErrInvalidCredentials
		}

		ctx.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Message: errorMessage,
			Error:   err.Error(),
		})
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: message.MsgLoginSuccess,
		Data:    authResponse,
	})
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
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Message: message.ErrUnauthorized,
		})
		return
	}

	// Call service
	userResponse, err := c.authService.GetUserByID(ctx.Request.Context(), userID.(string))
	if err != nil {
		statusCode := http.StatusInternalServer
		errorMessage := message.ErrInternalServer

		if errors.Is(err, repository.ErrUserNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrUserNotFound
		}

		ctx.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Message: errorMessage,
			Error:   err.Error(),
		})
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    userResponse,
	})
}

// Health check endpoint
func (c *AuthController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}
