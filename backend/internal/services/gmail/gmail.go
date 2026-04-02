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

		emailDate := extractEmailDate(full)
		result := s.processEmail(ctx, body, msg.Id, emailDate)
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

func (s *Service) SyncDateRange(ctx context.Context, startDate, endDate string) (imported, skipped, failed int) {
	log.Printf("syncing gmail for bank emails from %s to %s...", startDate, endDate)

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Printf("invalid start_date: %v", err)
		return
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Printf("invalid end_date: %v", err)
		return
	}
	end = end.AddDate(0, 0, 1)

	query := fmt.Sprintf("%s after:%d before:%d", s.query, start.Unix(), end.Unix())
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

		emailDate := extractEmailDate(full)
		result := s.processEmail(ctx, body, msg.Id, emailDate)
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

		emailDate := extractEmailDate(full)
		s.processEmail(ctx, body, msg.Id, emailDate)
	}
}

// sanitizeDate checks if the LLM-parsed date is reasonable compared to the
// email's received date. If the parsed date is more than 3 days away from the
// email date, the email date is used instead (the LLM likely hallucinated).
func sanitizeDate(parsedDate, emailDate string) string {
	if emailDate == "" {
		return parsedDate
	}
	pDate, pErr := time.Parse("2006-01-02", parsedDate)
	eDate, eErr := time.Parse("2006-01-02", emailDate)
	if pErr != nil || eErr != nil {
		if pErr != nil {
			return emailDate
		}
		return parsedDate
	}
	diff := pDate.Sub(eDate)
	if diff < 0 {
		diff = -diff
	}
	if diff > 3*24*time.Hour {
		log.Printf("date sanity check: LLM date %s is %d days from email date %s, using email date",
			parsedDate, int(diff.Hours()/24), emailDate)
		return emailDate
	}
	return parsedDate
}

func extractEmailDate(msg *gmail.Message) string {
	if msg.InternalDate > 0 {
		t := time.Unix(msg.InternalDate/1000, 0)
		return t.Format("2006-01-02")
	}
	return time.Now().Format("2006-01-02")
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
	if parsed.AccountLastFour != "" {
		var id, acctType string
		err := s.pool.QueryRow(ctx,
			"SELECT id, account_type FROM accounts WHERE user_id = $1 AND (last_four = $2 OR debit_card_last_four = $2)",
			s.userID, parsed.AccountLastFour,
		).Scan(&id, &acctType)
		if err == nil {
			switch acctType {
			case "credit_card":
				parsed.PaymentMethod = "credit_card"
			case "checking", "savings":
				if parsed.PaymentMethod == "" || parsed.PaymentMethod == "other" {
					parsed.PaymentMethod = "debit_card"
				}
			}
			return &id
		}
	}

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

// verifyTransferCategory checks whether a transaction tagged as "Transfer" by
// the LLM is actually an internal transfer between the user's own accounts at
// the same bank. It looks up from_account_last_four in the user's accounts and
// checks that both accounts share the same institution. If the from-account is
// not found or is at a different institution, the category is changed to
// "Income" (for credits) or "Other" (for debits).
func (s *Service) verifyTransferCategory(ctx context.Context, parsed *parser.ParsedTransaction, destAccountID *string) string {
	fromFour := parsed.FromAccountLastFour
	if fromFour == "" {
		fromFour = parsed.AccountLastFour
	}

	if fromFour == "" {
		log.Printf("transfer verification: no from-account last4 found, reclassifying")
		return s.fallbackCategory(parsed.Type)
	}

	var fromInstitution string
	err := s.pool.QueryRow(ctx,
		"SELECT institution FROM accounts WHERE user_id = $1 AND last_four = $2",
		s.userID, fromFour,
	).Scan(&fromInstitution)
	if err != nil {
		log.Printf("transfer verification: from-account %s not found in user's accounts, reclassifying as %s",
			fromFour, s.fallbackCategory(parsed.Type))
		return s.fallbackCategory(parsed.Type)
	}

	if destAccountID != nil {
		var destInstitution string
		err = s.pool.QueryRow(ctx,
			"SELECT institution FROM accounts WHERE id = $1",
			*destAccountID,
		).Scan(&destInstitution)
		if err == nil && !strings.EqualFold(fromInstitution, destInstitution) {
			log.Printf("transfer verification: from-account %s (%s) and dest-account (%s) are at different banks, reclassifying",
				fromFour, fromInstitution, destInstitution)
			return s.fallbackCategory(parsed.Type)
		}
	}

	log.Printf("transfer verification: confirmed internal transfer (from-account %s at %s)", fromFour, fromInstitution)
	return "Transfer"
}

var expenseCategories = map[string]bool{
	"groceries":        true,
	"dining":           true,
	"gas":              true,
	"transportation":   true,
	"shopping":         true,
	"bills & utilities": true,
	"rent & mortgage":  true,
	"healthcare":       true,
	"entertainment":    true,
	"subscriptions":    true,
	"atm":             true,
	"fees":            true,
}

func isExpenseCategory(category string) bool {
	return expenseCategories[strings.ToLower(category)]
}

func sanitizeTxnType(parsed *parser.ParsedTransaction) {
	cat := strings.ToLower(parsed.Category)
	desc := strings.ToLower(parsed.Description)
	method := strings.ToLower(parsed.PaymentMethod)

	// credit + expense category → always debit
	if strings.EqualFold(parsed.Type, "credit") && isExpenseCategory(parsed.Category) {
		log.Printf("type sanity check: overriding 'credit' → 'debit' (category %q is an expense)", parsed.Category)
		parsed.Type = "debit"
		return
	}

	// debit + income category → always credit
	if strings.EqualFold(parsed.Type, "debit") && cat == "income" {
		log.Printf("type sanity check: overriding 'debit' → 'credit' (category is Income)")
		parsed.Type = "credit"
		return
	}

	// card payment is always a debit (unless description says refund)
	if strings.EqualFold(parsed.Type, "credit") &&
		(method == "credit_card" || method == "debit_card") &&
		!strings.Contains(desc, "refund") {
		log.Printf("type sanity check: overriding 'credit' → 'debit' (payment method %q implies spending)", parsed.PaymentMethod)
		parsed.Type = "debit"
		return
	}

	// named merchant + credit + no sign of refund/deposit/income → debit
	if strings.EqualFold(parsed.Type, "credit") &&
		parsed.Merchant != "" &&
		cat != "income" && cat != "transfer" &&
		!strings.Contains(desc, "refund") &&
		!strings.Contains(desc, "deposit") &&
		!strings.Contains(desc, "received") {
		log.Printf("type sanity check: overriding 'credit' → 'debit' (merchant %q with no refund/deposit signal)", parsed.Merchant)
		parsed.Type = "debit"
		return
	}
}

func (s *Service) fallbackCategory(txnType string) string {
	if strings.EqualFold(txnType, "credit") {
		return "Income"
	}
	return "Other"
}

func (s *Service) processEmail(ctx context.Context, body, msgID, emailDate string) string {
	parsed, err := s.parserSvc.ParseWithDate(ctx, body, emailDate)
	if err != nil {
		s.logParseFailure(msgID, err, body)
		return "failed"
	}

	parsed.Date = sanitizeDate(parsed.Date, emailDate)

	if parsed.Amount == 0 {
		log.Printf("skipping non-transactional email [msg=%s]: no amount found (description: %s)", msgID, parsed.Description)
		return "skipped"
	}

	sanitizeTxnType(parsed)

	accountID := s.resolveAccountID(ctx, parsed)

	accountLastFour := ""
	if accountID != nil {
		var inactiveDate *time.Time
		_ = s.pool.QueryRow(ctx,
			"SELECT COALESCE(last_four, ''), inactive_date FROM accounts WHERE id = $1",
			*accountID,
		).Scan(&accountLastFour, &inactiveDate)

		if inactiveDate != nil {
			txnDate, err := time.Parse("2006-01-02", parsed.Date)
			if err == nil && !txnDate.Before(inactiveDate.Truncate(24*time.Hour)) {
				log.Printf("SKIPPED inactive account txn: account_id=%s inactive_date=%s | amount=%.2f merchant=%s date=%s type=%s category=%s method=%s confidence=%.0f%% description=%s account_last4=%s from_account_last4=%s email_msg_id=%s",
					*accountID, inactiveDate.Format("2006-01-02"),
					parsed.Amount, parsed.Merchant, parsed.Date, parsed.Type, parsed.Category,
					parsed.PaymentMethod, parsed.Confidence*100, parsed.Description,
					parsed.AccountLastFour, parsed.FromAccountLastFour, msgID)
				return "skipped"
			}
		}
	}

	hash := dedup.GenerateHash(parsed.Amount, emailDate, parsed.Merchant, parsed.Type, accountLastFour)
	exists, err := s.dedupSvc.Exists(ctx, hash)
	if err != nil {
		log.Printf("dedup check error: %v", err)
		return "failed"
	}
	if exists {
		log.Printf("duplicate transaction skipped (hash: %s)", hash[:12])
		return "skipped"
	}

	if strings.EqualFold(parsed.Category, "Transfer") {
		parsed.Category = s.verifyTransferCategory(ctx, parsed, accountID)
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
