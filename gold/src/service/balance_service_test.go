package service_test

import (
	"testing"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/repository"
	"github.com/getsnowflake/snowflake/gold/src/service"
	"github.com/getsnowflake/snowflake/gold/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBalanceService(t *testing.T) (*mocks.MockAccountRepo, *mocks.MockTransactionHistoryRepo, *gorm.DB, service.BalanceService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	accRepo := mocks.NewMockAccountRepo()
	histRepo := mocks.NewMockTransactionHistoryRepo()
	repos := &repository.Factory{
		Account:            accRepo,
		TransactionHistory: histRepo,
	}
	svc := service.NewBalanceService(repos)
	return accRepo, histRepo, db, svc
}

func TestGetBalance_Success(t *testing.T) {
	accRepo, _, db, svc := setupBalanceService(t)
	account := makeAccount()
	accRepo.Seed(account)

	resp, err := svc.GetBalance(db, account.UserID.String())
	require.NoError(t, err)
	assert.Equal(t, account.ID, resp.AccountID)
	assert.Equal(t, account.Balance, resp.Balance)
	assert.Equal(t, account.Status, resp.Status)
	assert.Equal(t, account.ReconciliationStatus, resp.ReconciliationStatus)
}

func TestGetBalance_NotFound(t *testing.T) {
	_, _, db, svc := setupBalanceService(t)
	_, err := svc.GetBalance(db, uuid.New().String())
	assert.ErrorIs(t, err, service.ErrAccountNotFound)
}

func TestGetBalanceHistory_Success(t *testing.T) {
	accRepo, histRepo, db, svc := setupBalanceService(t)
	account := makeAccount()
	accRepo.Seed(account)

	histRepo.Create(db, &models.TransactionHistory{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		AccountID:     account.ID,
		EntryType:     models.TransactionHistoryEntryTypeCredit,
		Amount:        1000,
		BalanceBefore: 0,
		BalanceAfter:  1000,
	})
	histRepo.Create(db, &models.TransactionHistory{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		AccountID:     account.ID,
		EntryType:     models.TransactionHistoryEntryTypeDebit,
		Amount:        200,
		BalanceBefore: 1000,
		BalanceAfter:  800,
	})

	resp, err := svc.GetBalanceHistory(db, account.UserID.String(), 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.Entries, 2)
}

func TestGetBalanceHistory_Empty(t *testing.T) {
	accRepo, _, db, svc := setupBalanceService(t)
	account := makeAccount()
	accRepo.Seed(account)

	resp, err := svc.GetBalanceHistory(db, account.UserID.String(), 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), resp.Total)
	assert.Empty(t, resp.Entries)
}

func TestGetBalanceHistory_NotFound(t *testing.T) {
	_, _, db, svc := setupBalanceService(t)
	_, err := svc.GetBalanceHistory(db, uuid.New().String(), 1, 20)
	assert.ErrorIs(t, err, service.ErrAccountNotFound)
}
