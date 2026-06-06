package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid; primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid; index; not null" json:"user_id"`
	TokenHash string    `gorm:"type:varchar(64); uniqueIndex; not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
}

func (t *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()
	return nil
}

func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
