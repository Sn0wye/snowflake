package repository

import (
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type accountRepo struct{}

func NewAccountRepo() AccountRepository {
	return &accountRepo{}
}

func (r *accountRepo) FindByUserID(db *gorm.DB, userID string) (*models.Account, error) {
	var account models.Account
	if err := db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) FindByID(db *gorm.DB, id uuid.UUID) (*models.Account, error) {
	var account models.Account
	if err := db.Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) FindByIDForUpdate(db *gorm.DB, id uuid.UUID) (*models.Account, error) {
	var account models.Account
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) Create(db *gorm.DB, account *models.Account) error {
	return db.Create(account).Error
}

func (r *accountRepo) Update(db *gorm.DB, account *models.Account) error {
	return db.Save(account).Error
}

func (r *accountRepo) All(db *gorm.DB) ([]models.Account, error) {
	var accounts []models.Account
	if err := db.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) UpdateReconciliationFields(db *gorm.DB, account *models.Account, fields map[string]interface{}) error {
	return db.Model(account).Updates(fields).Error
}
