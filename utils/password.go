package utils

import (

	"golang.org/x/crypto/bcrypt"
)

// ===========================
// PASSWORD HASHING FUNCTIONS
// ===========================

// HashPassword → generate bcrypt hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword → compare hash vs plain password
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}