package handlers

import (
	"context"
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
		`SELECT id, user_id, name, institution, account_type, last_synced_at, created_at, updated_at
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
			&a.LastSyncedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
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
		`INSERT INTO accounts (user_id, name, institution, account_type)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, req.Name, req.Institution, req.AccountType,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
