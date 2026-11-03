package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/manage-money/api/internal/jobs"
	"github.com/manage-money/api/internal/models"
	"github.com/manage-money/api/internal/storage"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrExpenseDayNotFound = errors.New("expense day not found")

type ExpenseDayService struct {
	db      *gorm.DB
	store   storage.BlobStore
	enqueue *jobs.Enqueuer
	ocr     *OCRService
	txSvc   *TransactionService
}

func NewExpenseDayService(db *gorm.DB, store storage.BlobStore, enqueue *jobs.Enqueuer, ocr *OCRService, txSvc *TransactionService) *ExpenseDayService {
	return &ExpenseDayService{db: db, store: store, enqueue: enqueue, ocr: ocr, txSvc: txSvc}
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func (s *ExpenseDayService) GetOrCreate(userID uuid.UUID, dateStr string) (models.ExpenseDayEntry, error) {
	d, err := parseDate(dateStr)
	if err != nil {
		return models.ExpenseDayEntry{}, err
	}
	var entry models.ExpenseDayEntry
	err = s.db.Preload("Lines").Where("user_id = ? AND entry_date = ?", userID, d).First(&entry).Error
	if err == nil {
		return entry, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entry, err
	}
	entry = models.ExpenseDayEntry{
		UserID:    userID,
		EntryDate: d,
		Status:    models.ExpenseDayDraft,
	}
	return entry, s.db.Create(&entry).Error
}

func (s *ExpenseDayService) List(userID uuid.UUID, fromStr, toStr string) ([]models.ExpenseDayEntry, error) {
	from, err := parseDate(fromStr)
	if err != nil {
		return nil, err
	}
	to, err := parseDate(toStr)
	if err != nil {
		return nil, err
	}
	var items []models.ExpenseDayEntry
	err = s.db.Preload("Lines").Where("user_id = ? AND entry_date >= ? AND entry_date <= ?", userID, from, to).
		Order("entry_date ASC").Find(&items).Error
	return items, err
}

func (s *ExpenseDayService) Get(userID, id uuid.UUID) (models.ExpenseDayEntry, error) {
	var entry models.ExpenseDayEntry
	err := s.db.Preload("Lines").Where("user_id = ? AND id = ?", userID, id).First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entry, ErrExpenseDayNotFound
	}
	return entry, err
}

func (s *ExpenseDayService) UploadImage(userID, id uuid.UUID, r io.Reader, size int64, mime string) (models.ExpenseDayEntry, error) {
	entry, err := s.Get(userID, id)
	if err != nil {
		return entry, err
	}
	ext := ".jpg"
	switch mime {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}
	key := storage.UserObjectKey(userID.String(), fmt.Sprintf("expense-days/%s%s", id.String(), ext))
	if err := s.store.Put(key, r, size, mime); err != nil {
		return entry, err
	}
	entry.ImageStorageKey = key
	entry.ImageMime = mime
	entry.Status = models.ExpenseDayProcessing
	entry.StatusMessage = ""
	if err := s.db.Save(&entry).Error; err != nil {
		return entry, err
	}
	if s.enqueue != nil {
		_ = s.enqueue.OCRExpenseDay(entry.ID)
	}
	return entry, nil
}

type LineInput struct {
	Name       string
	Amount     decimal.Decimal
	CategoryID *uuid.UUID
	SortOrder  int
	Confidence *float64
}

func (s *ExpenseDayService) UpdateLines(userID, id uuid.UUID, lines []LineInput) (models.ExpenseDayEntry, error) {
	entry, err := s.Get(userID, id)
	if err != nil {
		return entry, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("entry_id = ?", id).Delete(&models.ExpenseDayLineItem{}).Error; err != nil {
			return err
		}
		total := decimal.Zero
		for i, l := range lines {
			line := models.ExpenseDayLineItem{
				EntryID:    id,
				Name:       l.Name,
				Amount:     l.Amount,
				CategoryID: l.CategoryID,
				SortOrder:  l.SortOrder,
				Confidence: l.Confidence,
			}
			if line.SortOrder == 0 {
				line.SortOrder = i
			}
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
			total = total.Add(l.Amount)
		}
		status := entry.Status
		if status == models.ExpenseDayReady || status == models.ExpenseDayDraft {
			status = models.ExpenseDayReady
		}
		return tx.Model(&entry).Updates(map[string]any{
			"total_amount": total,
			"status":       status,
		}).Error
	})
	if err != nil {
		return entry, err
	}
	return s.Get(userID, id)
}

func (s *ExpenseDayService) Confirm(userID, id uuid.UUID, defaultCategoryID uuid.UUID) (models.ExpenseDayEntry, error) {
	entry, err := s.Get(userID, id)
	if err != nil {
		return entry, err
	}
	if len(entry.Lines) == 0 {
		return entry, errors.New("no line items")
	}
	occurredAt := entry.EntryDate.Add(12 * time.Hour)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for i := range entry.Lines {
			line := entry.Lines[i]
			if line.TransactionID != nil {
				continue
			}
			catID := defaultCategoryID
			if line.CategoryID != nil {
				catID = *line.CategoryID
			}
			txn := models.Transaction{
				UserID:     userID,
				CategoryID: catID,
				Amount:     line.Amount,
				Type:       models.TransactionExpense,
				Note:       line.Name,
				OccurredAt: occurredAt,
			}
			if err := tx.Create(&txn).Error; err != nil {
				return err
			}
			line.TransactionID = &txn.ID
			if err := tx.Save(&line).Error; err != nil {
				return err
			}
			if s.txSvc != nil {
				s.txSvc.ScheduleJobs(userID, occurredAt)
			}
		}
		entry.Status = models.ExpenseDayConfirmed
		return tx.Model(&entry).Update("status", models.ExpenseDayConfirmed).Error
	})
	if err != nil {
		return entry, err
	}
	return s.Get(userID, id)
}

func (s *ExpenseDayService) RunOCR(entryID uuid.UUID) error {
	var entry models.ExpenseDayEntry
	if err := s.db.First(&entry, "id = ?", entryID).Error; err != nil {
		return err
	}
	if entry.ImageStorageKey == "" {
		return errors.New("no image")
	}
	rc, _, err := s.store.Get(entry.ImageStorageKey)
	if err != nil {
		return err
	}
	defer rc.Close()
	result, err := s.ocr.ParseReceipt(rc, entry.ImageMime)
	if err != nil {
		entry.Status = models.ExpenseDayFailed
		entry.StatusMessage = err.Error()
		_ = s.db.Save(&entry).Error
		return err
	}
	raw, _ := json.Marshal(result)
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("entry_id = ?", entryID).Delete(&models.ExpenseDayLineItem{}).Error; err != nil {
			return err
		}
		total := decimal.Zero
		for i, item := range result.Items {
			amt, err := decimal.NewFromString(item.Amount)
			if err != nil {
				continue
			}
			conf := item.Confidence
			line := models.ExpenseDayLineItem{
				EntryID:    entryID,
				Name:       item.Name,
				Amount:     amt,
				SortOrder:  i,
				Confidence: &conf,
			}
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
			total = total.Add(amt)
		}
		entry.OCRRawJSON = raw
		entry.TotalAmount = total
		entry.Status = models.ExpenseDayReady
		entry.StatusMessage = ""
		return tx.Save(&entry).Error
	})
}

func (s *ExpenseDayService) MediaReader(userID uuid.UUID, key string) (io.ReadCloser, string, error) {
	prefix := fmt.Sprintf("users/%s/", userID.String())
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		return nil, "", errors.New("forbidden")
	}
	return s.store.Get(key)
}
