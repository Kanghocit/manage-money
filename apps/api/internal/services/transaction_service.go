package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/jobs"
	"github.com/manage-money/api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionService struct {
	db      *gorm.DB
	enqueue *jobs.Enqueuer
}

func NewTransactionService(db *gorm.DB, enqueue *jobs.Enqueuer) *TransactionService {
	return &TransactionService{db: db, enqueue: enqueue}
}

type ListTransactionsFilter struct {
	Month    string
	DateFrom string
	DateTo   string
	Category string
	Type     string
	Query    string
	Page     int
	PageSize int
}

func (s *TransactionService) List(userID uuid.UUID, f ListTransactionsFilter) ([]models.Transaction, int64, error) {
	q := s.db.Model(&models.Transaction{}).Where("user_id = ?", userID).Preload("Category")
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.Category != "" {
		if catID, err := uuid.Parse(f.Category); err == nil {
			q = q.Where("category_id = ?", catID)
		}
	}
	if f.DateFrom != "" && f.DateTo != "" {
		if from, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			if to, err := time.Parse("2006-01-02", f.DateTo); err == nil {
				end := to.Add(24 * time.Hour)
				q = q.Where("occurred_at >= ? AND occurred_at < ?", from.UTC(), end.UTC())
			}
		}
	} else if f.Month != "" {
		if t, err := time.Parse("2006-01", f.Month); err == nil {
			start, end := monthBounds(t.Year(), int(t.Month()))
			q = q.Where("occurred_at >= ? AND occurred_at < ?", start, end)
		}
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("note ILIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var items []models.Transaction
	err := q.Order("occurred_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

type CreateTransactionInput struct {
	CategoryID uuid.UUID
	Amount     decimal.Decimal
	Type       models.TransactionType
	Note       string
	OccurredAt time.Time
}

func (s *TransactionService) Create(userID uuid.UUID, in CreateTransactionInput) (models.Transaction, error) {
	tx := models.Transaction{
		UserID:     userID,
		CategoryID: in.CategoryID,
		Amount:     in.Amount,
		Type:       in.Type,
		Note:       in.Note,
		OccurredAt: in.OccurredAt.UTC(),
	}
	if err := s.db.Create(&tx).Error; err != nil {
		return models.Transaction{}, err
	}
	s.scheduleJobs(userID, in.OccurredAt)
	return s.Get(userID, tx.ID)
}

func (s *TransactionService) Get(userID, id uuid.UUID) (models.Transaction, error) {
	var tx models.Transaction
	err := s.db.Preload("Category").Where("user_id = ? AND id = ?", userID, id).First(&tx).Error
	return tx, err
}

func (s *TransactionService) Update(userID, id uuid.UUID, in CreateTransactionInput) (models.Transaction, error) {
	var tx models.Transaction
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&tx).Error; err != nil {
		return models.Transaction{}, err
	}
	tx.CategoryID = in.CategoryID
	tx.Amount = in.Amount
	tx.Type = in.Type
	tx.Note = in.Note
	tx.OccurredAt = in.OccurredAt.UTC()
	if err := s.db.Save(&tx).Error; err != nil {
		return models.Transaction{}, err
	}
	s.scheduleJobs(userID, in.OccurredAt)
	return s.Get(userID, id)
}

func (s *TransactionService) Delete(userID, id uuid.UUID) error {
	var tx models.Transaction
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&tx).Error; err != nil {
		return err
	}
	if err := s.db.Delete(&tx).Error; err != nil {
		return err
	}
	s.scheduleJobs(userID, tx.OccurredAt)
	return nil
}

func (s *TransactionService) ScheduleJobs(userID uuid.UUID, at time.Time) {
	s.scheduleJobs(userID, at)
}

func (s *TransactionService) scheduleJobs(userID uuid.UUID, at time.Time) {
	if s.enqueue == nil {
		return
	}
	y, m, _ := at.Date()
	_ = s.enqueue.RecalculateMonthly(userID, y, int(m))
}

func (s *TransactionService) SumByType(userID uuid.UUID, from, to time.Time, txType models.TransactionType) (decimal.Decimal, error) {
	var result struct {
		Total decimal.Decimal
	}
	err := s.db.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND type = ? AND occurred_at >= ? AND occurred_at < ?", userID, txType, from, to).
		Scan(&result).Error
	return result.Total, err
}

func (s *TransactionService) Recent(userID uuid.UUID, limit int) ([]models.Transaction, error) {
	if limit <= 0 {
		limit = 5
	}
	var items []models.Transaction
	err := s.db.Preload("Category").Where("user_id = ?", userID).Order("occurred_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

type DailyExpenseRow struct {
	Date  string          `json:"date"`
	Total decimal.Decimal `json:"total"`
	Count int64           `json:"count"`
}

func (s *TransactionService) DailyExpenseTotals(userID uuid.UUID, fromStr, toStr string) ([]DailyExpenseRow, error) {
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, err
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, err
	}
	endExclusive := to.Add(24 * time.Hour).UTC()
	fromUTC := from.UTC()
	var rows []DailyExpenseRow
	err = s.db.Model(&models.Transaction{}).
		Select(`TO_CHAR((occurred_at AT TIME ZONE 'UTC')::date, 'YYYY-MM-DD') as date, COALESCE(SUM(amount), 0) as total, COUNT(*) as count`).
		Where("user_id = ? AND type = ? AND occurred_at >= ? AND occurred_at < ?", userID, models.TransactionExpense, fromUTC, endExclusive).
		Group("DATE(occurred_at AT TIME ZONE 'UTC')").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

var ErrNotFound = errors.New("not found")
