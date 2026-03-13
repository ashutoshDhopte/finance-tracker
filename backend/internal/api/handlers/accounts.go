package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/models"
)

type AccountHandler struct {
	pool *pgxpool.Pool
}

func NewAccountHandler(pool *pgxpool.Pool) *AccountHandler {
	return &AccountHandler{pool: pool}
}

func (h *AccountHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.pool.Query(context.Background(),
		`SELECT id, user_id, name, institution, account_type, last_four, last_synced_at, created_at, updated_at
		 FROM accounts WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query accounts"})
		return
	}
	defer rows.Close()

	accs := []models.Account{}
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Institution, &a.AccountType,
			&a.LastFour, &a.LastSyncedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan account"})
			return
		}
		accs = append(accs, a)
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accs})
}

func (h *AccountHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	err := h.pool.QueryRow(context.Background(),
		`INSERT INTO accounts (user_id, name, institution, account_type, last_four)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, req.Name, req.Institution, req.AccountType, req.LastFour,
	).Scan(&id)
	if err != nil {
		log.Printf("create account error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *AccountHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.pool.Exec(context.Background(), `
		UPDATE accounts SET
			name = COALESCE($3, name),
			institution = COALESCE($4, institution),
			account_type = COALESCE($5, account_type),
			last_four = CASE WHEN $6::boolean THEN $7 ELSE last_four END,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		id, userID, req.Name, req.Institution, req.AccountType,
		req.LastFour != nil, req.LastFour,
	)
	if err != nil {
		log.Printf("update account error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account: " + err.Error()})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AccountHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	tag, err := h.pool.Exec(context.Background(),
		"DELETE FROM accounts WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account: " + err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
