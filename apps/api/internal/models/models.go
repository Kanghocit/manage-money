package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `gorm:"not null" json:"displayName"`
	Locale       string    `gorm:"default:vi" json:"locale"`
	AvatarURL    string    `json:"avatarUrl,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type Role struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `gorm:"uniqueIndex;not null" json:"name"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

type Category struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    *uuid.UUID   `gorm:"type:uuid;index" json:"userId,omitempty"`
	Name      string       `gorm:"not null" json:"name"`
	Icon      string       `json:"icon"`
	Color     string       `json:"color"`
	Type      CategoryType `gorm:"not null" json:"type"`
	IsSystem  bool         `gorm:"default:false" json:"isSystem"`
	CreatedAt time.Time    `json:"createdAt"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type Transaction struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID       `gorm:"type:uuid;index:idx_tx_user_occurred,priority:1;not null" json:"userId"`
	CategoryID uuid.UUID       `gorm:"type:uuid;not null" json:"categoryId"`
	Category   Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Amount     decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"amount"`
	Type       TransactionType `gorm:"not null" json:"type"`
	Note       string          `json:"note"`
	OccurredAt time.Time       `gorm:"index:idx_tx_user_occurred,priority:2;not null" json:"occurredAt"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type Budget struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;uniqueIndex:idx_budget_user_month,priority:1;not null" json:"userId"`
	Year      int          `gorm:"uniqueIndex:idx_budget_user_month,priority:2;not null" json:"year"`
	Month     int          `gorm:"uniqueIndex:idx_budget_user_month,priority:3;not null" json:"month"`
	Lines     []BudgetLine `gorm:"foreignKey:BudgetID" json:"lines,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BudgetLine struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	BudgetID     uuid.UUID       `gorm:"type:uuid;index;not null" json:"budgetId"`
	CategoryID   uuid.UUID       `gorm:"type:uuid;not null" json:"categoryId"`
	Category     Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	LimitAmount  decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"limitAmount"`
	SpentAmount  decimal.Decimal `gorm:"type:numeric(18,2);default:0" json:"spentAmount"`
}

func (bl *BudgetLine) BeforeCreate(tx *gorm.DB) error {
	if bl.ID == uuid.Nil {
		bl.ID = uuid.New()
	}
	return nil
}

type Goal struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID       `gorm:"type:uuid;index;not null" json:"userId"`
	Name          string          `gorm:"not null" json:"name"`
	TargetAmount  decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"targetAmount"`
	CurrentAmount decimal.Decimal `gorm:"type:numeric(18,2);default:0" json:"currentAmount"`
	Deadline      *time.Time      `json:"deadline,omitempty"`
	ImageURL      string          `json:"imageUrl,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (g *Goal) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

type UserSettings struct {
	UserID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Theme               string    `gorm:"default:system" json:"theme"`
	NotificationsEnabled bool     `gorm:"default:true" json:"notificationsEnabled"`
	DefaultCurrency     string    `gorm:"default:VND" json:"defaultCurrency"`
	Language            string    `gorm:"default:vi" json:"language"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type MonthlyAggregate struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID       `gorm:"type:uuid;uniqueIndex:idx_agg_user_month,priority:1;not null" json:"userId"`
	Year      int             `gorm:"uniqueIndex:idx_agg_user_month,priority:2;not null" json:"year"`
	Month     int             `gorm:"uniqueIndex:idx_agg_user_month,priority:3;not null" json:"month"`
	Income    decimal.Decimal `gorm:"type:numeric(18,2);default:0" json:"income"`
	Expense   decimal.Decimal `gorm:"type:numeric(18,2);default:0" json:"expense"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (m *MonthlyAggregate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type ExpenseDayStatus string

const (
	ExpenseDayDraft      ExpenseDayStatus = "draft"
	ExpenseDayProcessing ExpenseDayStatus = "processing"
	ExpenseDayReady      ExpenseDayStatus = "ready"
	ExpenseDayFailed     ExpenseDayStatus = "failed"
	ExpenseDayConfirmed  ExpenseDayStatus = "confirmed"
)

type ExpenseDayEntry struct {
	ID               uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           uuid.UUID        `gorm:"type:uuid;uniqueIndex:idx_expense_day_user_date,priority:1;not null" json:"userId"`
	EntryDate        time.Time        `gorm:"type:date;uniqueIndex:idx_expense_day_user_date,priority:2;not null" json:"entryDate"`
	Note             string           `json:"note"`
	Status           ExpenseDayStatus `gorm:"not null;default:draft" json:"status"`
	StatusMessage    string           `json:"statusMessage,omitempty"`
	ImageStorageKey  string           `json:"imageStorageKey,omitempty"`
	ImageMime        string           `json:"imageMime,omitempty"`
	OCRRawJSON       datatypes.JSON   `gorm:"type:jsonb" json:"ocrRawJson,omitempty"`
	TotalAmount      decimal.Decimal  `gorm:"type:numeric(18,2)" json:"totalAmount"`
	Lines            []ExpenseDayLineItem `gorm:"foreignKey:EntryID" json:"lines,omitempty"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

func (e *ExpenseDayEntry) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type ExpenseDayLineItem struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	EntryID       uuid.UUID        `gorm:"type:uuid;index;not null" json:"entryId"`
	Name          string           `gorm:"not null" json:"name"`
	Amount        decimal.Decimal  `gorm:"type:numeric(18,2);not null" json:"amount"`
	CategoryID    *uuid.UUID       `gorm:"type:uuid" json:"categoryId,omitempty"`
	TransactionID *uuid.UUID       `gorm:"type:uuid" json:"transactionId,omitempty"`
	SortOrder     int              `gorm:"default:0" json:"sortOrder"`
	Confidence    *float64         `json:"confidence,omitempty"`
}

func (l *ExpenseDayLineItem) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
