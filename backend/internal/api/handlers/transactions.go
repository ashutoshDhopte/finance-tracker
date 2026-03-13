package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/models"
)

type TransactionHandler struct {
	pool *pgxpool.Pool
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{pool: pool}
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT t.id, t.user_id, t.account_id, t.amount, t.currency,
		       t.merchant_name, t.merchant_raw, t.category_id, c.name as category_name,
		       t.transaction_date, t.posted_date, t.txn_type, t.source,
		       t.source_hash, t.ai_confidence, t.raw_text, t.notes,
		       t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1`

	args := []interface{}{userID}
	argIdx := 2

	if v := c.Query("start_date"); v != "" {
		query += fmt.Sprintf(" AND t.transaction_date >= $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("end_date"); v != "" {
		query += fmt.Sprintf(" AND t.transaction_date <= $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("category_id"); v != "" {
		query += fmt.Sprintf(" AND t.category_id = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("account_id"); v != "" {
		query += fmt.Sprintf(" AND t.account_id = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("txn_type"); v != "" {
		query += fmt.Sprintf(" AND t.txn_type = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("source"); v != "" {
		query += fmt.Sprintf(" AND t.source = $%d", argIdx)
		args = append(args, v)
		argIdx++
	}
	if v := c.Query("search"); v != "" {
		query += fmt.Sprintf(" AND (t.merchant_name ILIKE $%d OR t.notes ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+v+"%")
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY t.transaction_date DESC, t.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.pool.Query(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query transactions"})
		return
	}
	defer rows.Close()

	txns := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.Amount, &t.Currency,
			&t.MerchantName, &t.MerchantRaw, &t.CategoryID, &t.CategoryName,
			&t.TransactionDate, &t.PostedDate, &t.TxnType, &t.Source,
			&t.SourceHash, &t.AIConfidence, &t.RawText, &t.Notes,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan transaction"})
			return
		}
		txns = append(txns, t)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": txns, "count": len(txns)})
}

func (h *TransactionHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var t models.Transaction
	err := h.pool.QueryRow(context.Background(), `
		SELECT t.id, t.user_id, t.account_id, t.amount, t.currency,
		       t.merchant_name, t.merchant_raw, t.category_id, c.name as category_name,
		       t.transaction_date, t.posted_date, t.txn_type, t.source,
		       t.source_hash, t.ai_confidence, t.raw_text, t.notes,
		       t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.id = $1 AND t.user_id = $2`,
		id, userID,
	).Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.Amount, &t.Currency,
		&t.MerchantName, &t.MerchantRaw, &t.CategoryID, &t.CategoryName,
		&t.TransactionDate, &t.PostedDate, &t.TxnType, &t.Source,
		&t.SourceHash, &t.AIConfidence, &t.RawText, &t.Notes,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, t)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Source == "" {
		req.Source = "manual"
	}

	var id string
	err := h.pool.QueryRow(context.Background(), `
		INSERT INTO transactions (user_id, account_id, amount, currency, merchant_name,
		                         category_id, transaction_date, posted_date, txn_type, source, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`,
		userID, req.AccountID, req.Amount, req.Currency, req.MerchantName,
		req.CategoryID, req.TransactionDate, req.PostedDate, req.TxnType, req.Source, req.Notes,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req models.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Amount != nil {
		setParts = append(setParts, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *req.Amount)
		argIdx++
	}
	if req.MerchantName != nil {
		setParts = append(setParts, fmt.Sprintf("merchant_name = $%d", argIdx))
		args = append(args, *req.MerchantName)
		argIdx++
	}
	if req.CategoryID != nil {
		setParts = append(setParts, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *req.CategoryID)
		argIdx++
	}
	if req.TransactionDate != nil {
		setParts = append(setParts, fmt.Sprintf("transaction_date = $%d", argIdx))
		args = append(args, *req.TransactionDate)
		argIdx++
	}
	if req.TxnType != nil {
		setParts = append(setParts, fmt.Sprintf("txn_type = $%d", argIdx))
		args = append(args, *req.TxnType)
		argIdx++
	}
	if req.AccountID != nil {
		setParts = append(setParts, fmt.Sprintf("account_id = $%d", argIdx))
		args = append(args, *req.AccountID)
		argIdx++
	}
	if req.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, id, userID)
	query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d AND user_id = $%d",
		strings.Join(setParts, ", "), argIdx, argIdx+1)

	tag, err := h.pool.Exec(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update transaction"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	tag, err := h.pool.Exec(context.Background(),
		"DELETE FROM transactions WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
