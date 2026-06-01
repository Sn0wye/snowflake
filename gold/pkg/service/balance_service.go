package service

import (
	"github.com/Sn0wye/snowflake/gold/pkg/repository"
	"github.com/Sn0wye/snowflake/gold/src/dto"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"gorm.io/gorm"
)

type BalanceService interface {
	GetBalance(db *gorm.DB, userID string) (dto.BalanceResponse, error)
	GetBalanceHistory(db *gorm.DB, userID string, page, limit int) (dto.BalanceHistoryResponse, error)
}

type balanceService struct {
	repos *repository.Factory
}

func NewBalanceService(repos *repository.Factory) BalanceService {
	return &balanceService{repos: repos}
}

func (s *balanceService) GetBalance(db *gorm.DB, userID string) (dto.BalanceResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.BalanceResponse{}, mapNotFound(err, ErrAccountNotFound)
	}
	return dto.AccountToBalanceResponse(*account), nil
}

func (s *balanceService) GetBalanceHistory(db *gorm.DB, userID string, page, limit int) (dto.BalanceHistoryResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.BalanceHistoryResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	entries, total, err := s.repos.TransactionHistory.FindByAccountID(db, account.ID, page, limit)
	if err != nil {
		return dto.BalanceHistoryResponse{}, err
	}

	response := make([]dto.BalanceHistoryEntry, len(entries))
	for i, e := range entries {
		response[i] = dto.TransactionHistoryToEntry(e)
	}

	return dto.BalanceHistoryResponse{
		Entries: response,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *balanceService) ResolveAccount(db *gorm.DB, userID string) (*models.Account, error) {
	return s.repos.Account.FindByUserID(db, userID)
}
