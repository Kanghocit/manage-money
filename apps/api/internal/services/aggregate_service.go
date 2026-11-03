package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AggregateService struct {
	db *gorm.DB
}

func NewAggregateService(db *gorm.DB) *AggregateService {
	return &AggregateService{db: db}
}

func (s *AggregateService) Recalculate(userID uuid.UUID, year, month int) error {
	start, end := monthBounds(year, month)
	txSvc := NewTransactionService(s.db, nil)
	income, err := txSvc.SumByType(userID, start, end, models.TransactionIncome)
	if err != nil {
		return err
	}
	expense, err := txSvc.SumByType(userID, start, end, models.TransactionExpense)
	if err != nil {
		return err
	}
	var agg models.MonthlyAggregate
	err = s.db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&agg).Error
	if err == gorm.ErrRecordNotFound {
		agg = models.MonthlyAggregate{UserID: userID, Year: year, Month: month}
	}
	agg.Income = income
	agg.Expense = expense
	agg.UpdatedAt = time.Now().UTC()
	if agg.ID == uuid.Nil {
		return s.db.Create(&agg).Error
	}
	return s.db.Save(&agg).Error
}

func (s *AggregateService) SavingsRate(userID uuid.UUID, year, month int) decimal.Decimal {
	var agg models.MonthlyAggregate
	if err := s.db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&agg).Error; err != nil {
		return decimal.Zero
	}
	if agg.Income.IsZero() {
		return decimal.Zero
	}
	return agg.Income.Sub(agg.Expense).Div(agg.Income).Mul(decimal.NewFromInt(100))
}
