package service

import (
	"errors"

	"github.com/Sn0wye/snowflake/gold/src/dto"
	"gorm.io/gorm"
)

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountReconciliation = errors.New("account is under reconciliation review and cannot be modified")
	ErrAccountNotActive      = errors.New("account is not active")
	ErrDuplicateFlakeType    = errors.New("you already have an active flake key of this type")
	ErrFlakeLimitReached     = errors.New("maximum number of flake keys reached (5 typed keys per account)")
	ErrFlakeKeyConflict      = errors.New("this key value is already registered to another account")
	ErrFlakeNotFound         = errors.New("flake key not found")
	ErrFlakeAlreadyInactive  = errors.New("flake key is already inactive")
	ErrSelfTransfer          = errors.New("cannot transfer to your own account")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrDailyLimitExceeded    = errors.New("daily transaction limit exceeded")
	ErrDailyCountExceeded    = errors.New("maximum daily transaction count reached")
	ErrTransferFailed        = errors.New("transfer failed")
	ErrAmountTooLow          = errors.New("amount must be at least 1 centavo")
	ErrAmountTooHigh         = errors.New("amount exceeds maximum transaction limit")
	ErrForbidden             = errors.New("you can only deposit to your own account")
	ErrTransactionNotFound   = errors.New("transaction not found")
)

type IdempotentTransactionError struct {
	Response dto.TransactionResponse
}

func (e *IdempotentTransactionError) Error() string {
	return "transaction already exists (idempotent)"
}

func mapNotFound(err error, domainErr error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErr
	}
	return err
}
