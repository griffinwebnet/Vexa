package utils

import (
	"bufio"
	"os"
	"strings"
)

// IsUserAdmin checks if a user is in the sudo/wheel group
func IsUserAdmin(username string) bool {
	// Check /etc/group for sudo or wheel group membership
	groups := []string{"sudo", "wheel", "root"}
	
	file, err := os.Open("/etc/group")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		
		if len(parts) < 4 {
			continue
		}
		
		groupName := parts[0]
		members := strings.Split(parts[3], ",")
		
		// Check if this is an admin group
		isAdminGroup := false
		for _, g := range groups {
			if groupName == g {
				isAdminGroup = true
				break
			}
		}
		
		if !isAdminGroup {
			continue
		}
		
		// Check if user is a member
		for _, member := range members {
			if strings.TrimSpace(member) == username {
				return true
			}
		}
	}
	
	return username == "root"
}

