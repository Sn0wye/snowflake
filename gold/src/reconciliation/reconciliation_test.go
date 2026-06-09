package reconciliation_test

import (
	"database/sql"
	"testing"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/reconciliation"
	"github.com/getsnowflake/snowflake/gold/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReconDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Account{}, &models.TransactionHistory{}))
	return db
}

type reconRow struct {
	ReconciliationStatus sql.NullString
	LastReconciledAt     sql.NullString
	DiscrepancyDetectedAt sql.NullString
	DiscrepancyAmount    sql.NullInt64
}

func queryReconRow(t *testing.T, db *gorm.DB, id uuid.UUID) reconRow {
	t.Helper()
	var row reconRow
	err := db.Raw(`
		SELECT reconciliation_status, last_reconciled_at,
			   discrepancy_detected_at, discrepancy_amount
		FROM accounts WHERE id = ?
	`, id).Row().Scan(&row.ReconciliationStatus, &row.LastReconciledAt,
		&row.DiscrepancyDetectedAt, &row.DiscrepancyAmount)
	require.NoError(t, err)
	return row
}

func TestReconcile_ZeroActivity_SetsLastReconciledAt(t *testing.T) {
	db := setupReconDB(t)
	acc := &models.Account{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		Balance:              0,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}
	require.NoError(t, db.Create(acc).Error)

	job := reconciliation.NewJob(db, mocks.TestLogger())
	job.Run()

	row := queryReconRow(t, db, acc.ID)
	assert.Equal(t, "ok", row.ReconciliationStatus.String)
	assert.True(t, row.LastReconciledAt.Valid)
}

func TestReconcile_BalanceMatches_UpdatesReconciledAt(t *testing.T) {
	db := setupReconDB(t)
	acc := &models.Account{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		Balance:              1000,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}
	require.NoError(t, db.Create(acc).Error)
	require.NoError(t, db.Create(&models.TransactionHistory{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		AccountID:     acc.ID,
		EntryType:     models.TransactionHistoryEntryTypeCredit,
		Amount:        1000,
		BalanceBefore: 0,
		BalanceAfter:  1000,
	}).Error)

	job := reconciliation.NewJob(db, mocks.TestLogger())
	job.Run()

	row := queryReconRow(t, db, acc.ID)
	assert.Equal(t, "ok", row.ReconciliationStatus.String)
	assert.True(t, row.LastReconciledAt.Valid)
}

func TestReconcile_DiscrepancyDetected_FlagsAccount(t *testing.T) {
	db := setupReconDB(t)
	acc := &models.Account{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		Balance:              500,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}
	require.NoError(t, db.Create(acc).Error)
	require.NoError(t, db.Create(&models.TransactionHistory{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		AccountID:     acc.ID,
		EntryType:     models.TransactionHistoryEntryTypeCredit,
		Amount:        1000,
		BalanceBefore: 0,
		BalanceAfter:  1000,
	}).Error)

	job := reconciliation.NewJob(db, mocks.TestLogger())
	job.Run()

	row := queryReconRow(t, db, acc.ID)
	assert.Equal(t, "discrepancy", row.ReconciliationStatus.String)
	assert.True(t, row.DiscrepancyDetectedAt.Valid)
	assert.True(t, row.DiscrepancyAmount.Valid)
	assert.Equal(t, int64(500), row.DiscrepancyAmount.Int64)
	assert.True(t, row.LastReconciledAt.Valid)
}

func TestReconcile_PreviouslyFlagged_ResolvesToOK(t *testing.T) {
	db := setupReconDB(t)
	acc := &models.Account{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		Balance:              1000,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusDiscrepancy,
	}
	require.NoError(t, db.Create(acc).Error)
	require.NoError(t, db.Create(&models.TransactionHistory{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		AccountID:     acc.ID,
		EntryType:     models.TransactionHistoryEntryTypeCredit,
		Amount:        1000,
		BalanceBefore: 0,
		BalanceAfter:  1000,
	}).Error)

	job := reconciliation.NewJob(db, mocks.TestLogger())
	job.Run()

	row := queryReconRow(t, db, acc.ID)
	assert.Equal(t, "ok", row.ReconciliationStatus.String)
	assert.False(t, row.DiscrepancyDetectedAt.Valid)
	assert.False(t, row.DiscrepancyAmount.Valid)
	assert.True(t, row.LastReconciledAt.Valid)
}
