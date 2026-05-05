package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	QueueTransactionCreated   = "transaction.created"
	QueueTransactionCompleted = "transaction.completed"
	QueueTransactionFailed    = "transaction.failed"
)

type TransactionCreatedEvent struct {
	TransactionID     string  `json:"transaction_id"`
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	SenderAccountID   *string `json:"sender_account_id,omitempty"`
	ReceiverAccountID *string `json:"receiver_account_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

type TransactionCompletedEvent struct {
	TransactionID     string  `json:"transaction_id"`
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	SenderAccountID   *string `json:"sender_account_id,omitempty"`
	ReceiverAccountID *string `json:"receiver_account_id,omitempty"`
	CompletedAt       string  `json:"completed_at"`
}

type TransactionFailedEvent struct {
	TransactionID     string  `json:"transaction_id"`
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	SenderAccountID   *string `json:"sender_account_id,omitempty"`
	ReceiverAccountID *string `json:"receiver_account_id,omitempty"`
	Reason            string  `json:"reason"`
	FailedAt          string  `json:"failed_at"`
}

// Marshal serialises any event struct to JSON string.
func Marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("events.Marshal: %w", err)
	}
	return string(b), nil
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func strPtr(s string) *string { return &s }

func NewTransactionCompleted(txID, txType string, amount int64, senderID, receiverID *string, completedAt time.Time) TransactionCompletedEvent {
	e := TransactionCompletedEvent{
		TransactionID:     txID,
		Type:              txType,
		Amount:            amount,
		SenderAccountID:   senderID,
		ReceiverAccountID: receiverID,
		CompletedAt:       completedAt.UTC().Format(time.RFC3339),
	}
	_ = strPtr // suppress unused warning
	_ = nowUTC
	return e
}

func NewTransactionFailed(txID, txType string, amount int64, senderID, receiverID *string, reason string) TransactionFailedEvent {
	return TransactionFailedEvent{
		TransactionID:     txID,
		Type:              txType,
		Amount:            amount,
		SenderAccountID:   senderID,
		ReceiverAccountID: receiverID,
		Reason:            reason,
		FailedAt:          time.Now().UTC().Format(time.RFC3339),
	}
}
