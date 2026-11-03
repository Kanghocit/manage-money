package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BudgetService struct {
	db *gorm.DB
}

func NewBudgetService(db *gorm.DB) *BudgetService {
	return &BudgetService{db: db}
}

type BudgetLineInput struct {
	CategoryID  uuid.UUID
	LimitAmount decimal.Decimal
}

func (s *BudgetService) GetOrCreate(userID uuid.UUID, year, month int) (models.Budget, error) {
	var budget models.Budget
	err := s.db.Preload("Lines.Category").Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&budget).Error
	if err == nil {
		if err := s.refreshSpent(userID, &budget); err != nil {
			return budget, err
		}
		return budget, nil
	}
	if err != gorm.ErrRecordNotFound {
		return models.Budget{}, err
	}
	budget = models.Budget{UserID: userID, Year: year, Month: month}
	if err := s.db.Create(&budget).Error; err != nil {
		return models.Budget{}, err
	}
	return budget, nil
}

func (s *BudgetService) Upsert(userID uuid.UUID, year, month int, lines []BudgetLineInput) (models.Budget, error) {
	budget, err := s.GetOrCreate(userID, year, month)
	if err != nil {
		return models.Budget{}, err
	}
	return budget, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("budget_id = ?", budget.ID).Delete(&models.BudgetLine{}).Error; err != nil {
			return err
		}
		for _, line := range lines {
			bl := models.BudgetLine{
				BudgetID:    budget.ID,
				CategoryID:  line.CategoryID,
				LimitAmount: line.LimitAmount,
			}
			if err := tx.Create(&bl).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *BudgetService) refreshSpent(userID uuid.UUID, budget *models.Budget) error {
	start, end := monthBounds(budget.Year, budget.Month)
	for i := range budget.Lines {
		var spent decimal.Decimal
		err := s.db.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND category_id = ? AND type = ? AND occurred_at >= ? AND occurred_at < ?",
				userID, budget.Lines[i].CategoryID, models.TransactionExpense, start, end).
			Scan(&spent).Error
		if err != nil {
			return err
		}
		budget.Lines[i].SpentAmount = spent
	}
	return nil
}

func (s *BudgetService) TotalProgress(userID uuid.UUID, year, month int) (usedPercent int, totalLimit, totalSpent decimal.Decimal, err error) {
	budget, err := s.GetOrCreate(userID, year, month)
	if err != nil {
		return 0, decimal.Zero, decimal.Zero, err
	}
	for _, line := range budget.Lines {
		totalLimit = totalLimit.Add(line.LimitAmount)
		totalSpent = totalSpent.Add(line.SpentAmount)
	}
	if totalLimit.IsZero() {
		return 0, totalLimit, totalSpent, nil
	}
	pct := totalSpent.Div(totalLimit).Mul(decimal.NewFromInt(100)).IntPart()
	if pct > 100 {
		pct = 100
	}
	return int(pct), totalLimit, totalSpent, nil
}

func CurrentMonth() (int, int) {
	now := time.Now()
	return now.Year(), int(now.Month())
}
