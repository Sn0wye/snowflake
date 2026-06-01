package repository

import (
	"time"

	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionRepo struct{}

func NewTransactionRepo() TransactionRepository {
	return &transactionRepo{}
}

func (r *transactionRepo) FindByID(db *gorm.DB, id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := db.Where("id = ?", id).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepo) FindByIdempotencyKey(db *gorm.DB, key uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := db.Where("idempotency_key = ?", key).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepo) FindByAccountID(db *gorm.DB, accountID uuid.UUID, filter TransactionFilter) ([]models.Transaction, int64, error) {
	query := db.Model(&models.Transaction{}).
		Where("sender_account_id = ? OR receiver_account_id = ?", accountID, accountID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	var total int64
	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit

	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *transactionRepo) SumDailyAmount(db *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error) {
	var dailyAmount int64
	err := db.Model(&models.Transaction{}).
		Where("sender_account_id = ? AND status = ? AND created_at >= ?",
			accountID, models.TransactionStatusCompleted, startOfDay).
		Select("COALESCE(SUM(amount), 0)").Scan(&dailyAmount).Error
	return dailyAmount, err
}

func (r *transactionRepo) CountDaily(db *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error) {
	var count int64
	err := db.Model(&models.Transaction{}).
		Where("sender_account_id = ? AND status = ? AND created_at >= ?",
			accountID, models.TransactionStatusCompleted, startOfDay).
		Count(&count).Error
	return count, err
}

func (r *transactionRepo) Create(db *gorm.DB, transaction *models.Transaction) error {
	return db.Create(transaction).Error
}

func (r *transactionRepo) Save(db *gorm.DB, transaction *models.Transaction) error {
	return db.Save(transaction).Error
}
