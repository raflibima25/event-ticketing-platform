package utility

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTUtil handles JWT operations
type JWTUtil struct {
	secretKey string
	expiry    time.Duration
}

// NewJWTUtil creates new JWT utility instance
func NewJWTUtil(secretKey string, expiryStr string) (*JWTUtil, error) {
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return nil, err
	}

	return &JWTUtil{
		secretKey: secretKey,
		expiry:    expiry,
	}, nil
}

// GenerateToken generates new JWT token
func (j *JWTUtil) GenerateToken(userID, email, name, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiry)),
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

// GetExpiryDuration returns token expiry duration
func (j *JWTUtil) GetExpiryDuration() time.Duration {
	return j.expiry
}
