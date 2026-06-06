package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusClosed    AccountStatus = "closed"
)

type AccountReconciliationStatus string

const (
	AccountReconciliationStatusOK          AccountReconciliationStatus = "ok"
	AccountReconciliationStatusDiscrepancy AccountReconciliationStatus = "discrepancy"
)

type Account struct {
	ID                    uuid.UUID                   `gorm:"type:uuid; primaryKey" json:"id"`
	UserID                uuid.UUID                   `gorm:"type:uuid; not null; unique" json:"user_id"`
	Balance               int64                       `gorm:"type:bigint; default:0; not null" json:"balance"` // In cents
	LastReconciledAt      *time.Time                  `gorm:"type:timestamp" json:"last_reconciled_at,omitempty"`
	Status                AccountStatus               `gorm:"type:varchar(20); default:'active'; column:status; not null" json:"status"`
	ReconciliationStatus  AccountReconciliationStatus `gorm:"type:varchar(20); default:'ok'; not null" json:"reconciliation_status"`
	DiscrepancyDetectedAt *time.Time                  `gorm:"type:timestamp" json:"discrepancy_detected_at,omitempty"`
	DiscrepancyAmount     *int64                      `gorm:"type:bigint" json:"discrepancy_amount,omitempty"` // calculated - cached, in cents
	CreatedAt             time.Time                   `gorm:"autoCreateTime" json:"-"`
	UpdatedAt             time.Time                   `gorm:"autoUpdateTime" json:"-"`
}

type JSONAccount struct {
	ID                    uuid.UUID                   `json:"id"`
	UserID                uuid.UUID                   `json:"user_id"`
	Balance               int64                       `json:"balance"`
	LastReconciledAt      *time.Time                  `json:"last_reconciled_at,omitempty"`
	Status                AccountStatus               `json:"status"`
	ReconciliationStatus  AccountReconciliationStatus `json:"reconciliation_status"`
	DiscrepancyDetectedAt *time.Time                  `json:"discrepancy_detected_at,omitempty"`
	DiscrepancyAmount     *int64                      `json:"discrepancy_amount,omitempty"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()

	return nil
}

func (a *Account) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = time.Now()

	return nil
}
