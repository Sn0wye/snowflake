package repository

import (
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type flakeRepo struct{}

func NewFlakeRepo() FlakeRepository {
	return &flakeRepo{}
}

func (r *flakeRepo) FindByAccountID(db *gorm.DB, accountID uuid.UUID) ([]models.Flake, error) {
	var flakes []models.Flake
	if err := db.Where("account_id = ?", accountID).Find(&flakes).Error; err != nil {
		return nil, err
	}
	return flakes, nil
}

func (r *flakeRepo) FindByKeyValue(db *gorm.DB, keyValue string) (*models.Flake, error) {
	var flake models.Flake
	if err := db.Where("key_value = ? AND status = ?", keyValue, models.FlakeStatusActive).
		First(&flake).Error; err != nil {
		return nil, err
	}
	return &flake, nil
}

func (r *flakeRepo) FindByIDAndAccount(db *gorm.DB, id uuid.UUID, accountID uuid.UUID) (*models.Flake, error) {
	var flake models.Flake
	if err := db.Where("id = ? AND account_id = ?", id, accountID).First(&flake).Error; err != nil {
		return nil, err
	}
	return &flake, nil
}

func (r *flakeRepo) FindByTypeAndAccount(db *gorm.DB, keyType models.FlakeType, accountID uuid.UUID) (*models.Flake, error) {
	var flake models.Flake
	if err := db.Where("key_type = ? AND account_id = ? AND status = ?",
		keyType, accountID, models.FlakeStatusActive).First(&flake).Error; err != nil {
		return nil, err
	}
	return &flake, nil
}

func (r *flakeRepo) CountNonUnlimited(db *gorm.DB, accountID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&models.Flake{}).
		Where("account_id = ? AND status = ? AND key_type NOT IN ?",
			accountID, models.FlakeStatusActive,
			[]models.FlakeType{models.FlakeTypeRandom, models.FlakeTypeHandle}).
		Count(&count).Error
	return count, err
}

func (r *flakeRepo) Create(db *gorm.DB, flake *models.Flake) error {
	return db.Create(flake).Error
}

func (r *flakeRepo) UpdateStatus(db *gorm.DB, flake *models.Flake, status models.FlakeStatus) error {
	return db.Model(flake).Update("status", status).Error
}
