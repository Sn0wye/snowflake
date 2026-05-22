package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
)

var validProviders = map[OAuthProvider]bool{
	ProviderGoogle: true,
}

func (p OAuthProvider) IsValid() bool {
	return validProviders[p]
}

type OAuthAccount struct {
	ID         uuid.UUID     `gorm:"type:char(36); primaryKey" json:"id"`
	UserID     uuid.UUID     `gorm:"type:char(36); not null; index" json:"user_id"`
	Provider   OAuthProvider `gorm:"type:varchar(50); not null; uniqueIndex:idx_provider_provider_id" json:"provider"`
	ProviderID string        `gorm:"type:varchar(255); not null; uniqueIndex:idx_provider_provider_id" json:"provider_id"`
	CreatedAt  time.Time     `gorm:"autoCreateTime" json:"-"`
	UpdatedAt  time.Time     `gorm:"autoCreateTime" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (o *OAuthAccount) BeforeCreate(tx *gorm.DB) error {
	if !o.Provider.IsValid() {
		return fmt.Errorf("invalid OAuth provider: %s", o.Provider)
	}
	o.ID = uuid.New()
	return nil
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
