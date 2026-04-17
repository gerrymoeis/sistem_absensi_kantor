package main

import (
	"log"

	"absensi-app/internal/config"
	"absensi-app/internal/database"
	"absensi-app/internal/handler"
	"absensi-app/internal/middleware"
	"absensi-app/internal/repository"
	"absensi-app/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// securityHeaders adds security headers to all responses
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Only add HSTS if using HTTPS
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using defaults")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	absensiRepo := repository.NewAbsensiRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.Security.JWTSecret)
	absensiService := service.NewAbsensiService(absensiRepo, userRepo)
	logService := service.NewActivityLogService(activityLogRepo)
	adminService := service.NewAdminService(adminRepo, userRepo)
	userService := service.NewUserService(userRepo)
	exportService := service.NewExportService(adminRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logService)
	absensiHandler := handler.NewAbsensiHandler(absensiService, logService)
	adminHandler := handler.NewAdminHandler(adminService, userService, logService)
	exportHandler := handler.NewExportHandler(exportService, logService)

	// Setup Gin router
	// Set release mode by default for production
	gin.SetMode(gin.ReleaseMode)
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	
	// Configure trusted proxies (security)
	// Set to nil if not behind a proxy, or specify trusted proxy IPs
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	}
	
	// Add security headers
	router.Use(securityHeaders())
	
	// Apply IP restriction middleware globally
	router.Use(middleware.IPRestriction(cfg.Security.AllowedIPs))

	// Apply rate limiting only to API routes (not static files)
	// Removed global rate limiter for better performance

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("./web/templates/*")

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})
	router.GET("/login", authHandler.LoginPage)
	
	// Login endpoint with stricter rate limiting (5 req/min per IP)
	router.POST("/api/auth/login", middleware.LoginRateLimiter(), authHandler.Login)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	authorized.Use(middleware.APIRateLimiter()) // Rate limit authenticated routes
	{
		// Web pages
		authorized.GET("/dashboard", absensiHandler.DashboardPage)
		authorized.GET("/history", absensiHandler.HistoryPage)

		// API endpoints
		authorized.POST("/api/auth/logout", authHandler.Logout)
		authorized.GET("/api/auth/me", authHandler.Me)

		authorized.POST("/api/absensi/masuk", absensiHandler.ClockIn)
		authorized.POST("/api/absensi/pulang", absensiHandler.ClockOut)
		authorized.GET("/api/absensi/today", absensiHandler.GetToday)
		authorized.GET("/api/absensi/history", absensiHandler.GetHistory)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	admin.Use(middleware.AdminRequired())
	{
		// Admin pages
		admin.GET("/dashboard", adminHandler.DashboardPage)
	}

	// Admin API routes
	adminAPI := router.Group("/api/admin")
	adminAPI.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	adminAPI.Use(middleware.AdminRequired())
	adminAPI.Use(middleware.APIRateLimiter()) // Rate limit admin API
	{
		adminAPI.GET("/stats", adminHandler.GetStatistics)
		adminAPI.GET("/absensi", adminHandler.GetAllAbsensi)
		adminAPI.GET("/absensi/today", adminHandler.GetTodayAbsensi)
		adminAPI.GET("/absensi/user/:id", adminHandler.GetUserAbsensi)
		
		// User management
		adminAPI.GET("/users", adminHandler.GetAllUsers)
		adminAPI.POST("/users", adminHandler.CreateUser)
		adminAPI.GET("/users/:id", adminHandler.GetUser)
		adminAPI.PUT("/users/:id", adminHandler.UpdateUser)
		adminAPI.DELETE("/users/:id", adminHandler.DeleteUser)
		adminAPI.POST("/users/:id/reset-password", adminHandler.ResetPassword)
		
		// Activity logs
		adminAPI.GET("/logs", adminHandler.GetActivityLogs)
		adminAPI.GET("/logs/user/:id", adminHandler.GetUserActivityLogs)
		
		// Export
		adminAPI.GET("/export/excel", exportHandler.ExportExcel)
		adminAPI.GET("/export/excel/monthly", exportHandler.ExportExcelByMonth)
	}

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
