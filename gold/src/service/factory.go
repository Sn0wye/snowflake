package service

import (
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/src/repository"
)

type ServiceFactory struct {
	Balance     BalanceService
	Flake       FlakeService
	Transaction TransactionService
}

func NewServiceFactory(repos *repository.Factory, rmq *messaging.MessagingService, log *logger.Logger) *ServiceFactory {
	svc := &ServiceFactory{}
	svc.Balance = NewBalanceService(repos)
	svc.Flake = NewFlakeService(repos)
	svc.Transaction = NewTransactionService(repos, svc, rmq, log)
	return svc
}
