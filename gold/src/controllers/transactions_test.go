package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/gold/pkg/jwt"
	"github.com/getsnowflake/snowflake/gold/pkg/validator"
	"github.com/getsnowflake/snowflake/gold/src/controllers"
	"github.com/getsnowflake/snowflake/gold/src/dto"
	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/repository"
	"github.com/getsnowflake/snowflake/gold/src/service"
	"github.com/getsnowflake/snowflake/gold/test/mocks"
	"github.com/gofiber/fiber/v2"
	jwtLib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	validator.InitValidator(mocks.TestLogger())
}

func claimsMiddleware(c *fiber.Ctx) error {
	subject := c.Get("X-User-ID")
	c.Locals("claims", &jwt.Claims{
		RegisteredClaims: jwtLib.RegisteredClaims{
			Subject: subject,
			Issuer:  "snowflake-test",
		},
	})
	return c.Next()
}

func setupRealApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Account{}, &models.Transaction{},
		&models.TransactionHistory{}, &models.Flake{},
		&models.OutboxEvent{},
	))

	repos := repository.NewFactory()
	flakeSvc := service.NewFlakeService(repos)
	svcFactory := &service.ServiceFactory{}
	svcFactory.Flake = flakeSvc
	svcFactory.Transaction = service.NewTransactionService(repos, svcFactory, mocks.TestLogger())
	svcFactory.Balance = service.NewBalanceService(repos)

	txCtl := controllers.NewTransactionsController(db, nil, svcFactory.Transaction)
	balCtl := controllers.NewBalanceController(db, nil, svcFactory.Balance)

	app := fiber.New()

	app.Post("/transactions", claimsMiddleware, txCtl.CreateTransaction)
	app.Get("/transactions", claimsMiddleware, txCtl.GetTransactions)
	app.Get("/transactions/:id", claimsMiddleware, txCtl.GetTransactionByID)
	app.Post("/transactions/deposit", claimsMiddleware, txCtl.Deposit)
	app.Get("/balance", claimsMiddleware, balCtl.GetBalance)
	app.Get("/balance/history", claimsMiddleware, balCtl.GetBalanceHistory)

	return app, db
}

func seedAccountDB(t *testing.T, db *gorm.DB, userID uuid.UUID, balance int64) *models.Account {
	t.Helper()
	acc := &models.Account{
		UserID:               userID,
		Balance:              balance,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}
	require.NoError(t, db.Create(acc).Error)
	return acc
}

func makeJSON(v interface{}) *strings.Reader {
	b, _ := json.Marshal(v)
	return strings.NewReader(string(b))
}

func seedTxDB(t *testing.T, db *gorm.DB, tx *models.Transaction) *models.Transaction {
	t.Helper()
	if tx.IdempotencyKey == uuid.Nil {
		tx.IdempotencyKey = uuid.New()
	}
	if tx.Type == "" {
		tx.Type = models.TransactionTypeTransfer
	}
	if tx.Status == "" {
		tx.Status = models.TransactionStatusCompleted
	}
	require.NoError(t, db.Create(tx).Error)
	return tx
}

func getTransactions(t *testing.T, app *fiber.App, userID, query string) dto.PaginatedTransactionsResponse {
	t.Helper()
	req := httptest.NewRequest("GET", "/transactions"+query, nil)
	req.Header.Set("X-User-ID", userID)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var page dto.PaginatedTransactionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	return page
}

func TestGetTransactions_DirectionDebit_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)
	other := uuid.New()

	sent := seedTxDB(t, db, &models.Transaction{
		Amount:            1000,
		SenderAccountID:   &account.ID,
		ReceiverAccountID: &other,
	})
	seedTxDB(t, db, &models.Transaction{
		Type:              models.TransactionTypeDeposit,
		Amount:            2000,
		ReceiverAccountID: &account.ID,
	})

	page := getTransactions(t, app, account.UserID.String(), "?direction=debit")
	assert.Equal(t, int64(1), page.Total)
	require.Len(t, page.Transactions, 1)
	assert.Equal(t, sent.ID, page.Transactions[0].ID)
}

func TestGetTransactions_DirectionCredit_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)
	other := uuid.New()

	seedTxDB(t, db, &models.Transaction{
		Amount:            1000,
		SenderAccountID:   &account.ID,
		ReceiverAccountID: &other,
	})
	received := seedTxDB(t, db, &models.Transaction{
		Type:              models.TransactionTypeDeposit,
		Amount:            2000,
		ReceiverAccountID: &account.ID,
	})

	page := getTransactions(t, app, account.UserID.String(), "?direction=credit")
	assert.Equal(t, int64(1), page.Total)
	require.Len(t, page.Transactions, 1)
	assert.Equal(t, received.ID, page.Transactions[0].ID)
}

func TestGetTransactions_NoDirection_ReturnsBoth_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)
	other := uuid.New()

	seedTxDB(t, db, &models.Transaction{
		Amount:            1000,
		SenderAccountID:   &account.ID,
		ReceiverAccountID: &other,
	})
	seedTxDB(t, db, &models.Transaction{
		Type:              models.TransactionTypeDeposit,
		Amount:            2000,
		ReceiverAccountID: &account.ID,
	})

	page := getTransactions(t, app, account.UserID.String(), "")
	assert.Equal(t, int64(2), page.Total)
	assert.Len(t, page.Transactions, 2)
}

func TestGetTransactions_DateWindow_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	seedTxDB(t, db, &models.Transaction{
		Type: models.TransactionTypeDeposit, Amount: 1000,
		ReceiverAccountID: &account.ID,
		CreatedAt:         time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	})
	mid := seedTxDB(t, db, &models.Transaction{
		Type: models.TransactionTypeDeposit, Amount: 2000,
		ReceiverAccountID: &account.ID,
		CreatedAt:         time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	})
	seedTxDB(t, db, &models.Transaction{
		Type: models.TransactionTypeDeposit, Amount: 3000,
		ReceiverAccountID: &account.ID,
		CreatedAt:         time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
	})

	page := getTransactions(t, app, account.UserID.String(), "?start_date=2026-06-05&end_date=2026-06-15")
	assert.Equal(t, int64(1), page.Total)
	require.Len(t, page.Transactions, 1)
	assert.Equal(t, mid.ID, page.Transactions[0].ID)
}

func TestGetTransactions_EndDateInclusive_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	lateDay := seedTxDB(t, db, &models.Transaction{
		Type: models.TransactionTypeDeposit, Amount: 1000,
		ReceiverAccountID: &account.ID,
		CreatedAt:         time.Date(2026, 6, 15, 23, 30, 0, 0, time.UTC),
	})

	page := getTransactions(t, app, account.UserID.String(), "?end_date=2026-06-15")
	assert.Equal(t, int64(1), page.Total)
	require.Len(t, page.Transactions, 1)
	assert.Equal(t, lateDay.ID, page.Transactions[0].ID)
}

func TestGetTransactions_InvalidDate_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	req := httptest.NewRequest("GET", "/transactions?start_date=06-2026-15", nil)
	req.Header.Set("X-User-ID", account.UserID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetTransactions_StartAfterEnd_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	req := httptest.NewRequest("GET", "/transactions?start_date=2026-06-20&end_date=2026-06-10", nil)
	req.Header.Set("X-User-ID", account.UserID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetTransactions_InvalidDirection_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	req := httptest.NewRequest("GET", "/transactions?direction=sideways", nil)
	req.Header.Set("X-User-ID", account.UserID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTransaction_Success_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	sender := seedAccountDB(t, db, uuid.New(), 100000)
	receiver := seedAccountDB(t, db, uuid.New(), 50000)
	flake := &models.Flake{
		AccountID: receiver.ID,
		KeyType:   models.FlakeTypeEmail,
		KeyValue:  "receiver@test.com",
		Status:    models.FlakeStatusActive,
	}
	require.NoError(t, db.Create(flake).Error)

	req := httptest.NewRequest("POST", "/transactions", makeJSON(dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", sender.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var txResp dto.TransactionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&txResp))
	assert.Equal(t, models.TransactionStatusCompleted, txResp.Status)
	assert.Equal(t, int64(5000), txResp.Amount)
	assert.NotEqual(t, uuid.Nil, txResp.ID)
}

func TestCreateTransaction_InsufficientFunds_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	sender := seedAccountDB(t, db, uuid.New(), 100)
	receiver := seedAccountDB(t, db, uuid.New(), 50000)
	flake := &models.Flake{
		AccountID: receiver.ID,
		KeyType:   models.FlakeTypeEmail,
		KeyValue:  "receiver@test.com",
		Status:    models.FlakeStatusActive,
	}
	require.NoError(t, db.Create(flake).Error)

	req := httptest.NewRequest("POST", "/transactions", makeJSON(dto.CreateTransferRequest{
		ReceiverFlakeKey: "receiver@test.com",
		Amount:           5000,
		IdempotencyKey:   uuid.New(),
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", sender.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetTransactions_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	req := httptest.NewRequest("GET", "/transactions?page=1&limit=10", nil)
	req.Header.Set("X-User-ID", account.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var page dto.PaginatedTransactionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	assert.Equal(t, int64(0), page.Total)
}

func TestDeposit_Success_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 0)

	req := httptest.NewRequest("POST", "/transactions/deposit", makeJSON(dto.DepositRequest{
		AccountID:      account.ID,
		Amount:         10000,
		IdempotencyKey: uuid.New(),
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", account.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var txResp dto.TransactionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&txResp))
	assert.Equal(t, models.TransactionTypeDeposit, txResp.Type)
	assert.Equal(t, int64(10000), txResp.Amount)
	assert.NotEqual(t, uuid.Nil, txResp.ID)
}

func TestDeposit_Forbidden_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	owner := seedAccountDB(t, db, uuid.New(), 100000)
	other := seedAccountDB(t, db, uuid.New(), 0)

	req := httptest.NewRequest("POST", "/transactions/deposit", makeJSON(dto.DepositRequest{
		AccountID:      owner.ID,
		Amount:         10000,
		IdempotencyKey: uuid.New(),
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", other.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetBalance_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 5000)

	req := httptest.NewRequest("GET", "/balance", nil)
	req.Header.Set("X-User-ID", account.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var bal dto.BalanceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&bal))
	assert.Equal(t, account.ID, bal.AccountID)
	assert.Equal(t, int64(5000), bal.Balance)
}

func TestGetBalanceHistory_HTTP(t *testing.T) {
	app, db := setupRealApp(t)
	account := seedAccountDB(t, db, uuid.New(), 100000)

	req := httptest.NewRequest("GET", "/balance/history?page=1&limit=10", nil)
	req.Header.Set("X-User-ID", account.UserID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var hist dto.BalanceHistoryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&hist))
	assert.Equal(t, int64(0), hist.Total)
}
