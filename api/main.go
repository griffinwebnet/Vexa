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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	systemHandler := handlers.NewSystemHandler()
	userHandler := handlers.NewUserHandler()
	groupHandler := handlers.NewGroupHandler()
	domainHandler := handlers.NewDomainHandler()

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
		public.POST("/auth/login", authHandler.Login)
		public.GET("/system/status", systemHandler.SystemStatus)
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
		protected.POST("/domain/provision", domainHandler.ProvisionDomain)
		protected.GET("/domain/status", domainHandler.DomainStatus)
		protected.GET("/domain/info", domainHandler.GetDomainInfo)
		protected.PUT("/domain/configure", domainHandler.ConfigureDomain)

		// User management
		protected.GET("/users", userHandler.ListUsers)
		protected.POST("/users", userHandler.CreateUser)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		protected.DELETE("/users/:id", userHandler.DeleteUser)
		protected.POST("/users/:id/reset-password", handlers.ResetUserPassword)
		protected.POST("/users/:id/disable", handlers.DisableUser)
		protected.POST("/users/:id/enable", handlers.EnableUser)

		// Group management
		protected.GET("/groups", groupHandler.ListGroups)
		protected.POST("/groups", groupHandler.CreateGroup)
		protected.GET("/groups/:id", groupHandler.GetGroup)
		protected.PUT("/groups/:id", groupHandler.UpdateGroup)
		protected.DELETE("/groups/:id", groupHandler.DeleteGroup)
		protected.POST("/groups/:id/members", groupHandler.AddGroupMembers)
		protected.DELETE("/groups/:id/members", groupHandler.RemoveGroupMembers)

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

		// System services
		protected.GET("/system/services/:name", systemHandler.GetServiceStatus)

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
