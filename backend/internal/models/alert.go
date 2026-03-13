package models

import (
	"time"
)

type Alert struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Name            string     `json:"name"`
	CategoryID      *string    `json:"category_id,omitempty"`
	CategoryName    *string    `json:"category_name,omitempty"`
	Threshold       float64    `json:"threshold"`
	Period          string     `json:"period"`
	Enabled         bool       `json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateAlertRequest struct {
	Name       string  `json:"name" binding:"required"`
	CategoryID *string `json:"category_id,omitempty"`
	Threshold  float64 `json:"threshold" binding:"required"`
	Period     string  `json:"period" binding:"required,oneof=daily weekly biweekly monthly"`
}

type TriggeredAlert struct {
	Alert        Alert   `json:"alert"`
	CurrentSpend float64 `json:"current_spend"`
	Exceeded     bool    `json:"exceeded"`
}
