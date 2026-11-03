package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/manage-money/api/internal/auth"
	"github.com/manage-money/api/internal/config"
	"github.com/manage-money/api/internal/database"
	"github.com/manage-money/api/internal/handlers"
	"github.com/manage-money/api/internal/jobs"
	redisstore "github.com/manage-money/api/internal/redis"
	"github.com/manage-money/api/internal/services"
	"github.com/manage-money/api/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("seed: %v", err)
	}

	redis, err := redisstore.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis opt: %v", err)
	}
	enqueue := jobs.NewEnqueuer(redisOpt)
	defer enqueue.Close()

	blob, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	jwtSvc := auth.NewService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	ocrSvc := services.NewOCRService(cfg)
	txSvc := services.NewTransactionService(db, enqueue)
	expenseDaySvc := services.NewExpenseDayService(db, blob, enqueue, ocrSvc, txSvc)
	h := handlers.New(cfg, db, jwtSvc, redis, enqueue, expenseDaySvc)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	h.RegisterRoutes(r)
	log.Printf("API listening on :%s", cfg.APIPort)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatal(err)
	}
}
