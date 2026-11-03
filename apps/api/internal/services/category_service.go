package services

import (
	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) List(userID uuid.UUID) ([]models.Category, error) {
	var items []models.Category
	err := s.db.Where("is_system = ? OR user_id = ?", true, userID).Order("is_system DESC, name ASC").Find(&items).Error
	return items, err
}

type CategoryInput struct {
	Name  string
	Icon  string
	Color string
	Type  models.CategoryType
}

func (s *CategoryService) Create(userID uuid.UUID, in CategoryInput) (models.Category, error) {
	cat := models.Category{
		UserID:   &userID,
		Name:     in.Name,
		Icon:     in.Icon,
		Color:    in.Color,
		Type:     in.Type,
		IsSystem: false,
	}
	err := s.db.Create(&cat).Error
	return cat, err
}

func (s *CategoryService) Update(userID, id uuid.UUID, in CategoryInput) (models.Category, error) {
	var cat models.Category
	if err := s.db.Where("id = ? AND user_id = ? AND is_system = ?", id, userID, false).First(&cat).Error; err != nil {
		return models.Category{}, err
	}
	cat.Name = in.Name
	cat.Icon = in.Icon
	cat.Color = in.Color
	cat.Type = in.Type
	err := s.db.Save(&cat).Error
	return cat, err
}

func (s *CategoryService) Delete(userID, id uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ? AND is_system = ?", id, userID, false).Delete(&models.Category{}).Error
}
