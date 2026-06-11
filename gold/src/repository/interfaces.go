package repository

import (
	"time"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionFilter struct {
	Status string
	Type   string
	Page   int
	Limit  int
}

type AccountRepository interface {
	FindByUserID(db *gorm.DB, userID string) (*models.Account, error)
	FindByID(db *gorm.DB, id uuid.UUID) (*models.Account, error)
	FindByIDForUpdate(db *gorm.DB, id uuid.UUID) (*models.Account, error)
	Create(db *gorm.DB, account *models.Account) error
	Update(db *gorm.DB, account *models.Account) error
	All(db *gorm.DB) ([]models.Account, error)
	UpdateReconciliationFields(db *gorm.DB, account *models.Account, fields map[string]interface{}) error
	DebitBalance(db *gorm.DB, accountID uuid.UUID, amount int64) error
	CreditBalance(db *gorm.DB, accountID uuid.UUID, amount int64) error
}

type TransactionRepository interface {
	FindByID(db *gorm.DB, id uuid.UUID) (*models.Transaction, error)
	FindByIdempotencyKey(db *gorm.DB, key uuid.UUID) (*models.Transaction, error)
	FindByAccountID(db *gorm.DB, accountID uuid.UUID, filter TransactionFilter) ([]models.Transaction, int64, error)
	SumDailyAmount(db *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error)
	CountDaily(db *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error)
	Create(db *gorm.DB, transaction *models.Transaction) error
	Save(db *gorm.DB, transaction *models.Transaction) error
}

type TransactionHistoryRepository interface {
	Create(db *gorm.DB, history *models.TransactionHistory) error
	FindByAccountID(db *gorm.DB, accountID uuid.UUID, page, limit int) ([]models.TransactionHistory, int64, error)
	SumLedgerBalance(db *gorm.DB, accountID uuid.UUID) (int64, error)
}

type FlakeRepository interface {
	FindByAccountID(db *gorm.DB, accountID uuid.UUID) ([]models.Flake, error)
	FindByKeyValue(db *gorm.DB, keyValue string) (*models.Flake, error)
	FindByIDAndAccount(db *gorm.DB, id uuid.UUID, accountID uuid.UUID) (*models.Flake, error)
	FindByTypeAndAccount(db *gorm.DB, keyType models.FlakeType, accountID uuid.UUID) (*models.Flake, error)
	CountNonUnlimited(db *gorm.DB, accountID uuid.UUID) (int64, error)
	Create(db *gorm.DB, flake *models.Flake) error
	UpdateStatus(db *gorm.DB, flake *models.Flake, status models.FlakeStatus) error
}

type OutboxRepository interface {
	Create(db *gorm.DB, entry *models.OutboxEvent) error
	ClaimPending(db *gorm.DB, limit int) ([]models.OutboxEvent, error)
	MarkPublished(db *gorm.DB, id uuid.UUID) error
	MarkFailed(db *gorm.DB, id uuid.UUID, errMsg string) error
}

type Factory struct {
	Account            AccountRepository
	Transaction        TransactionRepository
	TransactionHistory TransactionHistoryRepository
	Flake              FlakeRepository
	Outbox             OutboxRepository
}

func NewFactory() *Factory {
	return &Factory{
		Account:            NewAccountRepo(),
		Transaction:        NewTransactionRepo(),
		TransactionHistory: NewTransactionHistoryRepo(),
		Flake:              NewFlakeRepo(),
		Outbox:             NewOutboxRepo(),
	}
}
