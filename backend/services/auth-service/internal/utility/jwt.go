package utility

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token type constants
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// JWTClaims represents JWT claims structure
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTUtil handles JWT operations
type JWTUtil struct {
	secretKey     string
	expiry        time.Duration
	refreshExpiry time.Duration
}

// NewJWTUtil creates new JWT utility instance
func NewJWTUtil(secretKey string, expiryStr string, refreshExpiryStr string) (*JWTUtil, error) {
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return nil, err
	}

	refreshExpiry, err := time.ParseDuration(refreshExpiryStr)
	if err != nil {
		return nil, err
	}

	return &JWTUtil{
		secretKey:     secretKey,
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateToken generates new JWT access token
func (j *JWTUtil) GenerateToken(userID, email, name, role string) (string, error) {
	return j.generateTokenWithType(userID, email, name, role, TokenTypeAccess, j.expiry)
}

// GenerateRefreshToken generates new JWT refresh token with longer expiry
func (j *JWTUtil) GenerateRefreshToken(userID, email, name, role string) (string, error) {
	return j.generateTokenWithType(userID, email, name, role, TokenTypeRefresh, j.refreshExpiry)
}

// generateTokenWithType generates a JWT token with specified type and expiry
func (j *JWTUtil) generateTokenWithType(userID, email, name, role, tokenType string, expiry time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates JWT token and returns claims
func (j *JWTUtil) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetExpiryDuration returns access token expiry duration
func (j *JWTUtil) GetExpiryDuration() time.Duration {
	return j.expiry
}

// GetRefreshExpiryDuration returns refresh token expiry duration
func (j *JWTUtil) GetRefreshExpiryDuration() time.Duration {
	return j.refreshExpiry
}
