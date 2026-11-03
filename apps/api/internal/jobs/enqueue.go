package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Enqueuer struct {
	client *asynq.Client
}

func NewEnqueuer(redisOpt asynq.RedisConnOpt) *Enqueuer {
	return &Enqueuer{client: asynq.NewClient(redisOpt)}
}

func (e *Enqueuer) Close() error {
	return e.client.Close()
}

func (e *Enqueuer) RecalculateMonthly(userID uuid.UUID, year, month int) error {
	payload, _ := json.Marshal(RecalculatePayload{
		UserID: userID.String(),
		Year:   year,
		Month:  month,
	})
	task := asynq.NewTask(TypeRecalculateAggregates, payload)
	_, err := e.client.Enqueue(task, asynq.ProcessIn(2*time.Second), asynq.Unique(5*time.Second))
	return err
}

func (e *Enqueuer) OCRExpenseDay(entryID uuid.UUID) error {
	payload, _ := json.Marshal(OCRExpenseDayPayload{EntryID: entryID.String()})
	task := asynq.NewTask(TypeOCRExpenseDay, payload)
	_, err := e.client.Enqueue(task)
	return err
}

func (e *Enqueuer) BudgetThreshold(userID, budgetID, categoryID uuid.UUID, percent int) error {
	payload, _ := json.Marshal(BudgetThresholdPayload{
		UserID:     userID.String(),
		BudgetID:   budgetID.String(),
		CategoryID: categoryID.String(),
		Percent:    percent,
	})
	task := asynq.NewTask(TypeBudgetThreshold, payload)
	_, err := e.client.Enqueue(task)
	return err
}
