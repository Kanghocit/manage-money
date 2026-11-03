package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/manage-money/api/internal/config"
	"github.com/manage-money/api/internal/database"
	"github.com/manage-money/api/internal/jobs"
	"github.com/manage-money/api/internal/services"
	"github.com/manage-money/api/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	aggSvc := services.NewAggregateService(db)
	blob, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	ocrSvc := services.NewOCRService(cfg)
	txSvc := services.NewTransactionService(db, nil)
	expenseDaySvc := services.NewExpenseDayService(db, blob, nil, ocrSvc, txSvc)
	srv := asynq.NewServer(redisOpt, asynq.Config{Concurrency: 5})
	mux := asynq.NewServeMux()

	mux.HandleFunc(jobs.TypeRecalculateAggregates, func(ctx context.Context, t *asynq.Task) error {
		var p jobs.RecalculatePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}
		userID, err := uuid.Parse(p.UserID)
		if err != nil {
			return err
		}
		return aggSvc.Recalculate(userID, p.Year, p.Month)
	})

	mux.HandleFunc(jobs.TypeOCRExpenseDay, func(ctx context.Context, t *asynq.Task) error {
		var p jobs.OCRExpenseDayPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}
		entryID, err := uuid.Parse(p.EntryID)
		if err != nil {
			return err
		}
		return expenseDaySvc.RunOCR(entryID)
	})

	mux.HandleFunc(jobs.TypeBudgetThreshold, func(ctx context.Context, t *asynq.Task) error {
		var p jobs.BudgetThresholdPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}
		log.Printf("budget threshold stub user=%s budget=%s category=%s percent=%d", p.UserID, p.BudgetID, p.CategoryID, p.Percent)
		return nil
	})

	log.Println("Asynq worker started")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
