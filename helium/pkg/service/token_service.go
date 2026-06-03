package service

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 30 * 24 * time.Hour
)

type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(db *gorm.DB, userID string) (string, error)
	RevokeAllUserRefreshTokens(db *gorm.DB, userID string) error
}

type tokenService struct {
	jwt *jwt.JWT
}

func newTokenService(j *jwt.JWT) TokenService {
	return &tokenService{jwt: j}
}

func (s *tokenService) GenerateAccessToken(userID string) (string, error) {
	return s.jwt.GenToken(userID, time.Now().Add(accessTokenDuration))
}

func (s *tokenService) GenerateRefreshToken(db *gorm.DB, userID string) (string, error) {
	tokenString, err := s.jwt.GenRefreshToken(userID, time.Now().Add(refreshTokenDuration))
	if err != nil {
		return "", errors.Wrap(err, "failed to generate refresh token")
	}

	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse user ID")
	}

	rt := models.RefreshToken{
		UserID:    parsedUserID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}

	if err := repository.NewRefreshTokenRepo().Create(db, &rt); err != nil {
		return "", errors.Wrap(err, "failed to store refresh token")
	}

	return tokenString, nil
}

func (s *tokenService) RevokeAllUserRefreshTokens(db *gorm.DB, userID string) error {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return errors.Wrap(err, "failed to parse user ID")
	}
	return repository.NewRefreshTokenRepo().DeleteAllByUserID(db, parsedUserID)
}
