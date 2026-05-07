package main

import (
	"context"
	"log"
	"time"

	"user-service/config"
	"user-service/internal/handler"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	//1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	//2. Connect MongoDB using config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(cfg.DBName)
	userCollection := db.Collection("users")

	//3. Initialize layers
	repo := repository.NewUserRepository(userCollection)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	healthHandler := handler.NewHealthHandler(client, cfg.AuthServiceURL)

	//4. Create Echo instance
	e := echo.New()

	//5. Middleware (inject secret)
	jwtMiddleware := middleware.JWTMiddleware(cfg.JWTSecret)

	//6. Routes
	userGroup := e.Group("/user", jwtMiddleware)
	userGroup.GET("/profile", h.GetProfile)
	userGroup.PUT("/profile", h.UpdateProfile)

	// internal service call (no auth)
	e.POST("/internal/user", h.CreateUser)
	// 2. Public ping route
	e.GET("/ping", healthHandler.Ping)

	// 3. Full health check route
	e.GET("/health", healthHandler.Health)

	// Kubernetes probes
	e.GET("/health/live", healthHandler.Live)   // ← new
	e.GET("/health/ready", healthHandler.Ready) // ← new

	//7. Start server
	log.Println("Server running on port:", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
