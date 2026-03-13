package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	ollamaURL string
	model     string
	client    *http.Client
}

type ParsedTransaction struct {
	Amount       float64 `json:"amount"`
	Merchant     string  `json:"merchant"`
	Date         string  `json:"date"`
	Type         string  `json:"transaction_type"`
	Category     string  `json:"suggested_category"`
	Confidence   float64 `json:"confidence"`
	Description  string  `json:"description,omitempty"`
}

func NewService(ollamaURL string) *Service {
	return &Service{
		ollamaURL: strings.TrimRight(ollamaURL, "/"),
		model:     "llama3.2:3b",
		client:    &http.Client{Timeout: 120 * time.Second},
	}
}

const systemPrompt = `You are a financial transaction parser. Given a bank email or SMS notification, extract the transaction details.

Respond ONLY with a valid JSON object (no markdown, no explanation) containing these fields:
- "amount": number (positive value, e.g. 47.23)
- "merchant": string (cleaned-up merchant name, e.g. "Trader Joe's" not "TRADER JOE'S #123")
- "date": string in YYYY-MM-DD format
- "transaction_type": "credit" or "debit"
- "suggested_category": one of: Groceries, Dining, Gas, Transportation, Shopping, Bills & Utilities, Rent & Mortgage, Healthcare, Entertainment, Subscriptions, Income, Transfer, ATM, Fees, Other
- "confidence": number between 0 and 1 indicating how confident you are in the parsing
- "description": brief one-line description of the transaction

If you cannot determine a field, use reasonable defaults. For amount, never return 0. For date, use today's date if unclear.`

func (s *Service) Parse(ctx context.Context, rawText string) (*ParsedTransaction, error) {
	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "Parse this bank notification:\n\n" + rawText},
		},
		"stream": false,
		"format": "json",
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ollamaURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode ollama response: %w", err)
	}

	var parsed ParsedTransaction
	if err := json.Unmarshal([]byte(ollamaResp.Message.Content), &parsed); err != nil {
		return nil, fmt.Errorf("parse LLM output as JSON: %w (raw: %s)", err, ollamaResp.Message.Content)
	}

	if parsed.Confidence == 0 {
		parsed.Confidence = 0.5
	}
	if parsed.Category == "" {
		parsed.Category = "Other"
	}
	if parsed.Type == "" {
		parsed.Type = "debit"
	}
	if parsed.Date == "" {
		parsed.Date = time.Now().Format("2006-01-02")
	}

	return &parsed, nil
}
