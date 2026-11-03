package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

type MonthlyTrend struct {
	Month   string `json:"month"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

func (s *StatsService) Trends(userID uuid.UUID, months int) ([]MonthlyTrend, error) {
	if months <= 0 || months > 12 {
		months = 6
	}
	now := time.Now()
	var out []MonthlyTrend
	for i := months - 1; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		y, m := t.Year(), int(t.Month())
		start, end := monthBounds(y, m)
		var inc, exp decimal.Decimal
		s.db.Model(&models.Transaction{}).Select("COALESCE(SUM(amount),0)").Where("user_id = ? AND type = ? AND occurred_at >= ? AND occurred_at < ?", userID, models.TransactionIncome, start, end).Scan(&inc)
		s.db.Model(&models.Transaction{}).Select("COALESCE(SUM(amount),0)").Where("user_id = ? AND type = ? AND occurred_at >= ? AND occurred_at < ?", userID, models.TransactionExpense, start, end).Scan(&exp)
		out = append(out, MonthlyTrend{
			Month:   t.Format("2006-01"),
			Income:  inc.StringFixed(0),
			Expense: exp.StringFixed(0),
		})
	}
	return out, nil
}

func (s *StatsService) SpendingStructure(userID uuid.UUID, month string) ([]StructureSlice, error) {
	var start, end time.Time
	if month != "" {
		if t, err := time.Parse("2006-01", month); err == nil {
			start, end = monthBounds(t.Year(), int(t.Month()))
		}
	}
	if start.IsZero() {
		now := time.Now()
		start, end = monthBounds(now.Year(), int(now.Month()))
	}
	dash := NewDashboardService(s.db, NewTransactionService(s.db, nil), NewGoalService(s.db), NewBudgetService(s.db), nil)
	return dash.buildStructure(userID, start, end), nil
}

type SettingsService struct {
	db *gorm.DB
}

func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) Get(userID uuid.UUID) (models.UserSettings, error) {
	var settings models.UserSettings
	err := s.db.FirstOrCreate(&settings, models.UserSettings{UserID: userID}).Error
	return settings, err
}

type SettingsPatch struct {
	Theme                *string
	NotificationsEnabled *bool
	DefaultCurrency      *string
	Language             *string
}

func (s *SettingsService) Patch(userID uuid.UUID, patch SettingsPatch) (models.UserSettings, error) {
	settings, err := s.Get(userID)
	if err != nil {
		return settings, err
	}
	if patch.Theme != nil {
		settings.Theme = *patch.Theme
	}
	if patch.NotificationsEnabled != nil {
		settings.NotificationsEnabled = *patch.NotificationsEnabled
	}
	if patch.DefaultCurrency != nil {
		settings.DefaultCurrency = *patch.DefaultCurrency
	}
	if patch.Language != nil {
		settings.Language = *patch.Language
	}
	err = s.db.Save(&settings).Error
	return settings, err
}

func (s *SettingsService) UpdateProfile(userID uuid.UUID, displayName, avatarURL string) (models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return user, err
	}
	if displayName != "" {
		user.DisplayName = displayName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	err := s.db.Save(&user).Error
	return user, err
}
