package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GoalService struct {
	db *gorm.DB
}

func NewGoalService(db *gorm.DB) *GoalService {
	return &GoalService{db: db}
}

type GoalInput struct {
	Name          string
	TargetAmount  decimal.Decimal
	CurrentAmount decimal.Decimal
	Deadline      *time.Time
	ImageURL      string
}

func (s *GoalService) List(userID uuid.UUID) ([]models.Goal, error) {
	var goals []models.Goal
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&goals).Error
	return goals, err
}

func (s *GoalService) Create(userID uuid.UUID, in GoalInput) (models.Goal, error) {
	g := models.Goal{
		UserID:        userID,
		Name:          in.Name,
		TargetAmount:  in.TargetAmount,
		CurrentAmount: in.CurrentAmount,
		Deadline:      in.Deadline,
		ImageURL:      in.ImageURL,
	}
	err := s.db.Create(&g).Error
	return g, err
}

func (s *GoalService) Update(userID, id uuid.UUID, in GoalInput) (models.Goal, error) {
	var g models.Goal
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&g).Error; err != nil {
		return models.Goal{}, err
	}
	g.Name = in.Name
	g.TargetAmount = in.TargetAmount
	g.CurrentAmount = in.CurrentAmount
	g.Deadline = in.Deadline
	g.ImageURL = in.ImageURL
	err := s.db.Save(&g).Error
	return g, err
}

func (s *GoalService) Contribute(userID, id uuid.UUID, amount decimal.Decimal) (models.Goal, error) {
	var g models.Goal
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&g).Error; err != nil {
		return models.Goal{}, err
	}
	g.CurrentAmount = g.CurrentAmount.Add(amount)
	err := s.db.Save(&g).Error
	return g, err
}

func (s *GoalService) Delete(userID, id uuid.UUID) error {
	return s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Goal{}).Error
}

func (s *GoalService) AverageProgress(userID uuid.UUID) int {
	goals, err := s.List(userID)
	if err != nil || len(goals) == 0 {
		return 0
	}
	var sum int64
	for _, g := range goals {
		if g.TargetAmount.IsZero() {
			continue
		}
		p := g.CurrentAmount.Div(g.TargetAmount).Mul(decimal.NewFromInt(100)).IntPart()
		if p > 100 {
			p = 100
		}
		sum += p
	}
	return int(sum / int64(len(goals)))
}
