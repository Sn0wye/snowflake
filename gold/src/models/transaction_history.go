package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionHistoryEntryType string

const (
	TransactionHistoryEntryTypeDebit  TransactionHistoryEntryType = "debit"
	TransactionHistoryEntryTypeCredit TransactionHistoryEntryType = "credit"
)

type TransactionHistory struct {
	ID            uuid.UUID                   `gorm:"type:uuid; primaryKey" json:"id"`
	TransactionID uuid.UUID                   `gorm:"type:uuid; not null; index" json:"transaction_id"`
	AccountID     uuid.UUID                   `gorm:"type:uuid; not null; index:idx_account_created" json:"account_id"`
	EntryType     TransactionHistoryEntryType `gorm:"type:varchar(10); not null" json:"entry_type"`
	Amount        int64                       `gorm:"type:bigint; not null" json:"amount"`         // In cents, always positive
	BalanceBefore int64                       `gorm:"type:bigint; not null" json:"balance_before"` // In cents
	BalanceAfter  int64                       `gorm:"type:bigint; not null" json:"balance_after"`  // In cents
	Description   string                      `gorm:"type:text" json:"description,omitempty"`
	CreatedAt     time.Time                   `gorm:"autoCreateTime; index:idx_account_created" json:"-"` // Immutable
}

type JSONTransactionHistory struct {
	ID            uuid.UUID                   `json:"id"`
	TransactionID uuid.UUID                   `json:"transaction_id"`
	AccountID     uuid.UUID                   `json:"account_id"`
	EntryType     TransactionHistoryEntryType `json:"entry_type"`
	Amount        int64                       `json:"amount"`
	BalanceBefore int64                       `json:"balance_before"`
	BalanceAfter  int64                       `json:"balance_after"`
	Description   string                      `json:"description,omitempty"`
}

func (th *TransactionHistory) BeforeCreate(tx *gorm.DB) error {
	th.ID = uuid.New()
	return nil
}

// TransactionHistory is immutable - no updates allowed
func (th *TransactionHistory) BeforeUpdate(tx *gorm.DB) error {
	return gorm.ErrInvalidData
}
