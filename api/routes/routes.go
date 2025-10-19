package routes

import (
	"bms-backend/internal/config"
	"bms-backend/internal/handlers"
	"bms-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeRoutes(router *gin.Engine, db *mongo.Database) {
	cfg := config.Load()

	// Initialize ALL handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(db)
	visitorHandler := handlers.NewVisitorHandler(db)
	maintenanceHandler := handlers.NewMaintenanceHandler(db)
	amenityHandler := handlers.NewAmenityHandler(db)
	noticeHandler := handlers.NewNoticeHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "building-management-system-society",
			"timestamp": "2025-10-02T12:15:00Z",
			"features":  []string{"multi-society", "society-segregation", "access-codes"},
		})
	})

	// Public routes
	api := router.Group("/api/v1")
	{
		// Society validation (public)
		api.POST("/society/validate", authHandler.ValidateSociety)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// QR code lookup (public for security guards)
		api.GET("/visitors/qr/:qrcode", visitorHandler.GetVisitorByQR)
	}

	// Protected routes (all require society context)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/stats", analyticsHandler.GetStats)
		}

		// User routes
		users := protected.Group("/users")
		{
			users.GET("/profile", authHandler.GetProfile)
			users.GET("/residents", middleware.RequireRole("secretary"), userHandler.GetResidents)
			users.GET("/stats", middleware.RequireRole("secretary", "security"), userHandler.GetStats)
			users.GET("/:id", userHandler.GetUserByID)
		}

		// Visitor routes (all society-aware)
		visitors := protected.Group("/visitors")
		{
			visitors.GET("", visitorHandler.GetVisitors)
			visitors.POST("", middleware.RequireRole("resident"), visitorHandler.CreateVisitor)
			visitors.GET("/pending", middleware.RequireRole("secretary", "security"), userHandler.GetPendingVisitors)
			visitors.GET("/:id", middleware.RequireRole("resident", "secretary", "security"), visitorHandler.GetVisitorByID)
			visitors.PUT("/:id/approve", middleware.RequireRole("secretary", "security"), visitorHandler.ApproveVisitor)
			visitors.PUT("/:id/checkin", middleware.RequireRole("security"), visitorHandler.CheckInVisitor)
			visitors.PUT("/:id/checkout", middleware.RequireRole("security"), visitorHandler.CheckOutVisitor)
		}

		// Maintenance routes (all society-aware)
		maintenance := protected.Group("/maintenance")
		{
			maintenance.GET("", maintenanceHandler.GetMaintenanceRecords)
			maintenance.GET("/:id", maintenanceHandler.GetMaintenanceByID)
			maintenance.POST("", middleware.RequireRole("secretary"), maintenanceHandler.CreateMaintenanceRecord)
			maintenance.POST("/pay", middleware.RequireRole("resident"), maintenanceHandler.PayMaintenance)
		}

		// Amenity routes (all society-aware)
		amenities := protected.Group("/amenities")
		{
			amenities.GET("", amenityHandler.GetAmenities)
			amenities.POST("/book", middleware.RequireRole("resident"), amenityHandler.BookAmenity)
			amenities.GET("/bookings", amenityHandler.GetBookings)
			amenities.PUT("/bookings/:id/cancel", amenityHandler.CancelBooking)
		}

		// Notice routes (all society-aware)
		notices := protected.Group("/notices")
		{
			notices.GET("", noticeHandler.GetNotices)
			notices.GET("/:id", noticeHandler.GetNoticeByID)
			notices.POST("", middleware.RequireRole("secretary"), noticeHandler.CreateNotice)
			notices.PUT("/:id", middleware.RequireRole("secretary"), noticeHandler.UpdateNotice)
			notices.DELETE("/:id", middleware.RequireRole("secretary"), noticeHandler.DeleteNotice)
		}
	}
}
