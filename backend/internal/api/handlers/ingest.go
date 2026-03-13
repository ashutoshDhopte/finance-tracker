package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/services/dedup"
	"github.com/ash/finance-tracker/backend/internal/services/parser"
)

type IngestHandler struct {
	pool      *pgxpool.Pool
	parserSvc *parser.Service
}

func NewIngestHandler(pool *pgxpool.Pool, parserSvc *parser.Service) *IngestHandler {
	return &IngestHandler{pool: pool, parserSvc: parserSvc}
}

func (h *IngestHandler) ImportCSV(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	bankType := c.DefaultPostForm("bank", "auto")
	accountID := c.PostForm("account_id")
	var accountIDPtr *string
	if accountID != "" {
		accountIDPtr = &accountID
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CSV file"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is empty or has no data rows"})
		return
	}

	if bankType == "auto" {
		bankType = detectBank(records[0])
	}
	log.Printf("CSV import: %d rows, bank=%s, headers=%v", len(records)-1, bankType, records[0])

	var imported, skipped, failed int
	dedupSvc := dedup.NewService(h.pool)

	for i, row := range records[1:] {
		txn, err := parseCSVRow(row, bankType)
		if err != nil {
			log.Printf("CSV row %d parse error: %v (row: %v)", i+1, err, row)
			failed++
			continue
		}

		hash := dedup.GenerateHash(txn.amount, txn.date, txn.merchant)
		exists, _ := dedupSvc.Exists(context.Background(), hash)
		if exists {
			skipped++
			continue
		}

		var categoryID *string
		if txn.category != "" {
			err = h.pool.QueryRow(context.Background(),
				"SELECT id FROM categories WHERE LOWER(name) = LOWER($1)",
				txn.category,
			).Scan(&categoryID)
			if err != nil {
				categoryID = nil
			}
		}
		if categoryID == nil && h.parserSvc != nil {
			desc := fmt.Sprintf("%s $%.2f on %s", txn.merchant, txn.amount, txn.date)
			parsed, parseErr := h.parserSvc.Parse(context.Background(), desc)
			if parseErr != nil {
				log.Printf("CSV row %d LLM categorization failed: %v", i+1, parseErr)
			} else {
				err = h.pool.QueryRow(context.Background(),
					"SELECT id FROM categories WHERE LOWER(name) = LOWER($1)",
					parsed.Category,
				).Scan(&categoryID)
				if err != nil {
					categoryID = nil
				}
			}
		}

		_, err = h.pool.Exec(context.Background(), `
			INSERT INTO transactions (user_id, account_id, amount, currency, merchant_name,
			                         category_id, transaction_date, txn_type,
			                         source, source_hash)
			VALUES ($1, $2, $3, 'USD', $4, $5, $6, $7, 'csv', $8)`,
			userID, accountIDPtr, txn.amount, txn.merchant, categoryID,
			txn.date, txn.txnType, hash,
		)
		if err != nil {
			log.Printf("CSV row %d insert error: %v", i+1, err)
			failed++
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "import complete",
		"filename": header.Filename,
		"imported": imported,
		"skipped":  skipped,
		"failed":   failed,
	})
}

type csvTransaction struct {
	amount   float64
	merchant string
	date     string
	txnType  string
	category string
}

func detectBank(headers []string) string {
	joined := strings.ToLower(strings.Join(headers, ","))
	if strings.Contains(joined, "post date") || strings.Contains(joined, "posting date") {
		return "chase"
	}
	if strings.Contains(joined, "withdrawal") || strings.Contains(joined, "deposit") {
		return "pnc"
	}
	return "generic"
}

func parseCSVRow(row []string, bankType string) (*csvTransaction, error) {
	switch bankType {
	case "chase":
		return parseChaseRow(row)
	case "pnc":
		return parsePNCRow(row)
	default:
		return parseGenericRow(row)
	}
}

// Chase CSV: Transaction Date, Post Date, Description, Category, Type, Amount, Memo
func parseChaseRow(row []string) (*csvTransaction, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("not enough columns")
	}

	date, err := parseDate(row[0])
	if err != nil {
		return nil, err
	}

	amountStr := strings.TrimSpace(row[5])
	amountStr = strings.ReplaceAll(amountStr, "$", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, err
	}

	txnType := "debit"
	if amount > 0 {
		txnType = "credit"
	}
	if amount < 0 {
		amount = -amount
	}

	cat := ""
	if len(row) >= 4 {
		cat = strings.TrimSpace(row[3])
	}

	return &csvTransaction{
		amount:   amount,
		merchant: strings.TrimSpace(row[2]),
		date:     date,
		txnType:  txnType,
		category: cat,
	}, nil
}

// PNC CSV: Date, Description, Withdrawals, Deposits, Balance
func parsePNCRow(row []string) (*csvTransaction, error) {
	if len(row) < 4 {
		return nil, fmt.Errorf("not enough columns")
	}

	date, err := parseDate(row[0])
	if err != nil {
		return nil, err
	}

	txnType := "debit"
	amountStr := strings.TrimSpace(row[2])
	if amountStr == "" || amountStr == "-" {
		amountStr = strings.TrimSpace(row[3])
		txnType = "credit"
	}

	amountStr = strings.ReplaceAll(amountStr, "$", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, err
	}
	if amount < 0 {
		amount = -amount
	}

	return &csvTransaction{
		amount:   amount,
		merchant: strings.TrimSpace(row[1]),
		date:     date,
		txnType:  txnType,
	}, nil
}

func parseGenericRow(row []string) (*csvTransaction, error) {
	if len(row) < 3 {
		return nil, fmt.Errorf("not enough columns")
	}

	date, err := parseDate(row[0])
	if err != nil {
		return nil, err
	}

	amountStr := strings.ReplaceAll(row[2], "$", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil {
		return nil, err
	}

	txnType := "debit"
	if amount > 0 {
		txnType = "credit"
	}
	if amount < 0 {
		amount = -amount
	}

	cat := ""
	if len(row) >= 4 {
		cat = strings.TrimSpace(row[3])
	}

	return &csvTransaction{
		amount:   amount,
		merchant: strings.TrimSpace(row[1]),
		date:     date,
		txnType:  txnType,
		category: cat,
	}, nil
}

func parseDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"01/02/2006",
		"1/2/2006",
		"2006-01-02",
		"01-02-2006",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("unparseable date: %s", s)
}
