package repository

import (
	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionHistoryRepo struct{}

func NewTransactionHistoryRepo() TransactionHistoryRepository {
	return &transactionHistoryRepo{}
}

func (r *transactionHistoryRepo) Create(db *gorm.DB, history *models.TransactionHistory) error {
	return db.Create(history).Error
}

func (r *transactionHistoryRepo) FindByAccountID(db *gorm.DB, accountID uuid.UUID, page, limit int) ([]models.TransactionHistory, int64, error) {
	var total int64
	if err := db.Model(&models.TransactionHistory{}).
		Where("account_id = ?", accountID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	var entries []models.TransactionHistory
	if err := db.
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *transactionHistoryRepo) SumLedgerBalance(db *gorm.DB, accountID uuid.UUID) (int64, error) {
	var result struct {
		Calculated int64
	}
	err := db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN entry_type = 'debit'  THEN amount ELSE 0 END), 0) AS calculated
		FROM transaction_histories
		WHERE account_id = ?
	`, accountID).Scan(&result).Error
	return result.Calculated, err
}
