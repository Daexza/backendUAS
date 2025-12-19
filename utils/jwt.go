package utils

import (
	"achievements-uas/app/models"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	ID          string   `json:"id"`
	Username        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

var ErrTokenInvalid = errors.New("token invalid")

func GenerateAccessToken(user *models.User, roleName string, permissions []string) (string, error) {
    // Ambil durasi dari env (misal: 15m)
    expireStr := os.Getenv("JWT_EXPIRE")
    duration, err := time.ParseDuration(expireStr)
    if err != nil {
        duration = time.Hour * 24 // default 24 jam
    }

    claims := JWTClaims{
        ID:          user.ID,
        Role:        roleName,
		Username:        user.Username,
        Permissions: permissions,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
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
		Username: user.Username,
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

