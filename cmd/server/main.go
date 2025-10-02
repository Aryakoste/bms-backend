package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bms-backend/api/routes"
	"bms-backend/internal/config"
	"bms-backend/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect(context.Background())

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8100",
		"*",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Add logging middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize routes
	routes.InitializeRoutes(router, db)

	// Create server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Building Management System - MULTI-SOCIETY VERSION")
		log.Printf("🔗 Server: http://localhost:%s", cfg.Port)
		log.Printf("📄 Health: http://localhost:%s/health", cfg.Port)
		log.Printf("🔐 API: http://localhost:%s/api/v1", cfg.Port)
		log.Printf("🏢 MULTI-SOCIETY FEATURES:")
		log.Printf("   • Society access codes")
		log.Printf("   • Data segregation by society")
		log.Printf("   • Society validation API")
		log.Printf("   • Society-aware authentication")
		log.Printf("   • Society-scoped data operations")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📴 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}
