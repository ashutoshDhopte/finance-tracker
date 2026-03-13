package models

import (
	"time"
)

type Account struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Name         string     `json:"name"`
	Institution  string     `json:"institution"`
	AccountType  string     `json:"account_type"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Institution string `json:"institution" binding:"required"`
	AccountType string `json:"account_type" binding:"required,oneof=checking savings credit_card investment"`
}
