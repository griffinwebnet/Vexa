package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vexa/api/handlers"
	"github.com/vexa/api/middleware"
)

// Global dev mode flag
var DevMode bool

const Version = "0.2.91"

func main() {
	// Parse command line flags
	flag.BoolVar(&DevMode, "dev", false, "Enable development mode with test credentials")
	flag.Parse()

	// Set Gin mode
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(DevMode)
	userHandler := handlers.NewUserHandler()
	groupHandler := handlers.NewGroupHandler()
	domainHandler := handlers.NewDomainHandler()
	computerHandler := handlers.NewComputerHandler()
	overlayHandler := handlers.NewOverlayHandler()

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
		public.GET("/auth/bootstrap-status", authHandler.BootstrapStatus)
		public.POST("/auth/bootstrap-admin", authHandler.BootstrapAdmin)
		public.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"version": Version,
				"service": "vexa-api",
			})
		})
		public.GET("/updates/check", handlers.CheckForUpdates)
		public.POST("/updates/upgrade", handlers.PerformUpgrade)
		public.GET("/domain/status", domainHandler.DomainStatus)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthRequired())
	protected.Use(middleware.ProvisioningGate())
	{
		// Domain management
		protected.POST("/domain/provision-with-output", domainHandler.ProvisionDomainWithOutput)
		protected.GET("/domain/info", domainHandler.GetDomainInfo)
		protected.PUT("/domain/configure", domainHandler.ConfigureDomain)

		// User management
		protected.GET("/users", userHandler.ListUsers)
		protected.POST("/users", userHandler.CreateUser)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		protected.DELETE("/users/:id", userHandler.DeleteUser)
		protected.POST("/users/:id/reset-password", handlers.ResetUserPassword)
		protected.POST("/users/:id/disable", userHandler.DisableUser)
		protected.POST("/users/:id/enable", userHandler.EnableUser)

		// Self-service endpoints
		protected.POST("/users/change-password", userHandler.ChangePassword)
		protected.POST("/users/update-profile", userHandler.UpdateProfile)

		// Group management
		protected.GET("/groups", groupHandler.ListGroups)
		protected.POST("/groups", groupHandler.CreateGroup)
		protected.GET("/groups/:id", groupHandler.GetGroup)
		protected.PUT("/groups/:id", groupHandler.UpdateGroup)
		protected.DELETE("/groups/:id", groupHandler.DeleteGroup)
		protected.POST("/groups/:id/members", groupHandler.AddGroupMembers)
		protected.DELETE("/groups/:id/members", groupHandler.RemoveGroupMembers)

		// Computer/Device management
		protected.GET("/computers", computerHandler.ListComputers)
		protected.GET("/computers/:id", computerHandler.GetComputer)
		protected.DELETE("/computers/:id", computerHandler.DeleteComputer)

		// DNS management
		dnsHandler := handlers.NewDNSHandler()
		protected.GET("/dns/status", dnsHandler.DNSStatus)
		protected.PUT("/dns/forwarders", dnsHandler.UpdateDNSForwarders)
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

		// Computer deployment
		deploymentHandler := handlers.NewDeploymentHandler()
		protected.GET("/deployment/scripts", deploymentHandler.GetDeploymentScripts)
		protected.POST("/deployment/generate", deploymentHandler.GenerateDeploymentCommand)
		protected.GET("/deployment/scripts/:script", deploymentHandler.ServeDeploymentScript)

		// Logs and auditing
		protected.GET("/audit/logs", handlers.GetAuditLogs)
		protected.GET("/audit/events", handlers.GetAuditEvents)

		// Overlay Networking (Headscale)
		protected.GET("/system/overlay-status", overlayHandler.GetOverlayStatus)
		protected.POST("/system/setup-overlay", overlayHandler.SetupOverlay)
		protected.POST("/system/test-fqdn", overlayHandler.TestFQDN)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Vexa API Server starting on port %s", port)
	if DevMode {
		log.Println("WARNING: Development mode enabled - test credentials are active!")
		log.Println("SECURITY: Never run with --dev flag in production!")
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
