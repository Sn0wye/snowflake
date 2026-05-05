package reconciliation

import (
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Job reconciles cached account balances against the ledger.
type Job struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewJob(db *gorm.DB, log *logger.Logger) *Job {
	return &Job{db: db, log: log}
}

// Start runs the reconciliation on the given interval (non-blocking).
func (j *Job) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			j.Run()
		}
	}()
}

// Run reconciles all accounts once.
func (j *Job) Run() {
	j.log.Info("Reconciliation job started")

	var accounts []models.Account
	if err := j.db.Find(&accounts).Error; err != nil {
		j.log.Error("Reconciliation: failed to fetch accounts", zap.Error(err))
		return
	}

	matched, flagged, errors := 0, 0, 0

	for _, account := range accounts {
		if err := j.reconcileAccount(account); err != nil {
			j.log.Error("Reconciliation: error processing account",
				zap.String("accountID", account.ID.String()),
				zap.Error(err),
			)
			errors++
		} else if account.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
			flagged++
		} else {
			matched++
		}
	}

	j.log.Info("Reconciliation job completed",
		zap.Int("matched", matched),
		zap.Int("flagged", flagged),
		zap.Int("errors", errors),
	)
}

type ledgerSum struct {
	Calculated int64
}

func (j *Job) reconcileAccount(account models.Account) error {
	var result ledgerSum

	// Calculate true balance from ledger
	err := j.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN entry_type = 'debit'  THEN amount ELSE 0 END), 0) AS calculated
		FROM transaction_histories
		WHERE account_id = ?
	`, account.ID).Scan(&result).Error
	if err != nil {
		return err
	}

	calculated := result.Calculated

	if calculated == account.Balance {
		// Clear any previous discrepancy flag
		if account.ReconciliationStatus != models.AccountReconciliationStatusOK {
			now := time.Now()
			_ = now
			return j.db.Model(&account).Updates(map[string]interface{}{
				"reconciliation_status":   models.AccountReconciliationStatusOK,
				"discrepancy_detected_at": nil,
				"discrepancy_amount":      nil,
				"last_reconciled_at":      time.Now(),
			}).Error
		}
		return j.db.Model(&account).Update("last_reconciled_at", time.Now()).Error
	}

	// Discrepancy detected — flag the account, do NOT touch balance
	discrepancy := calculated - account.Balance
	now := time.Now()

	j.log.Warn("Reconciliation discrepancy detected",
		zap.String("accountID", account.ID.String()),
		zap.Int64("cachedBalance", account.Balance),
		zap.Int64("calculatedBalance", calculated),
		zap.Int64("discrepancyAmount", discrepancy),
	)

	return j.db.Model(&account).Updates(map[string]interface{}{
		"reconciliation_status":   models.AccountReconciliationStatusDiscrepancy,
		"discrepancy_detected_at": now,
		"discrepancy_amount":      discrepancy,
		"last_reconciled_at":      now,
	}).Error
}
