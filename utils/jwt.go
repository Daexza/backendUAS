package utils

import (
	"achievements-uas/models"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	ID          string   `json:"id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

var ErrTokenInvalid = errors.New("token invalid")

// =====================================================
// ACCESS TOKEN – menerima PERMISSIONS dari AuthService
// =====================================================
func GenerateAccessToken(user *models.User, permissions []string) (string, error) {

	claims := JWTClaims{
		ID:          user.ID,
		Role:        user.RoleID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ParseAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func ValidateAndGetUserID(tokenStr string) (string, error) {
	claims, err := ParseAccessToken(tokenStr)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// =====================================================
// REFRESH TOKEN – TIDAK membawa permissions
// =====================================================
func GenerateRefreshToken(user *models.User) (string, error) {
	claims := JWTClaims{
		ID:   user.ID,
		Role: user.RoleID,
		Permissions: nil, // refresh token tidak butuh permission
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ParseRefreshToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// =====================================================
// BLACKLIST TOKEN
// =====================================================

var tokenBlacklist = make(map[string]time.Time)

func BlacklistToken(token string, exp time.Time) {
	tokenBlacklist[token] = exp
}

func IsTokenBlacklisted(token string) bool {
	exp, ok := tokenBlacklist[token]
	if !ok {
		return false
	}
	return time.Now().Before(exp)
}
