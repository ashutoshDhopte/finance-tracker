package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/models"
)

type AlertHandler struct {
	pool *pgxpool.Pool
}

func NewAlertHandler(pool *pgxpool.Pool) *AlertHandler {
	return &AlertHandler{pool: pool}
}

func (h *AlertHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.pool.Query(context.Background(), `
		SELECT a.id, a.user_id, a.name, a.category_id, c.name,
		       a.threshold, a.period, a.enabled, a.last_triggered_at,
		       a.created_at, a.updated_at
		FROM alerts a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.user_id = $1
		ORDER BY a.created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query alerts"})
		return
	}
	defer rows.Close()

	alerts := []models.Alert{}
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.CategoryID, &a.CategoryName,
			&a.Threshold, &a.Period, &a.Enabled, &a.LastTriggeredAt,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan alert"})
			return
		}
		alerts = append(alerts, a)
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (h *AlertHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	err := h.pool.QueryRow(context.Background(), `
		INSERT INTO alerts (user_id, name, category_id, threshold, period)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, req.Name, req.CategoryID, req.Threshold, req.Period,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *AlertHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req struct {
		Name      *string  `json:"name"`
		Threshold *float64 `json:"threshold"`
		Period    *string  `json:"period"`
		Enabled   *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.pool.Exec(context.Background(), `
		UPDATE alerts SET
			name = COALESCE($1, name),
			threshold = COALESCE($2, threshold),
			period = COALESCE($3, period),
			enabled = COALESCE($4, enabled),
			updated_at = $5
		WHERE id = $6 AND user_id = $7`,
		req.Name, req.Threshold, req.Period, req.Enabled, time.Now(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AlertHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	tag, err := h.pool.Exec(context.Background(),
		"DELETE FROM alerts WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete alert"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AlertHandler) Check(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.pool.Query(context.Background(), `
		SELECT a.id, a.user_id, a.name, a.category_id, c.name,
		       a.threshold, a.period, a.enabled, a.last_triggered_at,
		       a.created_at, a.updated_at
		FROM alerts a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.user_id = $1 AND a.enabled = true`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query alerts"})
		return
	}
	defer rows.Close()

	triggered := []models.TriggeredAlert{}
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.CategoryID, &a.CategoryName,
			&a.Threshold, &a.Period, &a.Enabled, &a.LastTriggeredAt,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}

		start := periodStart(a.Period)
		var spent float64
		q := `SELECT COALESCE(SUM(amount), 0) FROM transactions
		      WHERE user_id = $1 AND txn_type = 'debit' AND transaction_date >= $2`
		args := []interface{}{userID, start.Format("2006-01-02")}

		if a.CategoryID != nil {
			q += " AND category_id = $3"
			args = append(args, *a.CategoryID)
		}

		if err := h.pool.QueryRow(context.Background(), q, args...).Scan(&spent); err != nil {
			continue
		}

		if spent >= a.Threshold {
			triggered = append(triggered, models.TriggeredAlert{
				Alert:        a,
				CurrentSpend: spent,
				Exceeded:     true,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"triggered": triggered, "count": len(triggered)})
}

func periodStart(period string) time.Time {
	now := time.Now()
	switch period {
	case "daily":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "weekly":
		weekday := int(now.Weekday())
		return now.AddDate(0, 0, -weekday)
	case "biweekly":
		weekday := int(now.Weekday())
		start := now.AddDate(0, 0, -weekday)
		if now.Day() > 14 {
			return time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, now.Location())
		}
		return time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	case "monthly":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
}
