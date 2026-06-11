package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	ExchangeUserCreated       = "user.created"
	QueueUserCreatedAccount   = "user.created.account"
	QueueTransactionCreated   = "transaction.created"
	QueueTransactionCompleted = "transaction.completed"
	QueueTransactionReceived  = "transaction.received"
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

type TransactionReceivedEvent struct {
	TransactionID     string `json:"transaction_id"`
	Type              string `json:"type"`
	Amount            int64  `json:"amount"`
	ReceiverAccountID string `json:"receiver_account_id"`
	CompletedAt       string `json:"completed_at"`
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

func NewTransactionReceived(txID, txType string, amount int64, receiverID string, completedAt time.Time) TransactionReceivedEvent {
	return TransactionReceivedEvent{
		TransactionID:     txID,
		Type:              txType,
		Amount:            amount,
		ReceiverAccountID: receiverID,
		CompletedAt:       completedAt.UTC().Format(time.RFC3339),
	}
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
