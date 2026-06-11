package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusDead      OutboxStatus = "dead"
)

type OutboxEvent struct {
	ID          uuid.UUID    `gorm:"type:uuid; primaryKey" json:"id"`
	QueueName   string       `gorm:"type:varchar(255); not null" json:"queue_name"`
	Payload     string       `gorm:"type:text; not null" json:"payload"`
	Status      OutboxStatus `gorm:"type:varchar(20); not null; default:'pending'" json:"status"`
	Attempts    int          `gorm:"type:int; not null; default:0" json:"attempts"`
	LastError   *string      `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"-"`
	PublishedAt *time.Time   `gorm:"type:timestamptz" json:"published_at,omitempty"`
}

func (o *OutboxEvent) BeforeCreate(tx *gorm.DB) error {
	o.ID = uuid.New()
	return nil
}
