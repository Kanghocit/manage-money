package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken     = errors.New("email taken")
	ErrInvalidLogin   = errors.New("invalid login")
	ErrUserNotFound   = errors.New("user not found")
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(email, password, displayName string) (models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var count int64
	s.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return models.User{}, ErrEmailTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		Locale:       "vi",
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		settings := models.UserSettings{UserID: user.ID, Theme: "system", DefaultCurrency: "VND", Language: "vi"}
		if err := tx.Create(&settings).Error; err != nil {
			return err
		}
		var role models.Role
		if err := tx.Where("name = ?", "user").First(&role).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID).Error
	})
	if err != nil {
		return models.User{}, err
	}
	return loadUserWithRoles(s.db, user.ID)
}

func (s *AuthService) Login(email, password string) (models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var user models.User
	if err := s.db.Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrInvalidLogin
		}
		return models.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidLogin
	}
	return user, nil
}

func (s *AuthService) GetMe(userID uuid.UUID) (models.User, error) {
	user, err := loadUserWithRoles(s.db, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrUserNotFound
	}
	return user, err
}

func UserDTO(u models.User) map[string]any {
	return map[string]any{
		"id":          u.ID,
		"email":       u.Email,
		"displayName": u.DisplayName,
		"locale":      u.Locale,
		"avatarUrl":   u.AvatarURL,
		"role":        primaryRole(u),
	}
}
