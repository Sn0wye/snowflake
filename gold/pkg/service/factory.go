package service

import (
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/pkg/repository"
	"gorm.io/gorm"
)

type ServiceFactory struct {
	Balance     BalanceService
	Flake       FlakeService
	Transaction TransactionService
}

func NewServiceFactory(db *gorm.DB, repos *repository.Factory, rmq *messaging.MessagingService, log *logger.Logger) *ServiceFactory {
	svc := &ServiceFactory{}
	svc.Balance = NewBalanceService(repos)
	svc.Flake = NewFlakeService(repos)
	svc.Transaction = NewTransactionService(repos, svc, rmq, log)
	return svc
}
