package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/handlers"
	"github.com/vexa/api/middleware"
)

const Version = "0.0.2-prealpha"

func main() {
	// Set Gin mode
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "vexa-api",
			"version": Version,
		})
	})

	// Public routes
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", handlers.Login)
		public.GET("/system/status", handlers.SystemStatus)
		public.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"version": Version,
				"service": "vexa-api",
			})
		})
		public.GET("/updates/check", handlers.CheckForUpdates)
		public.POST("/updates/upgrade", handlers.PerformUpgrade)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthRequired())
	{
		// Domain management
		protected.POST("/domain/provision", handlers.ProvisionDomain)
		protected.GET("/domain/status", handlers.DomainStatus)
		protected.PUT("/domain/configure", handlers.ConfigureDomain)

		// User management
		protected.GET("/users", handlers.ListUsers)
		protected.POST("/users", handlers.CreateUser)
		protected.GET("/users/:id", handlers.GetUser)
		protected.PUT("/users/:id", handlers.UpdateUser)
		protected.DELETE("/users/:id", handlers.DeleteUser)
		protected.POST("/users/:id/reset-password", handlers.ResetUserPassword)
		protected.POST("/users/:id/disable", handlers.DisableUser)
		protected.POST("/users/:id/enable", handlers.EnableUser)

		// Group management
		protected.GET("/groups", handlers.ListGroups)
		protected.POST("/groups", handlers.CreateGroup)
		protected.GET("/groups/:id", handlers.GetGroup)
		protected.PUT("/groups/:id", handlers.UpdateGroup)
		protected.DELETE("/groups/:id", handlers.DeleteGroup)

		// Computer/Device management
		protected.GET("/computers", handlers.ListComputers)
		protected.GET("/computers/:id", handlers.GetComputer)
		protected.DELETE("/computers/:id", handlers.DeleteComputer)

		// DNS management
		protected.GET("/dns/zones", handlers.ListDNSZones)
		protected.GET("/dns/records", handlers.ListDNSRecords)
		protected.POST("/dns/records", handlers.CreateDNSRecord)
		protected.DELETE("/dns/records/:id", handlers.DeleteDNSRecord)

		// Domain Policies
		protected.GET("/domain/policies", handlers.GetDomainPolicies)
		protected.PUT("/domain/policies", handlers.UpdateDomainPolicies)

		// Organizational Units
		protected.GET("/domain/ous", handlers.GetOUList)
		protected.POST("/domain/ous", handlers.CreateOU)
		protected.DELETE("/domain/ous/:path", handlers.DeleteOU)

		// Logs and auditing
		protected.GET("/audit/logs", handlers.GetAuditLogs)
		protected.GET("/audit/events", handlers.GetAuditEvents)

		// Overlay Networking (Headscale)
		protected.GET("/system/overlay-status", handlers.GetOverlayStatus)
		protected.POST("/system/setup-overlay", handlers.SetupOverlay)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Vexa API Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
