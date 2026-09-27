package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ucup/taskmanager/internal/handler"
	"github.com/ucup/taskmanager/internal/repository"
	"github.com/ucup/taskmanager/internal/service"
	"github.com/ucup/taskmanager/pkg/cache"
	"github.com/ucup/taskmanager/pkg/db"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("backend/.env")

	dsn := getEnv("DATABASE_URL", "taskuser:taskpass@tcp(localhost:3306)/taskdb?parseTime=true&charset=utf8mb4&loc=Local")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	port := getEnv("PORT", "8080")

	database, err := db.ConnectMySQL(dsn)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	var c cache.Cache
	rc, err := cache.NewRedisCache(redisAddr, redisPassword, 0)
	if err != nil {
		log.Printf("WARNING: redis unavailable (%v) — running without cache", err)
	} else {
		c = rc
	}

	repo := repository.NewTaskRepository(database)
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc, c)

	r := gin.Default()
	// Permissive CORS for Expo dev client / web.
	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	})
	h.RegisterRoutes(r)

	log.Printf("API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
