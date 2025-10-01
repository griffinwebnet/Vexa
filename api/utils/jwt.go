package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

var jwtSecret string

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() string {
	if jwtSecret == "" {
		// Try to get from environment
		jwtSecret = os.Getenv("JWT_SECRET")
		
		// If not set, generate a random one (development only)
		if jwtSecret == "" {
			bytes := make([]byte, 32)
			if _, err := rand.Read(bytes); err != nil {
				panic("Failed to generate JWT secret")
			}
			jwtSecret = hex.EncodeToString(bytes)
		}
	}
	
	return jwtSecret
}

