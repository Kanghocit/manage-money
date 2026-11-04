package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/manage-money/api/internal/auth"
	"github.com/manage-money/api/internal/config"
	"github.com/manage-money/api/internal/jobs"
	"github.com/manage-money/api/internal/middleware"
	redisstore "github.com/manage-money/api/internal/redis"
	"github.com/manage-money/api/internal/models"
	"github.com/manage-money/api/internal/response"
	"github.com/manage-money/api/internal/services"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Handler struct {
	cfg         config.Config
	db          *gorm.DB
	jwt         *auth.Service
	redis       *redisstore.Store
	authSvc     *services.AuthService
	txSvc       *services.TransactionService
	catSvc      *services.CategoryService
	budgetSvc   *services.BudgetService
	goalSvc     *services.GoalService
	dashboardSvc *services.DashboardService
	statsSvc    *services.StatsService
	settingsSvc  *services.SettingsService
	expenseDaySvc *services.ExpenseDayService
}

func New(cfg config.Config, db *gorm.DB, jwt *auth.Service, redis *redisstore.Store, enqueue *jobs.Enqueuer, expenseDaySvc *services.ExpenseDayService) *Handler {
	authSvc := services.NewAuthService(db)
	txSvc := services.NewTransactionService(db, enqueue)
	catSvc := services.NewCategoryService(db)
	budgetSvc := services.NewBudgetService(db)
	goalSvc := services.NewGoalService(db)
	dashboardSvc := services.NewDashboardService(db, txSvc, goalSvc, budgetSvc, redis)
	statsSvc := services.NewStatsService(db)
	settingsSvc := services.NewSettingsService(db)
	return &Handler{
		cfg: cfg, db: db, jwt: jwt, redis: redis,
		authSvc: authSvc, txSvc: txSvc, catSvc: catSvc, budgetSvc: budgetSvc,
		goalSvc: goalSvc, dashboardSvc: dashboardSvc, statsSvc: statsSvc, settingsSvc: settingsSvc,
		expenseDaySvc: expenseDaySvc,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { response.OK(c, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")
	{
		authG := v1.Group("/auth")
		authG.POST("/register", h.Register)
		authG.POST("/login", h.Login)
		authG.POST("/logout", h.Logout)
		authG.POST("/refresh", h.Refresh)

		protected := v1.Group("")
		protected.Use(middleware.Auth(h.jwt, h.redis))
		protected.GET("/auth/me", h.Me)
		protected.GET("/dashboard/summary", h.DashboardSummary)
		protected.GET("/transactions", h.ListTransactions)
		protected.GET("/transactions/daily", h.ListTransactionsDaily)
		protected.POST("/transactions", h.CreateTransaction)
		protected.GET("/transactions/:id", h.GetTransaction)
		protected.PATCH("/transactions/:id", h.UpdateTransaction)
		protected.DELETE("/transactions/:id", h.DeleteTransaction)
		protected.GET("/categories", h.ListCategories)
		protected.POST("/categories", h.CreateCategory)
		protected.PATCH("/categories/:id", h.UpdateCategory)
		protected.DELETE("/categories/:id", h.DeleteCategory)
		protected.GET("/budgets/current", h.GetBudget)
		protected.PUT("/budgets/current", h.PutBudget)
		protected.GET("/goals", h.ListGoals)
		protected.POST("/goals", h.CreateGoal)
		protected.PATCH("/goals/:id", h.UpdateGoal)
		protected.DELETE("/goals/:id", h.DeleteGoal)
		protected.PATCH("/goals/:id/contribute", h.ContributeGoal)
		protected.GET("/stats/trends", h.StatsTrends)
		protected.GET("/stats/spending-structure", h.StatsStructure)
		protected.GET("/settings", h.GetSettings)
		protected.PATCH("/settings", h.PatchSettings)
		protected.PATCH("/settings/profile", h.PatchProfile)
		protected.POST("/expense-days", h.CreateExpenseDay)
		protected.GET("/expense-days", h.ListExpenseDays)
		protected.GET("/expense-days/:id", h.GetExpenseDay)
		protected.POST("/expense-days/:id/image", h.UploadExpenseDayImage)
		protected.PATCH("/expense-days/:id/lines", h.PatchExpenseDayLines)
		protected.POST("/expense-days/:id/confirm", h.ConfirmExpenseDay)
		protected.GET("/media/*key", h.ServeMedia)

		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		admin.GET("/ping", func(c *gin.Context) { response.OK(c, gin.H{"message": "admin ok"}) })
	}
}

func (h *Handler) setAuthCookies(c *gin.Context, pair *auth.TokenPair) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieAccess, pair.AccessToken, int(pair.AccessTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
	c.SetCookie(auth.CookieRefresh, pair.RefreshToken, int(pair.RefreshTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.CookieAccess, "", -1, "/", "", h.cfg.CookieSecure, true)
	c.SetCookie(auth.CookieRefresh, "", -1, "/", "", h.cfg.CookieSecure, true)
}

func userID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(middleware.UserIDKey)
	return id.(uuid.UUID)
}

type registerReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	user, err := h.authSvc.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		if err == services.ErrEmailTaken {
			response.Fail(c, http.StatusConflict, "EMAIL_TAKEN", "Email đã được sử dụng")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	role := services.UserDTO(user)["role"].(string)
	pair, err := h.jwt.IssueTokens(user.ID, user.Email, role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	h.setAuthCookies(c, pair)
	response.Created(c, services.UserDTO(user))
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	user, err := h.authSvc.Login(req.Email, req.Password)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email hoặc mật khẩu không đúng")
		return
	}
	role := services.UserDTO(user)["role"].(string)
	pair, err := h.jwt.IssueTokens(user.ID, user.Email, role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	h.setAuthCookies(c, pair)
	response.OK(c, services.UserDTO(user))
}

func (h *Handler) Logout(c *gin.Context) {
	if refresh, err := c.Cookie(auth.CookieRefresh); err == nil {
		if claims, err := h.jwt.ParseRefresh(refresh); err == nil && h.redis != nil {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				_ = h.redis.BlacklistRefresh(c.Request.Context(), claims.ID, ttl)
			}
		}
	}
	h.clearAuthCookies(c)
	response.OK(c, gin.H{"loggedOut": true})
}

func (h *Handler) Refresh(c *gin.Context) {
	refresh, err := c.Cookie(auth.CookieRefresh)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Thiếu refresh token")
		return
	}
	claims, err := h.jwt.ParseRefresh(refresh)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Refresh token không hợp lệ")
		return
	}
	if h.redis != nil {
		blacklisted, _ := h.redis.IsRefreshBlacklisted(c.Request.Context(), claims.ID)
		if blacklisted {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Refresh token đã bị thu hồi")
			return
		}
	}
	user, err := h.authSvc.GetMe(claims.UserID)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Người dùng không tồn tại")
		return
	}
	role := services.UserDTO(user)["role"].(string)
	pair, err := h.jwt.IssueTokens(user.ID, user.Email, role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	h.setAuthCookies(c, pair)
	response.OK(c, services.UserDTO(user))
}

func (h *Handler) Me(c *gin.Context) {
	user, err := h.authSvc.GetMe(userID(c))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy người dùng")
		return
	}
	response.OK(c, services.UserDTO(user))
}

func (h *Handler) DashboardSummary(c *gin.Context) {
	rangeStr := c.DefaultQuery("range", "30d")
	summary, err := h.dashboardSvc.Summary(userID(c), rangeStr)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, summary)
}

func (h *Handler) ListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	dateFrom := c.Query("from")
	dateTo := c.Query("to")
	if dateFrom == "" && c.Query("date") != "" {
		dateFrom = c.Query("date")
		dateTo = c.Query("date")
	}
	items, total, err := h.txSvc.List(userID(c), services.ListTransactionsFilter{
		Month: c.Query("month"), DateFrom: dateFrom, DateTo: dateTo,
		Category: c.Query("category"), Type: c.Query("type"),
		Query: c.Query("q"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (h *Handler) ListTransactionsDaily(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "from và to bắt buộc (YYYY-MM-DD)")
		return
	}
	rows, err := h.txSvc.DailyExpenseTotals(userID(c), from, to)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"date":  r.Date,
			"total": r.Total.StringFixed(0),
			"count": r.Count,
		})
	}
	response.OK(c, out)
}

type txReq struct {
	CategoryID string `json:"categoryId" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Note       string `json:"note"`
	OccurredAt string `json:"occurredAt"`
}

func parseTxReq(req txReq) (services.CreateTransactionInput, error) {
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return services.CreateTransactionInput{}, err
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return services.CreateTransactionInput{}, err
	}
	at := time.Now().UTC()
	if req.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, req.OccurredAt); err == nil {
			at = t.UTC()
		}
	}
	txType := models.TransactionType(req.Type)
	if txType != models.TransactionIncome && txType != models.TransactionExpense {
		return services.CreateTransactionInput{}, errors.New("invalid transaction type")
	}
	return services.CreateTransactionInput{
		CategoryID: catID, Amount: amount,
		Type: txType, Note: req.Note, OccurredAt: at,
	}, nil
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var req txReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	in, err := parseTxReq(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	tx, err := h.txSvc.Create(userID(c), in)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if h.redis != nil {
		_ = h.redis.InvalidateDashboard(c.Request.Context(), userID(c).String())
	}
	response.Created(c, tx)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	tx, err := h.txSvc.Get(userID(c), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy giao dịch")
		return
	}
	response.OK(c, tx)
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	var req txReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	in, err := parseTxReq(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	tx, err := h.txSvc.Update(userID(c), id, in)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if h.redis != nil {
		_ = h.redis.InvalidateDashboard(c.Request.Context(), userID(c).String())
	}
	response.OK(c, tx)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	if err := h.txSvc.Delete(userID(c), id); err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy giao dịch")
		return
	}
	if h.redis != nil {
		_ = h.redis.InvalidateDashboard(c.Request.Context(), userID(c).String())
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.catSvc.List(userID(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, items)
}

type catReq struct {
	Name  string `json:"name" binding:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Type  string `json:"type" binding:"required"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req catReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	cat, err := h.catSvc.Create(userID(c), services.CategoryInput{Name: req.Name, Icon: req.Icon, Color: req.Color, Type: models.CategoryType(req.Type)})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.Created(c, cat)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req catReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	cat, err := h.catSvc.Update(userID(c), id, services.CategoryInput{Name: req.Name, Icon: req.Icon, Color: req.Color, Type: models.CategoryType(req.Type)})
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy danh mục")
		return
	}
	response.OK(c, cat)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.catSvc.Delete(userID(c), id); err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy danh mục")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) GetBudget(c *gin.Context) {
	y, m := services.CurrentMonth()
	if yStr := c.Query("year"); yStr != "" {
		if v, err := strconv.Atoi(yStr); err == nil {
			y = v
		}
	}
	if mStr := c.Query("month"); mStr != "" {
		if v, err := strconv.Atoi(mStr); err == nil {
			m = v
		}
	}
	budget, err := h.budgetSvc.GetOrCreate(userID(c), y, m)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	pct, limit, spent, _ := h.budgetSvc.TotalProgress(userID(c), y, m)
	response.OK(c, gin.H{"budget": budget, "usedPercent": pct, "totalLimit": limit.StringFixed(0), "totalSpent": spent.StringFixed(0)})
}

type budgetPutReq struct {
	Lines []struct {
		CategoryID  string `json:"categoryId"`
		LimitAmount string `json:"limitAmount"`
	} `json:"lines"`
	Year  int `json:"year"`
	Month int `json:"month"`
}

func (h *Handler) PutBudget(c *gin.Context) {
	var req budgetPutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	y, m := req.Year, req.Month
	if y == 0 || m == 0 {
		y, m = services.CurrentMonth()
	}
	var lines []services.BudgetLineInput
	for _, l := range req.Lines {
		cid, err := uuid.Parse(l.CategoryID)
		if err != nil {
			continue
		}
		amt, err := decimal.NewFromString(l.LimitAmount)
		if err != nil {
			continue
		}
		lines = append(lines, services.BudgetLineInput{CategoryID: cid, LimitAmount: amt})
	}
	budget, err := h.budgetSvc.Upsert(userID(c), y, m, lines)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	budget, _ = h.budgetSvc.GetOrCreate(userID(c), y, m)
	response.OK(c, budget)
}

func (h *Handler) ListGoals(c *gin.Context) {
	items, err := h.goalSvc.List(userID(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, items)
}

type goalReq struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  string  `json:"targetAmount" binding:"required"`
	CurrentAmount string  `json:"currentAmount"`
	Deadline      *string `json:"deadline"`
	ImageURL      string  `json:"imageUrl"`
}

func parseGoalReq(req goalReq) (services.GoalInput, error) {
	target, err := decimal.NewFromString(req.TargetAmount)
	if err != nil {
		return services.GoalInput{}, err
	}
	current := decimal.Zero
	if req.CurrentAmount != "" {
		current, err = decimal.NewFromString(req.CurrentAmount)
		if err != nil {
			return services.GoalInput{}, err
		}
	}
	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		if t, err := time.Parse(time.RFC3339, *req.Deadline); err == nil {
			deadline = &t
		}
	}
	return services.GoalInput{Name: req.Name, TargetAmount: target, CurrentAmount: current, Deadline: deadline, ImageURL: req.ImageURL}, nil
}

func (h *Handler) CreateGoal(c *gin.Context) {
	var req goalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	in, err := parseGoalReq(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	g, err := h.goalSvc.Create(userID(c), in)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.Created(c, g)
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req goalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	in, err := parseGoalReq(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	g, err := h.goalSvc.Update(userID(c), id, in)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy mục tiêu")
		return
	}
	response.OK(c, g)
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.goalSvc.Delete(userID(c), id); err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy mục tiêu")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

type contributeReq struct {
	Amount string `json:"amount" binding:"required"`
}

func (h *Handler) ContributeGoal(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req contributeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "Số tiền không hợp lệ")
		return
	}
	g, err := h.goalSvc.Contribute(userID(c), id, amount)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy mục tiêu")
		return
	}
	response.OK(c, g)
}

func (h *Handler) StatsTrends(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	items, err := h.statsSvc.Trends(userID(c), months)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, items)
}

func (h *Handler) StatsStructure(c *gin.Context) {
	items, err := h.statsSvc.SpendingStructure(userID(c), c.Query("month"))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, items)
}

func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.settingsSvc.Get(userID(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, settings)
}

type settingsPatchReq struct {
	Theme                *string `json:"theme"`
	NotificationsEnabled *bool   `json:"notificationsEnabled"`
	DefaultCurrency      *string `json:"defaultCurrency"`
	Language             *string `json:"language"`
}

func (h *Handler) PatchSettings(c *gin.Context) {
	var req settingsPatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	settings, err := h.settingsSvc.Patch(userID(c), services.SettingsPatch{
		Theme: req.Theme, NotificationsEnabled: req.NotificationsEnabled,
		DefaultCurrency: req.DefaultCurrency, Language: req.Language,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, settings)
}

type profilePatchReq struct {
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

func (h *Handler) PatchProfile(c *gin.Context) {
	var req profilePatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	user, err := h.settingsSvc.UpdateProfile(userID(c), req.DisplayName, req.AvatarURL)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, services.UserDTO(user))
}
