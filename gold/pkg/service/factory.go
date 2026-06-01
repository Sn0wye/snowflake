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
	forms := &ServiceFactory{}
	forms.Balance = NewBalanceService(repos)
	forms.Flake = NewFlakeService(repos)
	forms.Transaction = NewTransactionService(repos, forms, rmq, log)
	return forms
}
