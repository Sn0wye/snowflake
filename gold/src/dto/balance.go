package dto

import (
	"time"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/google/uuid"
)

type BalanceResponse struct {
	AccountID            uuid.UUID                          `json:"account_id"`
	Balance              int64                              `json:"balance"` // in cents
	Status               models.AccountStatus               `json:"status"`
	ReconciliationStatus models.AccountReconciliationStatus `json:"reconciliation_status"`
} // @name BalanceResponse

type BalanceHistoryEntry struct {
	ID            uuid.UUID                          `json:"id"`
	TransactionID uuid.UUID                          `json:"transaction_id"`
	EntryType     models.TransactionHistoryEntryType `json:"entry_type"`
	Amount        int64                              `json:"amount"`
	BalanceBefore int64                              `json:"balance_before"`
	BalanceAfter  int64                              `json:"balance_after"`
	Description   string                             `json:"description,omitempty"`
	CreatedAt     time.Time                          `json:"created_at"`
} // @name BalanceHistoryEntry

type BalanceHistoryResponse struct {
	Entries []BalanceHistoryEntry `json:"entries"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
} // @name BalanceHistoryResponse

func AccountToBalanceResponse(a models.Account) BalanceResponse {
	return BalanceResponse{
		AccountID:            a.ID,
		Balance:              a.Balance,
		Status:               a.Status,
		ReconciliationStatus: a.ReconciliationStatus,
	}
}

func TransactionHistoryToEntry(th models.TransactionHistory) BalanceHistoryEntry {
	return BalanceHistoryEntry{
		ID:            th.ID,
		TransactionID: th.TransactionID,
		EntryType:     th.EntryType,
		Amount:        th.Amount,
		BalanceBefore: th.BalanceBefore,
		BalanceAfter:  th.BalanceAfter,
		Description:   th.Description,
		CreatedAt:     th.CreatedAt,
	}
}
