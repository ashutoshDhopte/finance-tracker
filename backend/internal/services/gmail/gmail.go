package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/html"
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
	failLog      *os.File
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

	failLog, err := os.OpenFile("parse_failures.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open parse_failures.log: %w", err)
	}

	return &Service{
		pool:         pool,
		parserSvc:    parserSvc,
		dedupSvc:     dedupSvc,
		gmailSvc:     gmailSvc,
		query:        query,
		pollInterval: pollInterval,
		userID:       userID,
		failLog:      failLog,
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func (s *Service) logParseFailure(msgID string, err error, body string) {
	log.Printf("PARSE FAILURE [msg=%s]: %v", msgID, err)
	entry := fmt.Sprintf(
		"=== PARSE FAILURE ===\nTime: %s\nMessage ID: %s\nError: %v\n\n--- EMAIL BODY ---\n%s\n--- END ---\n\n",
		time.Now().Format(time.RFC3339), msgID, err, body,
	)
	s.failLog.WriteString(entry)
}

func (s *Service) resolveAccountID(ctx context.Context, parsed *parser.ParsedTransaction) *string {
	method := strings.ToLower(parsed.PaymentMethod)

	if method == "zelle" || method == "debit_card" {
		var id string
		err := s.pool.QueryRow(ctx,
			"SELECT id FROM accounts WHERE user_id = $1 AND account_type = 'checking' LIMIT 1",
			s.userID,
		).Scan(&id)
		if err == nil {
			return &id
		}
		return nil
	}

	if parsed.AccountLastFour != "" {
		var id, accType string
		err := s.pool.QueryRow(ctx,
			"SELECT id, account_type FROM accounts WHERE user_id = $1 AND last_four = $2",
			s.userID, parsed.AccountLastFour,
		).Scan(&id, &accType)
		if err == nil {
			return &id
		}
	}

	var id string
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM accounts WHERE user_id = $1 AND account_type = 'checking' LIMIT 1",
		s.userID,
	).Scan(&id)
	if err == nil {
		return &id
	}
	return nil
}

func (s *Service) processEmail(ctx context.Context, body, msgID string) string {
	parsed, err := s.parserSvc.Parse(ctx, body)
	if err != nil {
		s.logParseFailure(msgID, err, body)
		return "failed"
	}

	if parsed.Amount == 0 {
		log.Printf("skipping non-transactional email [msg=%s]: no amount found (description: %s)", msgID, parsed.Description)
		return "skipped"
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

	accountID := s.resolveAccountID(ctx, parsed)

	confidence := float32(parsed.Confidence)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO transactions (user_id, account_id, amount, currency, merchant_name, merchant_raw,
		                         category_id, transaction_date, txn_type, source,
		                         source_hash, ai_confidence, raw_text)
		VALUES ($1, $2, $3, 'USD', $4, $5, $6, $7, $8, 'email', $9, $10, $11)`,
		s.userID, accountID, parsed.Amount, parsed.Merchant, parsed.Merchant,
		categoryID, parsed.Date, parsed.Type, hash, confidence, body,
	)
	if err != nil {
		log.Printf("insert transaction error: %v", err)
		return "failed"
	}

	log.Printf("saved transaction: $%.2f at %s on %s (method: %s, account_last4: %s, confidence: %.0f%%)",
		parsed.Amount, parsed.Merchant, parsed.Date, parsed.PaymentMethod, parsed.AccountLastFour, parsed.Confidence*100)
	return "imported"
}

func extractBody(msg *gmail.Message) string {
	if msg.Payload == nil {
		return ""
	}

	plain, htmlBody := extractParts(msg.Payload)
	if plain != "" {
		return plain
	}
	if htmlBody != "" {
		return stripHTML(htmlBody)
	}
	return ""
}

func extractParts(part *gmail.MessagePart) (plain, htmlBody string) {
	if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(data), ""
		}
	}
	if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			htmlBody = string(data)
		}
	}
	for _, sub := range part.Parts {
		p, h := extractParts(sub)
		if p != "" {
			return p, ""
		}
		if h != "" && htmlBody == "" {
			htmlBody = h
		}
	}
	return "", htmlBody
}

func stripHTML(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script" || n.Data == "head") {
			return
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				b.WriteString(text)
				b.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(b.String())
}
