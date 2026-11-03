package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	redisstore "github.com/manage-money/api/internal/redis"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DashboardService struct {
	db    *gorm.DB
	txSvc *TransactionService
	goalSvc *GoalService
	budgetSvc *BudgetService
	redis *redisstore.Store
}

func NewDashboardService(db *gorm.DB, tx *TransactionService, goals *GoalService, budget *BudgetService, redis *redisstore.Store) *DashboardService {
	return &DashboardService{db: db, txSvc: tx, goalSvc: goals, budgetSvc: budget, redis: redis}
}

type DashboardSummary struct {
	TotalIncome      string `json:"totalIncome"`
	TotalExpense     string `json:"totalExpense"`
	SavingsDeltaPct  int    `json:"savingsDeltaPct"`
	GoalProgressPct  int    `json:"goalProgressPct"`
	Trend            []TrendPoint `json:"trend"`
	Structure        []StructureSlice `json:"structure"`
	RecentTransactions []models.Transaction `json:"recentTransactions"`
	BudgetLines      []BudgetLineView `json:"budgetLines"`
}

type TrendPoint struct {
	Date    string `json:"date"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type StructureSlice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Color string `json:"color"`
}

type BudgetLineView struct {
	CategoryName string `json:"categoryName"`
	UsedPercent  int    `json:"usedPercent"`
	Spent        string `json:"spent"`
	Limit        string `json:"limit"`
	Color        string `json:"color"`
}

func (s *DashboardService) Summary(userID uuid.UUID, rangeStr string) (DashboardSummary, error) {
	if s.redis != nil {
		if cached, err := s.redis.GetDashboardCache(context.Background(), userID.String()); err == nil && len(cached) > 0 {
			var out DashboardSummary
			if json.Unmarshal(cached, &out) == nil {
				return out, nil
			}
		}
	}

	from := parseRange(rangeStr)
	to := time.Now().UTC()
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	mStart, mEnd := monthBounds(year, month)

	income, _ := s.txSvc.SumByType(userID, mStart, mEnd, models.TransactionIncome)
	expense, _ := s.txSvc.SumByType(userID, mStart, mEnd, models.TransactionExpense)

	prevStart := mStart.AddDate(0, -1, 0)
	prevEnd := mStart
	prevIncome, _ := s.txSvc.SumByType(userID, prevStart, prevEnd, models.TransactionIncome)
	prevExpense, _ := s.txSvc.SumByType(userID, prevStart, prevEnd, models.TransactionExpense)
	savingsDelta := calcDeltaPct(prevIncome.Sub(prevExpense), income.Sub(expense))

	trend := s.buildTrend(userID, from, to)
	structure := s.buildStructure(userID, mStart, mEnd)
	recent, _ := s.txSvc.Recent(userID, 5)
	budgetLines := s.buildBudgetLines(userID, year, month)
	goalPct := s.goalSvc.AverageProgress(userID)

	out := DashboardSummary{
		TotalIncome:        income.StringFixed(0),
		TotalExpense:       expense.StringFixed(0),
		SavingsDeltaPct:    savingsDelta,
		GoalProgressPct:    goalPct,
		Trend:              trend,
		Structure:          structure,
		RecentTransactions: recent,
		BudgetLines:        budgetLines,
	}
	if s.redis != nil {
		if b, err := json.Marshal(out); err == nil {
			_ = s.redis.CacheDashboard(context.Background(), userID.String(), b, 2*time.Minute)
		}
	}
	return out, nil
}

func calcDeltaPct(prev, curr decimal.Decimal) int {
	if prev.IsZero() {
		if curr.IsZero() {
			return 0
		}
		return 100
	}
	delta := curr.Sub(prev).Div(prev.Abs()).Mul(decimal.NewFromInt(100))
	return int(delta.IntPart())
}

func (s *DashboardService) buildTrend(userID uuid.UUID, from, to time.Time) []TrendPoint {
	var points []TrendPoint
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dayEnd := d.AddDate(0, 0, 1)
		inc, _ := s.txSvc.SumByType(userID, d, dayEnd, models.TransactionIncome)
		exp, _ := s.txSvc.SumByType(userID, d, dayEnd, models.TransactionExpense)
		points = append(points, TrendPoint{
			Date:    d.Format("2006-01-02"),
			Income:  inc.StringFixed(0),
			Expense: exp.StringFixed(0),
		})
	}
	return points
}

func (s *DashboardService) buildStructure(userID uuid.UUID, start, end time.Time) []StructureSlice {
	type row struct {
		Name  string
		Color string
		Total decimal.Decimal
	}
	var rows []row
	s.db.Table("transactions").
		Select("categories.name as name, categories.color as color, COALESCE(SUM(transactions.amount), 0) as total").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.user_id = ? AND transactions.type = ? AND transactions.occurred_at >= ? AND transactions.occurred_at < ?",
			userID, models.TransactionExpense, start, end).
		Group("categories.name, categories.color").
		Scan(&rows)
	var out []StructureSlice
	for _, r := range rows {
		out = append(out, StructureSlice{Name: r.Name, Value: r.Total.StringFixed(0), Color: r.Color})
	}
	return out
}

func (s *DashboardService) buildBudgetLines(userID uuid.UUID, year, month int) []BudgetLineView {
	budget, err := s.budgetSvc.GetOrCreate(userID, year, month)
	if err != nil {
		return nil
	}
	var views []BudgetLineView
	for _, line := range budget.Lines {
		pct := 0
		if !line.LimitAmount.IsZero() {
			pct = int(line.SpentAmount.Div(line.LimitAmount).Mul(decimal.NewFromInt(100)).IntPart())
		}
		name := ""
		color := ""
		if line.Category.ID != uuid.Nil {
			name = line.Category.Name
			color = line.Category.Color
		}
		views = append(views, BudgetLineView{
			CategoryName: name,
			UsedPercent:  pct,
			Spent:        line.SpentAmount.StringFixed(0),
			Limit:        line.LimitAmount.StringFixed(0),
			Color:        color,
		})
	}
	return views
}
