package models

import (
	"snowflake/pb"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:char(36); primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(255); not null" json:"name"`
	Username     string    `gorm:"type:varchar(255); not null" json:"username"`
	Email        string    `gorm:"type:varchar(255); unique; not null" json:"email"`
	Password     string    `gorm:"type:varchar(255); not null" json:"-"`
	AnnualIncome int       `gorm:"type:bigint; not null" json:"annual_income"`
	Debt         int       `gorm:"type:bigint; not null" json:"debt"`
	AssetsValue  int       `gorm:"type:bigint; not null" json:"assets_value"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    time.Time `gorm:"autoCreateTime" json:"-"`
}

type JSONUser struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	AnnualIncome int       `json:"annual_income"`
	Debt         int       `json:"debt"`
	AssetsValue  int       `json:"assets_value"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()

	return nil
}

func (u *User) ToGRPC() *pb.User {
	return &pb.User{
		Id:           u.ID.String(),
		Name:         u.Name,
		Username:     u.Username,
		Email:        u.Email,
		AnnualIncome: int64(u.AnnualIncome),
		Debt:         int64(u.Debt),
		AssetsValue:  int64(u.AssetsValue),
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.Format(time.RFC3339),
	}
}
