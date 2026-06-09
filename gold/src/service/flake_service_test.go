package service_test

import (
	"testing"

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

func setupFlakeService(t *testing.T) (*mocks.MockAccountRepo, *mocks.MockFlakeRepo, *gorm.DB, service.FlakeService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	accRepo := mocks.NewMockAccountRepo()
	flakeRepo := mocks.NewMockFlakeRepo()
	repos := &repository.Factory{
		Account: accRepo,
		Flake:   flakeRepo,
	}
	svc := service.NewFlakeService(repos)
	return accRepo, flakeRepo, db, svc
}

func TestCreateFlake_Success(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeEmail,
		KeyValue: "test@example.com",
	}
	resp, err := svc.CreateFlake(db, account.UserID.String(), req)
	require.NoError(t, err)
	assert.Equal(t, req.KeyType, resp.KeyType)
	assert.Equal(t, req.KeyValue, resp.KeyValue)
	assert.Equal(t, models.FlakeStatusActive, resp.Status)
	assert.Len(t, flakeRepo.All(), 1)
}

func TestCreateFlake_DuplicateType(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	flakeRepo.Seed(makeFlake(account.ID, "first@test.com"))

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeEmail,
		KeyValue: "second@test.com",
	}
	_, err := svc.CreateFlake(db, account.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrDuplicateFlakeType)
}

func TestCreateFlake_KeyConflict(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account1 := makeAccount()
	account2 := makeAccount()
	accRepo.Seed(account1, account2)
	flakeRepo.Seed(makeFlake(account2.ID, "shared@test.com"))

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeEmail,
		KeyValue: "shared@test.com",
	}
	_, err := svc.CreateFlake(db, account1.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrFlakeKeyConflict)
}

func TestCreateFlake_ReconciliationDiscrepancy(t *testing.T) {
	accRepo, _, db, svc := setupFlakeService(t)
	account := makeAccount()
	account.ReconciliationStatus = models.AccountReconciliationStatusDiscrepancy
	accRepo.Seed(account)

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeEmail,
		KeyValue: "blocked@test.com",
	}
	_, err := svc.CreateFlake(db, account.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrAccountReconciliation)
}

func TestCreateFlake_AccountNotFound(t *testing.T) {
	_, _, db, svc := setupFlakeService(t)
	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeEmail,
		KeyValue: "new@test.com",
	}
	_, err := svc.CreateFlake(db, uuid.New().String(), req)
	assert.ErrorIs(t, err, service.ErrAccountNotFound)
}

func TestGetFlakes_Success(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	flakeRepo.Seed(makeFlake(account.ID, "first@test.com"), makeFlake(account.ID, "second@test.com"))

	flakes, err := svc.GetFlakes(db, account.UserID.String())
	require.NoError(t, err)
	assert.Len(t, flakes, 2)
}

func TestGetFlakes_Empty(t *testing.T) {
	accRepo, _, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)

	flakes, err := svc.GetFlakes(db, account.UserID.String())
	require.NoError(t, err)
	assert.Empty(t, flakes)
}

func TestDeleteFlake_Success(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	f := makeFlake(account.ID, "delete@test.com")
	flakeRepo.Seed(f)

	resp, err := svc.DeleteFlake(db, account.UserID.String(), f.ID)
	require.NoError(t, err)
	assert.Equal(t, f.ID, resp.ID)

	updated, _ := flakeRepo.FindByIDAndAccount(db, f.ID, account.ID)
	assert.Equal(t, models.FlakeStatusInactive, updated.Status)
}

func TestDeleteFlake_AlreadyInactive(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	f := makeFlake(account.ID, "inactive@test.com")
	f.Status = models.FlakeStatusInactive
	flakeRepo.Seed(f)

	_, err := svc.DeleteFlake(db, account.UserID.String(), f.ID)
	assert.ErrorIs(t, err, service.ErrFlakeAlreadyInactive)
}

func TestDeleteFlake_NotFound(t *testing.T) {
	accRepo, _, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)

	_, err := svc.DeleteFlake(db, account.UserID.String(), uuid.New())
	assert.ErrorIs(t, err, service.ErrFlakeNotFound)
}

func TestCreateFlake_RandomUnlimited_CanCreateMultiple(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)

	for i := 0; i < 10; i++ {
		req := dto.CreateFlakeRequest{
			KeyType:  models.FlakeTypeRandom,
			KeyValue: uuid.New().String(),
		}
		_, err := svc.CreateFlake(db, account.UserID.String(), req)
		require.NoError(t, err, "random flake %d should succeed", i)
	}
	assert.Len(t, flakeRepo.All(), 10)
}

func TestCreateFlake_HandleDoesNotCountTowardLimit(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)

	types := []models.FlakeType{models.FlakeTypeEmail, models.FlakeTypePhone, models.FlakeTypeCPF, models.FlakeTypeCNPJ}
	for _, kt := range types {
		flakeRepo.Seed(&models.Flake{
			ID:        uuid.New(),
			AccountID: account.ID,
			KeyType:   kt,
			KeyValue:  uuid.New().String(),
			Status:    models.FlakeStatusActive,
		})
	}

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeHandle,
		KeyValue: "myhandle",
	}
	resp, err := svc.CreateFlake(db, account.UserID.String(), req)
	require.NoError(t, err)
	assert.Equal(t, models.FlakeTypeHandle, resp.KeyType)
}

func TestCreateFlake_HandleDuplicateIsRejected(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	flakeRepo.Seed(&models.Flake{
		ID:        uuid.New(),
		AccountID: account.ID,
		KeyType:   models.FlakeTypeHandle,
		KeyValue:  "myhandle",
		Status:    models.FlakeStatusActive,
	})

	req := dto.CreateFlakeRequest{
		KeyType:  models.FlakeTypeHandle,
		KeyValue: "otherhandle",
	}
	_, err := svc.CreateFlake(db, account.UserID.String(), req)
	assert.ErrorIs(t, err, service.ErrDuplicateFlakeType)
}

func TestPublicLookup_Success(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	flakeRepo.Seed(makeFlake(account.ID, "lookup@test.com"))

	resp, err := svc.PublicLookup(db, "lookup@test.com")
	require.NoError(t, err)
	assert.Equal(t, account.ID, resp.AccountID)
	assert.Equal(t, models.FlakeTypeEmail, resp.KeyType)
}

func TestPublicLookup_NotFound(t *testing.T) {
	_, _, db, svc := setupFlakeService(t)
	_, err := svc.PublicLookup(db, "nonexistent@test.com")
	assert.ErrorIs(t, err, service.ErrFlakeNotFound)
}

func TestResolveReceiver_Success(t *testing.T) {
	accRepo, flakeRepo, db, svc := setupFlakeService(t)
	account := makeAccount()
	accRepo.Seed(account)
	flakeRepo.Seed(makeFlake(account.ID, "receiver@test.com"))

	acc, flake, err := svc.ResolveReceiver(db, "receiver@test.com")
	require.NoError(t, err)
	assert.Equal(t, account.ID, acc.ID)
	assert.Equal(t, "receiver@test.com", flake.KeyValue)
}

func TestResolveReceiver_NotFound(t *testing.T) {
	_, _, db, svc := setupFlakeService(t)
	_, _, err := svc.ResolveReceiver(db, "nonexistent@test.com")
	assert.ErrorIs(t, err, service.ErrFlakeNotFound)
}
