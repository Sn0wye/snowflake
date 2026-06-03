package repository

import (
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type refreshTokenRepo struct{}

func NewRefreshTokenRepo() RefreshTokenRepository {
	return &refreshTokenRepo{}
}

func (r *refreshTokenRepo) FindByTokenHash(db *gorm.DB, tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := db.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepo) Create(db *gorm.DB, token *models.RefreshToken) error {
	return db.Create(token).Error
}

func (r *refreshTokenRepo) Delete(db *gorm.DB, token *models.RefreshToken) error {
	return db.Delete(token).Error
}

func (r *refreshTokenRepo) DeleteAllByUserID(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
