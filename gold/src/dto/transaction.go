package dto

import (
	"time"

	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
)

type CreateTransferRequest struct {
	ReceiverFlakeKey string    `json:"receiver_flake_key" validate:"required" example:"my@example.com"`
	Amount           int64     `json:"amount" validate:"required,min=1" example:"1000"`
	Description      string    `json:"description" example:"Lunch split"`
	IdempotencyKey   uuid.UUID `json:"idempotency_key" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
} // @name CreateTransferRequest

type DepositRequest struct {
	AccountID      uuid.UUID `json:"account_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount         int64     `json:"amount" validate:"required,min=1" example:"10000"`
	Description    string    `json:"description" example:"Initial deposit"`
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
} // @name DepositRequest

type TransactionResponse struct {
	ID                uuid.UUID                `json:"id"`
	Type              models.TransactionType   `json:"type"`
	Status            models.TransactionStatus `json:"status"`
	Amount            int64                    `json:"amount"`
	SenderAccountID   *uuid.UUID               `json:"sender_account_id,omitempty"`
	ReceiverAccountID *uuid.UUID               `json:"receiver_account_id,omitempty"`
	FlakeKeyUsed      string                   `json:"flake_key_used,omitempty"`
	Description       string                   `json:"description,omitempty"`
	IdempotencyKey    uuid.UUID                `json:"idempotency_key"`
	CompletedAt       *time.Time               `json:"completed_at,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
} // @name TransactionResponse

type PaginatedTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
} // @name PaginatedTransactionsResponse

func TransactionToResponse(t models.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:                t.ID,
		Type:              t.Type,
		Status:            t.Status,
		Amount:            t.Amount,
		SenderAccountID:   t.SenderAccountID,
		ReceiverAccountID: t.ReceiverAccountID,
		FlakeKeyUsed:      t.FlakeKeyUsed,
		Description:       t.Description,
		IdempotencyKey:    t.IdempotencyKey,
		CompletedAt:       t.CompletedAt,
		CreatedAt:         t.CreatedAt,
	}
}
