package jobs

const (
	TypeRecalculateAggregates = "recalculate:monthly_aggregates"
	TypeBudgetThreshold       = "notify:budget_threshold"
	TypeOCRExpenseDay         = "ocr:expense_day"
)

type RecalculatePayload struct {
	UserID string `json:"userId"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

type BudgetThresholdPayload struct {
	UserID     string `json:"userId"`
	BudgetID   string `json:"budgetId"`
	CategoryID string `json:"categoryId"`
	Percent    int    `json:"percent"`
}

type OCRExpenseDayPayload struct {
	EntryID string `json:"entryId"`
}
