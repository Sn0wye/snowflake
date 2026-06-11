package service

import (
	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"github.com/getsnowflake/snowflake/gold/src/repository"
)

type ServiceFactory struct {
	Balance     BalanceService
	Flake       FlakeService
	Transaction TransactionService
}

func NewServiceFactory(repos *repository.Factory, log *logger.Logger) *ServiceFactory {
	svc := &ServiceFactory{}
	svc.Balance = NewBalanceService(repos)
	svc.Flake = NewFlakeService(repos)
	svc.Transaction = NewTransactionService(repos, svc, log)
	return svc
}
