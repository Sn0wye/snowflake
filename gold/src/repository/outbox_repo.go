package repository

import (
	"time"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxOutboxAttempts = 5

type outboxRepo struct{}

func NewOutboxRepo() OutboxRepository {
	return &outboxRepo{}
}

func (r *outboxRepo) Create(db *gorm.DB, entry *models.OutboxEvent) error {
	return db.Create(entry).Error
}

func (r *outboxRepo) ClaimPending(db *gorm.DB, limit int) ([]models.OutboxEvent, error) {
	var entries []models.OutboxEvent
	err := db.Raw(`
		SELECT * FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`, limit).Scan(&entries).Error
	return entries, err
}

func (r *outboxRepo) MarkPublished(db *gorm.DB, id uuid.UUID) error {
	now := time.Now()
	return db.Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.OutboxStatusPublished,
			"published_at": now,
		}).Error
}

func (r *outboxRepo) MarkFailed(db *gorm.DB, id uuid.UUID, errMsg string) error {
	return db.Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"attempts":   gorm.Expr("attempts + 1"),
			"last_error": errMsg,
			"status": gorm.Expr(
				"CASE WHEN attempts + 1 >= ? THEN ?::varchar ELSE status END",
				maxOutboxAttempts, models.OutboxStatusDead,
			),
		}).Error
}
