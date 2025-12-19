package utils

import (
    "sync"
    "time"
)



// Gunakan huruf kecil agar private, tapi di level package
var (
    tokenBlacklist = make(map[string]time.Time)
	// Gunakan RWMutex agar lebih cepat saat banyak yang nge-check (Read)
    mux            sync.RWMutex 
)

func BlacklistToken(token string, expiration time.Time) {
    mux.Lock()
    defer mux.Unlock()
    tokenBlacklist[token] = expiration
}

func IsTokenBlacklisted(token string) bool {
    mux.RLock()
    defer mux.RUnlock()
    
    exp, exists := tokenBlacklist[token]
    if !exists {
        return false
    }

    // Jika sudah lewat waktu kadaluarsa aslinya, anggap tidak ada (atau hapus)
    if time.Now().After(exp) {
        return false
    }
    return true
}