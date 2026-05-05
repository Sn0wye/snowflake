package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeRefund     TransactionType = "refund"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

type Transaction struct {
	ID                uuid.UUID         `gorm:"type:char(36); primaryKey" json:"id"`
	Type              TransactionType   `gorm:"type:varchar(20); not null" json:"type"`
	Status            TransactionStatus `gorm:"type:varchar(20); not null; index" json:"status"`
	Amount            int64             `gorm:"type:bigint; not null" json:"amount"` // In cents
	SenderAccountID   *uuid.UUID        `gorm:"type:char(36); index" json:"sender_account_id,omitempty"`
	ReceiverAccountID *uuid.UUID        `gorm:"type:char(36); index" json:"receiver_account_id,omitempty"`
	FlakeKeyUsed      string            `gorm:"type:varchar(255)" json:"flake_key_used,omitempty"`
	Description       string            `gorm:"type:text" json:"description,omitempty"`
	IdempotencyKey    uuid.UUID         `gorm:"type:char(36); unique; not null" json:"idempotency_key"`
	Metadata          json.RawMessage   `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time         `gorm:"autoCreateTime; index" json:"-"`
	CompletedAt       *time.Time        `gorm:"type:timestamp" json:"completed_at,omitempty"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime" json:"-"`
}

type JSONTransaction struct {
	ID                uuid.UUID         `json:"id"`
	Type              TransactionType   `json:"type"`
	Status            TransactionStatus `json:"status"`
	Amount            int64             `json:"amount"`
	SenderAccountID   *uuid.UUID        `json:"sender_account_id,omitempty"`
	ReceiverAccountID *uuid.UUID        `json:"receiver_account_id,omitempty"`
	FlakeKeyUsed      string            `json:"flake_key_used,omitempty"`
	Description       string            `json:"description,omitempty"`
	IdempotencyKey    uuid.UUID         `json:"idempotency_key"`
	CompletedAt       *time.Time        `json:"completed_at,omitempty"`
	Metadata          json.RawMessage   `json:"metadata,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()

	return nil
}

func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()

	return nil
}
