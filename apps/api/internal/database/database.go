package database

import (
	"log"

	"github.com/manage-money/api/internal/config"
	"github.com/manage-money/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
		&models.BudgetLine{},
		&models.Goal{},
		&models.UserSettings{},
		&models.MonthlyAggregate{},
		&models.ExpenseDayEntry{},
		&models.ExpenseDayLineItem{},
	)
}

func Seed(db *gorm.DB, cfg config.Config) error {
	var userRole, adminRole models.Role
	if err := db.FirstOrCreate(&userRole, models.Role{Name: "user"}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&adminRole, models.Role{Name: "admin"}).Error; err != nil {
		return err
	}

	systemCategories := []models.Category{
		{Name: "Ăn uống", Icon: "utensils", Color: "#F97316", Type: models.CategoryExpense, IsSystem: true},
		{Name: "Di chuyển", Icon: "car", Color: "#3B82F6", Type: models.CategoryExpense, IsSystem: true},
		{Name: "Mua sắm", Icon: "shopping-bag", Color: "#EC4899", Type: models.CategoryExpense, IsSystem: true},
		{Name: "Học tập", Icon: "book", Color: "#8B5CF6", Type: models.CategoryExpense, IsSystem: true},
		{Name: "Khác", Icon: "more-horizontal", Color: "#6B7280", Type: models.CategoryExpense, IsSystem: true},
		{Name: "Lương", Icon: "wallet", Color: "#22C55E", Type: models.CategoryIncome, IsSystem: true},
		{Name: "Thưởng", Icon: "gift", Color: "#14B8A6", Type: models.CategoryIncome, IsSystem: true},
	}
	for _, cat := range systemCategories {
		var existing models.Category
		if err := db.Where("name = ? AND is_system = ?", cat.Name, true).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}

	if cfg.SeedAdminEmail != "" && cfg.SeedAdminPassword != "" {
		var admin models.User
		if err := db.Where("email = ?", cfg.SeedAdminEmail).First(&admin).Error; err == gorm.ErrRecordNotFound {
			hash, err := bcrypt.GenerateFromPassword([]byte(cfg.SeedAdminPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			admin = models.User{
				Email:        cfg.SeedAdminEmail,
				PasswordHash: string(hash),
				DisplayName:  "Admin",
				Locale:       "vi",
			}
			if err := db.Create(&admin).Error; err != nil {
				return err
			}
			if err := db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING", admin.ID, adminRole.ID).Error; err != nil {
				log.Printf("seed admin role: %v", err)
			}
		}
	}
	return nil
}
