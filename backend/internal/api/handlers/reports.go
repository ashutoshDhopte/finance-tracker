package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/models"
)

type ReportHandler struct {
	pool *pgxpool.Pool
}

func NewReportHandler(pool *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{pool: pool}
}

func (h *ReportHandler) Monthly(c *gin.Context) {
	userID := c.GetString("user_id")
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))

	start, err := time.Parse("2006-01", month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
		return
	}
	end := start.AddDate(0, 1, 0)

	summary, err := h.buildSummary(userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *ReportHandler) Biweekly(c *gin.Context) {
	userID := c.GetString("user_id")
	startDate := c.Query("start")
	endDate := c.Query("end")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end dates required"})
		return
	}

	summary, err := h.buildSummary(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *ReportHandler) Categories(c *gin.Context) {
	userID := c.GetString("user_id")
	from := c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	rows, err := h.pool.Query(context.Background(), `
		SELECT c.id, c.name, c.color, c.icon,
		       COALESCE(SUM(t.amount), 0) as total,
		       COUNT(t.id) as count
		FROM categories c
		LEFT JOIN transactions t ON t.category_id = c.id
		     AND t.user_id = $1
		     AND t.transaction_date >= $2
		     AND t.transaction_date <= $3
		     AND t.txn_type = 'debit'
		GROUP BY c.id, c.name, c.color, c.icon
		HAVING COUNT(t.id) > 0
		ORDER BY total DESC`,
		userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query categories"})
		return
	}
	defer rows.Close()

	cats := []models.CategorySummary{}
	for rows.Next() {
		var cs models.CategorySummary
		if err := rows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.Color, &cs.Icon, &cs.Total, &cs.Count); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan category summary"})
			return
		}
		cats = append(cats, cs)
	}

	c.JSON(http.StatusOK, gin.H{"categories": cats, "from": from, "to": to})
}

func (h *ReportHandler) Trends(c *gin.Context) {
	userID := c.GetString("user_id")
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	if months < 1 || months > 24 {
		months = 6
	}

	start := time.Now().AddDate(0, -months, 0)
	startStr := fmt.Sprintf("%s-01", start.Format("2006-01"))

	rows, err := h.pool.Query(context.Background(), `
		SELECT TO_CHAR(transaction_date, 'YYYY-MM') as month,
		       COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END), 0) as income,
		       COALESCE(SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END), 0) as expenses
		FROM transactions
		WHERE user_id = $1 AND transaction_date >= $2
		GROUP BY month
		ORDER BY month`,
		userID, startStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query trends"})
		return
	}
	defer rows.Close()

	trends := []models.TrendPoint{}
	for rows.Next() {
		var tp models.TrendPoint
		if err := rows.Scan(&tp.Month, &tp.TotalIncome, &tp.TotalExpenses); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan trend"})
			return
		}
		tp.Net = tp.TotalIncome - tp.TotalExpenses
		trends = append(trends, tp)
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends, "months": months})
}

func (h *ReportHandler) buildSummary(userID, startDate, endDate string) (*models.ReportSummary, error) {
	var totalIncome, totalExpenses float64
	err := h.pool.QueryRow(context.Background(), `
		SELECT
			COALESCE(SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date < $3`,
		userID, startDate, endDate,
	).Scan(&totalIncome, &totalExpenses)
	if err != nil {
		return nil, err
	}

	rows, err := h.pool.Query(context.Background(), `
		SELECT c.id, c.name, c.color, c.icon,
		       COALESCE(SUM(t.amount), 0), COUNT(t.id)
		FROM categories c
		JOIN transactions t ON t.category_id = c.id
		WHERE t.user_id = $1 AND t.transaction_date >= $2 AND t.transaction_date < $3
		GROUP BY c.id, c.name, c.color, c.icon
		ORDER BY SUM(t.amount) DESC`,
		userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := []models.CategorySummary{}
	for rows.Next() {
		var cs models.CategorySummary
		if err := rows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.Color, &cs.Icon, &cs.Total, &cs.Count); err != nil {
			return nil, err
		}
		cats = append(cats, cs)
	}

	return &models.ReportSummary{
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		Net:           totalIncome - totalExpenses,
		ByCategory:    cats,
	}, nil
}
