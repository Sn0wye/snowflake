package repository

import (
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(db *gorm.DB, id string) (*models.User, error)
	FindByEmail(db *gorm.DB, email string) (*models.User, error)
	Create(db *gorm.DB, user *models.User) error
	CountByUsername(db *gorm.DB, username string) (int64, error)
}

type RefreshTokenRepository interface {
	FindByTokenHash(db *gorm.DB, tokenHash string) (*models.RefreshToken, error)
	Create(db *gorm.DB, token *models.RefreshToken) error
	Delete(db *gorm.DB, token *models.RefreshToken) error
	DeleteAllByUserID(db *gorm.DB, userID uuid.UUID) error
}

type OAuthAccountRepository interface {
	FindByProviderAndProviderID(db *gorm.DB, provider models.OAuthProvider, providerID string) (*models.OAuthAccount, error)
	Create(db *gorm.DB, account *models.OAuthAccount) error
}

type Factory struct {
	User          UserRepository
	RefreshToken  RefreshTokenRepository
	OAuthAccount  OAuthAccountRepository
}

func NewFactory() *Factory {
	return &Factory{
		User:         NewUserRepo(),
		RefreshToken: NewRefreshTokenRepo(),
		OAuthAccount: NewOAuthAccountRepo(),
	}
}
