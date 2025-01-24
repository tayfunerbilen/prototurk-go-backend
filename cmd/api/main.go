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

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)

	// Initialize Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.JWT())

	// Routes
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authHandler.Me)
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
