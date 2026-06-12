package outbox

import (
	"context"
	"time"

	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"github.com/getsnowflake/snowflake/gold/src/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Publisher interface {
	Produce(queueName, message string) error
	ProduceWithHeaders(queueName, message string, correlationID string) error
}

type Worker struct {
	db           *gorm.DB
	publisher    Publisher
	repos        *repository.Factory
	log          *logger.Logger
	pollInterval time.Duration
	maxAttempts  int
}

func NewWorker(db *gorm.DB, publisher Publisher, repos *repository.Factory, log *logger.Logger) *Worker {
	return &Worker{
		db:           db,
		publisher:    publisher,
		repos:        repos,
		log:          log,
		pollInterval: 5 * time.Second,
		maxAttempts:  5,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Outbox worker shutting down")
			return
		case <-ticker.C:
			w.processOnce()
		}
	}
}

func (w *Worker) processOnce() {
	err := w.db.Transaction(func(tx *gorm.DB) error {
		entries, err := w.repos.Outbox.ClaimPending(tx, 20)
		if err != nil {
			w.log.Error("Outbox worker: failed to claim pending events", zap.Error(err))
			return err
		}

		for _, entry := range entries {
			correlationID := ""
			if entry.CorrelationID != nil {
				correlationID = *entry.CorrelationID
			}
			if pubErr := w.publisher.ProduceWithHeaders(entry.QueueName, entry.Payload, correlationID); pubErr != nil {
				w.log.Error("Outbox worker: failed to publish event",
					zap.String("eventID", entry.ID.String()),
					zap.String("queueName", entry.QueueName),
					zap.Error(pubErr),
				)
				if markErr := w.repos.Outbox.MarkFailed(tx, entry.ID, pubErr.Error(), w.maxAttempts); markErr != nil {
					w.log.Error("Outbox worker: failed to mark event as failed",
						zap.String("eventID", entry.ID.String()),
						zap.Error(markErr),
					)
				}
				continue
			}

			if markErr := w.repos.Outbox.MarkPublished(tx, entry.ID); markErr != nil {
				w.log.Error("Outbox worker: failed to mark event as published",
					zap.String("eventID", entry.ID.String()),
					zap.Error(markErr),
				)
			}
		}

		return nil
	})
	if err != nil {
		w.log.Error("Outbox worker: processOnce transaction failed", zap.Error(err))
	}
}
