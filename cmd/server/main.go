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
	faceRepo := repository.NewFaceRepository(db)
	leaveRequestRepo := repository.NewLeaveRequestRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.Security.JWTSecret)
	photoService := service.NewPhotoService("./data/photos")
	
	// Ensure photo directory exists
	if err := photoService.EnsureBasePathExists(); err != nil {
		log.Fatalf("Failed to create photo directory: %v", err)
	}
	
	logService := service.NewActivityLogService(activityLogRepo)
	adminService := service.NewAdminService(adminRepo, userRepo)
	userService := service.NewUserService(userRepo)
	exportService := service.NewExportService(adminRepo)

	// Initialize face service (if enabled)
	var faceService *service.FaceService
	var faceHandler *handler.FaceHandler
	if cfg.FaceRecognition.Enabled {
		var err error
		faceService, err = service.NewFaceService(cfg.FaceRecognition.ModelsPath, faceRepo, userRepo, cfg.Environment)
		if err != nil {
			log.Printf("Warning: Failed to initialize face service: %v", err)
			log.Println("Face recognition will be disabled")
		} else {
			defer faceService.Close()
			faceHandler = handler.NewFaceHandler(faceService, logService)
			if cfg.Environment == "development" {
				log.Println("Face recognition enabled (development mode - replay attack prevention disabled)")
			} else {
				log.Println("Face recognition enabled (production mode - replay attack prevention enabled)")
			}
		}
	}
	
	// Initialize absensi service (with face service if available)
	absensiService := service.NewAbsensiService(absensiRepo, userRepo, photoService, faceService)
	leaveRequestService := service.NewLeaveRequestService(leaveRequestRepo, absensiRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logService)
	absensiHandler := handler.NewAbsensiHandler(absensiService, logService)
	adminHandler := handler.NewAdminHandler(adminService, userService, logService)
	exportHandler := handler.NewExportHandler(exportService, logService)
	leaveRequestHandler := handler.NewLeaveRequestHandler(leaveRequestService, logService)

	// Setup Gin router
	// Set release mode by default for production
	gin.SetMode(gin.ReleaseMode)
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	
	// Smart logging: Always log requests, but control verbosity
	if cfg.Server.Mode == "debug" {
		router.Use(gin.Logger()) // Full debug logging with route list
	} else {
		router.Use(middleware.ProductionLogger()) // Clean production logging
	}
	
	// Configure trusted proxies (security)
	// Set to nil if not behind a proxy, or specify trusted proxy IPs
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	}
	
	// Add security headers
	router.Use(securityHeaders())
	
	// Apply IP restriction middleware globally
	router.Use(middleware.IPRestriction(cfg.Security.AllowedIPs))

	// Log rate limiting configuration
	if cfg.Environment == "development" {
		log.Println("Rate limiting: Development mode (Login: 50/min, API: 1000/min)")
	} else {
		log.Println("Rate limiting: Production mode (Login: 5/min, API: 300/min)")
	}

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
	
	// Web pages (no auth middleware - check token in JavaScript)
	router.GET("/dashboard", absensiHandler.DashboardPage)
	router.GET("/history", absensiHandler.HistoryPage)
	router.GET("/profile", authHandler.ProfilePage)
	
	// Admin page (protected with middleware for server-side validation)
	admin := router.Group("/admin")
	admin.Use(middleware.PageAuthRequired(cfg.Security.JWTSecret))
	admin.Use(middleware.PageAdminRequired())
	{
		admin.GET("", adminHandler.DashboardPage)           // /admin
		admin.GET("/", adminHandler.DashboardPage)          // /admin/
		admin.GET("/dashboard", adminHandler.DashboardPage) // /admin/dashboard (legacy)
	}
	
	// Login endpoint with stricter rate limiting
	// Development: 50 req/min, Production: 5 req/min per IP
	router.POST("/api/auth/login", middleware.LoginRateLimiter(cfg.Environment), authHandler.Login)
	
	// Face login endpoint (no rate limiting needed as face recognition already has rate limiting)
	router.POST("/api/auth/login-face", authHandler.LoginWithFace)
	
	// Face recognition for login (public endpoint - no auth required)
	if faceHandler != nil {
		router.POST("/api/face/recognize-login", faceHandler.RecognizeFace)
	}

	// Protected API routes (require authentication)
	authorized := router.Group("/api")
	authorized.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	authorized.Use(middleware.APIRateLimiter(cfg.Environment)) // Rate limit authenticated routes
	{
		// Auth endpoints
		authorized.POST("/auth/logout", authHandler.Logout)
		authorized.GET("/auth/me", authHandler.Me)
		authorized.POST("/auth/change-password", authHandler.ChangePassword)

		// Absensi endpoints
		authorized.POST("/absensi/masuk", absensiHandler.ClockIn)
		authorized.POST("/absensi/pulang", absensiHandler.ClockOut)
		authorized.GET("/absensi/today", absensiHandler.GetToday)
		authorized.GET("/absensi/history", absensiHandler.GetHistory)
		authorized.GET("/absensi/stats", absensiHandler.GetOwnStats)

		// Face recognition endpoints (if enabled)
		if faceHandler != nil {
			authorized.POST("/face/recognize", faceHandler.RecognizeFace)
			authorized.GET("/face/status", faceHandler.CheckEnrollmentStatus)
			authorized.POST("/face/self-enroll", faceHandler.SelfEnroll)
			authorized.POST("/face/self-enroll-comprehensive", faceHandler.SelfEnrollComprehensive) // New: 5-photo enrollment
		}

		// Leave request endpoints
		authorized.POST("/leave-requests", leaveRequestHandler.Create)
		authorized.GET("/leave-requests", leaveRequestHandler.GetUserRequests)
		authorized.GET("/leave-requests/:id", leaveRequestHandler.GetByID)
		authorized.DELETE("/leave-requests/:id", leaveRequestHandler.Delete)
	}

	// Admin routes (removed - moved to adminAPI)
	// admin := router.Group("/admin")
	// admin.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	// admin.Use(middleware.AdminRequired())
	// {
	// 	admin.GET("/dashboard", adminHandler.DashboardPage)
	// }

	// Admin API routes
	adminAPI := router.Group("/api/admin")
	adminAPI.Use(middleware.AuthRequired(cfg.Security.JWTSecret))
	adminAPI.Use(middleware.AdminRequired())
	adminAPI.Use(middleware.APIRateLimiter(cfg.Environment)) // Rate limit admin API
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

		// Face recognition management (if enabled)
		if faceHandler != nil {
			adminAPI.POST("/face/enroll", faceHandler.EnrollFace)
			adminAPI.DELETE("/face/:user_id", faceHandler.DeleteUserFaceData)
			adminAPI.GET("/face/stats", faceHandler.GetEncodingStats)
		}

		// Leave request management
		adminAPI.GET("/leave-requests", leaveRequestHandler.GetAllRequests)
		adminAPI.PUT("/leave-requests/:id/review", leaveRequestHandler.Review)
	}

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
