package repository

import (
	"time"

	"github.com/getsnowflake/snowflake/gold/src/models"
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
	query := db.Model(&models.Transaction{})

	switch filter.Direction {
	case "debit":
		query = query.Where("sender_account_id = ?", accountID)
	case "credit":
		query = query.Where("receiver_account_id = ?", accountID)
	default:
		query = query.Where("sender_account_id = ? OR receiver_account_id = ?", accountID, accountID)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at < ?", *filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
