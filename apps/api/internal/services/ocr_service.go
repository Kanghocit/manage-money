package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manage-money/api/internal/config"
)

type OCRService struct {
	cfg config.Config
}

func NewOCRService(cfg config.Config) *OCRService {
	return &OCRService{cfg: cfg}
}

type OCRLineItem struct {
	Name       string  `json:"name"`
	Amount     string  `json:"amount"`
	Confidence float64 `json:"confidence"`
}

type OCRResult struct {
	Items []OCRLineItem `json:"items"`
}

func (s *OCRService) ParseReceipt(r io.Reader, mime string) (OCRResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return OCRResult{}, err
	}
	if s.cfg.OpenAIAPIKey == "" {
		return mockOCRResult(), nil
	}
	return s.parseOpenAI(data, mime)
}

func mockOCRResult() OCRResult {
	return OCRResult{Items: []OCRLineItem{
		{Name: "Mục mẫu (cấu hình OPENAI_API_KEY để OCR thật)", Amount: "50000", Confidence: 0.5},
	}}
}

func (s *OCRService) parseOpenAI(image []byte, mime string) (OCRResult, error) {
	if mime == "" {
		mime = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(image)
	payload := map[string]any{
		"model": s.cfg.OpenAIModel,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": "Extract expense line items from this receipt image for a Vietnamese user. Return JSON: {\"items\":[{\"name\":\"...\",\"amount\":\"12345\",\"confidence\":0.9}]}. Amounts in VND as integer strings without currency symbol."},
					{"type": "image_url", "image_url": map[string]string{"url": fmt.Sprintf("data:%s;base64,%s", mime, b64)}},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return OCRResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.OpenAIAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OCRResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return OCRResult{}, fmt.Errorf("openai: %s", string(b))
	}
	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return OCRResult{}, err
	}
	if len(envelope.Choices) == 0 {
		return OCRResult{}, errors.New("empty openai response")
	}
	content := strings.TrimSpace(envelope.Choices[0].Message.Content)
	var result OCRResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return OCRResult{}, err
	}
	return result, nil
}
