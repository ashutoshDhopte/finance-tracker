package models

import (
	"time"
)

type Account struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Name              string     `json:"name"`
	Institution       string     `json:"institution"`
	AccountType       string     `json:"account_type"`
	LastFour          *string    `json:"last_four,omitempty"`
	DebitCardLastFour *string    `json:"debit_card_last_four,omitempty"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateAccountRequest struct {
	Name              string  `json:"name" binding:"required"`
	Institution       string  `json:"institution" binding:"required"`
	AccountType       string  `json:"account_type" binding:"required,oneof=checking savings credit_card investment"`
	LastFour          *string `json:"last_four,omitempty"`
	DebitCardLastFour *string `json:"debit_card_last_four,omitempty"`
}

type UpdateAccountRequest struct {
	Name              *string `json:"name,omitempty"`
	Institution       *string `json:"institution,omitempty"`
	AccountType       *string `json:"account_type,omitempty"`
	LastFour          *string `json:"last_four,omitempty"`
	DebitCardLastFour *string `json:"debit_card_last_four,omitempty"`
}
