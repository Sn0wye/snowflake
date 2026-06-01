package service

import (
	"errors"
	"sort"
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/events"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/pkg/repository"
	"github.com/Sn0wye/snowflake/gold/src/dto"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	minTransactionAmount  int64 = 1
	maxTransactionAmount  int64 = 10_000_000
	dailyTransactionLimit int64 = 5_000_000
	maxDailyTransactions  int64 = 100
)

type TransactionService interface {
	GetTransactions(db *gorm.DB, userID string, filter repository.TransactionFilter) (dto.PaginatedTransactionsResponse, error)
	GetTransactionByID(db *gorm.DB, userID string, id uuid.UUID) (dto.TransactionResponse, error)
	CreateTransaction(db *gorm.DB, userID string, req dto.CreateTransferRequest) (dto.TransactionResponse, error)
	Deposit(db *gorm.DB, userID string, req dto.DepositRequest) (dto.TransactionResponse, error)
}

type transactionService struct {
	repos *repository.Factory
	forms *ServiceFactory
	rmq   *messaging.MessagingService
	log   *logger.Logger
}

func NewTransactionService(repos *repository.Factory, forms *ServiceFactory, rmq *messaging.MessagingService, log *logger.Logger) TransactionService {
	return &transactionService{repos: repos, forms: forms, rmq: rmq, log: log}
}

func (s *transactionService) GetTransactions(db *gorm.DB, userID string, filter repository.TransactionFilter) (dto.PaginatedTransactionsResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.PaginatedTransactionsResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	transactions, total, err := s.repos.Transaction.FindByAccountID(db, account.ID, filter)
	if err != nil {
		return dto.PaginatedTransactionsResponse{}, err
	}

	response := make([]dto.TransactionResponse, len(transactions))
	for i, t := range transactions {
		response[i] = dto.TransactionToResponse(t)
	}

	return dto.PaginatedTransactionsResponse{
		Transactions: response,
		Total:        total,
		Page:         filter.Page,
		Limit:        filter.Limit,
	}, nil
}

func (s *transactionService) GetTransactionByID(db *gorm.DB, userID string, id uuid.UUID) (dto.TransactionResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.TransactionResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	transaction, err := s.repos.Transaction.FindByID(db, id)
	if err != nil {
		return dto.TransactionResponse{}, err
	}

	isSender := transaction.SenderAccountID != nil && *transaction.SenderAccountID == account.ID
	isReceiver := transaction.ReceiverAccountID != nil && *transaction.ReceiverAccountID == account.ID
	if !isSender && !isReceiver {
		return dto.TransactionResponse{}, gorm.ErrRecordNotFound
	}

	return dto.TransactionToResponse(*transaction), nil
}

func (s *transactionService) CreateTransaction(db *gorm.DB, userID string, req dto.CreateTransferRequest) (dto.TransactionResponse, error) {
	if req.Amount < minTransactionAmount {
		return dto.TransactionResponse{}, ErrAmountTooLow
	}
	if req.Amount > maxTransactionAmount {
		return dto.TransactionResponse{}, ErrAmountTooHigh
	}

	existing, err := s.repos.Transaction.FindByIdempotencyKey(db, req.IdempotencyKey)
	if err == nil {
		return dto.TransactionResponse{}, &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.TransactionResponse{}, err
	}

	senderAccount, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.TransactionResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	if senderAccount.Status != models.AccountStatusActive {
		return dto.TransactionResponse{}, ErrAccountNotActive
	}
	if senderAccount.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
		return dto.TransactionResponse{}, ErrAccountReconciliation
	}

	receiverAccount, _, err := s.forms.Flake.ResolveReceiver(db, req.ReceiverFlakeKey)
	if err != nil {
		return dto.TransactionResponse{}, err
	}

	if receiverAccount.Status != models.AccountStatusActive {
		return dto.TransactionResponse{}, ErrAccountNotActive
	}

	if senderAccount.ID == receiverAccount.ID {
		return dto.TransactionResponse{}, ErrSelfTransfer
	}

	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
	dailyAmount, err := s.repos.Transaction.SumDailyAmount(db, senderAccount.ID, startOfDay)
	if err != nil {
		return dto.TransactionResponse{}, err
	}
	if dailyAmount+req.Amount > dailyTransactionLimit {
		return dto.TransactionResponse{}, ErrDailyLimitExceeded
	}

	dailyCount, err := s.repos.Transaction.CountDaily(db, senderAccount.ID, startOfDay)
	if err != nil {
		return dto.TransactionResponse{}, err
	}
	if dailyCount >= maxDailyTransactions {
		return dto.TransactionResponse{}, ErrDailyCountExceeded
	}

	var result *models.Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		ids := []uuid.UUID{senderAccount.ID, receiverAccount.ID}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i].String() < ids[j].String()
		})

		locked := make(map[uuid.UUID]*models.Account)
		for _, id := range ids {
			acc, err := s.repos.Account.FindByIDForUpdate(tx, id)
			if err != nil {
				return err
			}
			locked[id] = acc
		}

		sender := locked[senderAccount.ID]
		receiver := locked[receiverAccount.ID]

		if sender.Balance < req.Amount {
			return ErrInsufficientFunds
		}

		now := time.Now()

		transaction := models.Transaction{
			Type:              models.TransactionTypeTransfer,
			Status:            models.TransactionStatusPending,
			Amount:            req.Amount,
			SenderAccountID:   &sender.ID,
			ReceiverAccountID: &receiver.ID,
			FlakeKeyUsed:      req.ReceiverFlakeKey,
			Description:       req.Description,
			IdempotencyKey:    req.IdempotencyKey,
		}
		if err := s.repos.Transaction.Create(tx, &transaction); err != nil {
			return err
		}

		if err := s.repos.TransactionHistory.Create(tx, &models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     sender.ID,
			EntryType:     models.TransactionHistoryEntryTypeDebit,
			Amount:        req.Amount,
			BalanceBefore: sender.Balance,
			BalanceAfter:  sender.Balance - req.Amount,
			Description:   req.Description,
		}); err != nil {
			return err
		}

		if err := s.repos.TransactionHistory.Create(tx, &models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     receiver.ID,
			EntryType:     models.TransactionHistoryEntryTypeCredit,
			Amount:        req.Amount,
			BalanceBefore: receiver.Balance,
			BalanceAfter:  receiver.Balance + req.Amount,
			Description:   req.Description,
		}); err != nil {
			return err
		}

		if err := tx.Model(sender).Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(receiver).Update("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			return err
		}

		transaction.Status = models.TransactionStatusCompleted
		transaction.CompletedAt = &now
		if err := s.repos.Transaction.Save(tx, &transaction); err != nil {
			return err
		}

		result = &transaction
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return dto.TransactionResponse{}, ErrInsufficientFunds
		}
		return dto.TransactionResponse{}, ErrTransferFailed
	}

	s.publishCompleted(result)

	return dto.TransactionToResponse(*result), nil
}

func (s *transactionService) Deposit(db *gorm.DB, userID string, req dto.DepositRequest) (dto.TransactionResponse, error) {
	if req.Amount < minTransactionAmount {
		return dto.TransactionResponse{}, ErrAmountTooLow
	}
	if req.Amount > maxTransactionAmount {
		return dto.TransactionResponse{}, ErrAmountTooHigh
	}

	existing, err := s.repos.Transaction.FindByIdempotencyKey(db, req.IdempotencyKey)
	if err == nil {
		return dto.TransactionResponse{}, &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.TransactionResponse{}, err
	}

	account, err := s.repos.Account.FindByID(db, req.AccountID)
	if err != nil {
		return dto.TransactionResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		return dto.TransactionResponse{}, err
	}
	if account.UserID != userIDParsed {
		return dto.TransactionResponse{}, ErrForbidden
	}
	if account.Status != models.AccountStatusActive {
		return dto.TransactionResponse{}, ErrAccountNotActive
	}

	var result *models.Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		acc, err := s.repos.Account.FindByIDForUpdate(tx, account.ID)
		if err != nil {
			return err
		}

		now := time.Now()

		transaction := models.Transaction{
			Type:              models.TransactionTypeDeposit,
			Status:            models.TransactionStatusPending,
			Amount:            req.Amount,
			ReceiverAccountID: &acc.ID,
			Description:       req.Description,
			IdempotencyKey:    req.IdempotencyKey,
		}
		if err := s.repos.Transaction.Create(tx, &transaction); err != nil {
			return err
		}

		if err := s.repos.TransactionHistory.Create(tx, &models.TransactionHistory{
			TransactionID: transaction.ID,
			AccountID:     acc.ID,
			EntryType:     models.TransactionHistoryEntryTypeCredit,
			Amount:        req.Amount,
			BalanceBefore: acc.Balance,
			BalanceAfter:  acc.Balance + req.Amount,
			Description:   req.Description,
		}); err != nil {
			return err
		}

		if err := tx.Model(acc).Update("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			return err
		}

		transaction.Status = models.TransactionStatusCompleted
		transaction.CompletedAt = &now
		if err := s.repos.Transaction.Save(tx, &transaction); err != nil {
			return err
		}

		result = &transaction
		return nil
	})

	if err != nil {
		return dto.TransactionResponse{}, ErrDepositFailed
	}

	s.publishCompleted(result)

	return dto.TransactionToResponse(*result), nil
}

func (s *transactionService) publishCompleted(t *models.Transaction) {
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
