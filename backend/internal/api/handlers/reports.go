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
	month := c.Query("month")
	accountID := c.Query("account_id")

	var startDate, endDate string
	if month != "" {
		start, err := time.Parse("2006-01", month)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
			return
		}
		end := start.AddDate(0, 1, 0)
		startDate = start.Format("2006-01-02")
		endDate = end.Format("2006-01-02")
	}

	summary, err := h.buildSummary(userID, startDate, endDate, accountID)
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
	accountID := c.Query("account_id")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end dates required"})
		return
	}

	summary, err := h.buildSummary(userID, startDate, endDate, accountID)
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
		WHERE c.name != 'Transfer'
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
	accountID := c.Query("account_id")
	if months < 1 || months > 24 {
		months = 6
	}

	cutoff := time.Now().AddDate(0, -months, 0).Format("2006-01-02")

	query := `
		SELECT TO_CHAR(t.transaction_date, 'YYYY-MM') as month,
		       COALESCE(SUM(CASE WHEN t.txn_type = 'credit' AND (c.name IS NULL OR c.name != 'Transfer') THEN t.amount ELSE 0 END), 0) as income,
		       COALESCE(SUM(CASE WHEN t.txn_type = 'debit'  AND (c.name IS NULL OR c.name != 'Transfer') THEN t.amount ELSE 0 END), 0) as expenses
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1 AND t.transaction_date >= $2`
	args := []interface{}{userID, cutoff}

	if accountID != "" {
		query += " AND t.account_id = $3"
		args = append(args, accountID)
	}

	query += " GROUP BY month ORDER BY month"

	rows, err := h.pool.Query(context.Background(), query, args...)
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

func (h *ReportHandler) buildSummary(userID, startDate, endDate, accountID string) (*models.ReportSummary, error) {
	var totalIncome, totalExpenses, totalTransfers float64
	allTime := startDate == "" || endDate == ""

	sumQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN txn_type = 'credit' AND (c.name IS NULL OR c.name != 'Transfer') THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN txn_type = 'debit'  AND (c.name IS NULL OR c.name != 'Transfer') THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.name = 'Transfer' THEN t.amount ELSE 0 END), 0)
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1`
	sumArgs := []interface{}{userID}
	argIdx := 2

	if !allTime {
		sumQuery += fmt.Sprintf(" AND t.transaction_date >= $%d AND t.transaction_date < $%d", argIdx, argIdx+1)
		sumArgs = append(sumArgs, startDate, endDate)
		argIdx += 2
	}
	if accountID != "" {
		sumQuery += fmt.Sprintf(" AND t.account_id = $%d", argIdx)
		sumArgs = append(sumArgs, accountID)
		argIdx++
	}

	err := h.pool.QueryRow(context.Background(), sumQuery, sumArgs...).Scan(&totalIncome, &totalExpenses, &totalTransfers)
	if err != nil {
		return nil, err
	}

	catQuery := `
		SELECT c.id, c.name, c.color, c.icon,
		       COALESCE(SUM(t.amount), 0), COUNT(t.id)
		FROM categories c
		JOIN transactions t ON t.category_id = c.id
		WHERE t.user_id = $1`
	catArgs := []interface{}{userID}
	catIdx := 2

	if !allTime {
		catQuery += fmt.Sprintf(" AND t.transaction_date >= $%d AND t.transaction_date < $%d", catIdx, catIdx+1)
		catArgs = append(catArgs, startDate, endDate)
		catIdx += 2
	}
	if accountID != "" {
		catQuery += fmt.Sprintf(" AND t.account_id = $%d", catIdx)
		catArgs = append(catArgs, accountID)
	}
	catQuery += " GROUP BY c.id, c.name, c.color, c.icon ORDER BY SUM(t.amount) DESC"

	rows, err := h.pool.Query(context.Background(), catQuery, catArgs...)
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

	accQuery := `
		SELECT a.id, a.name, a.institution, a.account_type, a.last_four,
		       COALESCE(SUM(CASE WHEN t.txn_type = 'credit' THEN t.amount ELSE 0 END), 0) as income,
		       COALESCE(SUM(CASE WHEN t.txn_type = 'debit' THEN t.amount ELSE 0 END), 0) as expenses
		FROM accounts a
		LEFT JOIN transactions t ON t.account_id = a.id`
	accArgs := []interface{}{userID}
	accIdx := 2

	if !allTime {
		accQuery += fmt.Sprintf(" AND t.transaction_date >= $%d AND t.transaction_date < $%d", accIdx, accIdx+1)
		accArgs = append(accArgs, startDate, endDate)
		accIdx += 2
	}

	accQuery += " WHERE a.user_id = $1"
	if accountID != "" {
		accQuery += fmt.Sprintf(" AND a.id = $%d", accIdx)
		accArgs = append(accArgs, accountID)
	}
	accQuery += " GROUP BY a.id, a.name, a.institution, a.account_type, a.last_four ORDER BY a.name"

	accRows, err := h.pool.Query(context.Background(), accQuery, accArgs...)
	if err != nil {
		return nil, err
	}
	defer accRows.Close()

	accs := []models.AccountSummary{}
	for accRows.Next() {
		var as models.AccountSummary
		if err := accRows.Scan(&as.AccountID, &as.AccountName, &as.Institution, &as.AccountType, &as.LastFour, &as.Income, &as.Expenses); err != nil {
			return nil, err
		}
		as.Net = as.Income - as.Expenses
		accs = append(accs, as)
	}

	return &models.ReportSummary{
		TotalIncome:    totalIncome,
		TotalExpenses:  totalExpenses,
		TotalTransfers: totalTransfers,
		Net:            totalIncome - totalExpenses,
		ByCategory:     cats,
		ByAccount:      accs,
	}, nil
}
