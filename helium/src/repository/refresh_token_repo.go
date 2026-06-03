package repository

import (
	"time"

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
		return nil, wrapNotFound(err)
	}
	return &token, nil
}

func (r *refreshTokenRepo) Create(db *gorm.DB, token *models.RefreshToken) error {
	return db.Create(token).Error
}

func (r *refreshTokenRepo) Delete(db *gorm.DB, token *models.RefreshToken) error {
	return db.Delete(token).Error
}

func (r *refreshTokenRepo) DeleteByTokenHash(db *gorm.DB, tokenHash string) (bool, error) {
	result := db.Where("token_hash = ? AND expires_at >= ?", tokenHash, time.Now()).
		Delete(&models.RefreshToken{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *refreshTokenRepo) DeleteAllByUserID(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
