package middleware

import (
	"sync"
	"time"
)

var tokenBlacklist = struct {
	sync.RWMutex
	data map[string]time.Time
}{data: make(map[string]time.Time)}

func BlacklistToken(token string, exp time.Time) {
	tokenBlacklist.Lock()
	tokenBlacklist.data[token] = exp
	tokenBlacklist.Unlock()
}

func IsTokenBlacklisted(token string) bool {
	tokenBlacklist.RLock()
	exp, ok := tokenBlacklist.data[token]
	tokenBlacklist.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		tokenBlacklist.Lock()
		delete(tokenBlacklist.data, token)
		tokenBlacklist.Unlock()
		return false
	}
	return true
}
