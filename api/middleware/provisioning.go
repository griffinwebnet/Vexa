package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/services"
)

// ProvisioningGate blocks non-admin users from accessing protected APIs
// when the system is unprovisioned. Admins are allowed (UI will route them
// to the wizard).
func ProvisioningGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow auth and domain provisioning endpoints to pass through
		path := c.FullPath()
		if path == "/api/v1/auth/login" ||
			path == "/api/v1/auth/bootstrap-status" ||
			path == "/api/v1/auth/bootstrap-admin" ||
			path == "/api/v1/domain/provision" ||
			path == "/api/v1/domain/provision-with-output" ||
			path == "/api/v1/domain/status" ||
			path == "/api/v1/version" ||
			path == "/health" {
			c.Next()
			return
		}

		domainService := services.NewDomainService()
		status, err := domainService.GetDomainStatus()
		if err == nil && status != nil && !status.Provisioned {
			// Read admin flag from auth middleware
			val, exists := c.Get("is_admin")
			isAdmin, _ := val.(bool)
			if !exists || !isAdmin {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "System is not provisioned. Admin access required.",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
