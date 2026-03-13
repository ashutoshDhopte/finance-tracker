package models

import (
	"time"
)

type Transaction struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	AccountID       *string    `json:"account_id,omitempty"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	MerchantName    *string    `json:"merchant_name,omitempty"`
	MerchantRaw     *string    `json:"merchant_raw,omitempty"`
	CategoryID      *string    `json:"category_id,omitempty"`
	CategoryName    *string    `json:"category_name,omitempty"`
	TransactionDate time.Time  `json:"transaction_date"`
	PostedDate      *time.Time `json:"posted_date,omitempty"`
	TxnType         string     `json:"txn_type"`
	Source          string     `json:"source"`
	SourceHash      *string    `json:"source_hash,omitempty"`
	AIConfidence    *float32   `json:"ai_confidence,omitempty"`
	RawText         *string    `json:"raw_text,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTransactionRequest struct {
	AccountID       *string  `json:"account_id,omitempty"`
	Amount          float64  `json:"amount" binding:"required"`
	Currency        string   `json:"currency"`
	MerchantName    *string  `json:"merchant_name,omitempty"`
	CategoryID      *string  `json:"category_id,omitempty"`
	TransactionDate string   `json:"transaction_date" binding:"required"`
	PostedDate      *string  `json:"posted_date,omitempty"`
	TxnType         string   `json:"txn_type" binding:"required,oneof=credit debit"`
	Source          string   `json:"source"`
	Notes           *string  `json:"notes,omitempty"`
}

type UpdateTransactionRequest struct {
	AccountID       *string  `json:"account_id,omitempty"`
	Amount          *float64 `json:"amount,omitempty"`
	MerchantName    *string  `json:"merchant_name,omitempty"`
	CategoryID      *string  `json:"category_id,omitempty"`
	TransactionDate *string  `json:"transaction_date,omitempty"`
	TxnType         *string  `json:"txn_type,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
}

type TransactionFilter struct {
	StartDate  string
	EndDate    string
	CategoryID string
	AccountID  string
	TxnType    string
	Source     string
	Search     string
	Limit      int
	Offset     int
}
