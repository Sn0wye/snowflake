package service

import (
	"errors"

	"github.com/getsnowflake/snowflake/gold/src/dto"
	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlakeService interface {
	CreateFlake(db *gorm.DB, userID string, req dto.CreateFlakeRequest) (dto.FlakeResponse, error)
	GetFlakes(db *gorm.DB, userID string) ([]dto.FlakeResponse, error)
	DeleteFlake(db *gorm.DB, userID string, flakeID uuid.UUID) (dto.FlakeResponse, error)
	PublicLookup(db *gorm.DB, keyValue string) (dto.LookupFlakeResponse, error)
	ResolveReceiver(db *gorm.DB, keyValue string) (*models.Account, *models.Flake, error)
}

type flakeService struct {
	repos *repository.Factory
}

func NewFlakeService(repos *repository.Factory) FlakeService {
	return &flakeService{repos: repos}
}

const maxTypedFlakeKeys = 5

var unlimitedKeyTypes = map[models.FlakeType]bool{
	models.FlakeTypeRandom: true,
	models.FlakeTypeHandle: true,
}

func (s *flakeService) CreateFlake(db *gorm.DB, userID string, req dto.CreateFlakeRequest) (dto.FlakeResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.FlakeResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	if account.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
		return dto.FlakeResponse{}, ErrAccountReconciliation
	}

	if !unlimitedKeyTypes[req.KeyType] {
		_, err := s.repos.Flake.FindByTypeAndAccount(db, req.KeyType, account.ID)
		if err == nil {
			return dto.FlakeResponse{}, ErrDuplicateFlakeType
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FlakeResponse{}, err
		}

		count, err := s.repos.Flake.CountNonUnlimited(db, account.ID)
		if err != nil {
			return dto.FlakeResponse{}, err
		}
		if count >= maxTypedFlakeKeys {
			return dto.FlakeResponse{}, ErrFlakeLimitReached
		}
	}

	_, err = s.repos.Flake.FindByKeyValue(db, req.KeyValue)
	if err == nil {
		return dto.FlakeResponse{}, ErrFlakeKeyConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.FlakeResponse{}, err
	}

	flake := &models.Flake{
		AccountID: account.ID,
		KeyType:   req.KeyType,
		KeyValue:  req.KeyValue,
		Status:    models.FlakeStatusActive,
	}

	if err := s.repos.Flake.Create(db, flake); err != nil {
		return dto.FlakeResponse{}, err
	}

	return dto.FlakeToResponse(*flake), nil
}

func (s *flakeService) GetFlakes(db *gorm.DB, userID string) ([]dto.FlakeResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return nil, mapNotFound(err, ErrAccountNotFound)
	}

	flakes, err := s.repos.Flake.FindByAccountID(db, account.ID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.FlakeResponse, len(flakes))
	for i, f := range flakes {
		response[i] = dto.FlakeToResponse(f)
	}

	return response, nil
}

func (s *flakeService) DeleteFlake(db *gorm.DB, userID string, flakeID uuid.UUID) (dto.FlakeResponse, error) {
	account, err := s.repos.Account.FindByUserID(db, userID)
	if err != nil {
		return dto.FlakeResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	flake, err := s.repos.Flake.FindByIDAndAccount(db, flakeID, account.ID)
	if err != nil {
		return dto.FlakeResponse{}, mapNotFound(err, ErrFlakeNotFound)
	}

	if flake.Status == models.FlakeStatusInactive {
		return dto.FlakeResponse{}, ErrFlakeAlreadyInactive
	}

	if err := s.repos.Flake.UpdateStatus(db, flake, models.FlakeStatusInactive); err != nil {
		return dto.FlakeResponse{}, err
	}

	return dto.FlakeToResponse(*flake), nil
}

func (s *flakeService) PublicLookup(db *gorm.DB, keyValue string) (dto.LookupFlakeResponse, error) {
	flake, err := s.repos.Flake.FindByKeyValue(db, keyValue)
	if err != nil {
		return dto.LookupFlakeResponse{}, mapNotFound(err, ErrFlakeNotFound)
	}

	account, err := s.repos.Account.FindByID(db, flake.AccountID)
	if err != nil {
		return dto.LookupFlakeResponse{}, mapNotFound(err, ErrAccountNotFound)
	}

	return dto.LookupFlakeResponse{
		AccountID: account.ID,
		KeyType:   flake.KeyType,
		KeyValue:  flake.KeyValue,
	}, nil
}

func (s *flakeService) ResolveReceiver(db *gorm.DB, keyValue string) (*models.Account, *models.Flake, error) {
	flake, err := s.repos.Flake.FindByKeyValue(db, keyValue)
	if err != nil {
		return nil, nil, mapNotFound(err, ErrFlakeNotFound)
	}

	account, err := s.repos.Account.FindByID(db, flake.AccountID)
	if err != nil {
		return nil, nil, mapNotFound(err, ErrAccountNotFound)
	}

	return account, flake, nil
}
