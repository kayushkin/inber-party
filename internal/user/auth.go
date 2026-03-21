package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles JWT token operations
type AuthService struct {
	secretKey []byte
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new auth service
func NewAuthService(secretKey string) *AuthService {
	if secretKey == "" {
		// Generate a random secret key for development
		secretKey = generateRandomKey()
	}
	return &AuthService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken creates a JWT token for a user
func (a *AuthService) GenerateToken(user *User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "inber-party",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

// ValidateToken validates and parses a JWT token
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token for an existing valid token
func (a *AuthService) RefreshToken(tokenString string, repo *Repository) (string, error) {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Get the user to ensure they still exist and are active
	user, err := repo.GetUserByID(claims.UserID)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !user.IsActive {
		return "", errors.New("user account is deactivated")
	}

	// Generate new token
	return a.GenerateToken(user)
}

// generateRandomKey generates a random key for JWT signing
func generateRandomKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		// Fallback to a static key for development
		return "inber-party-dev-secret-key-not-for-production-use"
	}
	return base64.URLEncoding.EncodeToString(key)
}
