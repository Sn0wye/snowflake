package repository

import (
	"github.com/getsnowflake/snowflake/helium/src/models"

	"gorm.io/gorm"
)

type oauthAccountRepo struct{}

func NewOAuthAccountRepo() OAuthAccountRepository {
	return &oauthAccountRepo{}
}

func (r *oauthAccountRepo) FindByProviderAndProviderID(db *gorm.DB, provider models.OAuthProvider, providerID string) (*models.OAuthAccount, error) {
	var account models.OAuthAccount
	if err := db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&account).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &account, nil
}

func (r *oauthAccountRepo) Create(db *gorm.DB, account *models.OAuthAccount) error {
	return db.Create(account).Error
}
