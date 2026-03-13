package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ash/finance-tracker/backend/internal/models"
)

type CategoryHandler struct {
	pool *pgxpool.Pool
}

func NewCategoryHandler(pool *pgxpool.Pool) *CategoryHandler {
	return &CategoryHandler{pool: pool}
}

func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.pool.Query(context.Background(),
		"SELECT id, name, icon, color, parent_id, created_at FROM categories ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query categories"})
		return
	}
	defer rows.Close()

	cats := []models.Category{}
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Color, &cat.ParentID, &cat.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan category"})
			return
		}
		cats = append(cats, cat)
	}

	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Icon == "" {
		req.Icon = "tag"
	}
	if req.Color == "" {
		req.Color = "#6B7280"
	}

	var id string
	err := h.pool.QueryRow(context.Background(),
		"INSERT INTO categories (name, icon, color, parent_id) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Name, req.Icon, req.Color, req.ParentID,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
