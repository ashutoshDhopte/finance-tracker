package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/ash/finance-tracker/backend/internal/services/dedup"
	"github.com/ash/finance-tracker/backend/internal/services/parser"
)

type Service struct {
	pool         *pgxpool.Pool
	parserSvc    *parser.Service
	dedupSvc     *dedup.Service
	gmailSvc     *gmail.Service
	query        string
	pollInterval time.Duration
	userID       string
}

func NewService(pool *pgxpool.Pool, parserSvc *parser.Service, dedupSvc *dedup.Service,
	credFile, tokenFile, query string, pollInterval time.Duration) (*Service, error) {

	gmailSvc, err := buildGmailService(credFile, tokenFile)
	if err != nil {
		return nil, fmt.Errorf("build gmail service: %w", err)
	}

	var userID string
	err = pool.QueryRow(context.Background(), "SELECT id FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("get user id: %w", err)
	}

	return &Service{
		pool:         pool,
		parserSvc:    parserSvc,
		dedupSvc:     dedupSvc,
		gmailSvc:     gmailSvc,
		query:        query,
		pollInterval: pollInterval,
		userID:       userID,
	}, nil
}

func buildGmailService(credFile, tokenFile string) (*gmail.Service, error) {
	credBytes, err := os.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(credBytes, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	tokBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("read token file (run auth flow first): %w", err)
	}

	tok := &oauth2.Token{}
	if err := parseJSON(tokBytes, tok); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	client := config.Client(context.Background(), tok)
	svc, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}

	return svc, nil
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *Service) Start(ctx context.Context) {
	log.Printf("gmail poller started (interval: %s, query: %s)", s.pollInterval, s.query)

	// Run immediately on start
	s.poll(ctx)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("gmail poller stopped")
			return
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

func (s *Service) SyncDays(ctx context.Context, days int) (imported, skipped, failed int) {
	log.Printf("syncing gmail for bank emails from last %d days...", days)

	query := fmt.Sprintf("%s newer_than:%dd", s.query, days)
	msgs, err := s.gmailSvc.Users.Messages.List("me").Q(query).MaxResults(500).Do()
	if err != nil {
		log.Printf("gmail list error: %v", err)
		return
	}

	log.Printf("found %d potential bank emails", len(msgs.Messages))

	for _, msg := range msgs.Messages {
		full, err := s.gmailSvc.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("gmail get message error: %v", err)
			failed++
			continue
		}

		body := extractBody(full)
		if body == "" {
			continue
		}

		result := s.processEmail(ctx, body, msg.Id)
		switch result {
		case "imported":
			imported++
		case "skipped":
			skipped++
		default:
			failed++
		}
	}

	log.Printf("sync complete: %d imported, %d skipped, %d failed", imported, skipped, failed)
	return
}

func (s *Service) poll(ctx context.Context) {
	log.Println("polling gmail for bank emails...")

	query := s.query + " newer_than:1h"
	msgs, err := s.gmailSvc.Users.Messages.List("me").Q(query).MaxResults(20).Do()
	if err != nil {
		log.Printf("gmail list error: %v", err)
		return
	}

	if len(msgs.Messages) == 0 {
		log.Println("no new bank emails found")
		return
	}

	log.Printf("found %d potential bank emails", len(msgs.Messages))

	for _, msg := range msgs.Messages {
		full, err := s.gmailSvc.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("gmail get message error: %v", err)
			continue
		}

		body := extractBody(full)
		if body == "" {
			continue
		}

		s.processEmail(ctx, body, msg.Id)
	}
}

func (s *Service) processEmail(ctx context.Context, body, msgID string) string {
	parsed, err := s.parserSvc.Parse(ctx, body)
	if err != nil {
		log.Printf("parse error for message %s: %v", msgID, err)
		return "failed"
	}

	hash := dedup.GenerateHash(parsed.Amount, parsed.Date, parsed.Merchant)
	exists, err := s.dedupSvc.Exists(ctx, hash)
	if err != nil {
		log.Printf("dedup check error: %v", err)
		return "failed"
	}
	if exists {
		log.Printf("duplicate transaction skipped (hash: %s)", hash[:12])
		return "skipped"
	}

	var categoryID *string
	err = s.pool.QueryRow(ctx,
		"SELECT id FROM categories WHERE LOWER(name) = LOWER($1)",
		parsed.Category,
	).Scan(&categoryID)
	if err != nil {
		log.Printf("category lookup failed for %q, using nil", parsed.Category)
		categoryID = nil
	}

	confidence := float32(parsed.Confidence)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO transactions (user_id, amount, currency, merchant_name, merchant_raw,
		                         category_id, transaction_date, txn_type, source,
		                         source_hash, ai_confidence, raw_text)
		VALUES ($1, $2, 'USD', $3, $4, $5, $6, $7, 'email', $8, $9, $10)`,
		s.userID, parsed.Amount, parsed.Merchant, parsed.Merchant,
		categoryID, parsed.Date, parsed.Type, hash, confidence, body,
	)
	if err != nil {
		log.Printf("insert transaction error: %v", err)
		return "failed"
	}

	log.Printf("saved transaction: $%.2f at %s on %s (confidence: %.0f%%)",
		parsed.Amount, parsed.Merchant, parsed.Date, parsed.Confidence*100)
	return "imported"
}

func extractBody(msg *gmail.Message) string {
	if msg.Payload == nil {
		return ""
	}

	// Try to get plain text from parts
	if parts := msg.Payload.Parts; len(parts) > 0 {
		for _, part := range parts {
			if part.MimeType == "text/plain" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
		// Fallback to first part
		if parts[0].Body != nil && parts[0].Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(parts[0].Body.Data)
			if err == nil {
				return string(data)
			}
		}
	}

	// Try body directly
	if msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	return ""
}
