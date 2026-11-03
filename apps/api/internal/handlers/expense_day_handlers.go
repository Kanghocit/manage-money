package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/manage-money/api/internal/models"
	"github.com/manage-money/api/internal/response"
	"github.com/manage-money/api/internal/services"
	"github.com/shopspring/decimal"
)

func (h *Handler) CreateExpenseDay(c *gin.Context) {
	var req struct {
		Date string `json:"date" binding:"required"`
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	entry, err := h.expenseDaySvc.GetOrCreate(userID(c), req.Date)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	if req.Note != "" {
		entry.Note = req.Note
		_ = h.db.Save(&entry).Error
	}
	response.Created(c, h.expenseDayDTO(entry))
}

func (h *Handler) ListExpenseDays(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "from và to bắt buộc (YYYY-MM-DD)")
		return
	}
	items, err := h.expenseDaySvc.List(userID(c), from, to)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, e := range items {
		out = append(out, h.expenseDayDTO(e))
	}
	response.OK(c, out)
}

func (h *Handler) GetExpenseDay(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	entry, err := h.expenseDaySvc.Get(userID(c), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy")
		return
	}
	response.OK(c, h.expenseDayDTO(entry))
}

func (h *Handler) UploadExpenseDayImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	file, err := c.FormFile("image")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "Thiếu file image")
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	defer f.Close()
	mime := file.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/jpeg"
	}
	entry, err := h.expenseDaySvc.UploadImage(userID(c), id, f, file.Size, mime)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, h.expenseDayDTO(entry))
}

func (h *Handler) PatchExpenseDayLines(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	var req struct {
		Lines []struct {
			Name       string  `json:"name"`
			Amount     string  `json:"amount"`
			CategoryID *string `json:"categoryId"`
		} `json:"lines"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	var lines []services.LineInput
	for i, l := range req.Lines {
		amt, err := decimal.NewFromString(l.Amount)
		if err != nil {
			continue
		}
		var catID *uuid.UUID
		if l.CategoryID != nil && *l.CategoryID != "" {
			parsed, err := uuid.Parse(*l.CategoryID)
			if err == nil {
				catID = &parsed
			}
		}
		lines = append(lines, services.LineInput{Name: l.Name, Amount: amt, CategoryID: catID, SortOrder: i})
	}
	entry, err := h.expenseDaySvc.UpdateLines(userID(c), id, lines)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.OK(c, h.expenseDayDTO(entry))
}

func (h *Handler) ConfirmExpenseDay(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", "ID không hợp lệ")
		return
	}
	var req struct {
		DefaultCategoryID string `json:"defaultCategoryId"`
	}
	_ = c.ShouldBindJSON(&req)
	catID := uuid.Nil
	if req.DefaultCategoryID != "" {
		catID, _ = uuid.Parse(req.DefaultCategoryID)
	}
	if catID == uuid.Nil {
		var cat models.Category
		if err := h.db.Where("is_system = ? AND name = ?", true, "Khác").First(&cat).Error; err == nil {
			catID = cat.ID
		}
	}
	entry, err := h.expenseDaySvc.Confirm(userID(c), id, catID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}
	response.OK(c, h.expenseDayDTO(entry))
}

func (h *Handler) ServeMedia(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	key, _ = url.PathUnescape(key)
	rc, mime, err := h.expenseDaySvc.MediaReader(userID(c), key)
	if err != nil {
		response.Fail(c, http.StatusForbidden, "FORBIDDEN", "Không truy cập được media")
		return
	}
	defer rc.Close()
	c.Header("Cache-Control", "private, max-age=300")
	c.DataFromReader(http.StatusOK, -1, mime, rc, nil)
}

func (h *Handler) expenseDayDTO(e models.ExpenseDayEntry) gin.H {
	imageURL := ""
	if e.ImageStorageKey != "" {
		imageURL = "/api/v1/media/" + url.PathEscape(e.ImageStorageKey)
	}
	lineDTOs := make([]gin.H, 0, len(e.Lines))
	for _, l := range e.Lines {
		lineDTOs = append(lineDTOs, gin.H{
			"id":            l.ID,
			"name":          l.Name,
			"amount":        l.Amount.StringFixed(0),
			"categoryId":    l.CategoryID,
			"transactionId": l.TransactionID,
			"sortOrder":     l.SortOrder,
			"confidence":    l.Confidence,
		})
	}
	return gin.H{
		"id":              e.ID,
		"entryDate":       e.EntryDate.Format("2006-01-02"),
		"note":            e.Note,
		"status":          e.Status,
		"statusMessage":   e.StatusMessage,
		"imageStorageKey": e.ImageStorageKey,
		"imageUrl":        imageURL,
		"totalAmount":     e.TotalAmount.StringFixed(0),
		"lines":           lineDTOs,
		"createdAt":       e.CreatedAt,
		"updatedAt":       e.UpdatedAt,
	}
}
