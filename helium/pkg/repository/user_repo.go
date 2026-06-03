package repository

import (
	"github.com/getsnowflake/snowflake/helium/src/models"

	"gorm.io/gorm"
)

type userRepo struct{}

func NewUserRepo() UserRepository {
	return &userRepo{}
}

func (r *userRepo) FindByID(db *gorm.DB, id string) (*models.User, error) {
	var user models.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &user, nil
}

func (r *userRepo) FindByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &user, nil
}

func (r *userRepo) Create(db *gorm.DB, user *models.User) error {
	return db.Create(user).Error
}

func (r *userRepo) CountByUsername(db *gorm.DB, username string) (int64, error) {
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
