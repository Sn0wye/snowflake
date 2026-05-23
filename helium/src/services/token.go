package services

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 30 * 24 * time.Hour
)

type TokenService struct {
	db *gorm.DB
	j  *jwt.JWT
}

func NewTokenService(db *gorm.DB, j *jwt.JWT) *TokenService {
	return &TokenService{db: db, j: j}
}

func (s *TokenService) GenerateAccessToken(userId string) (string, error) {
	return s.j.GenToken(userId, time.Now().Add(AccessTokenDuration))
}

func (s *TokenService) GenerateRefreshToken(userId string) (string, error) {
	tokenString, err := s.j.GenRefreshToken(userId, time.Now().Add(RefreshTokenDuration))
	if err != nil {
		return "", errors.Wrap(err, "failed to generate refresh token")
	}

	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse user ID")
	}

	rt := models.RefreshToken{
		UserID:    parsedUserID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.db.Create(&rt).Error; err != nil {
		return "", errors.Wrap(err, "failed to store refresh token")
	}

	return tokenString, nil
}

func (s *TokenService) RevokeAllUserRefreshTokens(userId string) error {
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return errors.Wrap(err, "failed to parse user ID")
	}

	return s.db.Where("user_id = ?", parsedUserID).Delete(&models.RefreshToken{}).Error
}
