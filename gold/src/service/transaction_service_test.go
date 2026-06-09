package service_test

import (
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/gold/src/dto"
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

func setupMockService() (*mocks.MockAccountRepo, *mocks.MockTransactionRepo, *mocks.MockTransactionHistoryRepo, *mocks.MockFlakeRepo, *gorm.DB, *repository.Factory, service.TransactionService) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	accRepo := mocks.NewMockAccountRepo()
	txRepo := mocks.NewMockTransactionRepo()
	histRepo := mocks.NewMockTransactionHistoryRepo()
	flakeRepo := mocks.NewMockFlakeRepo()
	repos := &repository.Factory{
		Account:            accRepo,
		Transaction:        txRepo,
		TransactionHistory: histRepo,
		Flake:              flakeRepo,
	}
	flakeSvc := service.NewFlakeService(repos)
	svcFactory := &service.ServiceFactory{}
	svcFactory.Flake = flakeSvc
	svc := service.NewTransactionService(repos, svcFactory, nil, mocks.TestLogger())
	return accRepo, txRepo, histRepo, flakeRepo, db, repos, svc
}

func makeAccount() *models.Account {
	return &models.Account{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		Balance:              100000,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}
}

func makeFlake(accountID uuid.UUID, keyValue string) *models.Flake {
	return &models.Flake{
		ID:        uuid.New(),
		AccountID: accountID,
		KeyType:   models.FlakeTypeEmail,
		KeyValue:  keyValue,
		Status:    models.FlakeStatusActive,
	}
}

func TestCreateTransfer_Success(t *testing.T) {
	accRepo, txRepo, histRepo, flakeRepo, db, _, svc := setupMockService()

	sender := makeAccount()
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		Description:      "Test transfer",
		IdempotencyKey:   uuid.New(),
	}

	resp, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	require.NoError(t, err)
	assert.Equal(t, req.Amount, resp.Amount)
	assert.Equal(t, models.TransactionStatusCompleted, resp.Status)
	assert.Equal(t, models.TransactionTypeTransfer, resp.Type)
	assert.Equal(t, sender.ID, *resp.SenderAccountID)
	assert.Equal(t, receiver.ID, *resp.ReceiverAccountID)

	updatedSender := accRepo.Lookup(sender.ID)
	require.NotNil(t, updatedSender)
	assert.Equal(t, int64(95000), updatedSender.Balance)

	updatedReceiver := accRepo.Lookup(receiver.ID)
	require.NotNil(t, updatedReceiver)
	assert.Equal(t, int64(105000), updatedReceiver.Balance)

	assert.Len(t, histRepo.All(), 2)

	var debit, credit *models.TransactionHistory
	entries := histRepo.All()
	for i, h := range entries {
		if h.EntryType == models.TransactionHistoryEntryTypeDebit {
			debit = &entries[i]
		} else {
			credit = &entries[i]
		}
	}
	require.NotNil(t, debit)
	require.NotNil(t, credit)
	assert.Equal(t, sender.ID, debit.AccountID)
	assert.Equal(t, int64(100000), debit.BalanceBefore)
	assert.Equal(t, int64(95000), debit.BalanceAfter)
	assert.Equal(t, int64(5000), debit.Amount)
	assert.Equal(t, receiver.ID, credit.AccountID)
	assert.Equal(t, int64(100000), credit.BalanceBefore)
	assert.Equal(t, int64(105000), credit.BalanceAfter)
	assert.Equal(t, int64(5000), credit.Amount)

	assert.Len(t, txRepo.All(), 1)
}

func TestCreateTransfer_InsufficientFunds(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	sender.Balance = 100
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrInsufficientFunds)
}

func TestCreateTransfer_Idempotent(t *testing.T) {
	accRepo, txRepo, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	idempotencyKey := uuid.New()
	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   idempotencyKey,
	}

	resp1, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	require.NoError(t, err)
	assert.Len(t, txRepo.All(), 1)

	_, err = svc.CreateTransaction(db, sender.UserID.String(), req)
	var idempotent *service.IdempotentTransactionError
	require.ErrorAs(t, err, &idempotent)
	assert.Equal(t, resp1.ID, idempotent.Response.ID)
	assert.Len(t, txRepo.All(), 1)
}

func TestCreateTransfer_SelfTransfer(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	flakeRepo.Seed(makeFlake(sender.ID, "self@test.com"))
	accRepo.Seed(sender)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "self@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrSelfTransfer)
}

func TestCreateTransfer_AmountTooLow(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           0,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrAmountTooLow)
}

func TestCreateTransfer_AmountTooHigh(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           10000001,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrAmountTooHigh)
}

func TestCreateTransfer_AccountNotActive(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	sender.Status = models.AccountStatusSuspended
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrAccountNotActive)
}

func TestCreateTransfer_ReconciliationDiscrepancy(t *testing.T) {
	accRepo, _, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	sender.ReconciliationStatus = models.AccountReconciliationStatusDiscrepancy
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrAccountReconciliation)
}

func TestCreateTransfer_DailyLimitExceeded(t *testing.T) {
	accRepo, txRepo, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	sender.Balance = 10000000
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	priorTx := &models.Transaction{
		ID:              uuid.New(),
		Type:            models.TransactionTypeTransfer,
		Status:          models.TransactionStatusCompleted,
		Amount:          5000000,
		SenderAccountID: &sender.ID,
		IdempotencyKey:  uuid.New(),
		CreatedAt:       time.Now(),
	}
	txRepo.Seed(priorTx)

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           100000,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrDailyLimitExceeded)
}

func TestCreateTransfer_DailyCountExceeded(t *testing.T) {
	accRepo, txRepo, _, flakeRepo, db, _, svc := setupMockService()
	sender := makeAccount()
	sender.Balance = 10000000
	receiver := makeAccount()
	flakeRepo.Seed(makeFlake(receiver.ID, "receiver@test.com"))
	accRepo.Seed(sender, receiver)

	for i := 0; i < 100; i++ {
		txRepo.Seed(&models.Transaction{
			ID:              uuid.New(),
			Type:            models.TransactionTypeTransfer,
			Status:          models.TransactionStatusCompleted,
			Amount:          100,
			SenderAccountID: &sender.ID,
			IdempotencyKey:  uuid.New(),
			CreatedAt:       time.Now(),
		})
	}

	req := dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           100,
		IdempotencyKey:   uuid.New(),
	}
	_, err := svc.CreateTransaction(db, sender.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrDailyCountExceeded)
}

func TestDeposit_Success(t *testing.T) {
	accRepo, txRepo, histRepo, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	req := dto.DepositRequest{
		AccountID:      account.ID,
		Amount:         5000,
		Description:    "Deposit",
		IdempotencyKey: uuid.New(),
	}

	resp, err := svc.Deposit(db, account.UserID.String(), req)
	require.NoError(t, err)
	assert.Equal(t, models.TransactionStatusCompleted, resp.Status)
	assert.Equal(t, models.TransactionTypeDeposit, resp.Type)
	assert.Equal(t, account.ID, *resp.ReceiverAccountID)

	updated := accRepo.Lookup(account.ID)
	require.NotNil(t, updated)
	assert.Equal(t, int64(105000), updated.Balance)

	entries := histRepo.All()
	assert.Len(t, entries, 1)
	assert.Equal(t, models.TransactionHistoryEntryTypeCredit, entries[0].EntryType)

	assert.Len(t, txRepo.All(), 1)
}

func TestDeposit_Forbidden(t *testing.T) {
	accRepo, _, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	req := dto.DepositRequest{
		AccountID:      account.ID,
		Amount:         5000,
		IdempotencyKey: uuid.New(),
	}
	_, err := svc.Deposit(db, uuid.New().String(), req)
	assert.ErrorIs(t, err, service.ErrForbidden)
}

func TestDeposit_Idempotent(t *testing.T) {
	accRepo, txRepo, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	idempotencyKey := uuid.New()
	req := dto.DepositRequest{
		AccountID:      account.ID,
		Amount:         5000,
		IdempotencyKey: idempotencyKey,
	}

	resp1, err := svc.Deposit(db, account.UserID.String(), req)
	require.NoError(t, err)

	_, err = svc.Deposit(db, account.UserID.String(), req)
	var idempotent *service.IdempotentTransactionError
	require.ErrorAs(t, err, &idempotent)
	assert.Equal(t, resp1.ID, idempotent.Response.ID)
	assert.Len(t, txRepo.All(), 1)
}

func TestGetTransactions_ByAccount(t *testing.T) {
	accRepo, txRepo, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	tx1 := &models.Transaction{
		ID:                uuid.New(),
		Type:              models.TransactionTypeTransfer,
		Status:            models.TransactionStatusCompleted,
		Amount:            1000,
		SenderAccountID:   &account.ID,
		ReceiverAccountID: ptrUUID(uuid.New()),
		IdempotencyKey:    uuid.New(),
	}
	tx2 := &models.Transaction{
		ID:                uuid.New(),
		Type:              models.TransactionTypeDeposit,
		Status:            models.TransactionStatusCompleted,
		Amount:            2000,
		ReceiverAccountID: &account.ID,
		IdempotencyKey:    uuid.New(),
	}
	txRepo.Seed(tx1, tx2)

	resp, err := svc.GetTransactions(db, account.UserID.String(), repository.TransactionFilter{Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.Transactions, 2)
}

func TestGetTransactions_NotFound(t *testing.T) {
	_, _, _, _, db, _, svc := setupMockService()
	_, err := svc.GetTransactions(db, uuid.New().String(), repository.TransactionFilter{Page: 1, Limit: 20})
	assert.ErrorIs(t, err, service.ErrAccountNotFound)
}

func TestGetTransactionByID_OwnTransaction(t *testing.T) {
	accRepo, txRepo, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	tx := &models.Transaction{
		ID:                uuid.New(),
		Type:              models.TransactionTypeTransfer,
		Status:            models.TransactionStatusCompleted,
		Amount:            1000,
		SenderAccountID:   &account.ID,
		ReceiverAccountID: ptrUUID(uuid.New()),
		IdempotencyKey:    uuid.New(),
	}
	txRepo.Seed(tx)

	resp, err := svc.GetTransactionByID(db, account.UserID.String(), tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.ID, resp.ID)
}

func TestGetTransactionByID_NotParticipant(t *testing.T) {
	accRepo, txRepo, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	other := makeAccount()
	accRepo.Seed(account, other)

	tx := &models.Transaction{
		ID:                uuid.New(),
		Type:              models.TransactionTypeTransfer,
		Status:            models.TransactionStatusCompleted,
		Amount:            1000,
		SenderAccountID:   ptrUUID(other.ID),
		ReceiverAccountID: ptrUUID(uuid.New()),
		IdempotencyKey:    uuid.New(),
	}
	txRepo.Seed(tx)

	_, err := svc.GetTransactionByID(db, account.UserID.String(), tx.ID)
	assert.ErrorIs(t, err, service.ErrTransactionNotFound)
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	accRepo, _, _, _, db, _, svc := setupMockService()
	account := makeAccount()
	accRepo.Seed(account)

	_, err := svc.GetTransactionByID(db, account.UserID.String(), uuid.New())
	assert.ErrorIs(t, err, service.ErrTransactionNotFound)
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
