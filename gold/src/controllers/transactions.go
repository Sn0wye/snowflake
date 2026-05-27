package controllers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/events"
	"github.com/Sn0wye/snowflake/gold/pkg/exceptions"
	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/src/dto"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/Sn0wye/snowflake/gold/src/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	minTransactionAmount  int64 = 1          // 1 cent
	maxTransactionAmount  int64 = 10_000_000 // R$ 100,000.00
	dailyTransactionLimit int64 = 5_000_000  // R$ 50,000.00
	maxDailyTransactions  int64 = 100

	defaultTransactionLimit = 20
	maxTransactionLimit     = 100
)

type TransactionsController interface {
	GetTransactions(ctx *fiber.Ctx) error
	GetTransactionByID(ctx *fiber.Ctx) error
	CreateTransaction(ctx *fiber.Ctx) error
	Deposit(ctx *fiber.Ctx) error
}

type transactionsController struct {
	db  *gorm.DB
	jwt *jwt.JWT
	rmq *messaging.MessagingService
	log *logger.Logger
}

func NewTransactionsController(db *gorm.DB, jwt *jwt.JWT, rmq *messaging.MessagingService, log *logger.Logger) TransactionsController {
	return &transactionsController{
		db:  db,
		jwt: jwt,
		rmq: rmq,
		log: log,
	}
}

// GetTransactions godoc
//
//	@Summary		/account/transactions
//	@Description	Get user account transactions with optional filters and pagination
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int									false	"Page number (default: 1)"
//	@Param			limit	query		int									false	"Items per page, max 100 (default: 20)"
//	@Param			status	query		string								false	"Filter by transaction status (pending, completed)"
//	@Param			type	query		string								false	"Filter by transaction type (transfer, deposit)"
//	@Success		200		{object}	dto.PaginatedTransactionsResponse	"PaginatedTransactionsResponse"
//	@Failure		404		{object}	exceptions.NotFoundError			"Account not found"
//	@Failure		500		{object}	exceptions.InternalServerError		"Failed to fetch transactions"
//	@Security		Bearer
//	@Router			/account/transactions [get]
//	@OperationId	getTransactions
func (s *transactionsController) GetTransactions(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	page := 1
	limit := defaultTransactionLimit
	if p := ctx.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := ctx.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > maxTransactionLimit {
				v = maxTransactionLimit
			}
			limit = v
		}
	}

	offset := (page - 1) * limit

	query := s.db.Model(&models.Transaction{}).
		Where("sender_account_id = ? OR receiver_account_id = ?", account.ID, account.ID)

	// Optional filters
	if status := ctx.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if txType := ctx.Query("type"); txType != "" {
		query = query.Where("type = ?", txType)
	}

	var total int64
	query.Count(&total)

	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return exceptions.InternalServer(ctx, "Failed to fetch transactions")
	}

	response := make([]dto.TransactionResponse, len(transactions))
	for i, t := range transactions {
		response[i] = dto.TransactionToResponse(t)
	}

	return ctx.Status(http.StatusOK).JSON(dto.PaginatedTransactionsResponse{
		Transactions: response,
		Total:        total,
		Page:         page,
		Limit:        limit,
	})
}

// GetTransactionByID godoc
//
//	@Summary		/account/transactions/{id}
//	@Description	Get a specific transaction by ID (user must be sender or receiver)
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string							true	"Transaction ID"
//	@Success		200	{object}	dto.TransactionResponse			"TransactionResponse"
//	@Failure		404	{object}	exceptions.NotFoundError		"Account or transaction not found"
//	@Failure		500	{object}	exceptions.InternalServerError	"Failed to fetch transaction"
//	@Security		Bearer
//	@Router			/account/transactions/{id} [get]
//	@OperationId	getTransactionByID
func (s *transactionsController) GetTransactionByID(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)
	id := ctx.Params("id")

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	var transaction models.Transaction
	if err := s.db.Where("id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exceptions.NotFound(ctx, "Transaction not found")
		}
		return exceptions.InternalServer(ctx, "Failed to fetch transaction")
	}

	// Verify the caller is sender or receiver
	isSender := transaction.SenderAccountID != nil && *transaction.SenderAccountID == account.ID
	isReceiver := transaction.ReceiverAccountID != nil && *transaction.ReceiverAccountID == account.ID
	if !isSender && !isReceiver {
		return exceptions.NotFound(ctx, "Transaction not found")
	}

	return ctx.Status(http.StatusOK).JSON(dto.TransactionToResponse(transaction))
}

// CreateTransaction godoc
//
//	@Summary		/account/transactions
//	@Description	Create a new transaction/transfer between accounts
//	@Description	Emits: `transaction.completed` event upon successful transfer completion.
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateTransferRequest		true	"Create Transfer Request"
//	@Success		201		{object}	dto.TransactionResponse			"TransactionResponse"
//	@Failure		400		{object}	exceptions.BadRequestError		"Invalid amount, account status, or limits exceeded"
//	@Failure		404		{object}	exceptions.NotFoundError		"Account or receiver flake key not found"
//	@Failure		500		{object}	exceptions.InternalServerError	"Transfer failed"
//	@Security		Bearer
//	@Router			/account/transactions [post]
//	@OperationId	createTransaction
func (s *transactionsController) CreateTransaction(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	body := new(dto.CreateTransferRequest)
	if err := utils.ParseRequest(ctx, body); err != nil {
		return err
	}

	// Validate amount range
	if body.Amount < minTransactionAmount {
		return exceptions.BadRequest(ctx, "Amount must be at least 1 centavo")
	}
	if body.Amount > maxTransactionAmount {
		return exceptions.BadRequest(ctx, "Amount exceeds maximum transaction limit of R$ 100,000.00")
	}

	// Idempotency: return existing transaction if key already used
	var existing models.Transaction
	if err := s.db.Where("idempotency_key = ?", body.IdempotencyKey).First(&existing).Error; err == nil {
		return ctx.Status(http.StatusOK).JSON(dto.TransactionToResponse(existing))
	}

	// Resolve sender account
	var senderAccount models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&senderAccount).Error; err != nil {
		return exceptions.NotFound(ctx, "Sender account not found")
	}

	if senderAccount.Status != models.AccountStatusActive {
		return exceptions.BadRequest(ctx, "Sender account is not active")
	}
	if senderAccount.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
		return exceptions.BadRequest(ctx, "Sender account is under reconciliation review")
	}

	// Resolve receiver via flake key
	var flake models.Flake
	if err := s.db.Where("key_value = ? AND status = ?", body.ReceiverFlakeKey, models.FlakeStatusActive).
		First(&flake).Error; err != nil {
		return exceptions.NotFound(ctx, "Receiver flake key not found or inactive")
	}

	var receiverAccount models.Account
	if err := s.db.Where("id = ?", flake.AccountID).First(&receiverAccount).Error; err != nil {
		return exceptions.NotFound(ctx, "Receiver account not found")
	}

	if receiverAccount.Status != models.AccountStatusActive {
		return exceptions.BadRequest(ctx, "Receiver account is not active")
	}

	// Cannot send to yourself
	if senderAccount.ID == receiverAccount.ID {
		return exceptions.BadRequest(ctx, "Cannot transfer to your own account")
	}

	// Check daily limits
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
	var dailyAmount int64
	s.db.Model(&models.Transaction{}).
		Where("sender_account_id = ? AND status = ? AND created_at >= ?",
			senderAccount.ID, models.TransactionStatusCompleted, startOfDay).
		Select("COALESCE(SUM(amount), 0)").Scan(&dailyAmount)

	if dailyAmount+body.Amount > dailyTransactionLimit {
		return exceptions.BadRequest(ctx, "Daily transaction limit of R$ 50,000.00 exceeded")
	}

	var dailyCount int64
	s.db.Model(&models.Transaction{}).
		Where("sender_account_id = ? AND status = ? AND created_at >= ?",
			senderAccount.ID, models.TransactionStatusCompleted, startOfDay).
		Count(&dailyCount)
	if dailyCount >= maxDailyTransactions {
		return exceptions.BadRequest(ctx, "Maximum daily transaction count (100) reached")
	}

	// Execute within a DB transaction with row locking
	var result *models.Transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Lock accounts in consistent UUID order to prevent deadlock
		ids := []uuid.UUID{senderAccount.ID, receiverAccount.ID}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i].String() < ids[j].String()
		})

		locked := make(map[uuid.UUID]*models.Account)
		for _, id := range ids {
			var acc models.Account
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", id).First(&acc).Error; err != nil {
				return err
			}
			locked[id] = &acc
		}

		sender := locked[senderAccount.ID]
		receiver := locked[receiverAccount.ID]

		// Re-check balance after lock
		if sender.Balance < body.Amount {
			return errors.New("insufficient funds")
		}

		now := time.Now()

		// Create transaction record
		transaction := models.Transaction{
			Type:              models.TransactionTypeTransfer,
			Status:            models.TransactionStatusPending,
			Amount:            body.Amount,
			SenderAccountID:   &sender.ID,
			ReceiverAccountID: &receiver.ID,
			FlakeKeyUsed:      body.ReceiverFlakeKey,
			Description:       body.Description,
			IdempotencyKey:    body.IdempotencyKey,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Debit entry for sender
		if err := tx.Create(&models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     sender.ID,
			EntryType:     models.TransactionHistoryEntryTypeDebit,
			Amount:        body.Amount,
			BalanceBefore: sender.Balance,
			BalanceAfter:  sender.Balance - body.Amount,
			Description:   body.Description,
		}).Error; err != nil {
			return err
		}

		// Credit entry for receiver
		if err := tx.Create(&models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     receiver.ID,
			EntryType:     models.TransactionHistoryEntryTypeCredit,
			Amount:        body.Amount,
			BalanceBefore: receiver.Balance,
			BalanceAfter:  receiver.Balance + body.Amount,
			Description:   body.Description,
		}).Error; err != nil {
			return err
		}

		// Update balances
		if err := tx.Model(sender).Update("balance", gorm.Expr("balance - ?", body.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(receiver).Update("balance", gorm.Expr("balance + ?", body.Amount)).Error; err != nil {
			return err
		}

		// Mark completed
		transaction.Status = models.TransactionStatusCompleted
		transaction.CompletedAt = &now
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}

		result = &transaction
		return nil
	})

	if err != nil {
		if err.Error() == "insufficient funds" {
			return exceptions.BadRequest(ctx, "Insufficient funds")
		}
		return exceptions.InternalServer(ctx, "Transfer failed")
	}

	s.publishCompleted(result)

	return ctx.Status(http.StatusCreated).JSON(dto.TransactionToResponse(*result))
}

// publishCompleted fires a transaction.completed event asynchronously.
// Failures are logged but do not affect the HTTP response.
func (s *transactionsController) publishCompleted(t *models.Transaction) {
	go func() {
		var senderStr, receiverStr *string
		if t.SenderAccountID != nil {
			v := t.SenderAccountID.String()
			senderStr = &v
		}
		if t.ReceiverAccountID != nil {
			v := t.ReceiverAccountID.String()
			receiverStr = &v
		}

		completedAt := time.Now()
		if t.CompletedAt != nil {
			completedAt = *t.CompletedAt
		}

		evt := events.NewTransactionCompleted(
			t.ID.String(),
			string(t.Type),
			t.Amount,
			senderStr,
			receiverStr,
			completedAt,
		)

		payload, err := events.Marshal(evt)
		if err != nil {
			s.log.Error("Failed to marshal transaction.completed event", zap.Error(err), zap.String("transactionID", t.ID.String()))
			return
		}

		if err := s.rmq.Produce(events.QueueTransactionCompleted, payload); err != nil {
			s.log.Error("Failed to publish transaction.completed event", zap.Error(err), zap.String("transactionID", t.ID.String()))
		}
	}()
}

// Deposit godoc
//
//	@Summary		/account/transactions/deposit
//	@Description	Deposit funds into an account
//	@Description	Emits: `transaction.completed` event upon successful deposit completion.
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.DepositRequest				true	"Deposit Request"
//	@Success		201		{object}	dto.TransactionResponse			"TransactionResponse"
//	@Failure		400		{object}	exceptions.BadRequestError		"Invalid amount or account status"
//	@Failure		403		{object}	exceptions.ForbiddenError		"Account does not belong to authenticated user"
//	@Failure		404		{object}	exceptions.NotFoundError		"Account not found"
//	@Failure		500		{object}	exceptions.InternalServerError	"Deposit failed"
//	@Router			/account/transactions/deposit [post]
//	@OperationId	deposit
func (s *transactionsController) Deposit(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	body := new(dto.DepositRequest)
	if err := utils.ParseRequest(ctx, body); err != nil {
		return err
	}

	if body.Amount < minTransactionAmount {
		return exceptions.BadRequest(ctx, "Amount must be at least 1 centavo")
	}
	if body.Amount > maxTransactionAmount {
		return exceptions.BadRequest(ctx, "Amount exceeds maximum deposit limit of R$ 100,000.00")
	}

	// Idempotency check
	var existing models.Transaction
	if err := s.db.Where("idempotency_key = ?", body.IdempotencyKey).First(&existing).Error; err == nil {
		return ctx.Status(http.StatusOK).JSON(dto.TransactionToResponse(existing))
	}

	var account models.Account
	if err := s.db.Where("id = ?", body.AccountID).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return exceptions.BadRequest(ctx, "Invalid user ID in token")
	}
	if account.UserID != userID {
		return exceptions.Forbidden(ctx, "You can only deposit to your own account")
	}
	if account.Status != models.AccountStatusActive {
		return exceptions.BadRequest(ctx, "Account is not active")
	}

	var result *models.Transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var acc models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", account.ID).First(&acc).Error; err != nil {
			return err
		}

		now := time.Now()

		transaction := models.Transaction{
			Type:              models.TransactionTypeDeposit,
			Status:            models.TransactionStatusPending,
			Amount:            body.Amount,
			ReceiverAccountID: &acc.ID,
			Description:       body.Description,
			IdempotencyKey:    body.IdempotencyKey,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Credit entry only (deposit from external source)
		if err := tx.Create(&models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     acc.ID,
			EntryType:     models.TransactionHistoryEntryTypeCredit,
			Amount:        body.Amount,
			BalanceBefore: acc.Balance,
			BalanceAfter:  acc.Balance + body.Amount,
			Description:   body.Description,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&acc).Update("balance", gorm.Expr("balance + ?", body.Amount)).Error; err != nil {
			return err
		}

		transaction.Status = models.TransactionStatusCompleted
		transaction.CompletedAt = &now
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}

		result = &transaction
		return nil
	})

	if err != nil {
		return exceptions.InternalServer(ctx, "Deposit failed")
	}

	s.publishCompleted(result)

	return ctx.Status(http.StatusCreated).JSON(dto.TransactionToResponse(*result))
}
