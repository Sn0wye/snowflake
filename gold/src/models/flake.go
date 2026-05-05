package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlakeType string

const (
	FlakeTypeEmail  FlakeType = "email"
	FlakeTypePhone  FlakeType = "phone"
	FlakeTypeCPF    FlakeType = "cpf"
	FlakeTypeCNPJ   FlakeType = "cnpj"
	FlakeTypeRandom FlakeType = "random"
	FlakeTypeHandle FlakeType = "handle"
)

type FlakeStatus string

const (
	FlakeStatusActive   FlakeStatus = "active"
	FlakeStatusInactive FlakeStatus = "inactive"
)

type Flake struct {
	ID        uuid.UUID   `gorm:"type:char(36); primaryKey" json:"id"`
	AccountID uuid.UUID   `gorm:"type:char(36); not null; index" json:"account_id"`
	KeyType   FlakeType   `gorm:"type:varchar(20); not null" json:"key_type"`
	KeyValue  string      `gorm:"type:varchar(255); unique; not null" json:"key_value"`
	Status    FlakeStatus `gorm:"type:varchar(20); not null; default:'active'" json:"status"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"-"`
}

type JSONFlake struct {
	ID        uuid.UUID   `json:"id"`
	AccountID uuid.UUID   `json:"account_id"`
	KeyType   FlakeType   `json:"key_type"`
	KeyValue  string      `json:"key_value"`
	Status    FlakeStatus `json:"status"`
}

func (f *Flake) BeforeCreate(tx *gorm.DB) error {
	f.ID = uuid.New()

	return nil
}

func (f *Flake) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()

	return nil
}
