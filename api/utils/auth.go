package utils

// AuthenticatePAM is a placeholder that only works on Linux
// On Windows, this will return false
// The actual PAM implementation should be in auth_linux.go with build tags
func AuthenticatePAM(username, password string) bool {
	// This is a stub - real implementation would be in auth_linux.go
	// with build constraints for Linux only
	return false
}

