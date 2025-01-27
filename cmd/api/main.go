package main

import (
	"log"
	"os"

	"prototurk/internal/database"
	"prototurk/internal/handlers"
	"prototurk/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database connection
	dbConfig := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}

	// Run migrations
	if err := database.RunMigrations(dbConfig); err != nil {
		log.Fatal("Error running migrations:", err)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	// Seed default admin
	if err := database.SeedDefaultAdmin(db); err != nil {
		log.Fatal("Error seeding default admin:", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	// Initialize Gin router
	router := gin.Default()

	// JWT secret'ı global olarak ekle
	router.Use(func(c *gin.Context) {
		c.Set("jwt_secret", os.Getenv("JWT_SECRET"))
		c.Set("db", db)
		c.Next()
	})

	// Routes
	api := router.Group("/api")
	{
		// User routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authHandler.Me)
			auth.PUT("/profile", authHandler.UpdateProfile)
			auth.PUT("/password", authHandler.UpdatePassword)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminJWT())
		{
			// Auth
			admin.POST("/login", adminHandler.Login)
			admin.GET("/me", adminHandler.Me)

			// CRUD
			admin.POST("", adminHandler.Create)
			admin.GET("", adminHandler.List)
			admin.GET("/:id", adminHandler.Get)
			admin.PUT("/:id", adminHandler.Update)
			admin.DELETE("/:id", adminHandler.Delete)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server error:", err)
	}
}
