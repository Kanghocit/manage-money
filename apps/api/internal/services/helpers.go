package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"gorm.io/gorm"
)

func primaryRole(user models.User) string {
	for _, r := range user.Roles {
		if r.Name == "admin" {
			return "admin"
		}
	}
	return "user"
}

func loadUserWithRoles(db *gorm.DB, id uuid.UUID) (models.User, error) {
	var user models.User
	err := db.Preload("Roles").First(&user, "id = ?", id).Error
	return user, err
}

func monthBounds(year, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end
}

func parseRange(rangeStr string) time.Time {
	now := time.Now().UTC()
	switch rangeStr {
	case "7d":
		return now.AddDate(0, 0, -7)
	case "60d":
		return now.AddDate(0, 0, -60)
	default:
		return now.AddDate(0, 0, -30)
	}
}
