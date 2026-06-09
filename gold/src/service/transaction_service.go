package service

import (
	"errors"
	"sort"
	"time"

	"github.com/getsnowflake/snowflake/gold/pkg/events"
	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"github.com/getsnowflake/snowflake/gold/src/dto"
	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/repository"
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

type EventPublisher interface {
	Produce(queueName, message string) error
}

type TransactionService interface {
	GetTransactions(db *gorm.DB, userID string, filter repository.TransactionFilter) (dto.PaginatedTransactionsResponse, error)
	GetTransactionByID(db *gorm.DB, userID string, id uuid.UUID) (dto.TransactionResponse, error)
	CreateTransaction(db *gorm.DB, userID string, req dto.CreateTransferRequest) (dto.TransactionResponse, error)
	Deposit(db *gorm.DB, userID string, req dto.DepositRequest) (dto.TransactionResponse, error)
}

type transactionService struct {
	repos     *repository.Factory
	svc       *ServiceFactory
	publisher EventPublisher
	log       *logger.Logger
}

func NewTransactionService(repos *repository.Factory, svc *ServiceFactory, publisher EventPublisher, log *logger.Logger) TransactionService {
	return &transactionService{repos: repos, svc: svc, publisher: publisher, log: log}
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
		return dto.TransactionResponse{}, mapNotFound(err, ErrTransactionNotFound)
	}

	isSender := transaction.SenderAccountID != nil && *transaction.SenderAccountID == account.ID
	isReceiver := transaction.ReceiverAccountID != nil && *transaction.ReceiverAccountID == account.ID
	if !isSender && !isReceiver {
		return dto.TransactionResponse{}, ErrTransactionNotFound
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

	senderAccount, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.TransactionResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	receiverAccount, _, err := s.svc.Flake.ResolveReceiver(db, req.ReceiverFlakeKey)
	if err != nil {
		return dto.TransactionResponse{}, err
	}

	var result *models.Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		existing, err := s.repos.Transaction.FindByIdempotencyKey(tx, req.IdempotencyKey)
		if err == nil {
			return &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

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

		if sender.Status != models.AccountStatusActive {
			return ErrAccountNotActive
		}
		if sender.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
			return ErrAccountReconciliation
		}
		if receiver.Status != models.AccountStatusActive {
			return ErrAccountNotActive
		}
		if sender.ID == receiver.ID {
			return ErrSelfTransfer
		}

		if sender.Balance < req.Amount {
			return ErrInsufficientFunds
		}

		startOfDay := time.Now().UTC().Truncate(24 * time.Hour)

		dailyAmount, err := s.repos.Transaction.SumDailyAmount(tx, senderAccount.ID, startOfDay)
		if err != nil {
			return err
		}
		if dailyAmount+req.Amount > dailyTransactionLimit {
			return ErrDailyLimitExceeded
		}

		dailyCount, err := s.repos.Transaction.CountDaily(tx, senderAccount.ID, startOfDay)
		if err != nil {
			return err
		}
		if dailyCount >= maxDailyTransactions {
			return ErrDailyCountExceeded
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
			if isUniqueViolation(err) {
				existing, fetchErr := s.repos.Transaction.FindByIdempotencyKey(tx, req.IdempotencyKey)
				if fetchErr != nil {
					return fetchErr
				}
				return &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
			}
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

		if err := s.repos.Account.DebitBalance(tx, sender.ID, req.Amount); err != nil {
			return err
		}
		if err := s.repos.Account.CreditBalance(tx, receiver.ID, req.Amount); err != nil {
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
		var idempotent *IdempotentTransactionError
		if errors.As(err, &idempotent) {
			return dto.TransactionResponse{}, idempotent
		}
		return dto.TransactionResponse{}, err
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

	account, err := s.repos.Account.FindByID(db, req.AccountID)
	if err != nil {
		return dto.TransactionResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		return dto.TransactionResponse{}, ErrInvalidTokenSubject
	}
	if account.UserID != userIDParsed {
		return dto.TransactionResponse{}, ErrForbidden
	}

	var result *models.Transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		existing, err := s.repos.Transaction.FindByIdempotencyKey(tx, req.IdempotencyKey)
		if err == nil {
			return &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		acc, err := s.repos.Account.FindByIDForUpdate(tx, account.ID)
		if err != nil {
			return err
		}

		if acc.Status != models.AccountStatusActive {
			return ErrAccountNotActive
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
			if isUniqueViolation(err) {
				existing, fetchErr := s.repos.Transaction.FindByIdempotencyKey(tx, req.IdempotencyKey)
				if fetchErr != nil {
					return fetchErr
				}
				return &IdempotentTransactionError{Response: dto.TransactionToResponse(*existing)}
			}
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

		if err := s.repos.Account.CreditBalance(tx, acc.ID, req.Amount); err != nil {
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
		var idempotent *IdempotentTransactionError
		if errors.As(err, &idempotent) {
			return dto.TransactionResponse{}, idempotent
		}
		return dto.TransactionResponse{}, err
	}

	s.publishCompleted(result)

	return dto.TransactionToResponse(*result), nil
}

func (s *transactionService) publishCompleted(t *models.Transaction) {
	if s.publisher == nil {
		return
	}
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

		if err := s.publisher.Produce(events.QueueTransactionCompleted, payload); err != nil {
			s.log.Error("Failed to publish transaction.completed event", zap.Error(err), zap.String("transactionID", t.ID.String()))
		}
	}()
}
