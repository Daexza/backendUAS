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

// Error untuk token invalid
var ErrTokenInvalid = errors.New("token invalid")

// ============================
// ACCESS TOKEN
// ============================
func GenerateAccessToken(user *models.User) (string, error) {
	claims := JWTClaims{
		ID:   user.ID,
		Role: user.RoleID,
		Permissions: []string{
			"user:manage", // contoh permission
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // 1 jam
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

// ============================
// REFRESH TOKEN
// ============================
func GenerateRefreshToken(user *models.User) (string, error) {
	claims := JWTClaims{
		ID:   user.ID,
		Role: user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 hari
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

// ============================
// BLACKLIST TOKEN (opsional)
// ============================
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
