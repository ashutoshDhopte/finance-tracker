package models

import (
	"time"
)

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Color     string    `json:"color"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	Icon     string  `json:"icon"`
	Color    string  `json:"color"`
	ParentID *string `json:"parent_id,omitempty"`
}
